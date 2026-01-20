// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package beam2

import (
	"context"
)

// Watcher defines hooks for observing and reacting to the lifecycle of a Beam value stream.
// Implementers can perform custom logic during initialization, on each update, and when canceled.
type Watcher[T any] interface {
	// Cancel is called when the watcher is terminated due to context cancellation.
	Cancel()

	// Init is called with the initial value. The seq parameter represents the sequence number
	// of the update. Called in the same goroutine where the watcher was added.
	//
	// Return true (done) to stop receiving updates after this call.
	Init(ctx context.Context, value *T, seq uint) bool

	// Update is called for each subsequent update to the value.
	// The seq parameter increments with each update.
	//
	// Return true (done) to stop receiving further updates.
	Update(ctx context.Context, value *T, seq uint) bool
}

func none() {}

func (b *beam[T, T2]) AddWatcher(ctx context.Context, w Watcher[T2]) (context.CancelFunc, bool) {
	watcher := newWatcher(ctx, b, w)
	ok := b.addWatcher(ctx, watcher)
	if !ok {
		return none, false
	}
	return watcher.Cancel, true
}
func (s *source[T]) AddWatcher(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool) {
	watcher := newWatcher(ctx, s, w)
	ok := s.addWatcher(ctx, watcher)
	if !ok {
		return none, false
	}
	return watcher.Cancel, true
}

type genericWatcher[T any] struct {
	init   func(context.Context, *T, uint) bool
	update func(context.Context, *T, uint) bool
	cancel func()
}

func (l *genericWatcher[T]) Init(ctx context.Context, value *T, seq uint) bool {
	return l.init(ctx, value, seq)
}

func (l *genericWatcher[T]) Update(ctx context.Context, v *T, seq uint) bool {
	return l.update(ctx, v, seq)
}

func (l *genericWatcher[T]) Cancel() {
	if l.cancel != nil {
		l.cancel()
	}
}

func sub[T any](b Beam[T], ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (context.CancelFunc, bool) {
	cancel, ok := b.AddWatcher(ctx, &genericWatcher[T]{
		init: func(ctx context.Context, v *T, _ uint) bool {
			return onValue(ctx, *v)
		},
		update: func(ctx context.Context, v *T, _ uint) bool {
			return onValue(ctx, *v)
		},
		cancel: onCancel,
	})
	return cancel, ok
}

func readAndSub[T any](b Beam[T], ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (*T, context.CancelFunc, bool) {
	var initVal *T
	cancel, ok := b.AddWatcher(ctx, &genericWatcher[T]{
		init: func(_ context.Context, v *T, _ uint) bool {
			initVal = v
			return onValue == nil
		},
		update: func(ctx context.Context, v *T, _ uint) bool {
			return onValue(ctx, *v)
		},
		cancel: onCancel,
	})
	if !ok {
		return nil, cancel, false
	}
	return initVal, cancel, true
}

func (b *beam[T, T2]) ReadAndSub(ctx context.Context, onValue func(context.Context, T2) bool) (T2, bool) {
	v, _, ok := readAndSub(b, ctx, onValue, nil)
	if !ok {
		return b.null, false
	}
	return *v, true
}

func (s *source[T]) ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool) {
	v, _, ok := readAndSub(s, ctx, onValue, nil)
	if !ok {
		return s.null, false
	}
	return *v, true
}

func (b *beam[T, T2]) XReadAndSub(ctx context.Context, onValue func(context.Context, T2) bool, onCancel func()) (T2, context.CancelFunc, bool) {
	v, cancel, ok := readAndSub(b, ctx, onValue, onCancel)
	if !ok {
		return b.null, cancel, false
	}
	return *v, cancel, ok
}

func (s *source[T]) XReadAndSub(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (T, context.CancelFunc, bool) {
	v, cancel, ok := readAndSub(s, ctx, onValue, onCancel)
	if !ok {
		return s.null, cancel, false
	}
	return *v, cancel, ok
}

func (b *beam[T, T2]) Read(ctx context.Context) (T2, bool) {
	v, _, ok := readAndSub(b, ctx, nil, nil)
	if !ok {
		return b.null, false
	}
	return *v, ok
}

func (s *source[T]) Read(ctx context.Context) (T, bool) {
	v, _, ok := readAndSub(s, ctx, nil, nil)
	if !ok {
		return s.null, false
	}
	return *v, ok
}

func (b *beam[T, T2]) XSub(ctx context.Context, onValue func(context.Context, T2) bool, onCancel func()) (context.CancelFunc, bool) {
	return sub(b, ctx, onValue, onCancel)
}

func (s *source[T]) XSub(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (context.CancelFunc, bool) {
	return sub(s, ctx, onValue, onCancel)
}

func (b *beam[T, T2]) Sub(ctx context.Context, onValue func(context.Context, T2) bool) bool {
	_, ok := sub(b, ctx, onValue, nil)
	return ok
}
func (s *source[T]) Sub(ctx context.Context, onValue func(context.Context, T) bool) bool {
	_, ok := sub(s, ctx, onValue, nil)
	return ok
}
