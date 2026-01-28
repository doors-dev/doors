package sh

import (
	"context"
	"sync"
	"sync/atomic"
)

type executable interface {
	execute(callback func())
}

type Guard interface {
	Release()
}

type Frame interface {
	Guard
	SimpleFrame
}

type SimpleFrame interface {
	Run(ctx context.Context, r Runtime, fun func(bool))
	Submit(ctx context.Context, r Runtime, fun func(bool))
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

func (f *baseFrame) Run(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(&run{
		runtime: s,
		fun:     fun,
		ctx:     ctx,
	})
}

func (f *baseFrame) Submit(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(&spawn{
		runtime: s,
		fun:     fun,
		ctx:     ctx,
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
	e.execute(f.callback)
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
		e.execute(f.callback)
		f.buffer[i] = nil
	}
	f.buffer = f.buffer[:0]
}

func (f *baseFrame) callback() {
	f.mu.Lock()
	f.counter -= 1
	completed := f.isCompleted()
	f.mu.Unlock()
	if !completed {
		return
	}
	f.onComplete()
}

func (f *baseFrame) isCompleted() bool {
	return f.released && f.counter == 0 && f.active
}

type run struct {
	runtime Runtime
	fun     func(bool)
	ctx     context.Context
}

func (s run) execute(callback func()) {
	s.runtime.Run(s.ctx, s.fun, callback)
}

type spawn struct {
	runtime Runtime
	fun     func(bool)
	ctx     context.Context
}

func (s spawn) execute(callback func()) {
	s.runtime.Submit(s.ctx, s.fun, callback)
}

type shreadFrame struct {
	mu        sync.Mutex
	next      *shreadFrame
	completed bool
	baseFrame
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
	baseFrame
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

func (j *joinedFrame) execute(callback func()) {
	j.register(callback)
}

type Shread struct {
	frame atomic.Pointer[shreadFrame]
}

func (s *Shread) Guard() Guard {
	return s.Frame()
}

func (s *Shread) Frame() Frame {
	var frame *shreadFrame
	frame = &shreadFrame{
		baseFrame: baseFrame{
			onComplete: frame.onComplete,
		},
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
	var joined *joinedFrame
	joined = &joinedFrame{
		joinCount: len(others) + 1,
		baseFrame: baseFrame{
			onComplete: joined.onComplete,
		},
	}
	first.schedule(joined)
	for _, frame := range others {
		frame.schedule(joined)
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

func (m *ValveFrame) schedule(e executable) {
	if m.active.Load() {
		e.execute(func() {})
		return
	}
	m.mu.Lock()
	if m.active.Load() {
		m.mu.Unlock()
		e.execute(func() {})
		return
	}
	m.buffer = append(m.buffer, e)
	m.mu.Unlock()
}

func (m *ValveFrame) Run(ctx context.Context, s Runtime, fun func(bool)) {
	m.schedule(run{runtime: s, ctx: ctx, fun: fun})
}

func (m *ValveFrame) Submit(ctx context.Context, s Runtime, fun func(bool)) {
	m.schedule(spawn{runtime: s, ctx: ctx, fun: fun})
}

var _ SimpleFrame = &ValveFrame{}
