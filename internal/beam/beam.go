// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package beam

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/shredder"
)

type Beam[T any] interface {
	// Sub subscribes to the value stream. The onValue callback is called immediately
	// with the current value (in the same goroutine), and again on every update.
	// Only instance runtime context is allowed.
	//
	// The subscription continues until:
	//   - The onValue function returns true (indicating done)
	//   - Unmount of any dynamic parent element
	//
	// Returns true if the subscription was successfully established;
	// false means the context was already canceled.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// XSub is an extended version of Sub that provides additional control.
	// It behaves the same as Sub, but also:
	//   - Accepts an onCancel callback, invoked when the subscription ends due to context cancellation
	//   - Returns a Cancel function for manual subscription termination
	//
	// Returns the Cancel function and a boolean indicating whether the subscription was established.
	XSub(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (context.CancelFunc, bool)

	// ReadAndSub returns the current value and then subscribes to future updates.
	// The onValue function is invoked on every subsequent update.
	// Only instance runtime context is allowed.
	//
	// Returns the initial value and a boolean:
	//   - If true, the value is valid and subscription was established
	//   - if false, the context was canceled or does not belong instance runtime.
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// ReadAndSubExt behaves like ReadAndSub with extended control options.
	// It provides the same functionality as ReadAndSub, but also:
	//   - Accepts an onCancel callback for handling cancellation events
	//   - Returns a Cancel function for manual termination
	//
	// Returns the initial value, Cancel function, and success boolean.
	// If the boolean is false, the value is undefined and no subscription was established.
	XReadAndSub(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (T, context.CancelFunc, bool)

	// Read returns the current value of the Beam without establishing a subscription.
	// Only instance runtime context is allowed.
	//
	// Returns the current value and a boolean:
	//   - If true, the value is valid
	//   - if false, the context was canceled or does not belong instance runtime.
	Read(ctx context.Context) (T, bool)

	// AddWatcher attaches a Watcher for full lifecycle control over subscription events.
	// Watchers receive separate callbacks for initialization, updates, and cancellation,
	// allowing for more sophisticated subscription management.
	// Only instance runtime context is allowed.
	//
	// Returns a Cancel function and a boolean indicating whether the watcher was added.
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
