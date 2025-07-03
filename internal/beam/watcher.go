package beam

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/node"
)

type watcher[T any] struct {
	beam Beam[T]
	w    Watcher[T]
	init chan struct{}
	done atomic.Bool
	s    *node.Screen
	id   uint
	seq  uint
	ctx  context.Context
}

func newWatcher[T any](beam Beam[T], w Watcher[T]) *watcher[T] {
	return &watcher[T]{
		beam: beam,
		w:    w,
		init: make(chan struct{}, 0),
		done: atomic.Bool{},
		s:    nil,
		id:   0,
		seq:  0,
	}
}

func (w *watcher[T]) GetId() uint {
	return w.id
}

func (w *watcher[T]) Cancel() {
	if !w.done.CompareAndSwap(false, true) {
		return
	}
	w.w.Cancel()
	w.s.UnregWatcher(w.id)
}

func (w *watcher[T]) Sync(ctx context.Context, seq uint, c *common.Collector) {
	<-w.init
	if w.done.Load() {
		return
	}
	v, updated := w.beam.sync(seq, c)
	if v == nil {
		panic("update sync logic error:" + fmt.Sprint(seq))
	}
	if !updated {
		return
	}
	ctx = ctxwg.Infect(ctx, w.ctx)
	done := w.w.Update(ctx, v, seq)
	if done {
		w.done.Store(true)
		w.s.UnregWatcher(w.id)
		return
	}
	return
}

func (w *watcher[T]) Init(ctx context.Context, s *node.Screen, id uint, seq uint) func() {
	w.seq = seq
	w.id = id
	w.s = s
	w.ctx = ctx
	v, _ := w.beam.sync(seq, nil)
	if v == nil {
		panic("init sync logic error:" + fmt.Sprint(seq))
	}
	return func() {
		defer close(w.init)
		done := w.w.Init(ctx, v, seq)
		if done {
			w.done.Store(true)
			w.s.UnregWatcher(w.id)
		}
	}
}
