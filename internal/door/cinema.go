package door

import (
	"context"
	"log"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type Watcher interface {
	GetId() uint
	Sync(context.Context, uint, *common.FuncCollector)
	Cancel()
	Init(context.Context, *Screen, uint, uint) func()
}

type screenCinema interface {
	tryKill(uint64)
	isRoot() bool
	doorId() uint64
	newThread() *shredder.Thread
}

func newScreen(id uint64, cinema screenCinema, coreThread *shredder.Thread) *Screen {
	return &Screen{
		id:         id,
		mu:         sync.Mutex{},
		thread:     cinema.newThread(),
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


func (s *Screen) sync(syncThread *shredder.Thread, ctx context.Context, seq uint, c *common.FuncCollector) {
	s.thread.WriteInstant(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		var children []*Screen
		t.Write(func(t *shredder.Thread) {
			if t == nil {
				return
			}
			var watchers []Watcher
			watchers, children = s.commit(seq)
			for _, w := range watchers {
				t.Read(func(t *shredder.Thread) {
					w.Sync(ctx, seq, c)
				})
			}
		}, shredder.R(s.coreThread))
		t.Write(func(t *shredder.Thread) {
			if t == nil {
				return
			}
			for _, screen := range children {
				screen.sync(syncThread, ctx, seq, c)
			}
		})
	}, shredder.R(syncThread))
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

func newCinema(parent *Cinema, inst instance, coreThread *shredder.Thread, doorId uint64) *Cinema {
	return &Cinema{
		mu:         sync.Mutex{},
		coreThread: coreThread,
		inst:       inst,
		parent:     parent,
		screens:    make(map[uint64]*Screen),
		id:         doorId,
	}
}

type Cinema struct {
	mu         sync.Mutex
	killed     bool
	coreThread *shredder.Thread
	inst       instance
	parent     *Cinema
	screens    map[uint64]*Screen
	id         uint64
}

func (ss *Cinema) newThread() *shredder.Thread {
	return ss.inst.Thread()
}

func (ss *Cinema) doorId() uint64 {
	return ss.id
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
		screen = newScreen(id, ss, ss.coreThread)
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
	ss.coreThread.Read(func(t *shredder.Thread) {
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
	})
}

func (ss *Cinema) InitSync(syncThread *shredder.Thread, ctx context.Context, id uint64, seq uint, c *common.FuncCollector) {
	if !ss.isRoot() {
		log.Fatal("Only root door can init sync")
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	s, ok := ss.screens[id]
	if !ok {
		return
	}
	s.sync(syncThread, ctx, seq, c)
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
