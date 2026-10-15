// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctex

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type tagContext struct {
	context.Context
	tag string
}

func (c tagContext) Value(key any) any {
	return c.tag
}

type userKey struct{}

var internalKey = new(int)

const (
	tagLife   = "life"
	tagValues = "values"
)

func viewKeys() map[string]any {
	return map[string]any{
		"core":     common.KeyCore,
		"session":  common.KeySession,
		"frame":    common.KeyFrame,
		"replace":  common.KeyHistoryReplace,
		"document": common.KeyDocument,
		"user":     userKey{},
		"internal": internalKey,
	}
}

func checkRouting(t *testing.T, name string, ctx context.Context, want map[string]any) {
	t.Helper()
	for k, key := range viewKeys() {
		if got := ctx.Value(key); got != want[k] {
			t.Errorf("%s: key %s: got %#v, want %#v", name, k, got, want[k])
		}
	}
}

// Each kind routes every key class to life, values, nothing or computed frames.
func TestViewRouting(t *testing.T) {
	life := tagContext{context.Background(), tagLife}
	values := tagContext{context.Background(), tagValues}
	checkRouting(t, "render", NewRenderContext(life, values), map[string]any{
		"core": tagLife, "session": tagLife, "frame": tagLife, "replace": tagLife, "document": tagLife,
		"user": tagValues, "internal": tagValues,
	})
	checkRouting(t, "detached", NewDetachedContext(values, life, false), map[string]any{
		"core": tagValues, "session": tagValues, "frame": Frames{detached: true}, "replace": tagValues, "document": nil,
		"user": tagValues, "internal": tagValues,
	})
	checkRouting(t, "settle", NewDetachedContext(values, life, true), map[string]any{
		"core": tagValues, "session": tagValues, "frame": Frames{}, "replace": tagValues, "document": nil,
		"user": tagValues, "internal": tagValues,
	})
	checkRouting(t, "infected", NewInfectedContext(values, life), map[string]any{
		"core": tagLife, "session": tagLife, "frame": nil, "replace": tagValues, "document": tagLife,
		"user": tagValues, "internal": tagValues,
	})
}

// Every kind takes Done, Err and Deadline from life, never from values.
func TestViewLifetime(t *testing.T) {
	deadline := time.Now().Add(time.Hour)
	for name, build := range viewBuilders() {
		life, cancel := context.WithDeadline(context.Background(), deadline)
		values, cancelValues := context.WithCancel(context.Background())
		cancelValues()
		ctx := build(values, life)
		if d, ok := ctx.Deadline(); !ok || !d.Equal(deadline) {
			t.Errorf("%s: expected the life deadline", name)
		}
		if ctx.Err() != nil || ctx.Done() != life.Done() {
			t.Errorf("%s: expected the life done channel and no error", name)
		}
		cancel()
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Errorf("%s: expected the life error, got %v", name, ctx.Err())
		}
	}
}

func viewBuilders() map[string]func(values, life context.Context) context.Context {
	return map[string]func(values, life context.Context) context.Context{
		"render": func(values, life context.Context) context.Context {
			return NewRenderContext(life, values)
		},
		"detached": func(values, life context.Context) context.Context {
			return NewDetachedContext(values, life, false)
		},
		"settle": func(values, life context.Context) context.Context {
			return NewDetachedContext(values, life, true)
		},
		"infected": func(values, life context.Context) context.Context {
			return NewInfectedContext(values, life)
		},
	}
}

// A render context unwraps only a render context, and a nil user leaves the system context.
func TestNewRenderContext(t *testing.T) {
	life := tagContext{context.Background(), tagLife}
	first := tagContext{context.Background(), "first"}
	second := tagContext{context.Background(), "second"}
	if NewRenderContext(life, nil) != context.Context(life) {
		t.Fatal("expected the system context without a user context")
	}
	outer := NewRenderContext(life, first)
	inner := NewRenderContext(outer, second)
	if inner.Value(common.KeyCore) != tagLife || inner.Value(userKey{}) != "second" {
		t.Fatal("expected a render context to replace the user context of a render context")
	}
	if NewRenderContext(outer, nil) != context.Context(life) {
		t.Fatal("expected a render context without a user context to unwrap to the system context")
	}
	detached := NewDetachedContext(first, life, false)
	wrapped := NewRenderContext(detached, second)
	if wrapped.Value(common.KeyCore) != "first" || wrapped.Value(common.KeyDocument) != nil || wrapped.Value(userKey{}) != "second" {
		t.Fatal("expected a render context to keep a detached system context whole")
	}
	if NewRenderContext(detached, nil) != detached {
		t.Fatal("expected a detached context not to be unwrapped")
	}
}

// The document mark goes to the system side of a render context, or wraps any other context.
func TestNewDocumentContext(t *testing.T) {
	values := context.WithValue(context.Background(), userKey{}, "user")
	render := NewDocumentContext(NewRenderContext(context.Background(), values))
	if render.Value(common.KeyDocument) != true || render.Value(userKey{}) != "user" {
		t.Fatal("expected a document render context with the user values")
	}
	if UserContext(render, nil) != values {
		t.Fatal("expected the document render context to keep its user context")
	}
	plain := NewDocumentContext(values)
	if plain.Value(common.KeyDocument) != true || plain.Value(userKey{}) != "user" {
		t.Fatal("expected a document mark on a plain context")
	}
	detached := NewDocumentContext(NewDetachedContext(values, context.Background(), false))
	if detached.Value(common.KeyDocument) != true || !GetFrames(detached).Detached() {
		t.Fatal("expected a document mark over a detached context")
	}
}

// The user context is the values of a render context, nil for the parent itself, and ctx otherwise.
func TestUserContext(t *testing.T) {
	parent := context.WithValue(context.Background(), userKey{}, "parent")
	values := context.WithValue(context.Background(), userKey{}, "user")
	if UserContext(NewRenderContext(parent, values), parent) != values {
		t.Fatal("expected the user context of a render context")
	}
	if UserContext(parent, parent) != nil {
		t.Fatal("expected no user context for the parent")
	}
	if UserContext(values, parent) != values {
		t.Fatal("expected a foreign context to be the user context")
	}
	detached := NewDetachedContext(values, parent, false)
	if UserContext(detached, parent) != detached {
		t.Fatal("expected a detached context to be the user context as a whole")
	}
}

// Nested views route through each other: a detached infected context, a render context in a detached one.
func TestViewNested(t *testing.T) {
	page := context.WithValue(context.WithValue(context.Background(), common.KeyCore, testOwner("a")), common.KeyDocument, true)
	scope, _ := NewAfterFrame(context.WithValue(context.Background(), userKey{}, "source"), "a", 3)
	infected := NewInfectedContext(scope, page)
	detached := NewDetachedContext(infected, context.Background(), false)
	if detached.Value(common.KeyCore) != testOwner("a") || detached.Value(common.KeyDocument) != nil || detached.Value(userKey{}) != "source" {
		t.Fatal("expected a detached infected context to read the target core and source values without the document mark")
	}
	if GetFrames(detached) != (Frames{detached: true}) {
		t.Fatal("expected only the detached mark")
	}
	settle := NewDetachedContext(infected, context.Background(), true)
	if GetFrames(settle) != GetFrames(scope).Settle() {
		t.Fatal("expected the settle frames of the infected frames")
	}
	render := NewRenderContext(detached, context.WithValue(context.Background(), userKey{}, "render"))
	if render.Value(common.KeyCore) != testOwner("a") || render.Value(common.KeyDocument) != nil || GetFrames(render) != (Frames{detached: true}) || render.Value(userKey{}) != "render" {
		t.Fatal("expected a render context to read doors keys through the detached context")
	}
	reinfected := NewInfectedContext(render, page)
	if reinfected.Value(common.KeyDocument) != true || reinfected.Value(userKey{}) != "render" || !GetFrames(reinfected).Detached() {
		t.Fatal("expected an infected render context to take the target document and the source detached frames")
	}
}

// Cause of a view reads the cancel state through Value, so it follows the values side for every kind.
func TestViewCause(t *testing.T) {
	cause := errors.New("cause")
	for name, build := range viewBuilders() {
		life, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		ctx := build(context.Background(), life)
		if got := context.Cause(ctx); !errors.Is(got, context.Canceled) {
			t.Errorf("%s: expected the error of ctx, got %v", name, got)
		}
		same := build(life, life)
		if got := context.Cause(same); got != cause {
			t.Errorf("%s: expected the cause of life shared with values, got %v", name, got)
		}
	}
}

// Children and after funcs on every kind attach to life without a goroutine each.
func TestViewChildrenNoGoroutines(t *testing.T) {
	for name, build := range viewBuilders() {
		t.Run(name, func(t *testing.T) {
			life, cancel := context.WithCancel(context.Background())
			values := context.WithValue(context.Background(), userKey{}, "user")
			checkChildren(t, build(values, life), cancel)
		})
	}
}

// Frame, document and replace marks over every kind stay on the view, so children still attach without a goroutine.
func TestViewMarksNoGoroutines(t *testing.T) {
	marks := map[string]func(context.Context) context.Context{
		"after": func(ctx context.Context) context.Context {
			ctx, _ = NewAfterFrame(ctx, "a", 1)
			return ctx
		},
		"batch": func(ctx context.Context) context.Context {
			ctx, _ = NewBatch(ctx, 1)
			return ctx
		},
		"sync": func(ctx context.Context) context.Context {
			return SyncFrameInsert(ctx, shredder.FreeFrame{}, shredder.FreeFrame{})
		},
		"document": NewDocumentContext,
		"replace":  NewReplaceContext,
	}
	for name, build := range viewBuilders() {
		for mark, insert := range marks {
			t.Run(name+"/"+mark, func(t *testing.T) {
				life, cancel := context.WithCancel(context.Background())
				values := context.WithValue(context.Background(), userKey{}, "user")
				ctx := insert(build(values, life))
				if ctx.Value(userKey{}) != "user" {
					t.Fatal("expected the user values through the mark")
				}
				checkChildren(t, ctx, cancel)
			})
		}
	}
}

// Marks over a plain context read as set and leave every other key to it.
func TestViewMarksPlain(t *testing.T) {
	base := context.WithValue(context.Background(), userKey{}, "user")
	ctx, after := NewAfterFrame(base, "a", 3)
	ctx = NewReplaceContext(NewDocumentContext(ctx))
	fs := GetFrames(ctx)
	if fs.after != after || fs.inst != "a" || fs.group != 3 {
		t.Fatal("expected the inserted frames")
	}
	if ctx.Value(common.KeyDocument) != true || ctx.Value(common.KeyHistoryReplace) != true || ctx.Value(userKey{}) != "user" {
		t.Fatal("expected the marks and the user values")
	}
	if base.Value(common.KeyFrame) != nil || base.Value(common.KeyDocument) != nil {
		t.Fatal("expected the base context unchanged")
	}
}

// Every mark is a new context, and comparing a marked context never compares the user context.
func TestViewMarksIdentity(t *testing.T) {
	user := uncomparableContext{context.Background(), []string{"user"}}
	parent := NewRenderContext(context.Background(), user)
	replaced := NewReplaceContext(parent)
	if UserContext(replaced, parent) != replaced {
		t.Fatal("expected a marked render context to be the user context as a whole")
	}
	if NewReplaceContext(replaced) == replaced || NewDocumentContext(NewDocumentContext(user)) == NewDocumentContext(user) {
		t.Fatal("expected a new context for every mark")
	}
	framed, _ := NewAfterFrame(parent, "a", 0)
	if UserContext(framed, parent) != framed {
		t.Fatal("expected a framed render context to be the user context as a whole")
	}
}

type uncomparableContext struct {
	context.Context
	tags []string
}

func checkChildren(t *testing.T, ctx context.Context, cancel func()) {
	t.Helper()
	const n = 100
	before := runtime.NumGoroutine()
	children := make([]context.Context, 0, n)
	for range n {
		child, stop := context.WithCancel(ctx)
		t.Cleanup(stop)
		children = append(children, child)
	}
	ran := make(chan struct{}, n)
	for range n {
		context.AfterFunc(ctx, func() { ran <- struct{}{} })
	}
	if grown := runtime.NumGoroutine() - before; grown > 10 {
		t.Fatalf("expected no goroutine per child, grew by %d", grown)
	}
	cancel()
	for i, child := range children {
		select {
		case <-child.Done():
		case <-time.After(5 * time.Second):
			t.Fatalf("child %d not canceled", i)
		}
	}
	for i := range n {
		select {
		case <-ran:
		case <-time.After(5 * time.Second):
			t.Fatalf("after func %d did not run", i)
		}
	}
}

// Infected frames stay whole only for the target's own instance; a target without a core, a foreign one, or detached frames without a source core get the sync frames only.
func TestViewInfectOwners(t *testing.T) {
	scope, _ := NewAfterFrame(context.Background(), "a", 3)
	scope = SyncFrameInsert(scope, shredder.FreeFrame{}, shredder.FreeFrame{})
	syncOnly := Frames{sync: shredder.FreeFrame{}, init: shredder.FreeFrame{}}
	pageA := context.WithValue(context.Background(), common.KeyCore, testOwner("a"))
	pageB := context.WithValue(context.Background(), common.KeyCore, testOwner("b"))
	if got := GetFrames(NewInfectedContext(scope, context.Background())); got != syncOnly {
		t.Fatalf("expected sync frames without an owner core, got %+v", got)
	}
	if got := GetFrames(NewInfectedContext(scope, pageA)); got != GetFrames(scope) {
		t.Fatalf("expected whole frames in the same instance, got %+v", got)
	}
	if got := GetFrames(NewInfectedContext(scope, pageB)); got != syncOnly {
		t.Fatalf("expected sync frames in a foreign instance, got %+v", got)
	}
	withCore := NewDetachedContext(context.WithValue(scope, common.KeyCore, testOwner("a")), context.Background(), false)
	if got := GetFrames(NewInfectedContext(withCore, pageA)); got != (Frames{detached: true}) {
		t.Fatalf("expected detached frames to reach the source core's instance, got %+v", got)
	}
	if got := GetFrames(NewInfectedContext(withCore, pageB)); got != (Frames{}) {
		t.Fatalf("expected detached frames cut in a foreign instance, got %+v", got)
	}
	noCore := NewDetachedContext(scope, context.Background(), false)
	if got := GetFrames(NewInfectedContext(noCore, pageA)); got != (Frames{}) {
		t.Fatalf("expected detached frames without a source core to be cut, got %+v", got)
	}
}
