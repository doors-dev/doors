package sh

import (
	"sync"
	"sync/atomic"
)

type executable interface {
	Execute(callback func())
}

type Frame interface {
	Release()
	SimpleFrame
}

type SimpleFrame interface {
	Run(Spawner, func())
	AnyFrame
}

type AnyFrame interface {
	schedule(executable)
}


type Releaser interface {
	Release()
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
	active bool
}

func (m *ValveFrame) Activate() {
	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		panic("frame: already active")
	}
	m.active = true
	m.mu.Unlock()
	for i, e := range m.buffer {
		m.schedule(e)
		m.buffer[i] = nil
	}
}

func (m *ValveFrame) Run(s Spawner, fun func()) {
	m.mu.Lock()
	if !m.active {
		m.buffer = append(m.buffer, &simpleExecutable{
			fun:     fun,
			spawner: s,
		})
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()
	m.schedule(&simpleExecutable{
		spawner: s,
		fun:     fun,
	})

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

/*
func (s *Shread) Frame() Frame {

}

/*
type Executable interface {
	Execute(callback func())
}

type innerFrame struct {
	next    *innerFrame
	counter int
}

type Frame struct {
	mu    sync.Mutex
	inner innerFrame
}

func (f *Frame) Slide() {
	f.mu.Lock()
	defer f.mu.Unlock()
}

type Framer interface {
	submit(exe Executable)
}

type JoinedFrames struct {
	mu        sync.Mutex
	callbacks []func()
	counter   int
}

func (j *JoinedFrames) report() {

}

func (j *JoinedFrames) submit(exe Executable) {

}

/*
import (
	"sync"
	"sync/atomic"

	"github.com/gammazero/deque"
)
type frameState int

const (
	pending frameState = iota
	active
	done
)

type Frame struct {
	mu       sync.Mutex
	released bool
	counter  int
	parents  []*Frame
}

func (f *Frame)

func (f *Frame) done() bool {
	return f.released && f.counter == 0
}

func (f *Frame) report() {
	f.mu.Lock()
	f.counter -= 1
	done := f.done()
	f.mu.Unlock()
	if !done {
		return
	}
	f.reportToParents()
}

func (f *Frame) reportToParents() {
	for _, parent := range f.parents {
		parent.report()
	}
}

func (f *Frame) Release() {
	f.mu.Lock()
	if f.released {
		f.mu.Unlock()
		panic("frame: already released")
	}
	f.released = true
	done := f.done()
	f.mu.Unlock()
	if !done {
		return
	}
	f.reportToParents()
}

/*
import (
	"sync"

	"github.com/gammazero/deque"
)

type shreadState int

const (
	pending shreadState = iota
	active
	reported
)
jk
type Spawner interface {
	Spawn(func()) bool
}

type executable interface {
	Execute(onDone func())
}

type Shread struct {
	mu          sync.Mutex
	queue       deque.Deque[*Shread]
	buffer      []executable
	state       shreadState
	taskCounter int
	parent      *Shread
}

func (s *Shread) Activate() {
	if s.parent != nil {
		panic("Branched shreads cannot be manually activated")
	}
	s.activate()
}

type Release = func()

func (s *Shread) ReleasedBranch() (*Shread, bool) {
	b, r, ok := s.Branch()
	if ok {
		r()
	}
	return b, ok
}

func (s *Shread) Branch() (*Shread, Release, bool) {
	s.mu.Lock()
	if s.state == reported {
		s.mu.Unlock()
		return nil, nil, false
	}
	sh := &Shread{
		parent: s,
	}
	s.queue.PushBack(sh)
	activate := s.queue.Len() == 1
	s.mu.Unlock()
	if activate {
		sh.activate()
	}
	return sh, sync.OnceFunc(func() {
		sh.reportTask()
	}), true
}

func (s *Shread) reportTask() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.taskCounter -= 1
	s.tryReport()
}

func (s *Shread) tryReport() {
	if s.taskCounter != -1 || s.queue.Len() != 0 {
		return
	}
	s.state = reported
	s.parent.reportShread()
}

func (s *Shread) reportShread() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue.PopFront()
	s.tryReport()
	if s.queue.Len() == 0 {
		return
	}
	s.queue.At(0).activate()
}

func (s *Shread) add(task executable) bool {
	s.mu.Lock()
	switch s.state {
	case active:
		s.taskCounter += 1
		s.mu.Unlock()
		task.Execute(s.reportTask)
		return true
	case reported:
		s.mu.Unlock()
		return false
	case pending:
		s.buffer = append(s.buffer, task)
		s.mu.Unlock()
		return true
	default:
		panic("shread: invalid state")
	}
}

func (s *Shread) activate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != pending {
		panic("Shread is already active or done")
	}
	s.state = active
	for i, f := range s.buffer {
		s.taskCounter += 1
		f.Execute(s.reportTask)
		s.buffer[i] = nil
	}
	s.buffer = s.buffer[:0]
	if s.queue.Len() != 0 {
		s.queue.At(0).activate()
	}
}

func Run(spawner Spawner, f func(bool), shreads ...*Shread) {
	m := multitask{
		f:       f,
		spawner: spawner,
		counter: len(shreads),
	}
	for _, s := range shreads {
		if !s.add(&m) {
			m.cancel()
			return
		}
	}
}

type multitask struct {
	mu       sync.Mutex
	f        func(bool)
	counter  int
	onDone   []func()
	spawner  Spawner
	canceled bool
}

func (m *multitask) Execute(onDone func()) {
	m.mu.Lock()
	if m.canceled {
		m.mu.Unlock()
		onDone()
		return
	}
	m.counter -= 1
	m.onDone = append(m.onDone, onDone)
	m.mu.Unlock()
	if m.counter != 0 {
		return
	}
	m.spawn()
}

func (m *multitask) spawn() {
	if m.spawner == nil {
		m.run(true)
		return
	}
	ok := m.spawner.Spawn(func() {
		m.run(true)
	})
	if !ok {
		m.run(false)
	}
}

func (m *multitask) run(ok bool) {
	defer m.done()
	m.f(ok)
}

func (m *multitask) done() {
	for _, f := range m.onDone {
		f()
	}
}

func (m *multitask) cancel() {
	m.mu.Lock()
	m.canceled = true
	m.mu.Unlock()
	m.done()
	m.f(false)
	m.f = nil
	m.spawner = nil
}

/*

type Shread struct {
	mu      sync.Mutex
	parent  *Shread
	queue   deque.Deque[*Shread]
	buffer  []executable
	state   shreadState
	counter int
}



func (s *Shread) report() {
}

func (s *Shread) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != active {
		panic("Shread is not active")
	}
	s.state = done
	if s.counter == 0 && s.parent != nil {
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
}   */
