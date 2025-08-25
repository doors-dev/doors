// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package beam

import (
	"context"
	"errors"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/shredder"
)

type instance interface {
	Thread() *shredder.Thread
	Cinema() *door.Cinema
	NewId() uint64
}

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
	// Returns the completion ch<D-s>annel
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
	//
	// Returns the completion channel
	XMutate(context.Context, func(T) T) <-chan error

	// Latest returns the most recently set or mutated value without requiring a context.
	// This provides direct access to the current state and is not affected by
	// context cancellation and Door tree state, unlike Read.
	//
	// WARNING: Latest() does not participate in render cycle consistency guarantees.
	// Use Read() to ensure consistent values across the component tree.
	Latest() T
}

type source[T any] struct {
	null   T
	seq    uint
	values map[uint]*T
	inst   instance
	id     uint64
	init   sync.Once
	distinct    func(new T, old T) bool
	mu     sync.RWMutex
}

func NewSourceBeamExt[T any](init T, distinct func(new T, old T) bool) SourceBeam[T] {
	return &source[T]{
		seq: 1,
		values: map[uint]*T{
			1: &init,
		},
		inst: nil,
		id:   0,
		init: sync.Once{},
		distinct:  distinct,
	}
}

func NewSourceBeam[T comparable](init T) SourceBeam[T] {
	upd := func(new T, old T) bool {
		return new != old
	}
	return NewSourceBeamExt(init, upd)
}

func (s *source[T]) Latest() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.values[s.seq]
}

func (s *source[T]) sync(seq uint, _ *common.FuncCollector) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[seq]
	return value, ok
}

func (s *source[T]) update(ctx context.Context, v *T) <-chan error {
	return s.applyMutation(ctx, func(l *T) (*T, bool) {
		if s.distinct != nil && !s.distinct(*l, *v) {
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
		if s.distinct != nil && !s.distinct(*l, new) {
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
	defer s.mu.Unlock()
	ch, ok := common.ResultChannel(ctx, "SourceBeam mutation")
	if !ok {
		return ch
	}
	ctx = common.ClearBlockingCtx(ctx)
	new, update := m(s.values[s.seq])
	if !update {
		close(ch)
		return ch
	}
	s.seq += 1
	seq := s.seq
	s.values[seq] = new
	if s.inst == nil {
		delete(s.values, seq-1)
		ch <- nil
		close(ch)
		return ch
	}
	cinema := s.inst.Cinema()
	done := ctxwg.Add(ctx)
	syncThread := s.inst.Thread()
	c := common.NewFuncCollector()
	cinema.InitSync(syncThread, ctx, s.id, seq, c)
	syncThread.WriteStarving(func(t *shredder.Thread) {
		defer done()
		if t == nil {
			ch <- errors.New("instance ended")
			close(ch)
			return
		}
		c.Apply()
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.values, seq-1)
		ch <- nil
		close(ch)
	})
	return ch

}

func (s *source[T]) addWatcher(ctx context.Context, w door.Watcher) bool {
	cinema := ctx.Value(common.DoorCtxKey).(door.Core).Cinema()
	inst := ctx.Value(common.InstanceCtxKey).(instance)
	s.init.Do(func() {
		s.inst = inst
		s.id = inst.NewId()
	})
	return cinema.AddWatcher(ctx, s.id, w, s.seq)
}
