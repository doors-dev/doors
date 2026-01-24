package door2

import (
	"sync"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

var bufferPool = sync.Pool{
	New: func() any {
		return &deque.Deque[any]{}
	},
}

func newPipe() *pipe {
	return &pipe{
		buffer:  bufferPool.Get().(*deque.Deque[any]),
	}
}

type Pipe interface {
	SendTo(gox.Printer)
}

type pipe struct {
	mu      sync.Mutex
	closed  bool
	parent  *pipe
	buffer  *deque.Deque[any]
	tracker *tracker
	frame   sh.Frame
	printer gox.Printer
}

func (r *pipe) SendTo(printer gox.Printer) {
	if r.parent != nil {
		panic("Can't initiate printing with owned renderer")
	}
	r.print(printer)
}

func (r *pipe) Send(job gox.Job) error {
	switch job := job.(type) {
	case *node:
		job.render(r)
	case *gox.JobComp:
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		newRenderer := r.branch()
		newRenderer.frame.Run(r.tracker.root.spawner, func() {
			defer newRenderer.close()
			comp.Main().Print(ctx, newRenderer)
		})
	default:
		r.job(job)
	}
	return nil
}


func (r *pipe) print(printer gox.Printer) {
	stack := []*pipe{r}
	closed := false
main:
	for len(stack) != 0 {
		rr := stack[len(stack)-1]
		rr.mu.Lock()
		for rr.buffer.Len() != 0 {
			next := rr.buffer.PopFront()
			switch next := next.(type) {
			case gox.Job:
				printer.Send(next)
			case *pipe:
				rr.mu.Unlock()
				stack = append(stack, next)
				continue main
			}
		}
		closed = rr.closed
		if closed {
			bufferPool.Put(rr.buffer)
			rr.buffer = nil
			rr.mu.Unlock()
			stack[len(stack)-1] = nil
			stack = stack[:len(stack)-1]
			continue
		}
		rr.printer = printer
		rr.mu.Unlock()
		return
	}
	if !closed || r.parent == nil {
		return
	}
	r.parent.print(printer)
}

func (r *pipe) close() {
	r.mu.Lock()
	if r.closed {
		panic("renderer is already closed")
	}
	r.closed = true
	done := r.printer != nil && r.buffer.Len() == 0
	r.mu.Unlock()
	if !done || r.parent == nil {
		return
	}
	r.parent.print(r.printer)
}

func (r *pipe) job(job gox.Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		panic("render is closed")
	}
	if r.printer != nil {
		r.printer.Send(job)
		return
	}
	r.buffer.PushBack(job)
}

func (r *pipe) branch() *pipe {
	r.mu.Lock()
	if r.closed {
		panic("render is closed")
	}
	newRenderer := newPipe()
	newRenderer.tracker = r.tracker
	newRenderer.frame = r.frame
	newRenderer.parent = r
	if r.printer != nil {
		printer := r.printer
		r.printer = nil
		r.mu.Unlock()
		newRenderer.print(printer)
		return newRenderer
	}
	r.buffer.PushBack(newRenderer)
	r.mu.Unlock()
	return newRenderer
}
