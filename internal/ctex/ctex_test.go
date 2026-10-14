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

func TestFreeHelpers(t *testing.T) {
	base := context.Background()
	if IsDetachedCtx(base) {
		t.Fatal("expected base context to be non-free")
	}

	free := NewDetachedContext(base, base, false)
	if !IsDetachedCtx(free) {
		t.Fatal("expected free context flag")
	}
}

func TestFrameHelpers(t *testing.T) {
	base := context.Background()
	if _, ok := AfterFrame(base); ok {
		t.Fatal("expected no frame in base context")
	}
	if typ := reflect.TypeOf(GetFrames(base).Call()); typ != reflect.TypeOf(shredder.FreeFrame{}) {
		t.Fatalf("expected free frame without injected context, got %v", typ)
	}

	ctx, after := AfterFrameInsert(base, 0, 0, "a")
	got, ok := AfterFrame(ctx)
	if !ok || got != after {
		t.Fatal("expected inserted frame to be retrievable")
	}
	if GetFrames(ctx).Call() != after {
		t.Fatal("expected frame helper to expose inserted frame")
	}

	target := context.WithValue(context.Background(), common.KeyCore, testOwner("a"))
	infected := FrameInfect(ctx, target)
	got, ok = AfterFrame(infected)
	if !ok || got != after {
		t.Fatal("expected frame to infect a context of the same instance")
	}
	if infected.Value(common.KeyCore) != testOwner("a") {
		t.Fatal("expected frame infect to preserve existing target values")
	}
	noFrame := FrameInfect(base, target)
	if _, ok := AfterFrame(noFrame); ok {
		t.Fatal("expected no frame when source has none")
	}
	if noFrame.Value(common.KeyCore) != testOwner("a") {
		t.Fatal("expected infect to expose target core value")
	}
}

type testOwner string

func (o testOwner) InstanceID() string {
	return string(o)
}

// A context of another instance gets only the sync frames: no after frame, no batch.
func TestFrameInfectForeignInstance(t *testing.T) {
	ctx, _ := AfterFrameInsert(context.Background(), 3, 7, "a")
	ctx = SyncFrameInsert(ctx, shredder.FreeFrame{}, shredder.FreeFrame{})
	foreign := FrameInfect(ctx, context.WithValue(context.Background(), common.KeyCore, testOwner("b")))
	if _, ok := AfterFrame(foreign); ok {
		t.Fatal("expected no after frame in a foreign instance")
	}
	fs := GetFrames(foreign)
	if fs.Group("a") != 0 || fs.Group("b") != 0 || fs.Batch("a") != 0 || fs.Batch("b") != 0 || fs.BatchFrame("a") != nil || fs.BatchGuard("a") {
		t.Fatal("expected no group and no batch in a foreign instance")
	}
	if fs.sync == nil || fs.init == nil {
		t.Fatal("expected the sync frames to cross instances")
	}
	same := GetFrames(FrameInfect(ctx, context.WithValue(context.Background(), common.KeyCore, testOwner("a"))))
	if same.Group("a") != 3 || same.Batch("a") != 7 || !same.BatchGuard("a") || same.BatchFrame("a") == nil {
		t.Fatal("expected the group and the batch to reach a context of the same instance")
	}
}

// A group holds nothing, a batch has guard and frame, a nested after frame inherits both, and a new batch keeps the group.
func TestGroupAndBatch(t *testing.T) {
	plain, _ := AfterFrameInsert(context.Background(), 5, 0, "a")
	if fs := GetFrames(plain); fs.Group("a") != 5 || fs.Group("b") != 0 || fs.Batch("a") != 0 || fs.BatchGuard("a") || fs.BatchFrame("a") != nil {
		t.Fatal("expected a group without a batch")
	}
	batched, _ := AfterFrameInsert(context.Background(), 5, 6, "a")
	fs := GetFrames(batched)
	if fs.Group("a") != 5 || fs.Batch("a") != 6 || !fs.BatchGuard("a") || fs.BatchFrame("a") == nil {
		t.Fatal("expected a batch to carry group, guard and frame")
	}
	alone, _ := AfterFrameInsert(plain, 0, 8, "a")
	if got := GetFrames(alone); got.Group("a") != 0 || got.Batch("a") != 8 || !got.BatchGuard("a") {
		t.Fatal("expected an explicit batch not to inherit the group")
	}
	fresh, _ := AfterFrameInsert(batched, 7, 0, "a")
	if got := GetFrames(fresh); got.Group("a") != 7 || got.Batch("a") != 0 || got.BatchFrame("a") != nil || got.BatchGuard("a") {
		t.Fatal("expected an explicit group not to inherit the batch")
	}
	preserved := NewDetachedContext(batched, batched, true)
	if got := GetFrames(preserved); got.Group("a") != 5 || got.Batch("a") != 6 || got.BatchFrame("a") != fs.BatchFrame("a") || got.BatchGuard("a") {
		t.Fatal("expected a preserved context to keep group and batch and drop the guard")
	}
	if got := GetFrames(NewDetachedContext(batched, batched, false)); got.Group("a") != 0 || got.Batch("a") != 0 {
		t.Fatal("expected a detached context to cut group and batch")
	}
	nested, _ := AfterFrameInsert(preserved, 0, 0, "a")
	if got := GetFrames(nested); got.Group("a") != 5 || got.Batch("a") != 6 || got.BatchFrame("a") != fs.BatchFrame("a") || got.BatchGuard("a") {
		t.Fatal("expected a nested after frame to keep group and batch as it found them")
	}
	inherited, _ := AfterFrameInsert(batched, 0, 0, "a")
	if got := GetFrames(inherited); got.Group("a") != 5 || got.Batch("a") != 6 || !got.BatchGuard("a") {
		t.Fatal("expected a nested after frame to inherit the guard")
	}
	other, _ := AfterFrameInsert(batched, 0, 0, "b")
	if got := GetFrames(other); got.Group("b") != 0 || got.Batch("b") != 0 || got.BatchFrame("b") != nil {
		t.Fatal("expected no inheritance across instances")
	}
	opened, after, frame, inserted := BatchFrameInsert(plain, 9, "a")
	if got := GetFrames(opened); inserted || after != GetFrames(plain).after || got.Group("a") != 5 || got.Batch("a") != 9 || !got.BatchGuard("a") || got.BatchFrame("a") != frame {
		t.Fatal("expected a new batch to keep the group and the after frame")
	}
	late, _, lateFrame, inserted := BatchFrameInsert(preserved, 10, "a")
	if got := GetFrames(late); !inserted || got.Group("a") != 5 || got.Batch("a") != 10 || got.BatchFrame("a") != lateFrame || lateFrame == fs.BatchFrame("a") {
		t.Fatal("expected a new batch on a preserved context to bring its own frames and keep the group")
	}
	foreign, _, _, _ := BatchFrameInsert(plain, 11, "b")
	if got := GetFrames(foreign); got.Group("b") != 0 || got.Batch("b") != 11 {
		t.Fatal("expected a new batch of another instance to drop the group")
	}
}

func TestBatchHelpers(t *testing.T) {
	base := context.Background()
	if got := GetFrames(base).Batch("a"); got != 0 {
		t.Fatalf("expected no batch in base context, got %d", got)
	}

	ctx, after := AfterFrameInsert(base, 0, 7, "a")
	if got, ok := AfterFrame(ctx); !ok || got != after {
		t.Fatal("expected the after frame next to the batch")
	}
	if got := GetFrames(ctx).Batch("a"); got != 7 {
		t.Fatalf("expected batch 7 for the owning instance, got %d", got)
	}
	if got := GetFrames(ctx).Batch("b"); got != 0 {
		t.Fatalf("expected no batch for a foreign instance, got %d", got)
	}

	synced := SyncFrameInsert(ctx, shredder.FreeFrame{}, shredder.FreeFrame{})
	if got := GetFrames(synced).Batch("a"); got != 7 {
		t.Fatalf("expected sync insert to keep the batch, got %d", got)
	}

	detached := NewDetachedContext(ctx, ctx, false)
	if got := GetFrames(detached).Batch("a"); got != 0 {
		t.Fatalf("expected detached context to drop the batch, got %d", got)
	}

	preserved := NewDetachedContext(ctx, ctx, true)
	if !IsDetachedCtx(preserved) {
		t.Fatal("expected the preserving context to stay detached")
	}
	if got := GetFrames(preserved).Batch("a"); got != 7 {
		t.Fatalf("expected detached context to preserve the batch, got %d", got)
	}
	if got := GetFrames(preserved).Batch("b"); got != 0 {
		t.Fatalf("expected no preserved batch for a foreign instance, got %d", got)
	}
	if _, ok := AfterFrame(preserved); ok {
		t.Fatal("expected the preserving context to carry no after frame")
	}
	if got := GetFrames(NewDetachedContext(base, base, true)).Batch("a"); got != 0 {
		t.Fatalf("expected no batch to preserve from a plain context, got %d", got)
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
