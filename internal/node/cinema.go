package node

import (
	"context"
	"log"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type Watcher interface {
	GetId() uint
	Sync(context.Context, uint, *shredder.Collector[func()])
	Cancel()
	Init(context.Context, *Screen, uint, uint) func()
}

type screenCinema interface {
	tryKill(uint64)
	isRoot() bool
}

func newScreen(id uint64, cinema screenCinema, thread *shredder.Thread, coreThread *shredder.Thread) *Screen {
	return &Screen{
		id:         id,
		mu:         sync.Mutex{},
		thread:     thread,
		coreThread: coreThread,
		watchers:   make(map[uint]Watcher),
		children:   common.NewSet[*Screen](),
		cinema:     cinema,
	}
}

type Screen struct {
	id         uint64
	counter    uint
	mu         sync.Mutex
	thread     *shredder.Thread
	coreThread *shredder.Thread
	watchers   map[uint]Watcher
	parent     *Screen
	children   common.Set[*Screen]
	seq        uint
	delMark    bool
	cinema     screenCinema
}

func (s *Screen) kill(init bool) {
	s.thread.Kill(func() {
		if init && !s.cinema.isRoot() {
			s.parent.removeChild(s)
		}
		s.mu.Lock()
        wathchers := s.watchersSlice()
		s.mu.Unlock()
		for _, watcher := range wathchers {
			watcher.Cancel()
		}
	})
}

func (s *Screen) tryKill() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isEmpty() {
		return false
	}
	return s.thread.Kill(func() {
		if !s.cinema.isRoot() {
			s.parent.removeChild(s)
		}
	})
}

func (s *Screen) sync(ctx context.Context, seq uint, sourceCollector *shredder.Collector[func()]) {
	sourceCollector.Read(func(c *shredder.Collector[func()]) {
		if c == nil {
			return
		}
		var children []*Screen
		c.Write(func(c *shredder.Collector[func()]) {
			if c == nil {
				return
			}
			var watchers []Watcher
			watchers, children = s.commit(seq)
			for _, w := range watchers {
				c.Read(func(c *shredder.Collector[func()]) {
					w.Sync(ctx, seq, c)
				})
			}
		}, shredder.W(s.thread), shredder.R(s.coreThread))
		c.Write(func(c *shredder.Collector[func()]) {
			if c == nil {
				return
			}
			for _, screen := range children {
				screen.sync(ctx, seq, sourceCollector)
			}
		})
	})
}

func (s *Screen) watchersSlice() []Watcher {
	slice := make([]Watcher, len(s.watchers))
	i := 0
	for id := range s.watchers {
		slice[i] = s.watchers[id]
		i += 1
	}
    return slice
}

func (s *Screen) commit(seq uint) ([]Watcher, []*Screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq = seq
	return s.watchersSlice(), s.children.Slice()
}

func (s *Screen) addWatcher(ctx context.Context, w Watcher) func() {
	s.mu.Lock()
	s.counter += 1
	s.watchers[s.counter] = w
	init := w.Init(ctx, s, s.counter, s.seq)
	s.mu.Unlock()
	return init
}

func (s *Screen) addChild(screen *Screen) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.children.Add(screen)
	screen.init(s, s.seq)
}

func (s *Screen) init(parent *Screen, seq uint) {
	s.parent = parent
	s.seq = seq
}

func (s *Screen) UnregWatcher(id uint) {
	s.mu.Lock()
	if s.thread.Killed() {
		s.mu.Unlock()
		return
	}
	delete(s.watchers, id)
	empty := s.isEmpty()
	s.mu.Unlock()
	if empty {
		s.cinema.tryKill(s.id)
	}
}

func (s *Screen) removeChild(screen *Screen) {
	s.mu.Lock()
	if s.thread.Killed() {
		s.mu.Unlock()
		return
	}
	s.children.Remove(screen)
	empty := s.isEmpty()
	s.mu.Unlock()
	if empty {
		s.cinema.tryKill(s.id)
	}
}

func (s *Screen) isEmpty() bool {
	return len(s.watchers) == 0 && s.children.IsEmpty()
}

func newCinema(parent *Cinema, inst instance, coreThread *shredder.Thread) *Cinema {
	return &Cinema{
		mu:         sync.Mutex{},
		coreThread: coreThread,
		inst:       inst,
		parent:     parent,
		screens:    make(map[uint64]*Screen),
	}
}

type Cinema struct {
	mu         sync.Mutex
	killed     bool
	coreThread *shredder.Thread
	inst       instance
	parent     *Cinema
	screens    map[uint64]*Screen
}

func (ss *Cinema) AddWatcher(ctx context.Context, screenId uint64, w Watcher, lastSeq uint) bool {
	ss.mu.Lock()
	s, ok := ss.ensure(screenId, lastSeq)
	if !ok {
		ss.mu.Unlock()
		return false
	}
	init := s.addWatcher(ctx, w)
	defer init()
	ss.mu.Unlock()
	return true
}

func (ss *Cinema) addChild(id uint64, screen *Screen, lastSeq uint) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	s, ok := ss.ensure(id, lastSeq)
	if !ok {
		return false
	}
	s.addChild(screen)
	return true
}

func (ss *Cinema) ensure(id uint64, lastSeq uint) (*Screen, bool) {
	if ss.killed {
		return nil, false
	}
	screen, ok := ss.screens[id]
	if !ok {
		screen = newScreen(id, ss, ss.inst.Thread(), ss.coreThread)
		ss.screens[id] = screen
		if !ss.isRoot() {
			ok := ss.parent.addChild(id, screen, lastSeq)
			if !ok {
				return nil, false
			}
		} else {
			screen.init(nil, lastSeq)
		}
	}
	return screen, true
}

func (ss *Cinema) tryKill(id uint64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.killed {
		return
	}
	screen, has := ss.screens[id]
	if !has {
		return
	}
	killed := screen.tryKill()
	if killed {
		delete(ss.screens, id)
	}
}

func (ss *Cinema) InitSync(ctx context.Context, id uint64, seq uint, c *shredder.Collector[func()]) {
	if !ss.isRoot() {
		log.Fatal("Only root node can init sync")
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	s, ok := ss.screens[id]
	if !ok {
		return
	}
	s.sync(ctx, seq, c)
}

func (ss *Cinema) kill(init bool) {
	ss.mu.Lock()
	if ss.killed {
		ss.mu.Unlock()
		return
	}
	ss.killed = true
	ss.mu.Unlock()
	for id := range ss.screens {
		ss.screens[id].kill(init)
	}
}

func (ss *Cinema) isRoot() bool {
	return ss.parent == nil
}
