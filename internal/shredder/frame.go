package shredder

import (
	"context"
	"log/slog"
	"sync"
)

type executable interface {
	execute(callback func(error))
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
		slog.Warn(
			"attempted to schedule on completed frame; use ctx := doors.Free(ctx) for background operations and Doors API calls from goroutines",
		)
		e.execute(func(error) {})
		return
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

func (f *baseFrame) callback(error) {
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

func (s run) execute(callback func(error)) {
	s.runtime.Run(s.ctx, s.fun, callback)
}

type spawn struct {
	runtime Runtime
	fun     func(bool)
	ctx     context.Context
}

func (s spawn) execute(callback func(error)) {
	s.runtime.Submit(s.ctx, s.fun, callback)
}
