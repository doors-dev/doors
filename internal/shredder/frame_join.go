package shredder

import "sync"

type joinedFrame struct {
	mu        sync.Mutex
	callbacks []func(error)
	joinCount int
	baseFrame
}

func (f *joinedFrame) onComplete() {
	for _, callback := range f.callbacks {
		callback(nil)
	}
}

func (f *joinedFrame) register(callback func(error)) {
	f.mu.Lock()
	f.callbacks = append(f.callbacks, callback)
	ready := len(f.callbacks) == f.joinCount
	f.mu.Unlock()
	if !ready {
		return
	}
	f.activate()
}

func (j *joinedFrame) execute(callback func(error)) {
	j.register(callback)
}

func Join(release bool, frames ...AnyFrame) Frame {
	if len(frames) == 0 {
		panic("join must have frames")
	}
	joined := &joinedFrame{
		joinCount: len(frames),
	}
	joined.baseFrame.onComplete = joined.onComplete
	for _, frame := range frames {
		frame.schedule(joined)
		if release {
			if g, ok := frame.(Guard); ok {
				g.Release()
			}
		}
	}
	return joined
}
