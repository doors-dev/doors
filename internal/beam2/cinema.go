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
	"github.com/doors-dev/doors/internal/sh"
)

type Door interface {
	NewFrame() sh.Frame
	Context() context.Context
}

type parentScreen interface {
	removeSub(*screen)
}

type screen struct {
	mu       sync.Mutex
	cinema   *cinema
	sourceId common.ID
	shread   sh.Shread
	subs     common.Set[*screen]
	watchers common.Set[*watcher]
	parent   parentScreen
	seq      uint
}

func (s *screen) sync(ctx context.Context, cleanFrame sh.SimpleFrame, sourceFrame sh.SimpleFrame, seq uint, isStopped func() bool) {

	doorFrame := s.cinema.door.NewFrame()
	defer doorFrame.Release()

	screenFrame := s.shread.Frame()
	defer screenFrame.Release()

	syncFrame := sh.Join(sourceFrame, doorFrame, screenFrame)
	defer syncFrame.Release()

	syncFrame.Run(s.cinema.ctx(), s.cinema.runtime, func(ok bool) {
		if !ok {
			return
		}

		syncShread := sh.Shread{}

		watchers, subs := s.commit(seq)

		watchersFrame := syncShread.Frame()
		defer watchersFrame.Release()

		syncWatchersFrame := sh.Join(syncFrame, watchersFrame)
		defer syncWatchersFrame.Release()

		childrenFrame := syncShread.Frame()
		defer childrenFrame.Release()

		syncChildrenFrame := sh.Join(syncFrame, childrenFrame)
		defer syncWatchersFrame.Release()

		for _, watcher := range watchers {
			syncWatchersFrame.Submit(s.cinema.ctx(), s.cinema.runtime, func(ok bool) {
				if !ok {
					return
				}
				watcher.sync(ctx, seq, cleanFrame)
			})
		}

		syncChildrenFrame.Run(s.cinema.ctx(), s.cinema.runtime, func(ok bool) {
			if !ok {
				return
			}
			if isStopped() {
				return
			}
			for _, screen := range subs {
				screen.sync(ctx, cleanFrame, sourceFrame, seq, isStopped)
			}
		})
	})

}

func (s *screen) commit(seq uint) ([]*watcher, []*screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq = seq
	return s.watchers.Slice(), s.subs.Slice()
}

func (s *screen) init(parent parentScreen, seq uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.parent = parent
	s.seq = seq
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

func (s *screen) tryRemove() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isEmpty() {
		return false
	}
	s.parent.removeSub(s)
	return true
}

func (s *screen) isEmpty() bool {
	return s.watchers.Len() == 0 && s.subs.Len() == 0
}

func (s *screen) removeWatcher(w *watcher) {
	if s.cinema.isKilled() {
		return
	}
	s.mu.Lock()
	if s.cinema.isKilled() {
		s.mu.Unlock()
		return
	}
	s.watchers.Remove(w)
	if s.isEmpty() {
		defer s.cinema.tryRemove(s.sourceId)
	}
	s.mu.Unlock()
}

func (s *screen) addWatcher(w *watcher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.watchers.Add(w)
	w.initSeq = s.seq
	w.screen = s
}

func (s *screen) addSub(sub *screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs.Add(sub)
	sub.init(s, s.seq)
}

func (s *screen) removeSub(sub *screen) {
	if s.cinema.isKilled() {
		return
	}
	s.mu.Lock()
	if s.cinema.isKilled() {
		s.mu.Unlock()
		return
	}
	s.subs.Remove(sub)
	if s.isEmpty() {
		defer s.cinema.tryRemove(s.sourceId)
	}
	s.mu.Unlock()
}

type Cinema = *cinema

func NewCinema(parent Cinema, door Door, runtime sh.Runtime) Cinema {
	return &cinema{
		parent:  parent,
		door:    door,
		runtime: runtime,
	}
}

type cinema struct {
	mu      sync.Mutex
	parent  *cinema
	door    Door
	runtime sh.Runtime
	screens map[common.ID]*screen
}

func (c *cinema) isKilled() bool {
	return c.door.Context().Err() != nil
}

func (c *cinema) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, screen := range c.screens {
		screen.cancel()
	}
	clear(c.screens)

}

func (c *cinema) ctx() context.Context {
	return c.door.Context()
}

func (c *cinema) tryRemove(sourceId common.ID) {
	frame := c.door.NewFrame()
	defer frame.Release()
	frame.Run(c.ctx(), c.runtime, func(ok bool) {
		if !ok {
			return
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		scr, ok := c.screens[sourceId]
		if !ok {
			return
		}
		if scr.tryRemove() {
			delete(c.screens, sourceId)
		}
	})
}

func (c *cinema) addWatcher(src anySource, w *watcher) bool {
	c.mu.Lock()
	s, ok := c.getScreen(src)
	if !ok {
		c.mu.Unlock()
		return true
	}
	s.addWatcher(w)
	c.mu.Unlock()
	w.init()
	return true
}

func (c *cinema) getScreen(src anySource) (*screen, bool) {
	if c.isKilled() {
		return nil, false
	}
	scr, ok := c.screens[src.getID()]
	if ok {
		return scr, true
	}
	scr = &screen{
		subs:     common.NewSet[*screen](),
		sourceId: src.getID(),
		cinema:   c,
	}
	c.screens[src.getID()] = scr
	if c.parent == nil {
		src.addSub(scr)
	} else {
		if !c.parent.addSub(src, scr) {
			return nil, false
		}
	}
	return scr, true
}

func (c *cinema) addSub(src anySource, scr *screen) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.getScreen(src)
	if !ok {
		return false
	}
	s.addSub(scr)
	return true
}
