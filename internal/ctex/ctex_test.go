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

func TestBlockingHelpers(t *testing.T) {
	base := context.Background()
	if IsBlockingCtx(base) {
		t.Fatal("expected base context to be non-blocking")
	}

	blocking := SetBlockingCtx(base)
	if !IsBlockingCtx(blocking) {
		t.Fatal("expected blocking context flag")
	}
	if cleared := ClearBlockingCtx(base); cleared != base {
		t.Fatal("expected clear to preserve non-blocking context")
	}
	if IsBlockingCtx(ClearBlockingCtx(blocking)) {
		t.Fatal("expected cleared context to become non-blocking")
	}

	LogBlockingWarning(base, "beam", "recv")
	LogBlockingWarning(blocking, "beam", "recv")
}

func TestFrameHelpers(t *testing.T) {
	base := context.Background()
	if _, ok := AfterFrame(base); ok {
		t.Fatal("expected no frame in base context")
	}
	if typ := reflect.TypeOf(Frame(base)); typ != reflect.TypeOf(shredder.FreeFrame{}) {
		t.Fatalf("expected free frame without injected context, got %v", typ)
	}

	ctx, after := FrameInsert(base)
	got, ok := AfterFrame(ctx)
	if !ok || got != after {
		t.Fatal("expected inserted frame to be retrievable")
	}
	if Frame(ctx) != after {
		t.Fatal("expected frame helper to expose inserted frame")
	}

	target := context.WithValue(context.Background(), KeyBlocking, true)
	infected := FrameInfect(ctx, target)
	got, ok = AfterFrame(infected)
	if !ok || got != after {
		t.Fatal("expected frame to infect target context")
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
