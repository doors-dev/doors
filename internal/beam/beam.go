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

// Beam is a read-only reactive value that can be read, subscribed to, or
// watched.
type Beam[T any] interface {
	// Effect returns the current value and rerenders the closest dynamic parent
	// when the value changes.
	Effect(ctx context.Context) (T, bool)

	// Sub subscribes to the value stream. onValue is called immediately with the
	// current value in the same goroutine, and again on every update.
	// Only instance runtime context is allowed.
	//
	// The subscription continues until:
	//   - onValue returns true, or
	//   - a dynamic parent is unmounted.
	//
	// It returns false if the context was already canceled or does not belong to
	// an instance runtime.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// ReadAndSub returns the current value and then subscribes to future
	// updates. onValue is called only for subsequent updates.
	// Only instance runtime context is allowed.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// Read returns the current value without creating a subscription.
	// Only instance runtime context is allowed.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	Read(ctx context.Context) (T, bool)

	// AddWatcher attaches a low-level watcher for separate init, update, and
	// cancellation callbacks. Only instance runtime context is allowed.
	AddWatcher(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool)

	addWatcher(ctx context.Context, w *watcher) bool
	sync(uint, uint, shredder.SimpleFrame) (*T, bool)
}

type entry[T any] struct {
	value   *T
	prev    uint
	updated bool
}

func NewBeamEqual[T1 any, T2 any](source Beam[T1], cast func(T1) T2, equal func(new T2, old T2) bool) Beam[T2] {
	if equal == nil {
		equal = func(T2, T2) bool {
			return false
		}
	}
	return &beam[T1, T2]{
		source: source,
		values: make(map[uint]entry[T2]),
		mu:     sync.Mutex{},
		cast:   cast,
		equal:  equal,
	}
}

func NewBeam[T1 any, T2 comparable](source Beam[T1], cast func(T1) T2) Beam[T2] {
	equal := func(new T2, old T2) bool {
		return new == old
	}
	return &beam[T1, T2]{
		source: source,
		values: make(map[uint]entry[T2]),
		mu:     sync.Mutex{},
		cast:   cast,
		equal:  equal,
	}
}

type beam[T1 any, T2 any] struct {
	source Beam[T1]
	values map[uint]entry[T2]
	mu     sync.Mutex
	cast   func(T1) T2
	equal  func(new T2, old T2) bool
	null   T2
}

func (b *beam[T1, T2]) addWatcher(ctx context.Context, w *watcher) bool {
	return b.source.addWatcher(ctx, w)
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

func (b *beam[T1, T2]) sync(prev uint, seq uint, after shredder.SimpleFrame) (*T2, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.syncEntry(prev, seq, after)
}
