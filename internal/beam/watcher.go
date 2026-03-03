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
	"fmt"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type initResult int

const (
	initContinue initResult = iota
	initWatch
	initDone
)

type innerWatcher interface {
	init(id int, ctx context.Context, seq uint) initResult
	sync(id int, ctx context.Context, seq uint, cleanFrame shredder.SimpleFrame) bool
	cancel()
}

func newWatcher(inner innerWatcher) *watcher {
	return &watcher{
		screens:   common.NewSet[*screen](),
		inner:     inner,
		initGuard: make(chan struct{}),
	}
}

type watcher struct {
	mu        sync.Mutex
	done      bool
	inner     innerWatcher
	initGuard chan struct{}
	screens   common.Set[*screen]
	id        int
}

func (w *watcher) register(screen *screen) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.done {
		panic("Can't be done before all screens are registered")
	}
	w.screens.Add(screen)
}

func (w *watcher) unregister() {
	for s := range w.screens {
		s.removeWatcher(w)
	}
}

func (w *watcher) Cancel() {
	w.mu.Lock()
	if w.done {
		w.mu.Unlock()
		return
	}
	w.done = true
	w.mu.Unlock()
	w.unregister()
	w.inner.cancel()
}

func (w *watcher) init(ctx context.Context, seq uint) {
	w.mu.Lock()
	if w.done {
		w.mu.Unlock()
		panic("Can't be done before all screens are initialized")
	}
	res := w.inner.init(w.id, ctx, seq)
	switch res {
	case initContinue:
		close(w.initGuard)
		w.mu.Unlock()
	case initWatch:
		close(w.initGuard)
		w.mu.Unlock()
	case initDone:
		close(w.initGuard)
		w.done = true
		w.mu.Unlock()
		w.unregister()
	}
}

func (w *watcher) sync(ctx context.Context, seq uint, cleanFrame shredder.SimpleFrame) {
	<-w.initGuard
	w.mu.Lock()
	if w.done {
		w.mu.Unlock()
		return
	}
	done := w.inner.sync(w.id, ctx, seq, cleanFrame)
	w.done = done
	w.mu.Unlock()
	if done {
		w.unregister()
	}
}

func newSingleWatcher[T any](beam Beam[T], w Watcher[T]) innerWatcher {
	return &singleWatcher[T]{
		beam: beam,
		w:    w,
	}
}

type singleWatcher[T any] struct {
	beam Beam[T]
	w    Watcher[T]
	ctx  context.Context
	seq  uint
}

func (s *singleWatcher[T]) cancel() {
	s.w.Cancel()
}

func (s *singleWatcher[T]) init(_id int, ctx context.Context, seq uint) initResult {
	s.ctx = ctx
	s.seq = seq
	v, _ := s.beam.sync(0, seq, nil)
	if v == nil {
		panic("init sync logic error:" + fmt.Sprint(seq))
	}
	if s.w.Watch(ctx, *v) {
		return initDone
	}
	return initWatch
}

func (s *singleWatcher[T]) sync(_id int, ctx context.Context, seq uint, cleanFrame shredder.SimpleFrame) bool {
	v, updated := s.beam.sync(s.seq, seq, cleanFrame)
	s.seq = seq
	if v == nil {
		panic("update sync logic error:" + fmt.Sprint(seq))
	}
	if !updated {
		return false
	}
	ctx = ctex.WgInfect(ctx, s.ctx)
	return s.w.Watch(ctx, *v)
}

var _ innerWatcher = &singleWatcher[any]{}
