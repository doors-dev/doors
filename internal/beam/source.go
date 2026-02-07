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
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type Source[T any] interface {
	Beam[T]

	// Update sets a new value and propagates it to all subscribers and derived beams.
	// The update is applied only if it passes the source's distinct function and provided
	// context is valid
	//
	Update(context.Context, T)

	// XUpdate performs an update and returns a channel that signals when the update
	// has been fully propagated to all subscribers. This allows coordination of
	// dependent operations that must wait for the update to complete.
	//
	// The returned channel receives nil on successful propagation or an error if
	// provided context is invalid or instance ended before propagation finished.
	//
	// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
	//
	// Returns the completion channel
	XUpdate(context.Context, T) <-chan error

	// Mutate allows modifying the current value using the provided function.
	// The function receives a copy of the current value and returns a new one.
	// The mutation is applied only if the result passes the source's distinct function.
	// Return of copy without modification will do nothing (if distinct function != nil)

	Mutate(context.Context, func(T) T)

	// XMutate performs a mutation and returns a channel that signals when the mutation
	// has been fully propagated to all subscribers. This allows coordination of
	// dependent operations that must wait for the mutation to complete.
	//
	// The returned channel receives nil on successful propagation or an error if
	// provided context is invalid or instance ended before propagation finished.
	// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
	//
	// Returns the completion channel
	XMutate(context.Context, func(T) T) <-chan error

	// Latest returns the most recently set or mutated value without requiring a context.
	// This provides direct access to the current state and is not affected by
	// context cancellation and doors tree state, unlike Read.
	//
	// WARNING: Latest() does not participate in render cycle consistency guarantees.
	// Use Read() to ensure consistent values across the component tree.
	Latest() T

	// DisableSkipping makes data propagation continue even if a new value
	// is issued. Useful, if you use beam as a communication channel
	// and want all data to be delivered to subscribers.
	DisableSkipping()
}

type anySource interface {
	getID() common.ID
	addSub(s *screen)
	removeSub(s *screen)
}

type source[T any] struct {
	id     common.ID
	seq    uint
	values map[uint]*T
	equal  func(new T, old T) bool
	mu     sync.RWMutex
	noSkip bool
	subs   common.Set[*screen]
	null   T
}

func (s *source[T]) getID() common.ID {
	return s.id
}

func (s *source[T]) addSub(sc *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.subs == nil {
		s.subs = common.NewSet[*screen]()
	}
	s.subs.Add(sc)
	sc.init(s, s.seq)
}

func (s *source[T]) removeSub(sc *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs.Remove(sc)
}

func NewSourceEqual[T any](init T, equal func(new T, old T) bool) Source[T] {
	return &source[T]{
		id:  common.NewID(),
		seq: 1,
		values: map[uint]*T{
			1: &init,
		},
		subs:  common.NewSet[*screen](),
		equal: equal,
	}
}

func equal[T comparable](new T, old T) bool {
	return new == old
}

func NewSource[T comparable](init T) Source[T] {
	return NewSourceEqual(init, equal)
}

func (s *source[T]) DisableSkipping() {
	s.noSkip = true
}

func (s *source[T]) Latest() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.values[s.seq]
}

func (s *source[T]) sync(seq uint, _ shredder.SimpleFrame) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[seq]
	return value, ok
}

func (s *source[T]) XUpdate(ctx context.Context, v T) <-chan error {
	ctex.LogBlockingWarning(ctx, "SourceBeam", "XUpdate")
	return s.mutateOrUpdate(ctx, nil, &v)
}

func (s *source[T]) Update(ctx context.Context, v T) {
	s.mutateOrUpdate(ctx, nil, &v)
}

func (s *source[T]) XMutate(ctx context.Context, m func(T) T) <-chan error {
	ctex.LogBlockingWarning(ctx, "SourceBeam", "XMutate")
	return s.mutateOrUpdate(ctx, m, nil)
}

func (s *source[T]) Mutate(ctx context.Context, m func(T) T) {
	s.mutateOrUpdate(ctx, m, nil)
}

func (s *source[T]) mutateOrUpdate(ctx context.Context, mut func(T) T, value *T) <-chan error {
	s.mu.Lock()
	ch := make(chan error, 1)
	ctex.LogCanceled(ctx, "SourceBeam mutation")
	ctx = ctex.ClearBlockingCtx(ctx)

	seq, commited := s.commit(mut, value)
	if !commited {
		s.mu.Unlock()
		close(ch)
		return ch
	}

	if len(s.subs) == 0 {
		s.cleanBefore(seq)
		s.mu.Unlock()
		ch <- nil
		close(ch)
		return ch
	}

	done := ctex.WgAdd(ctx)

	stopped := atomic.Bool{}

	isStopped := func() bool {
		if s.noSkip {
			return false
		}
		if stopped.Load() {
			return true
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if seq == s.seq {
			return false
		}
		stopped.Store(true)
		return true
	}

	sh := shredder.Thread{}
	syncFrame := sh.Frame()
	checkFrame := sh.Frame()
	cleanFrame := &shredder.ValveFrame{}

	for _, sub := range s.subs.Slice() {
		sub.sync(true, ctx, cleanFrame, syncFrame, seq, isStopped)
	}

	syncFrame.Release()

	s.mu.Unlock()

	checkFrame.Run(nil, nil, func(bool) {
		defer done()
		ch <- nil
		close(ch)
		if stopped.Load() {
			return
		}
		checkFrame.Release()
	})

	cleanFrame.Run(nil, nil, func(bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.cleanBefore(seq)
	})

	return ch
}

func (s *source[T]) commit(mut func(T) T, value *T) (uint, bool) {
	prev := s.values[s.seq]
	var next *T
	switch true {
	case mut != nil:
		updated := mut(*prev)
		next = &updated
	case value != nil:
		next = value
	default:
		panic("SourceBeam: no value or mutation provided")
	}
	if s.equal != nil && s.equal(*prev, *next) {
		return 0, false
	}
	s.seq += 1
	seq := s.seq
	s.values[seq] = next
	return seq, true
}

func (s *source[T]) cleanBefore(seq uint) {
	for oldSeq := range s.values {
		if oldSeq < seq {
			delete(s.values, oldSeq)
		}
	}

}

type Core interface {
	Cinema() Cinema
}

func (s *source[T]) addWatcher(ctx context.Context, w *watcher) bool {
	core := ctx.Value(ctex.KeyCore).(Core)
	core.Cinema().addWatcher(s, w)
	return true
}
