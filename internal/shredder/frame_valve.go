package shredder

import (
	"context"
	"sync"
	"sync/atomic"
)

type ValveFrame struct {
	mu     sync.Mutex
	buffer []executable
	active atomic.Bool
}

func (f *ValveFrame) Activate() {
	f.mu.Lock()
	if f.active.Load() {
		f.mu.Unlock()
		return
	}
	f.active.Store(true)
	buf := f.buffer
	f.buffer = nil
	f.mu.Unlock()
	for i, e := range buf {
		f.schedule(e)
		buf[i] = nil
	}
}

func (f *ValveFrame) schedule(e executable) {
	if f.active.Load() {
		e.execute(func(error) {})
		return
	}
	f.mu.Lock()
	if f.active.Load() {
		f.mu.Unlock()
		e.execute(func(error) {})
		return
	}
	f.buffer = append(f.buffer, e)
	f.mu.Unlock()
}

func (f *ValveFrame) Run(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(run{runtime: s, ctx: ctx, fun: fun})
}

func (f *ValveFrame) Submit(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: s, ctx: ctx, fun: fun})
}

var _ SimpleFrame = &ValveFrame{}
