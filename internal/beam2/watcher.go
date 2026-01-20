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
	"fmt"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/sh"
)

type anyInnerWatcher interface {
	setContext(context.Context)
	init(uint) bool
	sync(context.Context, uint, sh.SimpleFrame) bool
	cancel()
}

type innerWatcher[T any] struct {
	ctx  context.Context
	beam Beam[T]
	w    Watcher[T]
}

func (s *innerWatcher[T]) setContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *innerWatcher[T]) cancel() {
	s.w.Cancel()
}

func (s *innerWatcher[T]) init(seq uint) bool {
	v, _ := s.beam.sync(seq, nil)
	if v == nil {
		panic("init sync logic error:" + fmt.Sprint(seq))
	}
	return s.w.Init(s.ctx, v, seq)
}

func (s *innerWatcher[T]) sync(ctx context.Context, seq uint, after sh.SimpleFrame) bool {
	v, updated := s.beam.sync(seq, after)
	if v == nil {
		panic("update sync logic error:" + fmt.Sprint(seq))
	}
	if !updated {
		return false
	}
	ctx = ctxwg.Infect(ctx, s.ctx)
	return s.w.Update(ctx, v, seq)
}

type watcher struct {
	inner     anyInnerWatcher
	initGuard chan struct{}
	initSeq   uint
	done      atomic.Bool
	screen    *screen
	ctx       context.Context
}

func newWatcher[T any](ctx context.Context, beam Beam[T], w Watcher[T]) *watcher {
	return &watcher{
		inner: &innerWatcher[T]{
			beam: beam,
			w:    w,
			ctx:  ctx,
		},
		initGuard: make(chan struct{}),
	}
}

func (w *watcher) Cancel() {
	if !w.done.CompareAndSwap(false, true) {
		return
	}
	w.inner.cancel()
	w.screen.removeWatcher(w)
}

func (w *watcher) sync(ctx context.Context, seq uint, after sh.SimpleFrame) {
	<-w.initGuard
	if w.done.Load() {
		return
	}
	if w.inner.sync(ctx, seq, after) {
		if !w.done.CompareAndSwap(false, true) {
			return
		}
		w.screen.removeWatcher(w)
	}
}

func (w *watcher) init() {
	if w.initSeq == 0 {
		panic("watcher init seq is not set")
	}
	done := w.inner.init(w.initSeq)
	remove := done && w.done.CompareAndSwap(false, done)
	close(w.initGuard)
	if remove {
		w.screen.removeWatcher(w)
	}
}
