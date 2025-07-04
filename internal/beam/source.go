package beam

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/node"
	"github.com/doors-dev/doors/internal/shredder"
)

type instance interface {
	Thread() *shredder.Thread
	Cinema() *node.Cinema
	NewId() uint64
}

type SourceBeam[T any] interface {
	Beam[T]
	// Update sets a new value in the source and propagates it to all subscribers and derived beams.
	//
	// Returns true if the context is still valid and the update was accepted;
	// false if the context was canceled before the update.
	Update(context.Context, T) bool
	// XUpdate performs an extended update of the value and returns a channel
	// that signals when the update has been fully propagated to all subscribers.
	//
	// If there are no active subscribers, the channel is closed immediately.
	// If the instance has been terminated, the returned error will be non-nil.
	//
	// Returns:
	//   - A receive-only channel that yields a single error or closes silently.
	//   - A boolean indicating whether the update was accepted (context still valid).
	XUpdate(context.Context, T) (<-chan error, bool)
	// Mutate allows modifying the current value using the provided function.
	// The function receives a pointer to the copy of a current value and returns true to apply the changes.
	//
	// If the context is canceled before mutation, the function is not invoked and false is returned.
	Mutate(context.Context, func(*T) bool) bool
	// XMutate performs an extended mutation. It behaves like Mutate, but also returns
	// a channel that signals when the mutation has been fully propagated to all subscribers.
	//
	// If the mutation function is not applied (returns false or context is canceled), the channel is closed.
	// If no subscribers are present, the channel is closed immediately.
	//
	// Returns:
	//   - A receive-only channel that yields a single error or closes silently.
	//   - A boolean indicating whether the mutation was accepted.
	XMutate(context.Context, func(*T) bool) (<-chan error, bool)
	// Latest returns the most recently set or mutated value.
	// Unlike Read, this does not depend on context and always returns latest value directly.
	Latest() T
}

type source[T any] struct {
	null   T
	seq    uint
	values map[uint]*T
	inst   instance
	id     uint64
	init   sync.Once
	upd    func(new T, old T) bool
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
		upd:  distinct,
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
	value, _ := s.values[s.seq]
	return *value
}

func (s *source[T]) sync(seq uint, _ *common.FuncCollector) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[seq]
	return value, ok
}

func (s *source[T]) update(ctx context.Context, v *T) (<-chan error, bool) {
	return s.applyMutation(ctx, func(l *T) (*T, bool) {
		if s.upd != nil && !s.upd(*l, *v) {
			return nil, false
		}
		return v, true
	})
}

func (s *source[T]) XUpdate(ctx context.Context, v T) (<-chan error, bool) {
	common.LogBlockingWarning(ctx, "SourceBeam", "XUpdate")
	return s.update(ctx, &v)
}

func (s *source[T]) Update(ctx context.Context, v T) bool {
	_, ok := s.update(ctx, &v)
	return ok
}

func (s *source[T]) mutate(ctx context.Context, m func(*T) bool) (<-chan error, bool) {
	return s.applyMutation(ctx, func(l *T) (*T, bool) {
		copy := *l
		apply := m(&copy)
		if !apply || (s.upd != nil && !s.upd(*l, copy)) {
			return nil, false
		}
		return &copy, true
	})
}

func (s *source[T]) XMutate(ctx context.Context, m func(*T) bool) (<-chan error, bool) {
	common.LogBlockingWarning(ctx, "SourceBeam", "XMutate")
	return s.mutate(ctx, m)
}
func (s *source[T]) Mutate(ctx context.Context, m func(*T) bool) bool {
	_, ok := s.mutate(ctx, m)
	return ok
}

func (s *source[T]) applyMutation(ctx context.Context, m func(*T) (*T, bool)) (<-chan error, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan error, 1)
	ctx = common.ClearBlockingCtx(ctx)
	new, update := m(s.values[s.seq])
	if !update {
		close(ch)
		return ch, true
	}
	s.seq += 1
	seq := s.seq
	s.values[seq] = new
	if s.inst == nil {
		delete(s.values, seq-1)
		ch <- nil
		close(ch)
		return ch, true
	}
	cinema := s.inst.Cinema()
	done := ctxwg.Add(ctx)
	syncThread := s.inst.Thread()
	c := common.NewFuncCollector()
	cinema.InitSync(syncThread, ctx, s.id, seq, c)
	syncThread.WriteStarving(func(t *shredder.Thread) {
		defer done()
		c.Apply()
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.values, seq-1)
		ch <- nil
		close(ch)
	})
	return ch, true

}

func (s *source[T]) addWatcher(ctx context.Context, w node.Watcher) bool {
	cinema := ctx.Value(common.CinemaCtxKey).(*node.Cinema)
	inst := ctx.Value(common.InstanceCtxKey).(instance)
	s.init.Do(func() {
		s.inst = inst
		s.id = inst.NewId()
	})
	return cinema.AddWatcher(ctx, s.id, w, s.seq)
}
