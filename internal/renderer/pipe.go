package renderer

import (
	"sync"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/gammazero/deque"
)

type pipe struct {
	mu     sync.Mutex
	innie  deque.Deque[any]
	signal chan struct{}
	closed bool
	root   *Root
	thread *sh.Thread
}

func (p *pipe) put(a any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.innie.PushBack(a)
	if p.signal != nil {
		close(p.signal)
		p.signal = nil
	}
}

func (p *pipe) close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.signal != nil {
		close(p.signal)
		p.signal = nil
	}
	p.closed = true
}

func (p *pipe) get() (any, bool) {
	p.mu.Lock()
	if p.innie.Len() == 0 {
		if p.closed {
			p.mu.Unlock()
			return nil, false
		}
		ch := make(chan struct{}, 1)
		p.signal = ch
		p.mu.Unlock()
		return p.get()
	}
	defer p.mu.Unlock()
	return p.innie.PopFront(), true
}
