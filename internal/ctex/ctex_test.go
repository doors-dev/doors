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
	"bytes"
	"context"
	"log/slog"
	"reflect"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

func TestStoreOperations(t *testing.T) {
	store := NewStore()

	if got := store.Load("missing"); got != nil {
		t.Fatalf("expected nil load for missing key, got %#v", got)
	}
	if prev := store.Save("k", "v1"); prev != nil {
		t.Fatalf("expected nil previous value, got %#v", prev)
	}
	if got := store.Load("k"); got != "v1" {
		t.Fatalf("unexpected stored value: %#v", got)
	}
	if prev := store.Save("k", "v2"); prev != "v1" {
		t.Fatalf("unexpected previous value: %#v", prev)
	}
	if got := store.Init("k", func() any { return "ignored" }); got != "v2" {
		t.Fatalf("expected init to reuse existing value, got %#v", got)
	}
	if got := store.Init("new", func() any { return "created" }); got != "created" {
		t.Fatalf("unexpected init value: %#v", got)
	}
	if removed := store.Remove("new"); removed != "created" {
		t.Fatalf("unexpected removed value: %#v", removed)
	}
	if removed := store.Remove("missing"); removed != nil {
		t.Fatalf("expected nil remove for missing key, got %#v", removed)
	}
}

func TestStoreInitWaitsForInFlightValue(t *testing.T) {
	store := NewStore()
	entered := make(chan struct{})
	release := make(chan struct{})
	first := make(chan any, 1)
	loaded := make(chan any, 1)
	second := make(chan any, 1)
	var calls atomic.Int32

	go func() {
		first <- store.Init("k", func() any {
			calls.Add(1)
			close(entered)
			<-release
			return "created"
		})
	}()
	<-entered

	go func() {
		loaded <- store.Load("k")
	}()
	go func() {
		second <- store.Init("k", func() any {
			calls.Add(1)
			return "ignored"
		})
	}()

	requireNoStoreValue(t, loaded)
	requireNoStoreValue(t, second)
	close(release)

	if got := requireStoreValue(t, first); got != "created" {
		t.Fatalf("unexpected first init value: %#v", got)
	}
	if got := requireStoreValue(t, loaded); got != "created" {
		t.Fatalf("unexpected load value: %#v", got)
	}
	if got := requireStoreValue(t, second); got != "created" {
		t.Fatalf("unexpected second init value: %#v", got)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected constructor to run once, ran %d times", got)
	}
}

func TestStoreInitPanicReleasesInFlightValue(t *testing.T) {
	store := NewStore()
	entered := make(chan struct{})
	release := make(chan struct{})
	panicked := make(chan any, 1)
	loaded := make(chan any, 1)
	retried := make(chan any, 1)

	go func() {
		defer func() {
			panicked <- recover()
		}()
		store.Init("k", func() any {
			close(entered)
			<-release
			panic("boom")
		})
	}()
	<-entered

	go func() {
		loaded <- store.Load("k")
	}()
	go func() {
		retried <- store.Init("k", func() any {
			return "retry"
		})
	}()

	requireNoStoreValue(t, loaded)
	requireNoStoreValue(t, retried)
	close(release)

	if got := requireStoreValue(t, panicked); got != "boom" {
		t.Fatalf("expected init panic to propagate, got %#v", got)
	}
	if got := requireStoreValue(t, loaded); got != nil {
		t.Fatalf("expected released failed init to read nil, got %#v", got)
	}
	retryValue := requireStoreValue(t, retried)
	if retryValue != "retry" {
		t.Fatalf("unexpected retry value after failed init: %#v", retryValue)
	}

	saved := make(chan any, 1)
	go func() {
		saved <- store.Save("k", "saved")
	}()
	if got := requireStoreValue(t, saved); got != retryValue {
		t.Fatalf("expected save after failed init to return previous value %#v, got %#v", retryValue, got)
	}
	if got := store.Load("k"); got != "saved" {
		t.Fatalf("expected save after failed init to publish value, got %#v", got)
	}
}

func TestStoreInitRetriesAfterPanic(t *testing.T) {
	store := NewStore()
	func() {
		defer func() {
			if r := recover(); r != "boom" {
				t.Fatalf("expected init panic to propagate, got %#v", r)
			}
		}()
		store.Init("k", func() any {
			panic("boom")
		})
	}()

	if got := store.Init("k", func() any { return "retry" }); got != "retry" {
		t.Fatalf("expected init after panic to retry constructor, got %#v", got)
	}
}

func requireStoreValue(t *testing.T, ch <-chan any) any {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for store value")
		return nil
	}
}

func requireNoStoreValue(t *testing.T, ch <-chan any) {
	t.Helper()
	select {
	case v := <-ch:
		t.Fatalf("expected store operation to wait, got %#v", v)
	case <-time.After(25 * time.Millisecond):
	}
}

// A detached context is marked detached, also through a wrapper; a fake one and a plain context are not.
func TestFreeHelpers(t *testing.T) {
	base := context.Background()
	if GetFrames(base).Detached() {
		t.Fatal("expected base context to be non-detached")
	}
	free := NewDetachedContext(base, base, false)
	if !GetFrames(free).Detached() {
		t.Fatal("expected detached context flag")
	}
	timed, cancel := context.WithTimeout(free, time.Minute)
	defer cancel()
	if !GetFrames(timed).Detached() {
		t.Fatal("expected a wrapper of a detached context to stay detached")
	}
	if GetFrames(NewDetachedContext(base, base, true)).Detached() {
		t.Fatal("expected a fake detached context not to be detached")
	}
}

func TestFrameHelpers(t *testing.T) {
	base := context.Background()
	if _, ok := GetFrames(base).AfterFrame(); ok {
		t.Fatal("expected no frame in base context")
	}
	if typ := reflect.TypeOf(GetFrames(base).Call()); typ != reflect.TypeOf(shredder.FreeFrame{}) {
		t.Fatalf("expected free frame without injected context, got %v", typ)
	}
	ctx, after := NewAfterFrame(base, "a", 0)
	got, ok := GetFrames(ctx).AfterFrame()
	if !ok || got != after || GetFrames(ctx).Call() != after {
		t.Fatal("expected the scope's after frame to be retrievable")
	}
}

type testOwner string

func (o testOwner) InstanceID() string {
	return string(o)
}

// Owner means an after frame, a batch frame a batch valve; Settle keeps only the ids.
func TestFramesMethods(t *testing.T) {
	if fs := GetFrames(context.Background()); fs.Owner() || fs.Detached() || fs.Group() != 0 || fs.Batch() != 0 || !batchNone(fs) {
		t.Fatal("expected empty frames in a base context")
	}
	scope, after := NewAfterFrame(context.Background(), "a", 5)
	fs := GetFrames(scope)
	if !fs.Owner() || fs.Detached() || fs.Group() != 5 || fs.Batch() != 0 || fs.Inst() != "a" || !batchNone(fs) {
		t.Fatal("expected a scope to own work without a batch")
	}
	batched, valve := NewBatch(scope, 9)
	fs = GetFrames(batched)
	if !fs.Owner() || batchNone(fs) || fs.Group() != 5 || fs.Batch() != 9 || batchOf(fs) != valve || fs.Call() != after {
		t.Fatal("expected a batch on the scope to keep its after frame and group")
	}
	fs.detached = true
	settle := fs.Settle()
	if settle.Owner() || settle.Detached() || !batchNone(settle) || settle.Group() != 5 || settle.Batch() != 9 || settle.Inst() != "a" || settle.sync != nil || settle.init != nil {
		t.Fatal("expected settle frames to carry the ids only")
	}
}

// A new scope keeps the sync frames of ctx and takes the given group, nothing else.
func TestNewAfterFrameInheritsSyncOnly(t *testing.T) {
	sync := shredder.FreeFrame{}
	outer, _ := NewAfterFrame(context.Background(), "a", 3)
	outer, _ = NewBatch(outer, 4)
	outer = SyncFrameInsert(outer, sync, sync)
	inner, after := NewAfterFrame(outer, "a", 0)
	fs := GetFrames(inner)
	if fs.after != after || fs.Group() != 0 || fs.Batch() != 0 || !batchNone(fs) || fs.Inst() != "a" {
		t.Fatal("expected a new scope to inherit no ids and no valve")
	}
	if fs.sync != sync || fs.init != sync {
		t.Fatal("expected a new scope to keep the sync frames")
	}
	settled, _ := NewAfterFrame(NewDetachedContext(outer, outer, true), "a", 0)
	if fs := GetFrames(settled); fs.Group() != 0 || fs.Batch() != 0 || !batchNone(fs) {
		t.Fatal("expected a new scope on a settle context to inherit no ids")
	}
	grouped, _ := NewAfterFrame(outer, "a", 7)
	if fs := GetFrames(grouped); fs.Group() != 7 || fs.Batch() != 0 || !batchNone(fs) {
		t.Fatal("expected a new scope to take the given group only")
	}
}

// A new batch keeps the scope's group and after frame and brings its own id and valve.
func TestNewBatchKeepsGroup(t *testing.T) {
	scope, after := NewAfterFrame(context.Background(), "a", 5)
	first, v1 := NewBatch(scope, 8)
	second, v2 := NewBatch(scope, 9)
	f1, f2 := GetFrames(first), GetFrames(second)
	if f1.Group() != 5 || f2.Group() != 5 || f1.Batch() != 8 || f2.Batch() != 9 {
		t.Fatal("expected each batch to keep the group and carry its id")
	}
	if v1 == v2 || batchOf(f1) != v1 || batchOf(f2) != v2 || f1.Call() != after || f2.Call() != after {
		t.Fatal("expected each batch to bring its own valve on the same after frame")
	}
	if fs := GetFrames(scope); fs.Batch() != 0 || !batchNone(fs) {
		t.Fatal("expected a batch not to change the scope's context")
	}
}

// A fake detached context carries the settle frames and is not detached; a real one carries only the detached mark, which a new scope drops.
func TestDetachedFrames(t *testing.T) {
	scope, _ := NewAfterFrame(context.Background(), "a", 5)
	batched, _ := NewBatch(scope, 6)
	batched = SyncFrameInsert(batched, shredder.FreeFrame{}, shredder.FreeFrame{})
	fake := NewDetachedContext(batched, batched, true)
	if got := GetFrames(fake); got != GetFrames(batched).Settle() || got.Detached() {
		t.Fatalf("expected the settle frames of ctx, got %+v", got)
	}
	if got := GetFrames(NewDetachedContext(context.Background(), context.Background(), true)); got != (Frames{}) {
		t.Fatalf("expected empty frames from a plain context, got %+v", got)
	}
	detached := NewDetachedContext(batched, batched, false)
	if got := GetFrames(detached); got != (Frames{detached: true}) {
		t.Fatalf("expected only the detached mark, got %+v", got)
	}
	scoped, _ := NewAfterFrame(detached, "a", 0)
	if GetFrames(scoped).Detached() {
		t.Fatal("expected a new scope on a detached context not to be detached")
	}
}

// Infecting a context of the same instance keeps the frames, the detached mark included; another instance gets only the sync frames.
func TestFrameInfect(t *testing.T) {
	scope, after := NewAfterFrame(context.Background(), "a", 3)
	batched, valve := NewBatch(scope, 7)
	batched = SyncFrameInsert(batched, shredder.FreeFrame{}, shredder.FreeFrame{})
	pageA := context.WithValue(context.Background(), common.KeyCore, testOwner("a"))
	pageB := context.WithValue(context.Background(), common.KeyCore, testOwner("b"))
	same := NewInfectedContext(batched, pageA)
	if got := GetFrames(same); got != GetFrames(batched) || got.Call() != after || batchOf(got) != valve {
		t.Fatal("expected the frames to reach a context of the same instance unchanged")
	}
	if same.Value(common.KeyCore) != testOwner("a") {
		t.Fatal("expected frame infect to preserve existing target values")
	}
	foreign := GetFrames(NewInfectedContext(batched, pageB))
	if foreign.Owner() || !batchNone(foreign) || foreign.Group() != 0 || foreign.Batch() != 0 || foreign.Inst() != "" {
		t.Fatal("expected no after frame, ids or valve in a foreign instance")
	}
	if foreign.sync == nil || foreign.init == nil {
		t.Fatal("expected the sync frames to cross instances")
	}
	detached := NewDetachedContext(pageA, pageA, false)
	if !GetFrames(NewInfectedContext(detached, pageA)).Detached() {
		t.Fatal("expected the detached mark to reach a context of the same instance")
	}
	if GetFrames(NewInfectedContext(detached, pageB)).Detached() {
		t.Fatal("expected no detached mark in a foreign instance")
	}
	noFrame := NewInfectedContext(context.Background(), pageA)
	if _, ok := GetFrames(noFrame).AfterFrame(); ok || noFrame.Value(common.KeyCore) != testOwner("a") {
		t.Fatal("expected no frame when source has none, and the target core")
	}
}

func TestLogCanceled(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	LogCanceled(context.Background(), "update")
	if buf.Len() != 0 {
		t.Fatalf("expected no log for a live context, got %q", buf.String())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	LogCanceled(ctx, "update")
	// the first frame outside the doors module is the test runner
	if !regexp.MustCompile(`caller=\S*testing/testing\.go:\d+`).MatchString(buf.String()) {
		t.Fatalf("expected caller outside the doors module, got %q", buf.String())
	}
}

func batchNone(f Frames) bool {
	_, ok := f.BatchFrame()
	return !ok
}

func batchOf(f Frames) *shredder.ValveFrame {
	v, _ := f.BatchFrame()
	return v
}
