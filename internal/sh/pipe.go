package sh

import (
	"sync"

	"github.com/gammazero/deque"
)

var queuePool = sync.Pool{
	New: func() any {
		return &deque.Deque[any]{}
	},
}

type Queue = *queue

func NewQueue() Queue {
	q := queuePool.Get().(*deque.Deque[any])
	return &queue{
		innie: q,
	}
}

type queue struct {
	mu     sync.Mutex
	innie  *deque.Deque[any]
	signal chan struct{}
	closed bool
}

func (p Queue) Put(a any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		panic("pipe closed")
	}
	p.innie.PushBack(a)
	if p.signal != nil {
		close(p.signal)
		p.signal = nil
	}
}

func (p Queue) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		panic("pipe closed")
	}
	if p.signal != nil {
		close(p.signal)
		p.signal = nil
	}
	p.closed = true
}

func (p Queue) Get() (any, bool) {
	p.mu.Lock()
	if p.innie == nil {
		return nil, false
	}
	if p.innie.Len() == 0 {
		if p.closed {
			queuePool.Put(p.innie)
			p.innie = nil
			p.mu.Unlock()
			return nil, false
		}
		ch := make(chan struct{}, 1)
		p.signal = ch
		p.mu.Unlock()
		return p.Get()
	}
	defer p.mu.Unlock()
	return p.innie.PopFront(), true
}
