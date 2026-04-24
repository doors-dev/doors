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
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type anySource interface {
	getID() common.ID
	addSub(s *screen)
	removeSub(s *screen)
}

var _ anySource = (*SourceBeam[any])(nil)
var _ Beam[any] = (*SourceBeam[any])(nil)

type SourceBeam[T any] struct {
	id     common.ID
	seq    uint
	values map[uint]*T
	equal  func(new T, old T) bool
	mu     sync.RWMutex
	noSkip bool
	subs   common.Set[*screen]
	null   T
}

func (s *SourceBeam[T]) getID() common.ID {
	return s.id
}

func (s *SourceBeam[T]) addSub(sc *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.subs == nil {
		s.subs = common.NewSet[*screen]()
	}
	s.subs.Add(sc)
	sc.init(s, s.seq)
}

func (s *SourceBeam[T]) removeSub(sc *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs.Remove(sc)
}

func neverEqual[T any](T, T) bool {
	return false
}

// NewSourceEqual creates a SourceBeam with a custom equality function.
func NewSourceEqual[T any](init T, equal func(new T, old T) bool) *SourceBeam[T] {
	if equal == nil {
		equal = neverEqual[T]
	}
	return &SourceBeam[T]{
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

// NewSource creates a SourceBeam that uses `==` to suppress equal updates.
func NewSource[T comparable](init T) *SourceBeam[T] {
	return NewSourceEqual(init, equal)
}

// DisableSkipping forces every committed value to propagate.
func (s *SourceBeam[T]) DisableSkipping() {
	s.noSkip = true
}

// Get returns the latest committed value.
func (s *SourceBeam[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.values[s.seq]
}

func (s *SourceBeam[T]) sync(prev uint, seq uint, _ shredder.SimpleFrame) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[seq]
	if !ok {
		return nil, false
	}
	if prev == 0 {
		return value, true
	}
	if prev == seq-1 {
		return value, true
	}
	prevValue, ok := s.values[prev]
	if !ok {
		return value, true
	}
	return value, !s.equal(*value, *prevValue)
}

// XUpdate behaves like [Source.Update] and returns a completion channel.
func (s *SourceBeam[T]) XUpdate(ctx context.Context, v T) <-chan error {
	ctex.LogFreeWarning(ctx, "SourceBeam", "XUpdate")
	return s.mutateOrUpdate(ctx, nil, &v)
}

// Update stores v and starts propagation.
func (s *SourceBeam[T]) Update(ctx context.Context, v T) {
	s.mutateOrUpdate(ctx, nil, &v)
}

// XMutate behaves like [Source.Mutate] and returns a completion channel.
func (s *SourceBeam[T]) XMutate(ctx context.Context, m func(T) T) <-chan error {
	ctex.LogFreeWarning(ctx, "SourceBeam", "XMutate")
	return s.mutateOrUpdate(ctx, m, nil)
}

// Mutate updates the value by applying m to the current value.
func (s *SourceBeam[T]) Mutate(ctx context.Context, m func(T) T) {
	s.mutateOrUpdate(ctx, m, nil)
}

func (s *SourceBeam[T]) mutateOrUpdate(ctx context.Context, mut func(T) T, value *T) <-chan error {
	s.mu.Lock()
	ch := make(chan error, 1)
	ctex.LogCanceled(ctx, "SourceBeam mutation")
	ctx = ctex.ClearFreeCtx(ctx)

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
	ctxFrame := ctex.GetFrames(ctx).Call()

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
	syncFrame := shredder.Join(true, sh.Frame())
	checkFrame := shredder.Join(true, ctxFrame, sh.Frame())
	cleanFrame := &shredder.ValveFrame{}

	for _, sub := range s.subs.Slice() {
		sub.sync(true, ctx, cleanFrame, syncFrame, seq, isStopped)
	}

	syncFrame.Release()

	s.mu.Unlock()

	checkFrame.Run(nil, nil, func(bool) {
		ch <- nil
		close(ch)
		if stopped.Load() {
			return
		}
		cleanFrame.Activate()
	})

	checkFrame.Release()

	cleanFrame.Run(nil, nil, func(bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.cleanBefore(seq)
	})

	return ch
}

func (s *SourceBeam[T]) commit(mut func(T) T, value *T) (uint, bool) {
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

func (s *SourceBeam[T]) cleanBefore(seq uint) {
	for oldSeq := range s.values {
		if oldSeq < seq {
			delete(s.values, oldSeq)
		}
	}

}

type Core interface {
	Cinema() Cinema
}

func (s *SourceBeam[T]) addWatcher(ctx context.Context, w *watcher) bool {
	core, ok := ctx.Value(ctex.KeyCore).(Core)
	if !ok {
		return false
	}
	return core.Cinema().addWatcher(s, w)
}
