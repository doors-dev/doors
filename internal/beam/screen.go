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

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type screen struct {
	mu               sync.Mutex
	sourceID         common.ID
	thread           shredder.Thread
	watcherSyncGuard shredder.ReadBlockingWriteThread
	cinema           *cinema
	parent           parentScreen
	seq              uint
	watchers         common.Set[*watcher]
	subs             common.Set[*screen]
	removeScheduled  bool
}

func (s *screen) init(parent parentScreen, seq uint) {
	s.parent = parent
	s.seq = seq
}

func (s *screen) addWatcher(w *watcher) (uint, shredder.Frame) {
	frame := shredder.Join(true, s.watcherSyncGuard.Read(), s.cinema.ReadFrame())
	s.mu.Lock()
	seq := s.seq
	if s.watchers == nil {
		s.watchers = common.NewSet[*watcher]()
	}
	s.watchers.Add(w)
	w.register(s)
	s.mu.Unlock()
	return seq, frame
}

func (s *screen) removeWatcher(w *watcher) {
	s.mu.Lock()
	if s.cinema.isKilled() {
		s.mu.Unlock()
		return
	}
	s.watchers.Remove(w)
	sheduleRemove := s.needRemove()
	s.mu.Unlock()
	if !sheduleRemove {
		return
	}
	s.scheduleRemove()
}

func (s *screen) addSub(sub *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.subs == nil {
		s.subs = common.NewSet[*screen]()
	}
	s.subs.Add(sub)
	sub.init(s, s.seq)
}

func (s *screen) removeSub(sub *screen) {
	s.mu.Lock()
	if s.cinema.isKilled() {
		s.mu.Unlock()
		return
	}
	s.subs.Remove(sub)
	sheduleRemove := s.needRemove()
	s.mu.Unlock()
	if !sheduleRemove {
		return
	}
	s.scheduleRemove()
}

func (s *screen) sync(init bool, ctx context.Context, cleanFrame shredder.SimpleFrame, sourceFrame shredder.SimpleFrame, seq uint, isStopped func() bool) {
	syncFrame := shredder.Join(true, sourceFrame, s.cinema.door.ReadFrame(), s.thread.Frame(), s.cinema.ReadFrame())
	defer syncFrame.Release()
	schedule := syncFrame.Run
	if init {
		schedule = syncFrame.Submit
	}
	schedule(s.cinema.ctx(), s.cinema.runtime, func(ok bool) {
		if !ok {
			return
		}
		var watchers []*watcher
		var subs []*screen
		syncThread := shredder.Thread{}
		writeFrame, readFrame := s.watcherSyncGuard.Write()
		readFrame = shredder.Join(true, readFrame)
		commitFrame := shredder.Join(true, syncThread.Frame(), syncFrame, writeFrame)
		watchersFrame := shredder.Join(true, syncFrame, syncThread.Frame(), readFrame)
		childerenFrame := shredder.Join(true, syncFrame, syncThread.Frame())

		commitFrame.Run(s.cinema.ctx(), s.cinema.runtime, func(b bool) {
			if !b {
				return
			}
			watchers, subs = s.commit(seq)
		})
		commitFrame.Release()

		watchersFrame.Run(s.cinema.ctx(), s.cinema.runtime, func(b bool) {
			if !b {
				return
			}
			for _, watcher := range watchers {
				watcherCtx := ctex.FrameInfect(ctx, s.cinema.ctx())
				watcherCtx = ctex.SyncFrameInsert(watcherCtx, readFrame)
				watcherFrame := shredder.Join(false, watchersFrame, watcher.syncFrame())
				watcherFrame.Submit(s.cinema.ctx(), s.cinema.runtime, func(ok bool) {
					if !ok {
						return
					}
					watcher.sync(watcherCtx, seq, cleanFrame)
				})
				watcherFrame.Release()
			}
		})
		watchersFrame.Release()

		childerenFrame.Run(s.cinema.ctx(), s.cinema.runtime, func(ok bool) {
			if !ok {
				return
			}
			if isStopped() {
				return
			}
			for _, screen := range subs {
				screen.sync(false, ctx, cleanFrame, sourceFrame, seq, isStopped)
			}

		})
		childerenFrame.Release()

	})
}

func (s *screen) scheduleRemove() {
	frame := s.cinema.writeFrame()
	defer frame.Release()
	frame.Run(s.cinema.ctx(), s.cinema.runtime, func(bool) {
		s.cinema.tryRemove(s.sourceID)
	})
}

func (s *screen) isEmpty() bool {
	return s.watchers.Len() == 0 && s.subs.Len() == 0
}

func (s *screen) commit(seq uint) ([]*watcher, []*screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq = seq
	return s.watchers.Slice(), s.subs.Slice()
}

func (s *screen) needRemove() bool {
	sheduleRemove := s.isEmpty() && !s.removeScheduled
	if sheduleRemove {
		s.removeScheduled = true
	}
	return sheduleRemove
}

func (s *screen) tryRemove() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isEmpty() {
		s.removeScheduled = false
		return false
	}
	s.parent.removeSub(s)
	return true
}

func (s *screen) cancel() {
	s.mu.Lock()
	w := s.watchers.Slice()
	s.watchers.Clear()
	s.subs.Clear()
	s.mu.Unlock()
	for _, w := range w {
		w.Cancel()
	}
	s.parent.removeSub(s)
}
