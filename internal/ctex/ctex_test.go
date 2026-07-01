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
	"reflect"
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
	if IsFreeCtx(base) {
		t.Fatal("expected base context to be non-free")
	}

	free := NewFreeContext(base, base)
	if !IsFreeCtx(free) {
		t.Fatal("expected free context flag")
	}

	LogFreeWarning(base, "beam", "recv")
	LogFreeWarning(free, "beam", "recv")
}

func TestFrameHelpers(t *testing.T) {
	base := context.Background()
	if _, ok := AfterFrame(base); ok {
		t.Fatal("expected no frame in base context")
	}
	if typ := reflect.TypeOf(GetFrames(base).Call()); typ != reflect.TypeOf(shredder.FreeFrame{}) {
		t.Fatalf("expected free frame without injected context, got %v", typ)
	}

	ctx, after := AfterFrameInsert(base)
	got, ok := AfterFrame(ctx)
	if !ok || got != after {
		t.Fatal("expected inserted frame to be retrievable")
	}
	if GetFrames(ctx).Call() != after {
		t.Fatal("expected frame helper to expose inserted frame")
	}

	target := context.WithValue(context.Background(), common.KeyCore, "core")
	infected := FrameInfect(ctx, target)
	got, ok = AfterFrame(infected)
	if !ok || got != after {
		t.Fatal("expected frame to infect target context")
	}
	if infected.Value(common.KeyCore) != "core" {
		t.Fatal("expected frame infect to preserve existing target values")
	}
	noFrame := FrameInfect(base, target)
	if _, ok := AfterFrame(noFrame); ok {
		t.Fatal("expected no frame when source has none")
	}
	if noFrame.Value(common.KeyCore) != "core" {
		t.Fatal("expected infect to expose target core value")
	}
}

func TestLogCanceled(t *testing.T) {
	LogCanceled(context.Background(), "update")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	LogCanceled(ctx, "update")
}
