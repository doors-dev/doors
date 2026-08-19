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

type Beamer[T any] interface {
	Get() T
	Watch(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool)
	addWatcher(ctx context.Context, w *watcher) bool
	sync(uint, uint, shredder.SimpleFrame) (*T, bool)
	oldest() uint
}

type entry[T any] struct {
	value   *T
	prev    uint
	updated bool
}

type Beam[T1, T2 any] = *beam[T1, T2]

func NewBeam[T1 any, T2 any](source Beamer[T1], get func(T1) T2, equal func(new T2, old T2) bool) Beam[T1, T2] {
	if equal == nil {
		equal = NeverEqual
	}
	return &beam[T1, T2]{
		beam:   source,
		values: make(map[uint]entry[T2]),
		get:    get,
		equal:  equal,
	}
}

var _ Beamer[any] = (*beam[any, any])(nil)

type beam[T1 any, T2 any] struct {
	beam   Beamer[T1]
	values map[uint]entry[T2]
	mu     sync.Mutex
	get    func(T1) T2
	equal  func(new T2, old T2) bool
	null   T2
}

func (b *beam[T1, T2]) Get() T2 {
	return b.get(b.beam.Get())
}

func (b *beam[T1, T2]) addWatcher(ctx context.Context, w *watcher) bool {
	return b.beam.addWatcher(ctx, w)
}

func (b *beam[T1, T2]) syncEntry(prev, seq uint, after shredder.SimpleFrame) (v *T2, u bool) {
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
		e.updated = !b.equal(*e.value, *prevValue.value)
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
	} else {
		oldest := b.beam.oldest()
		for s := range b.values {
			if s < oldest {
				delete(b.values, s)
			}
		}
	}
	sourceVal, updated := b.beam.sync(prev, seq, after)
	if sourceVal == nil {
		return nil, false
	}
	if !updated {
		prevValue, has := b.values[prev]
		if has {
			return prevValue.value, false
		}
		value := b.get(*sourceVal)
		b.values[seq] = entry[T2]{
			value:   &value,
			prev:    prev,
			updated: false,
		}
		return &value, false
	}
	newValue := b.get(*sourceVal)
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

func (b *beam[T1, T2]) sync(prev uint, seq uint, after shredder.SimpleFrame) (*T2, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.syncEntry(prev, seq, after)
}

func (b *beam[T1, T2]) oldest() uint {
	return b.beam.oldest()
}
