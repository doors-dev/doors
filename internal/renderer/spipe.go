package renderer
/*
import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

var pipePool = sync.Pool{
	New: func() any {
		return &spipe{}
	},
}

func getPipe(p parent, t *shredder.Thread) *spipe {
	q := pipePool.Get().(*spipe)
	q.parent = p
	return q
}

func putPipe(q *spipe) {
	q.parent = nil
	q.thread = nil
	q.signal = nil
	q.closed = false
	q.innie.Clear()
	
	pipePool.Put(q)
}

type parent interface {
	thread() *shredder.Thread
	addChild(t *Tracker)
}

type spipe struct {
	mu     sync.Mutex
	innie  deque.Deque[any]
	signal chan struct{}
	closed bool
	thread *shredder.Thread
}

func (p *spipe) door(t *Tracker) {
}

func (p *spipe) comp(ctx context.Context, comp gox.Comp) {
	qn := getPipe(p.thread)
	p.put(qn)
	shredder.Run(func(t *shredder.Thread) {
		defer qn.close()
		comp.Main().Print(ctx, qn)
	}, shredder.R(p.thread))
}

func (p *spipe) Send(j gox.Job) error {
	switch job := j.(type) {
	case *trackerJob:
		p.door((*Tracker)(job))
	case *gox.JobComp:
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		p.comp(ctx, comp)
	default:
		p.put(j)
	}
	return nil
}

func (p *spipe) put(a any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.innie.PushBack(a)
	if p.signal != nil {
		close(p.signal)
		p.signal = nil
	}
}

func (p *spipe) close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.signal != nil {
		close(p.signal)
		p.signal = nil
	}
	p.closed = true
}

func (p *spipe) get() (any, bool) {
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
} */
