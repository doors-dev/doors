package door

import (
	"sync"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/doors/internal/shredder2"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

func newPipe() *pipe {
	return &pipe{}
}

type pipe struct {
	mu     sync.Mutex
	innie  deque.Deque[any]
	signal chan struct{}
	closed bool
	parent parent
	frame  sh.Frame
}


func (p *pipe) Send(job gox.Job) error {
	switch job := job.(type) {
	case *node:
		p.parent.addChild(job)
		job.render(p.parent, p)
	case *gox.JobComp:
		newPipe := newPipe()
		newPipe.parent = p.parent
		newPipe.thread = p.thread
		p.put(newPipe)
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		sh.Run(nil, func(ok bool) {
			comp.Main().Print(ctx, newPipe)
		}, p.shread)
		/*
			sh.Run(func(t *sh.Thread) {
				comp.Main().Print(ctx, newPipe)
			}, p.thread.R()) */
	default:
		p.put(job)
	}
	return nil
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
