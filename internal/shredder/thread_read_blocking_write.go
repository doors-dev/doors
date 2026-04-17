package shredder

import (
	"sync"
)

type ReadBlockingWriteThread struct {
	mu       sync.Mutex
	read     *baseFrame
	nextRead *baseFrame
	write    *baseFrame
}

func (f *ReadBlockingWriteThread) init() {
	if f.read != nil {
		return
	}
	f.read = &baseFrame{
		onComplete: f.complete,
	}
	f.read.activate()
}

func (f *ReadBlockingWriteThread) Read() Frame {
	f.mu.Lock()
	f.init()
	defer f.mu.Unlock()
	return Join(false, f.read)
}

func (f *ReadBlockingWriteThread) complete() {
	f.mu.Lock()
	f.read = f.nextRead
	f.nextRead = nil
	f.write.activate()
}

func (f *ReadBlockingWriteThread) Write() (write Frame, read Frame) {
	f.mu.Lock()
	f.init()
	if f.write != nil {
		f.mu.Unlock()
		panic("blocking frame contract violation: blocking frame is already issued")
	}
	f.nextRead = &baseFrame{
		onComplete: f.complete,
	}
	f.write = &baseFrame{
		onComplete: func() {
			f.write = nil
			f.mu.Unlock()
			f.read.activate()
		},
	}
	read = Join(false, f.nextRead)
	write = f.write
	f.mu.Unlock()
	f.read.Release()
	return
}
