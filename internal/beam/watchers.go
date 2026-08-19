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
)

// Watcher is a subscriber to a value stream.
type Watcher[T any] interface {
	// Cancel runs when the subscription ends, whether the context was canceled
	// or the returned cancel function was called.
	Cancel()
	// Watch receives the current value on the calling goroutine, then every
	// update. Return true to end the subscription.
	Watch(ctx context.Context, value T) bool
}

func none() {}

func (b *beam[T1, T2]) Watch(ctx context.Context, w Watcher[T2]) (context.CancelFunc, bool) {
	watcher := newWatcher(newSingleWatcher(b, w))
	ok := b.addWatcher(ctx, watcher)
	if !ok {
		return none, false
	}
	return watcher.Cancel, true
}
func (s *source[T]) Watch(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool) {
	watcher := newWatcher(newSingleWatcher(s, w))
	ok := s.addWatcher(ctx, watcher)
	if !ok {
		return none, false
	}
	return watcher.Cancel, true
}

type genericWatcher[T any] struct {
	watch  func(context.Context, T) bool
	cancel func()
}

func (l *genericWatcher[T]) Watch(ctx context.Context, value T) bool {
	return l.watch(ctx, value)
}

func (l *genericWatcher[T]) Cancel() {
	if l.cancel == nil {
		return
	}
	l.cancel()
}

func sub[T any](b Beamer[T], ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (context.CancelFunc, bool) {
	cancel, ok := b.Watch(ctx, &genericWatcher[T]{
		watch: func(ctx context.Context, v T) bool {
			return onValue(ctx, v)
		},
		cancel: onCancel,
	})
	return cancel, ok
}

func readAndSub[T any](b Beamer[T], ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (*T, context.CancelFunc, bool) {
	var init *T
	started := false
	cancel, ok := b.Watch(ctx, &genericWatcher[T]{
		watch: func(ctx context.Context, v T) bool {
			if started {
				return onValue(ctx, v)
			}
			started = true
			init = new(T)
			*init = v
			return onValue == nil
		},
		cancel: onCancel,
	})
	if !ok {
		return nil, cancel, false
	}
	v := init
	init = nil
	return v, cancel, true
}

func (b *beam[T1, T2]) ReadAndSub(ctx context.Context, onValue func(context.Context, T2) bool) (T2, bool) {
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

func (b *beam[T1, T2]) Read(ctx context.Context) (T2, bool) {
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

func (b *beam[T1, T2]) Sub(ctx context.Context, onValue func(context.Context, T2) bool) bool {
	_, ok := sub(b, ctx, onValue, nil)
	return ok
}

func (s *source[T]) Sub(ctx context.Context, onValue func(context.Context, T) bool) bool {
	_, ok := sub(s, ctx, onValue, nil)
	return ok
}
