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
	"fmt"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/shredder"
)

type watcherResult int

const (
	watch watcherResult = iota
	done
)

type innerWatcher interface {
	init(ctx context.Context, seq uint) watcherResult
	sync(ctx context.Context, seq uint, cleanFrame shredder.SimpleFrame) watcherResult
	cancel()
}

func newWatcher(inner innerWatcher) *watcher {
	return &watcher{
		inner: inner,
	}
}

const (
	watcherInit int32 = iota
	watcherReady
	wathcherSync
	watcherDone
)

type watcher struct {
	state     atomic.Int32
	inner     innerWatcher
	initGuard shredder.ValveFrame
	screen    *screen
}

func (w *watcher) register(screen *screen) {
	w.screen = screen
}

func (w *watcher) unregister() {
	w.screen.removeWatcher(w)
}

func (w *watcher) Cancel() {
	prev := w.state.Swap(watcherDone)
	if prev != watcherReady {
		return
	}
	w.unregister()
	w.inner.cancel()
}

func (w *watcher) syncFrame() shredder.AnyFrame {
	return &w.initGuard
}

func (w *watcher) init(ctx context.Context, seq uint) {
	res := w.inner.init(ctx, seq)
	switch res {
	case watch:
		ok := w.state.CompareAndSwap(watcherInit, watcherReady)
		w.initGuard.Activate()
		if ok {
			return
		}
		w.unregister()
		w.inner.cancel()
	case done:
		w.state.Store(watcherDone)
		w.initGuard.Activate()
		w.unregister()
	}
}

func (w *watcher) sync(ctx context.Context, seq uint, cleanFrame shredder.SimpleFrame) {
	ok := w.state.CompareAndSwap(watcherReady, wathcherSync)
	if !ok {
		return
	}
	res := w.inner.sync(ctx, seq, cleanFrame)
	if res == done {
		w.state.Store(watcherDone)
		w.unregister()
		return
	}
	if w.state.CompareAndSwap(wathcherSync, watcherReady) {
		return
	}
	w.unregister()
	w.inner.cancel()
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
	seq  uint
}

func (s *singleWatcher[T]) cancel() {
	s.w.Cancel()
}

func (s *singleWatcher[T]) init(ctx context.Context, seq uint) watcherResult {
	s.seq = seq
	v, _ := s.beam.sync(0, seq, nil)
	if v == nil {
		panic("init sync logic error:" + fmt.Sprint(seq))
	}
	if s.w.Watch(ctx, *v) {
		return done
	}
	return watch
}

func (s *singleWatcher[T]) sync(ctx context.Context, seq uint, cleanFrame shredder.SimpleFrame) watcherResult {
	v, updated := s.beam.sync(s.seq, seq, cleanFrame)
	s.seq = seq
	if v == nil {
		panic("update sync logic error:" + fmt.Sprint(seq))
	}
	if !updated {
		return watch
	}
	if s.w.Watch(ctx, *v) {
		return done
	}
	return watch
}
