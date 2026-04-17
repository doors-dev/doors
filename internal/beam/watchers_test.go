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
	"fmt"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type testCore struct {
	cinema Cinema
}

func (c testCore) Cinema() Cinema {
	return c.cinema
}

type testDoor struct {
	ctx    context.Context
	thread shredder.Thread
}

func (d *testDoor) ReadFrame() shredder.Frame {
	return d.thread.Frame()
}

func (d *testDoor) Context() context.Context {
	return d.ctx
}

type noopShutdown struct{}

func (noopShutdown) Shutdown() {}

type stubSyncSource[T any] struct {
	syncCalls int
	syncFunc  func(prev, seq uint, after shredder.SimpleFrame) (*T, bool)
}

func (s *stubSyncSource[T]) Sub(context.Context, func(context.Context, T) bool) bool {
	return false
}

func (s *stubSyncSource[T]) XSub(context.Context, func(context.Context, T) bool, func()) (context.CancelFunc, bool) {
	return none, false
}

func (s *stubSyncSource[T]) ReadAndSub(context.Context, func(context.Context, T) bool) (T, bool) {
	var zero T
	return zero, false
}

func (s *stubSyncSource[T]) XReadAndSub(context.Context, func(context.Context, T) bool, func()) (T, context.CancelFunc, bool) {
	var zero T
	return zero, none, false
}

func (s *stubSyncSource[T]) Read(context.Context) (T, bool) {
	var zero T
	return zero, false
}

func (s *stubSyncSource[T]) AddWatcher(context.Context, Watcher[T]) (context.CancelFunc, bool) {
	return none, false
}

func (s *stubSyncSource[T]) addWatcher(context.Context, *watcher) bool {
	return false
}

func (s *stubSyncSource[T]) sync(prev, seq uint, after shredder.SimpleFrame) (*T, bool) {
	s.syncCalls++
	if s.syncFunc == nil {
		return nil, false
	}
	return s.syncFunc(prev, seq, after)
}

func newBeamContext(t *testing.T) context.Context {
	t.Helper()
	runtime := shredder.NewRuntime(context.Background(), 1, noopShutdown{})
	t.Cleanup(runtime.Cancel)

	door := &testDoor{ctx: context.Background()}
	cinema := NewCinema(nil, door, runtime)
	ctx := context.WithValue(context.Background(), ctex.KeyCore, testCore{cinema: cinema})
	door.ctx = ctx
	return ctx
}

func newNestedBeamContexts(t *testing.T) (context.Context, context.Context) {
	t.Helper()

	runtime := shredder.NewRuntime(context.Background(), 1, noopShutdown{})
	t.Cleanup(runtime.Cancel)

	parentDoor := &testDoor{ctx: context.Background()}
	parentCinema := NewCinema(nil, parentDoor, runtime)
	parentCtx := context.WithValue(context.Background(), ctex.KeyCore, testCore{cinema: parentCinema})
	parentDoor.ctx = parentCtx

	childDoor := &testDoor{ctx: context.Background()}
	childCinema := NewCinema(parentCinema, childDoor, runtime)
	childCtx := context.WithValue(context.Background(), ctex.KeyCore, testCore{cinema: childCinema})
	childDoor.ctx = childCtx

	return parentCtx, childCtx
}

func expectErr(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err, ok := <-ch:
		if !ok {
			t.Fatal("expected channel result, got closed channel")
		}
		return err
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for error result")
		return nil
	}
}

func expectInt(t *testing.T, ch <-chan int) int {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for int value")
		return 0
	}
}

func expectString(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for string value")
		return ""
	}
}

func expectSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cancel signal")
	}
}

func TestSourceXUpdateAndXMutate(t *testing.T) {
	source := NewSourceEqual(0, nil)

	if err := expectErr(t, source.XUpdate(context.Background(), 1)); err != nil {
		t.Fatal(err)
	}
	if got := source.Get(); got != 1 {
		t.Fatal("unexpected source value after XUpdate:", got)
	}

	if err := expectErr(t, source.XMutate(context.Background(), func(v int) int {
		return v + 1
	})); err != nil {
		t.Fatal(err)
	}
	if got := source.Get(); got != 2 {
		t.Fatal("unexpected source value after XMutate:", got)
	}
}

func TestBeamWatcherExtendedAPIs(t *testing.T) {
	ctx := newBeamContext(t)
	source := NewSource(1)
	derived := NewBeam(source, func(v int) string {
		return fmt.Sprintf("v:%d", v)
	})

	readAndSubUpdates := make(chan string, 1)
	initial, ok := derived.ReadAndSub(ctx, func(ctx context.Context, value string) bool {
		readAndSubUpdates <- value
		return true
	})
	if !ok {
		t.Fatal("expected ReadAndSub to subscribe")
	}
	if initial != "v:1" {
		t.Fatal("unexpected initial derived value:", initial)
	}

	source.Update(ctx, 2)
	if got := expectString(t, readAndSubUpdates); got != "v:2" {
		t.Fatal("unexpected derived update:", got)
	}

	sourceReadAndSubUpdates := make(chan int, 1)
	sourceReadAndSubCanceled := make(chan struct{}, 1)
	sourceInitial, sourceCancel, ok := source.XReadAndSub(ctx, func(ctx context.Context, value int) bool {
		sourceReadAndSubUpdates <- value
		return false
	}, func() {
		close(sourceReadAndSubCanceled)
	})
	if !ok {
		t.Fatal("expected source XReadAndSub to subscribe")
	}
	if sourceInitial != 2 {
		t.Fatal("unexpected initial source value:", sourceInitial)
	}

	source.Update(ctx, 3)
	if got := expectInt(t, sourceReadAndSubUpdates); got != 3 {
		t.Fatal("unexpected source XReadAndSub update:", got)
	}
	sourceCancel()
	expectSignal(t, sourceReadAndSubCanceled)

	derivedReadAndSubUpdates := make(chan string, 1)
	derivedReadAndSubCanceled := make(chan struct{}, 1)
	derivedInitial, derivedCancel, ok := derived.XReadAndSub(ctx, func(ctx context.Context, value string) bool {
		derivedReadAndSubUpdates <- value
		return false
	}, func() {
		close(derivedReadAndSubCanceled)
	})
	if !ok {
		t.Fatal("expected derived XReadAndSub to subscribe")
	}
	if derivedInitial != "v:3" {
		t.Fatal("unexpected initial derived XReadAndSub value:", derivedInitial)
	}

	source.Update(ctx, 4)
	if got := expectString(t, derivedReadAndSubUpdates); got != "v:4" {
		t.Fatal("unexpected derived XReadAndSub update:", got)
	}
	derivedCancel()
	expectSignal(t, derivedReadAndSubCanceled)

	sourceSubUpdates := make(chan int, 2)
	sourceSubCanceled := make(chan struct{}, 1)
	sourceSubCancel, ok := source.XSub(ctx, func(ctx context.Context, value int) bool {
		sourceSubUpdates <- value
		return false
	}, func() {
		close(sourceSubCanceled)
	})
	if !ok {
		t.Fatal("expected source XSub to subscribe")
	}
	if got := expectInt(t, sourceSubUpdates); got != 4 {
		t.Fatal("unexpected initial source XSub value:", got)
	}

	derivedSubUpdates := make(chan string, 2)
	derivedSubCanceled := make(chan struct{}, 1)
	derivedSubCancel, ok := derived.XSub(ctx, func(ctx context.Context, value string) bool {
		derivedSubUpdates <- value
		return false
	}, func() {
		close(derivedSubCanceled)
	})
	if !ok {
		t.Fatal("expected derived XSub to subscribe")
	}
	if got := expectString(t, derivedSubUpdates); got != "v:4" {
		t.Fatal("unexpected initial derived XSub value:", got)
	}

	source.Update(ctx, 5)
	if got := expectInt(t, sourceSubUpdates); got != 5 {
		t.Fatal("unexpected source XSub update:", got)
	}
	if got := expectString(t, derivedSubUpdates); got != "v:5" {
		t.Fatal("unexpected derived XSub update:", got)
	}

	sourceSubCancel()
	expectSignal(t, sourceSubCanceled)
	derivedSubCancel()
	expectSignal(t, derivedSubCanceled)

	noCoreCancel, ok := derived.XSub(context.Background(), func(context.Context, string) bool {
		return false
	}, nil)
	if ok {
		t.Fatal("expected XSub without a Doors context to fail")
	}
	noCoreCancel()
}

func TestNewBeamEqualDefaultsNilComparator(t *testing.T) {
	source := &stubSyncSource[int]{}
	derived := NewBeamEqual(source, func(v int) string {
		return fmt.Sprintf("v:%d", v)
	}, nil)

	typed, ok := derived.(*beam[int, string])
	if !ok {
		t.Fatal("expected NewBeamEqual to return internal beam implementation")
	}
	if typed.equal == nil {
		t.Fatal("expected nil comparator to be replaced with a default implementation")
	}
	if typed.equal("same", "same") {
		t.Fatal("expected default comparator to treat equal values as changed")
	}
}

func TestBeamReadSubHelpersAndEquality(t *testing.T) {
	none()

	ctx := newBeamContext(t)
	source := NewSourceEqual(1, func(new int, old int) bool {
		return new == old
	})
	source.DisableSkipping()
	source.Mutate(ctx, func(v int) int {
		return v + 1
	})
	if got := source.Get(); got != 2 {
		t.Fatal("unexpected source value after Mutate:", got)
	}

	derived := NewBeamEqual(source, func(v int) int {
		return v % 2
	}, func(new int, old int) bool {
		return new == old
	})

	if got, ok := source.Read(ctx); !ok || got != 2 {
		t.Fatal("unexpected source Read result:", got, ok)
	}
	if got, ok := derived.Read(ctx); !ok || got != 0 {
		t.Fatal("unexpected beam Read result:", got, ok)
	}

	sourceUpdates := make(chan int, 1)
	sourceInitial, ok := source.ReadAndSub(ctx, func(ctx context.Context, value int) bool {
		sourceUpdates <- value
		return true
	})
	if !ok {
		t.Fatal("expected source ReadAndSub to subscribe")
	}
	if sourceInitial != 2 {
		t.Fatal("unexpected source initial value:", sourceInitial)
	}

	beamUpdates := make(chan int, 2)
	if !derived.Sub(ctx, func(ctx context.Context, value int) bool {
		beamUpdates <- value
		return value == 1
	}) {
		t.Fatal("expected beam Sub to subscribe")
	}
	if got := expectInt(t, beamUpdates); got != 0 {
		t.Fatal("unexpected initial beam Sub value:", got)
	}

	if !source.Sub(ctx, func(ctx context.Context, value int) bool {
		return value == 3
	}) {
		t.Fatal("expected source Sub to subscribe")
	}

	source.Update(ctx, 3)
	if got := expectInt(t, sourceUpdates); got != 3 {
		t.Fatal("unexpected source ReadAndSub update:", got)
	}
	if got := expectInt(t, beamUpdates); got != 1 {
		t.Fatal("unexpected beam Sub update:", got)
	}
}

func TestBeamSyncEntryCachedBranches(t *testing.T) {
	t.Run("prev zero reuses cached value as updated", func(t *testing.T) {
		source := &stubSyncSource[int]{}
		value := 11
		b := &beam[int, int]{
			source: source,
			values: map[uint]entry[int]{
				7: {
					value:   &value,
					prev:    3,
					updated: false,
				},
			},
			equal: func(new int, old int) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(0, 7, nil)
		if got != &value {
			t.Fatal("expected cached value pointer on first read")
		}
		if !updated {
			t.Fatal("expected first read of cached seq to be treated as updated")
		}
		if source.syncCalls != 0 {
			t.Fatal("expected cached read to avoid source sync")
		}
	})

	t.Run("same prev returns cached updated flag", func(t *testing.T) {
		source := &stubSyncSource[int]{}
		value := 17
		b := &beam[int, int]{
			source: source,
			values: map[uint]entry[int]{
				9: {
					value:   &value,
					prev:    4,
					updated: false,
				},
			},
			equal: func(new int, old int) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(4, 9, nil)
		if got != &value {
			t.Fatal("expected cached value pointer when prev matches")
		}
		if updated {
			t.Fatal("expected cached updated flag to be reused")
		}
		if source.syncCalls != 0 {
			t.Fatal("expected cached read to avoid source sync")
		}
	})

	t.Run("missing prev entry falls back to cached value as updated", func(t *testing.T) {
		source := &stubSyncSource[int]{}
		value := 23
		b := &beam[int, int]{
			source: source,
			values: map[uint]entry[int]{
				11: {
					value:   &value,
					prev:    5,
					updated: false,
				},
			},
			equal: func(new int, old int) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(7, 11, nil)
		if got != &value {
			t.Fatal("expected cached value pointer when prev entry is missing")
		}
		if !updated {
			t.Fatal("expected missing prev entry to force updated")
		}
		if source.syncCalls != 0 {
			t.Fatal("expected cached read to avoid source sync")
		}
	})

	t.Run("rebinds cached seq to a different prev when both entries are stored", func(t *testing.T) {
		source := &stubSyncSource[int]{}
		prevValue := 0
		value := 1
		b := &beam[int, int]{
			source: source,
			values: map[uint]entry[int]{
				1: {
					value:   &prevValue,
					prev:    0,
					updated: true,
				},
				3: {
					value:   &value,
					prev:    2,
					updated: true,
				},
			},
			equal: func(new int, old int) bool {
				return new == old
			},
		}

		got, _ := b.syncEntry(1, 3, nil)
		if got != &value {
			t.Fatal("expected cached seq value to be reused during prev recompute")
		}
		entry := b.values[3]
		if entry.prev != 1 {
			t.Fatal("expected cached seq entry to be rebound to the new prev:", entry.prev)
		}
		if source.syncCalls != 0 {
			t.Fatal("expected cached prev recompute to avoid source sync")
		}
	})
}

func TestBeamSyncEntryStaleEqualBranches(t *testing.T) {
	t.Run("stale source equal reuses previous cached value", func(t *testing.T) {
		sourceValue := 8
		source := &stubSyncSource[int]{
			syncFunc: func(prev, seq uint, after shredder.SimpleFrame) (*int, bool) {
				if prev != 2 || seq != 5 {
					t.Fatalf("unexpected sync request prev=%d seq=%d", prev, seq)
				}
				return &sourceValue, false
			},
		}

		prevValue := "cached"
		b := &beam[int, string]{
			source: source,
			values: map[uint]entry[string]{
				2: {
					value:   &prevValue,
					prev:    1,
					updated: true,
				},
			},
			cast: func(v int) string {
				t.Fatal("cast should not run when previous beam value is still cached")
				return ""
			},
			equal: func(new string, old string) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(2, 5, nil)
		if got != &prevValue {
			t.Fatal("expected stale equal sync to reuse previous cached beam value")
		}
		if updated {
			t.Fatal("expected stale equal sync to report no update")
		}
		if _, has := b.values[5]; has {
			t.Fatal("expected no new cached entry when previous beam value can be reused")
		}
	})

	t.Run("stale source equal rebuilds value when previous beam entry is gone", func(t *testing.T) {
		sourceValue := 9
		source := &stubSyncSource[int]{
			syncFunc: func(prev, seq uint, after shredder.SimpleFrame) (*int, bool) {
				if prev != 3 || seq != 6 {
					t.Fatalf("unexpected sync request prev=%d seq=%d", prev, seq)
				}
				return &sourceValue, false
			},
		}

		b := &beam[int, string]{
			source: source,
			values: map[uint]entry[string]{},
			cast: func(v int) string {
				return fmt.Sprintf("v:%d", v)
			},
			equal: func(new string, old string) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(3, 6, nil)
		if got == nil || *got != "v:9" {
			t.Fatal("expected stale equal sync to rebuild current beam value")
		}
		if updated {
			t.Fatal("expected rebuilt stale equal sync to report no update")
		}
		entry, has := b.values[6]
		if !has {
			t.Fatal("expected rebuilt stale equal sync to cache the current seq")
		}
		if entry.prev != 3 || entry.updated {
			t.Fatal("unexpected rebuilt entry metadata:", entry.prev, entry.updated)
		}
		if entry.value == nil || *entry.value != "v:9" {
			t.Fatal("unexpected rebuilt entry value")
		}
	})
}

func TestBeamSyncEntrySourceBranches(t *testing.T) {
	t.Run("missing source value returns nil", func(t *testing.T) {
		source := &stubSyncSource[int]{
			syncFunc: func(prev, seq uint, after shredder.SimpleFrame) (*int, bool) {
				return nil, false
			},
		}
		b := &beam[int, string]{
			source: source,
			values: map[uint]entry[string]{},
			cast: func(v int) string {
				t.Fatal("cast should not run when source value is missing")
				return ""
			},
			equal: func(new string, old string) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(2, 4, nil)
		if got != nil {
			t.Fatal("expected nil when source sync has no value")
		}
		if updated {
			t.Fatal("expected missing source value to report no update")
		}
	})

	t.Run("updated source without previous beam value stores new entry", func(t *testing.T) {
		sourceValue := 10
		source := &stubSyncSource[int]{
			syncFunc: func(prev, seq uint, after shredder.SimpleFrame) (*int, bool) {
				if prev != 2 || seq != 7 {
					t.Fatalf("unexpected sync request prev=%d seq=%d", prev, seq)
				}
				return &sourceValue, true
			},
		}
		b := &beam[int, string]{
			source: source,
			values: map[uint]entry[string]{},
			cast: func(v int) string {
				return fmt.Sprintf("v:%d", v)
			},
			equal: func(new string, old string) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(2, 7, nil)
		if got == nil || *got != "v:10" {
			t.Fatal("expected updated source to build a new beam value")
		}
		if !updated {
			t.Fatal("expected updated source without prev entry to report update")
		}
		entry, has := b.values[7]
		if !has {
			t.Fatal("expected new seq entry to be cached")
		}
		if entry.prev != 2 || !entry.updated {
			t.Fatal("unexpected cached entry metadata:", entry.prev, entry.updated)
		}
		if entry.value == nil || *entry.value != "v:10" {
			t.Fatal("unexpected cached entry value")
		}
	})

	t.Run("updated source equal to previous beam value reuses previous pointer", func(t *testing.T) {
		sourceValue := 12
		source := &stubSyncSource[int]{
			syncFunc: func(prev, seq uint, after shredder.SimpleFrame) (*int, bool) {
				return &sourceValue, true
			},
		}
		prevValue := "v:12"
		b := &beam[int, string]{
			source: source,
			values: map[uint]entry[string]{
				3: {
					value:   &prevValue,
					prev:    2,
					updated: true,
				},
			},
			cast: func(v int) string {
				return fmt.Sprintf("v:%d", v)
			},
			equal: func(new string, old string) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(3, 8, nil)
		if got != &prevValue {
			t.Fatal("expected equal updated source to reuse previous beam pointer")
		}
		if updated {
			t.Fatal("expected equal updated source to report no update")
		}
		entry, has := b.values[8]
		if !has {
			t.Fatal("expected equal updated source to cache current seq")
		}
		if entry.value != &prevValue || entry.prev != 3 || entry.updated {
			t.Fatal("unexpected equal-entry cache state:", entry.prev, entry.updated)
		}
	})

	t.Run("updated source changed from previous beam value stores new pointer", func(t *testing.T) {
		sourceValue := 13
		source := &stubSyncSource[int]{
			syncFunc: func(prev, seq uint, after shredder.SimpleFrame) (*int, bool) {
				return &sourceValue, true
			},
		}
		prevValue := "v:12"
		b := &beam[int, string]{
			source: source,
			values: map[uint]entry[string]{
				4: {
					value:   &prevValue,
					prev:    3,
					updated: false,
				},
			},
			cast: func(v int) string {
				return fmt.Sprintf("v:%d", v)
			},
			equal: func(new string, old string) bool {
				return new == old
			},
		}

		got, updated := b.syncEntry(4, 9, nil)
		if got == nil || *got != "v:13" {
			t.Fatal("expected changed updated source to build a new value")
		}
		if !updated {
			t.Fatal("expected changed updated source to report update")
		}
		if got == &prevValue {
			t.Fatal("expected changed updated source to allocate a new pointer")
		}
		entry, has := b.values[9]
		if !has {
			t.Fatal("expected changed updated source to cache current seq")
		}
		if entry.prev != 4 || !entry.updated {
			t.Fatal("unexpected changed-entry cache metadata:", entry.prev, entry.updated)
		}
		if entry.value == nil || *entry.value != "v:13" {
			t.Fatal("unexpected changed-entry cached value")
		}
	})
}

func TestBeamSyncEntryRecomputesCachedPrevAcrossDoorTree(t *testing.T) {
	parentCtx, childCtx := newNestedBeamContexts(t)
	source := NewSourceEqual(0, func(new int, old int) bool {
		return new == old
	})
	derived := NewBeamEqual(source, func(v int) int {
		return v % 2
	}, func(new int, old int) bool {
		return new == old
	})
	typed, ok := derived.(*beam[int, int])
	if !ok {
		t.Fatal("expected internal beam implementation")
	}

	parentUpdates := make(chan int, 4)
	childUpdates := make(chan int, 4)
	triggered := false

	if !derived.Sub(parentCtx, func(ctx context.Context, value int) bool {
		parentUpdates <- value
		if value == 1 && !triggered {
			triggered = true
			source.Update(ctx, 3)
		}
		return false
	}) {
		t.Fatal("expected parent derived sub to register")
	}
	if !derived.Sub(childCtx, func(ctx context.Context, value int) bool {
		childUpdates <- value
		return false
	}) {
		t.Fatal("expected child derived sub to register")
	}

	if got := expectInt(t, parentUpdates); got != 0 {
		t.Fatal("unexpected parent initial derived value:", got)
	}
	if got := expectInt(t, childUpdates); got != 0 {
		t.Fatal("unexpected child initial derived value:", got)
	}

	source.Update(parentCtx, 1)
	if got := expectInt(t, parentUpdates); got != 1 {
		t.Fatal("expected parent to observe the first odd derived value:", got)
	}

	deadline := time.Now().Add(time.Second)
	for {
		typed.mu.Lock()
		entry, has := typed.values[3]
		prev := uint(0)
		if has {
			prev = entry.prev
		}
		typed.mu.Unlock()
		if has && prev == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for cached seq to be recomputed against child prev")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
