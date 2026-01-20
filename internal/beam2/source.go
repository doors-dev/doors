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
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/sh"
)


type SourceBeam[T any] interface {
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
	null   T
	seq    uint
	values map[uint]*T
	init   sync.Once
	equal  func(new T, old T) bool
	mu     sync.RWMutex
	noSkip bool
	subs   common.Set[*screen]
}

func (s *source[T]) getID() common.ID {
	return s.id
}

func (s *source[T]) addSub(sc *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs.Add(sc)
	sc.init(s, s.seq)
}

func (s *source[T]) removeSub(sc *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs.Remove(sc)
}

func NewSourceBeamEqual[T any](init T, equal func(new T, old T) bool) SourceBeam[T] {
	return &source[T]{
		id:  common.NewID(),
		seq: 1,
		values: map[uint]*T{
			1: &init,
		},
		equal: equal,
	}
}

func equal[T comparable](new T, old T) bool {
	return new == old
}

func NewSourceBeam[T comparable](init T) SourceBeam[T] {
	return NewSourceBeamEqual(init, equal)
}

func (s *source[T]) DisableSkipping() {
	s.noSkip = true
}

func (s *source[T]) Latest() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.values[s.seq]
}

func (s *source[T]) sync(seq uint, _ sh.SimpleFrame) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[seq]
	return value, ok
}

func (s *source[T]) update(ctx context.Context, v *T) <-chan error {
	return s.applyMutation(ctx, func(l *T) (*T, bool) {
		if s.equal != nil && s.equal(*l, *v) {
			return nil, false
		}
		return v, true
	})
}

func (s *source[T]) XUpdate(ctx context.Context, v T) <-chan error {
	common.LogBlockingWarning(ctx, "SourceBeam", "XUpdate")
	return s.update(ctx, &v)
}

func (s *source[T]) Update(ctx context.Context, v T) {
	s.update(ctx, &v)
}

func (s *source[T]) mutate(ctx context.Context, m func(T) T) <-chan error {
	return s.applyMutation(ctx, func(l *T) (*T, bool) {
		new := m(*l)
		if s.equal != nil && s.equal(*l, new) {
			return nil, false
		}
		return &new, true
	})
}

func (s *source[T]) XMutate(ctx context.Context, m func(T) T) <-chan error {
	common.LogBlockingWarning(ctx, "SourceBeam", "XMutate")
	return s.mutate(ctx, m)
}
func (s *source[T]) Mutate(ctx context.Context, m func(T) T) {
	s.mutate(ctx, m)
}

func (s *source[T]) applyMutation(ctx context.Context, m func(*T) (*T, bool)) <-chan error {
	s.mu.Lock()
	ch := make(chan error, 1)
	common.LogCanceled(ctx, "SourceBeam mutation")
	ctx = common.ClearBlockingCtx(ctx)
	new, update := m(s.values[s.seq])
	if !update {
		close(ch)
		s.mu.Unlock()
		return ch
	}
	s.seq += 1
	seq := s.seq
	s.values[seq] = new
	if len(s.subs) == 0 {
		delete(s.values, seq-1)
		s.mu.Unlock()
		ch <- nil
		close(ch)
		return ch
	}
	subs := s.subs.Slice()
	s.mu.Unlock()

	done := ctxwg.Add(ctx)
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

	sh := sh.Shread{}
	syncFrame := sh.Frame()
	defer syncFrame.Release()

	checkFrame := sh.Frame()

	finalFrame := sh.Frame()

	for _, sub := range subs {
		syncFrame.Run(nil, func() {
			sub.sync(ctx, finalFrame, syncFrame, seq, isStopped)
		})
	}

	checkFrame.Run(nil, func() {
		defer done()
		ch <- nil
		close(ch)
		if stopped.Load() {
			return
		}
		checkFrame.Release()
	})

	finalFrame.Run(nil, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for oldSeq := range s.values {
			if oldSeq < seq {
				delete(s.values, oldSeq)
			}
		}
	})

	return ch
}

type Core interface {
	Cinema() Cinema
}

func (s *source[T]) addWatcher(ctx context.Context, w *watcher) bool {
	core := ctx.Value(common.CtxKeyCore).(Core)
	core.Cinema().addWatcher(s, w)
	return true
}
