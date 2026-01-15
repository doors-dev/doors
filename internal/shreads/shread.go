package shreads

import (
	"sync"
	"sync/atomic"

	"github.com/gammazero/deque"
)

type shreadState int

const (
	pending shreadState = iota
	active
	done
)

type executable interface {
	Execute()
}

type Shread struct {
	mu      sync.Mutex
	parent  *Shread
	queue   deque.Deque[*Shread]
	buffer  []executable
	state   shreadState
	counter int
}

func (s *Shread) activate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = active
	for i, f := range s.buffer {
		f.Execute()
		s.buffer[i] = nil
	}
	s.buffer = s.buffer[:0]
	if s.counter != 0 {
		s.queue.PopFront().activate()
	}
}

func (s *Shread) report() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter -= 1
	if s.counter == 0 && s.state == done {
		s.parent.report()
		return
	}
	next := s.queue.PopFront()
	next.activate()
}

func (s *Shread) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == done {
		panic("shread: already closed")
	}
	s.state = done
	if s.counter == 0 {
		s.parent.report()
	}
}

func (s *Shread) add(task executable) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state {
	case active:
		task.Execute()
		return true
	case done:
		return false
	case pending:
		s.buffer = append(s.buffer, task)
		return true
	default:
		panic("shread: invalid state")
	}
}

func (s *Shread) Branch() (*Shread, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == done {
		return nil, false
	}
	sh := &Shread{
		parent: s,
	}
	s.counter += 1
	if s.counter == 1 && s.state == active {
		sh.activate()
	} else {
		s.queue.PushBack(sh)
	}
	return sh, true
}

type Spawner interface {
	Spawn(func())
}

func Run(spawner Spawner, f func(bool), shreads ...*Shread) {
	m := multitask{
		f: f,
		s: spawner,
	}
	m.c.Store(int32(len(shreads)))
	for _, s := range shreads {
		if !s.add(&m) {
			m.f = nil
			m.s = nil
			f(false)
			return
		}
	}
}

type multitask struct {
	c atomic.Int32
	f func(bool)
	s Spawner
}

func (m *multitask) Execute() {
	if m.c.Add(-1) != 0 {
		return
	}
	if m.s == nil {
		m.f(true)
		return
	}
	m.s.Spawn(func() {
		m.f(true)
	})
}

/*

type Shread = *shread


type shread struct {
	mu      sync.Mutex
	parent  *shread
	innie   deque.Deque[*shread]
	buffer  []func()
	state   shreadState
	spawner Spawner
	counter int
}

func (s *shread) activate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = active
	for i, f := range s.buffer {
		s.spawner.Spawn(f)
		s.buffer[i] = nil
	}
	s.buffer = s.buffer[:0]
}

func (s *shread) report() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter -= 1
	if s.counter == 0 && s.state == done {
		s.parent.report()
		return
	}
	next := s.innie.PopFront()
	next.activate()
}

func (s Shread) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == done {
		panic("shread: already closed")
	}
	s.state = done
	if s.counter == 0 {
		s.parent.report()
	}
}

func (s Shread) Run(f func(), spawn bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state {
	case active:
		if spawn {
			s.spawner.Spawn(f)
		} else {
			f()
		}
		return true
	case done:
		return false
	case pending:
		s.buffer = append(s.buffer, f)
		return true
	default:
		panic("shread: invalid state")
	}
}

func (s Shread) Branch() (Shread, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == done {
		return nil, false
	}
	sh := &shread{
		parent: s,
	}
	s.counter += 1
	if s.counter == 1 && s.state == active {
		sh.activate()
	} else {
		s.innie.PushBack(sh)
	}
	return sh, true
} */
