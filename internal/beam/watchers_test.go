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

	"github.com/doors-dev/doors/internal/shredder"
)

type stubSyncSource[T any] struct {
	syncCalls int
	syncFunc  func(prev, seq uint, after shredder.SimpleFrame) (*T, bool)
}

func (s *stubSyncSource[T]) Sub(context.Context, func(context.Context, T) bool) bool {
	return false
}

func (s *stubSyncSource[T]) ReadAndSub(context.Context, func(context.Context, T) bool) (T, bool) {
	var zero T
	return zero, false
}

func (s *stubSyncSource[T]) Read(context.Context) (T, bool) {
	var zero T
	return zero, false
}

func (s *stubSyncSource[T]) Watch(context.Context, Watcher[T]) (context.CancelFunc, bool) {
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

func TestSourceXUpdateAndXMutate(t *testing.T) {
	source := NewSourceEqual(0, nil)

	if err := <-source.XUpdate(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if got := source.Get(); got != 1 {
		t.Fatal("unexpected source value after XUpdate:", got)
	}

	if err := <-source.XMutate(context.Background(), func(v int) int {
		return v + 1
	}); err != nil {
		t.Fatal(err)
	}
	if got := source.Get(); got != 2 {
		t.Fatal("unexpected source value after XMutate:", got)
	}
}

func TestNewBeamEqualDefaultsNilComparator(t *testing.T) {
	source := &stubSyncSource[int]{}
	derived := NewBeamEqual(source, func(v int) string {
		return fmt.Sprintf("v:%d", v)
	}, nil)

	if derived.equal == nil {
		t.Fatal("expected nil comparator to be replaced with a default implementation")
	}
	if derived.equal("same", "same") {
		t.Fatal("expected default comparator to treat equal values as changed")
	}
}

func TestBeamSyncEntryCachedBranches(t *testing.T) {
	t.Run("prev zero reuses cached value as updated", func(t *testing.T) {
		source := &stubSyncSource[int]{}
		value := 11
		b := &DerivedBeam[int, int]{
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
		b := &DerivedBeam[int, int]{
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
		b := &DerivedBeam[int, int]{
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
		b := &DerivedBeam[int, int]{
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
		b := &DerivedBeam[int, string]{
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

		b := &DerivedBeam[int, string]{
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
		b := &DerivedBeam[int, string]{
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
		b := &DerivedBeam[int, string]{
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
		b := &DerivedBeam[int, string]{
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
		b := &DerivedBeam[int, string]{
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
