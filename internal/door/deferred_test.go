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

package door

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
	"weak"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/gox"
)

type doorOp = func(d *Door, ctx context.Context, content any) <-chan error

// onClean registers f like doors.OnClean.
func onClean(ctx context.Context, f func()) {
	c := ctx.Value(common.KeyCore).(core.Core)
	c.Door().CleanFrame().Run(context.Background(), c.Instance().Runtime(), func(bool) {
		f()
	})
}

// gated renders text once gate is closed.
func gated(gate <-chan struct{}, s string) gox.Elem {
	return func(cur gox.Cursor) error {
		<-gate
		return cur.Text(s)
	}
}

func recvOp(t *testing.T, chs <-chan (<-chan error)) <-chan error {
	t.Helper()
	select {
	case ch := <-chs:
		return ch
	case <-time.After(10 * time.Second):
		t.Fatal("the clean callback never issued the operation")
		return nil
	}
}

// expectApplied waits for an operation channel to report scheduled and
// applied, then close.
func expectApplied(t *testing.T, ch <-chan error) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	var got []error
	for {
		select {
		case err, ok := <-ch:
			if !ok {
				if len(got) != 2 || got[0] != nil || got[1] != nil {
					t.Fatalf("expected the operation to apply, got %v", got)
				}
				return
			}
			got = append(got, err)
		case <-timeout:
			t.Fatalf("operation channel did not close, got %v", got)
		}
	}
}

// expectCanceled waits for an operation channel to report context.Canceled
// and close.
func expectCanceled(t *testing.T, ch <-chan error) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	canceled := false
	for {
		select {
		case err, ok := <-ch:
			if !ok {
				if !canceled {
					t.Fatal("operation was not canceled")
				}
				return
			}
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("operation failed: %v", err)
				}
				canceled = true
			}
		case <-timeout:
			t.Fatal("operation channel did not close")
		}
	}
}

// An operation on the same Door issued by the replaced content's clean
// callback supersedes the deferred operation that drops that content.
func TestDeferredCleanCallbackUpdatesDoor(t *testing.T) {
	cases := []struct {
		name     string
		deferred doorOp
		update   doorOp
	}{
		{"inner", (*Door).DeferredInner, (*Door).Inner},
		{"outer", (*Door).DeferredOuter, (*Door).Outer},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inst := newTestInstance(t)
			d := &Door{}
			updates := make(chan (<-chan error), 1)
			d.Inner(context.Background(), func(cur gox.Cursor) error {
				onClean(cur.Context(), func() {
					updates <- c.update(d, context.Background(), text("after"))
				})
				return cur.Text("skeleton")
			})
			renderTestPage(t, inst, func(cur gox.Cursor) error {
				return cur.Comp(d)
			})
			gate := make(chan struct{})
			ch := c.deferred(d, context.Background(), gated(gate, "heavy"))
			close(gate)
			expectCanceled(t, ch)
			expectApplied(t, recvOp(t, updates))
			expectApplied(t, d.Inner(context.Background(), text("later")))
		})
	}
}

// The same for an operation on an ancestor Door.
func TestDeferredCleanCallbackUpdatesAncestor(t *testing.T) {
	cases := []struct {
		name     string
		deferred doorOp
	}{
		{"inner", (*Door).DeferredInner},
		{"outer", (*Door).DeferredOuter},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inst := newTestInstance(t)
			parent := &Door{}
			d := &Door{}
			updates := make(chan (<-chan error), 1)
			d.Inner(context.Background(), func(cur gox.Cursor) error {
				onClean(cur.Context(), func() {
					updates <- parent.Inner(context.Background(), text("after"))
				})
				return cur.Text("skeleton")
			})
			parent.Inner(context.Background(), func(cur gox.Cursor) error {
				return cur.Comp(d)
			})
			renderTestPage(t, inst, func(cur gox.Cursor) error {
				return cur.Comp(parent)
			})
			gate := make(chan struct{})
			ch := c.deferred(d, context.Background(), gated(gate, "heavy"))
			close(gate)
			expectCanceled(t, ch)
			expectApplied(t, recvOp(t, updates))
			expectApplied(t, parent.Inner(context.Background(), text("later")))
		})
	}
}

type panicWatcher struct{}

func (panicWatcher) Watch(context.Context, int) bool {
	return false
}

func (panicWatcher) Cancel() {
	panic("watcher cancel")
}

// A panic while a deferred operation drops the replaced content ends the
// instance, not the process.
func TestDeferredCleanPanicEndsInstance(t *testing.T) {
	cases := []struct {
		name     string
		deferred doorOp
	}{
		{"inner", (*Door).DeferredInner},
		{"outer", (*Door).DeferredOuter},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inst := newTestInstance(t)
			src := beam.NewSource(0, beam.DefaultEqual[int], false)
			d := &Door{}
			d.Inner(context.Background(), func(cur gox.Cursor) error {
				if _, ok := src.Watch(cur.Context(), panicWatcher{}); !ok {
					return errors.New("watch refused")
				}
				return cur.Text("skeleton")
			})
			renderTestPage(t, inst, func(cur gox.Cursor) error {
				return cur.Comp(d)
			})
			killed := make(chan struct{})
			inst.root.outer.CleanFrame().Run(context.Background(), nil, func(bool) {
				close(killed)
			})
			gate := make(chan struct{})
			ch := c.deferred(d, context.Background(), gated(gate, "heavy"))
			close(gate)
			expectCanceled(t, ch)
			select {
			case <-inst.runtime.Context().Done():
			case <-time.After(10 * time.Second):
				t.Fatal("the instance survived the panic")
			}
			select {
			case <-killed:
			case <-time.After(10 * time.Second):
				t.Fatal("the root was not killed")
			}
		})
	}
}

// A panic while a plain operation cleans the content it replaces ends the
// instance and still closes the operation channel with context.Canceled.
func TestPlainCleanPanicReportsCanceled(t *testing.T) {
	cases := []struct {
		name string
		prev doorOp
		op   func(d *Door, ctx context.Context) <-chan error
	}{
		{"outer", (*Door).Inner, func(d *Door, ctx context.Context) <-chan error { return d.Outer(ctx, text("newer")) }},
		{"reload over outer", (*Door).Outer, (*Door).Reload},
		{"static", (*Door).Inner, func(d *Door, ctx context.Context) <-chan error { return d.Static(ctx, text("newer")) }},
		{"inner over outer", (*Door).Outer, func(d *Door, ctx context.Context) <-chan error { return d.Inner(ctx, text("newer")) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inst := newTestInstance(t)
			src := beam.NewSource(0, beam.DefaultEqual[int], false)
			d := &Door{}
			c.prev(d, context.Background(), gox.Elem(func(cur gox.Cursor) error {
				if _, ok := src.Watch(cur.Context(), panicWatcher{}); !ok {
					return errors.New("watch refused")
				}
				return cur.Text("old")
			}))
			renderTestPage(t, inst, func(cur gox.Cursor) error {
				return cur.Comp(d)
			})
			expectCanceled(t, c.op(d, context.Background()))
		})
	}
}

// A deferred render superseded while the content it replaces panics in its
// cleanup still cleans its own content.
func TestDeferredSupersedeCleansOwnContentOnPrevPanic(t *testing.T) {
	inst := newTestInstance(t)
	src := beam.NewSource(0, beam.DefaultEqual[int], false)
	d := &Door{}
	d.Inner(context.Background(), func(cur gox.Cursor) error {
		if _, ok := src.Watch(cur.Context(), panicWatcher{}); !ok {
			return errors.New("watch refused")
		}
		return cur.Text("skeleton")
	})
	renderTestPage(t, inst, func(cur gox.Cursor) error {
		return cur.Comp(d)
	})
	gate := make(chan struct{})
	defer close(gate)
	entered := make(chan struct{})
	cleaned := make(chan struct{})
	d.DeferredInner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		onClean(cur.Context(), func() { close(cleaned) })
		close(entered)
		<-gate
		return cur.Text("heavy")
	}))
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("the deferred render never started")
	}
	d.Inner(context.Background(), text("newer"))
	select {
	case <-cleaned:
	case <-time.After(10 * time.Second):
		t.Fatal("the deferred content was not cleaned")
	}
}

// Once a deferred op has cleaned the content it replaced, the Door keeps no
// reference to that content.
func TestDeferredReleasesReplacedContent(t *testing.T) {
	cases := []struct {
		name     string
		first    doorOp
		deferred doorOp
	}{
		{"inner", (*Door).Inner, (*Door).DeferredInner},
		{"inner over outer", (*Door).Outer, (*Door).DeferredInner},
		{"outer", (*Door).Inner, (*Door).DeferredOuter},
		{"static", (*Door).Inner, (*Door).DeferredStatic},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inst := newTestInstance(t)
			d := &Door{}
			c.first(d, context.Background(), text("skeleton"))
			renderTestPage(t, inst, func(cur gox.Cursor) error {
				return cur.Comp(d)
			})
			replaced := weak.Make(d.node.Load().tracker)
			expectApplied(t, c.deferred(d, context.Background(), text("heavy")))
			deadline := time.Now().Add(5 * time.Second)
			for replaced.Value() != nil {
				if time.Now().After(deadline) {
					t.Fatal("the Door kept the replaced content")
				}
				runtime.GC()
			}
			runtime.KeepAlive(d)
		})
	}
}
