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

package beam

import (
	"context"
	"slices"
	"sync"
	"testing"
)

func freeLen[T any](s Source[T]) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.freeSubs)
}

func TestFreeSubReceivesUpdates(t *testing.T) {
	src := NewSource(0, DefaultEqual[int], false)
	var mu sync.Mutex
	var got []int
	ok := src.Sub(context.Background(), func(_ context.Context, v int) bool {
		mu.Lock()
		got = append(got, v)
		mu.Unlock()
		return false
	})
	if !ok {
		t.Fatal("sub refused on background ctx")
	}
	if err := <-src.Update(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if err := <-src.Update(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Fatalf("expected [0 1 2], got %v", got)
	}
}

func TestFreeSubRead(t *testing.T) {
	src := NewSource(7, DefaultEqual[int], false)
	v, ok := src.Read(context.Background())
	if !ok || v != 7 {
		t.Fatalf("expected 7 true, got %d %v", v, ok)
	}
	if n := freeLen(src); n != 0 {
		t.Fatalf("expected read to leave no subs, got %d", n)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, ok := src.Read(ctx); ok {
		t.Fatal("expected read to refuse canceled ctx")
	}
}

func TestFreeSubRefusesCanceledCtx(t *testing.T) {
	src := NewSource(0, DefaultEqual[int], false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if src.Sub(ctx, func(context.Context, int) bool { return false }) {
		t.Fatal("expected sub to refuse canceled ctx")
	}
}

func TestFreeSubEndsWhenDone(t *testing.T) {
	src := NewSource(0, DefaultEqual[int], false)
	var mu sync.Mutex
	calls := 0
	src.Sub(context.Background(), func(_ context.Context, v int) bool {
		mu.Lock()
		calls++
		mu.Unlock()
		return v == 1
	})
	if err := <-src.Update(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if n := freeLen(src); n != 0 {
		t.Fatalf("expected done sub removed, got %d", n)
	}
	<-src.Update(context.Background(), 2)
	mu.Lock()
	defer mu.Unlock()
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

type recordingWatcher struct {
	mu       sync.Mutex
	values   []int
	canceled bool
}

func (r *recordingWatcher) Watch(_ context.Context, v int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values = append(r.values, v)
	return false
}

func (r *recordingWatcher) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canceled = true
}

func (r *recordingWatcher) isCanceled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.canceled
}

func TestFreeSubCancelFunc(t *testing.T) {
	src := NewSource(0, DefaultEqual[int], false)
	w := &recordingWatcher{}
	cancel, ok := src.Watch(context.Background(), w)
	if !ok {
		t.Fatal("watch refused on background ctx")
	}
	cancel()
	if !w.isCanceled() {
		t.Fatal("expected watcher cancel")
	}
	if n := freeLen(src); n != 0 {
		t.Fatalf("expected canceled sub removed, got %d", n)
	}
	<-src.Update(context.Background(), 1)
	w.mu.Lock()
	defer w.mu.Unlock()
	if !slices.Equal(w.values, []int{0}) {
		t.Fatalf("expected only initial value, got %v", w.values)
	}
}

func TestFreeSubLazyCtxCleanup(t *testing.T) {
	src := NewSource(0, DefaultEqual[int], false)
	w := &recordingWatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	if _, ok := src.Watch(ctx, w); !ok {
		t.Fatal("watch refused")
	}
	cancel()
	if n := freeLen(src); n != 1 {
		t.Fatalf("expected sub kept until propagation, got %d", n)
	}
	if err := <-src.Update(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if n := freeLen(src); n != 0 {
		t.Fatalf("expected dead sub removed on propagation, got %d", n)
	}
	if !w.isCanceled() {
		t.Fatal("expected watcher cancel on propagation")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if !slices.Equal(w.values, []int{0}) {
		t.Fatalf("expected only initial value, got %v", w.values)
	}
}

func TestFreeSubPanicAborts(t *testing.T) {
	src := NewSource(0, DefaultEqual[int], false)
	src.Sub(context.Background(), func(_ context.Context, v int) bool {
		if v == 1 {
			panic("subscriber boom")
		}
		return false
	})
	if err := <-src.Update(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if n := freeLen(src); n != 0 {
		t.Fatalf("expected panicking sub removed, got %d", n)
	}
	if err := <-src.Update(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	if got := src.Get(); got != 2 {
		t.Fatalf("expected source to keep working, got %d", got)
	}
}

func TestFreeSubSerialOrder(t *testing.T) {
	src := NewSource(0, nil, true)
	var mu sync.Mutex
	var got []int
	src.Sub(context.Background(), func(_ context.Context, v int) bool {
		mu.Lock()
		got = append(got, v)
		mu.Unlock()
		return false
	})
	for i := 1; i <= 20; i++ {
		if err := <-src.Update(context.Background(), i); err != nil {
			t.Fatal(err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 21 {
		t.Fatalf("expected 21 values, got %d: %v", len(got), got)
	}
	if !slices.IsSorted(got) {
		t.Fatalf("expected ordered delivery, got %v", got)
	}
}

func TestFreeSubOnBeam(t *testing.T) {
	src := NewSource(1, DefaultEqual[int], false)
	derived := NewBeam(src, func(v int) int { return v * 10 }, DefaultEqual[int])
	var mu sync.Mutex
	var got []int
	ok := derived.Sub(context.Background(), func(_ context.Context, v int) bool {
		mu.Lock()
		got = append(got, v)
		mu.Unlock()
		return false
	})
	if !ok {
		t.Fatal("beam sub refused on background ctx")
	}
	if err := <-src.Update(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !slices.Equal(got, []int{10, 20}) {
		t.Fatalf("expected [10 20], got %v", got)
	}
}
