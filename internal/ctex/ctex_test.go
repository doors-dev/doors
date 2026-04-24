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
	"testing"

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

func TestFreeHelpers(t *testing.T) {
	base := context.Background()
	if IsFreeCtx(base) {
		t.Fatal("expected base context to be non-free")
	}

	free := NewFreeContext(base, base)
	if !IsFreeCtx(free) {
		t.Fatal("expected free context flag")
	}
	if cleared := ClearFreeCtx(base); cleared != base {
		t.Fatal("expected clear to preserve non-free context")
	}
	if IsFreeCtx(ClearFreeCtx(free)) {
		t.Fatal("expected cleared context to become non-free")
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

	target := context.WithValue(context.Background(), KeyCore, "core")
	infected := FrameInfect(ctx, target)
	got, ok = AfterFrame(infected)
	if !ok || got != after {
		t.Fatal("expected frame to infect target context")
	}
	if infected.Value(KeyCore) != "core" {
		t.Fatal("expected frame infect to preserve existing target values")
	}
	if FrameInfect(base, target) != target {
		t.Fatal("expected infect without frame to preserve target context")
	}
}

func TestLogCanceled(t *testing.T) {
	LogCanceled(context.Background(), "update")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	LogCanceled(ctx, "update")
}
