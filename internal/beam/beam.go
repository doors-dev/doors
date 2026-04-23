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
	"sync"

	"github.com/doors-dev/doors/internal/shredder"
)

type Beam[T any] interface {
	Watch(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool)
	addWatcher(ctx context.Context, w *watcher) bool
	sync(uint, uint, shredder.SimpleFrame) (*T, bool)
}

type entry[T any] struct {
	value   *T
	prev    uint
	updated bool
}

func NewBeamEqual[T1 any, T2 any](source Beam[T1], cast func(T1) T2, equal func(new T2, old T2) bool) *DerivedBeam[T1, T2] {
	if equal == nil {
		equal = neverEqual[T2]
	}
	return &DerivedBeam[T1, T2]{
		source: source,
		values: make(map[uint]entry[T2]),
		mu:     sync.Mutex{},
		cast:   cast,
		equal:  equal,
	}
}

func NewBeam[T1 any, T2 comparable](source Beam[T1], cast func(T1) T2) *DerivedBeam[T1, T2] {
	equal := func(new T2, old T2) bool {
		return new == old
	}
	return &DerivedBeam[T1, T2]{
		source: source,
		values: make(map[uint]entry[T2]),
		mu:     sync.Mutex{},
		cast:   cast,
		equal:  equal,
	}
}

var _ Beam[any] = (*DerivedBeam[any, any])(nil)

type DerivedBeam[T1 any, T2 any] struct {
	source Beam[T1]
	values map[uint]entry[T2]
	mu     sync.Mutex
	cast   func(T1) T2
	equal  func(new T2, old T2) bool
	null   T2
}

func (b *DerivedBeam[T1, T2]) addWatcher(ctx context.Context, w *watcher) bool {
	return b.source.addWatcher(ctx, w)
}

func (b *DerivedBeam[T1, T2]) syncEntry(prev, seq uint, after shredder.SimpleFrame) (v *T2, u bool) {
	e, has := b.values[seq]
	if has {
		if prev == 0 {
			return e.value, true
		}
		if e.prev == prev {
			return e.value, e.updated
		}
		prevValue, has := b.values[prev]
		if !has {
			return e.value, true
		}
		e.updated = b.equal(*e.value, *prevValue.value)
		e.prev = prev
		b.values[seq] = e
		return e.value, e.updated
	}
	if after != nil {
		after.Run(nil, nil, func(bool) {
			b.mu.Lock()
			defer b.mu.Unlock()
			for s := range b.values {
				if s < seq {
					delete(b.values, s)
				}
			}
		})
	}
	sourceVal, updated := b.source.sync(prev, seq, after)
	if sourceVal == nil {
		return nil, false
	}
	if !updated {
		prevValue, has := b.values[prev]
		if has {
			return prevValue.value, false
		}
		value := b.cast(*sourceVal)
		b.values[seq] = entry[T2]{
			value:   &value,
			prev:    prev,
			updated: false,
		}
		return &value, false
	}
	newValue := b.cast(*sourceVal)
	prevValue, has := b.values[prev]
	if !has {
		b.values[seq] = entry[T2]{
			value:   &newValue,
			prev:    prev,
			updated: true,
		}
		return &newValue, true
	}
	updated = !b.equal(newValue, *prevValue.value)
	if !updated {
		b.values[seq] = entry[T2]{
			value:   prevValue.value,
			prev:    prev,
			updated: false,
		}
		return prevValue.value, false
	}
	b.values[seq] = entry[T2]{
		value:   &newValue,
		prev:    prev,
		updated: true,
	}
	return &newValue, true
}

func (b *DerivedBeam[T1, T2]) sync(prev uint, seq uint, after shredder.SimpleFrame) (*T2, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.syncEntry(prev, seq, after)
}
