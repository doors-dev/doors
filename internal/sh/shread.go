package sh

import (
	"sync"
	"sync/atomic"
)

type executable interface {
	Execute(callback func())
}

type Guard interface {
	Release()
}

type Frame interface {
	Guard
	SimpleFrame
}

type SimpleFrame interface {
	Run(Spawner, func())
	AnyFrame
}

type AnyFrame interface {
	schedule(executable)
}


type baseFrame struct {
	mu         sync.Mutex
	active     bool
	released   bool
	counter    int
	onComplete func()
	buffer     []executable
}

func (f *baseFrame) Run(spawner Spawner, fun func()) {
	f.schedule(&simpleExecutable{
		spawner: spawner,
		fun:     fun,
	})
}

func (f *baseFrame) Release() {
	f.mu.Lock()
	if f.released {
		f.mu.Unlock()
		return
	}
	f.released = true
	completed := f.isCompleted()
	f.mu.Unlock()
	if !completed {
		return
	}
	f.onComplete()
}

func (f *baseFrame) schedule(e executable) {
	f.mu.Lock()
	if f.isCompleted() {
		f.mu.Unlock()
		panic("can's schedule on completed frames")
	}
	f.counter += 1
	if !f.active {
		f.buffer = append(f.buffer, e)
		f.mu.Unlock()
		return
	}
	f.mu.Unlock()
	f.execute(e)
}

func (f *baseFrame) activate() {
	f.mu.Lock()
	if f.active {
		f.mu.Unlock()
		panic("can't activate an already active frame")
	}
	if f.isCompleted() {
		f.mu.Unlock()
		panic("can't activate a completed frame")
	}
	f.active = true
	done := f.isCompleted()
	f.mu.Unlock()
	if done {
		f.onComplete()
		return
	}
	for i, e := range f.buffer {
		f.execute(e)
		f.buffer[i] = nil
	}
	f.buffer = f.buffer[:0]
}

func (f *baseFrame) execute(e executable) {
	e.Execute(func() {
		f.mu.Lock()
		f.counter -= 1
		completed := f.isCompleted()
		f.mu.Unlock()
		if !completed {
			return
		}
		f.onComplete()
	})
}

func (f *baseFrame) isCompleted() bool {
	return f.released && f.counter == 0 && f.active
}

type simpleExecutable struct {
	spawner Spawner
	fun     func()
}

func (s *simpleExecutable) Execute(callback func()) {
	if s.spawner == nil {
		defer callback()
		s.fun()
		return
	}
	s.spawner.Spawn(func() {
		defer callback()
		s.fun()
	})
}

type shreadFrame struct {
	mu        sync.Mutex
	next      *shreadFrame
	completed bool
	*baseFrame
}

func (f *shreadFrame) setNext(next *shreadFrame) {
	f.mu.Lock()
	completed := f.completed
	f.next = next
	f.mu.Unlock()
	if !completed {
		return
	}
	next.activate()
}

func (f *shreadFrame) onComplete() {
	f.mu.Lock()
	f.completed = true
	next := f.next
	f.mu.Unlock()
	if next == nil {
		return
	}
	next.activate()
}

type joinedFrame struct {
	mu        sync.Mutex
	callbacks []func()
	joinCount int
	*baseFrame
}

func (f *joinedFrame) onComplete() {
	for _, callback := range f.callbacks {
		callback()
	}
}

func (f *joinedFrame) register(callback func()) {
	f.mu.Lock()
	f.callbacks = append(f.callbacks, callback)
	ready := len(f.callbacks) == f.joinCount
	f.mu.Unlock()
	if !ready {
		return
	}
	f.activate()
}

type joinedExecution joinedFrame

func (j *joinedExecution) Execute(callback func()) {
	(*joinedFrame)(j).register(callback)
}

type Shread struct {
	frame atomic.Pointer[shreadFrame]
}

func (s *Shread) Guard() Guard {
	return s.Frame()
}

func (s *Shread) Frame() Frame {
	frame := &shreadFrame{}
	frame.baseFrame = &baseFrame{
		onComplete: frame.onComplete,
	}
	prev := s.frame.Swap(frame)
	if prev == nil {
		frame.activate()
	} else {
		prev.setNext(frame)
	}
	return frame
}

func Join(first AnyFrame, others ...AnyFrame) Frame {
	joined := &joinedFrame{
		joinCount: len(others) + 1,
	}
	joined.baseFrame = &baseFrame{
		onComplete: joined.onComplete,
	}
	first.schedule((*joinedExecution)(joined))
	for _, frame := range others {
		frame.schedule((*joinedExecution)(joined))
	}
	return joined
}

type ValveFrame struct {
	mu     sync.Mutex
	buffer []executable
	active atomic.Bool
}

func (m *ValveFrame) Activate() {
	m.mu.Lock()
	if m.active.Load() {
		m.mu.Unlock()
		panic("frame: already active")
	}
	m.active.Store(true)
	buf := m.buffer
	m.buffer = nil
	m.mu.Unlock()
	for i, e := range buf {
		m.schedule(e)
		buf[i] = nil
	}
}

func (m *ValveFrame) Run(s Spawner, fun func()) {
	if m.active.Load() {
		m.schedule(&simpleExecutable{spawner: s, fun: fun})
		return
	}
	m.mu.Lock()
	if !m.active.Load() {
		m.buffer = append(m.buffer, &simpleExecutable{spawner: s, fun: fun})
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()
	m.schedule(&simpleExecutable{spawner: s, fun: fun})
}

func (m *ValveFrame) schedule(e executable) {
	e.Execute(func() {})
}

var _ SimpleFrame = &ValveFrame{}

type FreeFrame struct{}

func (f FreeFrame) Run(s Spawner, fun func()) {
	f.schedule(&simpleExecutable{
		spawner: s,
		fun:     fun,
	})
}

func (f FreeFrame) schedule(e executable) {
	e.Execute(func() {})
}

var _ SimpleFrame = FreeFrame{}
