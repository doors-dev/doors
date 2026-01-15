package renderer

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)
/*
type door struct {
	inner container
}

func (d *door) Job(ctx context.Context) gox.Job {
	return d.inner.job(ctx)
}

func (d *door) Proxy(ctx context.Context, p gox.Printer) gox.Proxy {
	return d.inner.proxy(ctx, p)
}

var _ gox.JobProvider = &door{}
var _ gox.ProxyProvider = &door{}

type container struct {
	mu      sync.Mutex
	tracker *Tracker
	content any
}

func (c *container) replace(content any) {
	c.mu.Lock()
	defer func() {
		c.tracker = nil
		c.mu.Unlock()
	}()
	c.content = content
	c.tracker.replace(content)
}

func (c *container) update(content any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tracker = c.tracker.inherit()
	c.tracker.update(content)
}

func (c *container) job(ctx context.Context) gox.Job {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tracker == nil {
		if c.content == nil {
			return nil
		}
		return gox.NewJobComp(ctx, gox.Elem(func(ctx context.Context, cursor gox.Cursor) error {
			return cursor.Any(ctx, c.content)
		}))
	}
	job := (*trackerJob)(c.tracker)
	return job
}

/*
func (c *container) proxy(ctx context.Context, p gox.Printer) gox.Proxy {

} */

/*
var trackerPool = sync.Pool{
	New: func() any {
		return &Tracker{
			children: common.NewSet[*container](),
		}
	},
}

type trackerJob Tracker

func (t *trackerJob) Context() context.Context {
	return t.Context()
}

func (t *trackerJob) Output(w io.Writer) error {
	return errors.New("Door is used outside framework's render pipeline")

}

type Tracker struct {
	id       uint64
	parent   parent
	ctx      context.Context
	cancel   context.CancelFunc
	th   *shredder.Thread
	mu       sync.Mutex
	children common.Set[*Tracker]
}

func (t *Tracker) render(parent parent) *pipe {
	t.th = thread.Spawner().NewThead()
	shredder.Run(func(t *shredder.Thread) {

	}, thread.R(), t.th.W())
}

func (t *Tracker) addChild(c *Tracker) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.children.Add(c)
}

func (t *Tracker) thread() *shredder.Thread {
	return t.th
}

func (t *Tracker) job() {

}

func (t *Tracker) replace(content any) {

}

func (x *Tracker) update(content any) {
	shredder.Run(func(t *shredder.Thread) {
		q := getPipe(t)
		q.Send(gox.NewJobComp(x.ctx, gox.Elem(func(ctx context.Context, cursor gox.Cursor) error {
			return cursor.Any(ctx, content)
		})))
		shredder.Run(func(t *shredder.Thread) {
		}, t.Ws())
	}, x.thread.W())
}

func (t *Tracker) inherit() (tn *Tracker) {
	tn = trackerPool.Get().(*Tracker)
	if t == nil {
		return
	}
	tn.id = t.id
	tn.pctx = tn.ctx
	tn.release()
	return
}

func (t *Tracker) release() {
	t.id = 0
	t.pctx = nil
	t.ctx = nil
	t.thread = nil
	t.children.Clear()
	t.content = nil
	trackerPool.Put(t)
}

func (t *Tracker) parentContext() context.Context {
	if t == nil {
		return nil
	}
	return t.ctx
}

func (t *Tracker) Context() context.Context {
	if t == nil {
		return nil
	}
	return t.ctx
}

func (t *Tracker) Id() uint64 {
	if t == nil {
		return 0
	}
	return t.id
}

func (t *Tracker) unmount() {
	if t == nil {
		return
	}
}

type container struct {
	mu      sync.Mutex
	id      uint64
	tracker *tracker
	content gox.Comp
}

func (c *container) unmount() {
	c.lock()
	defer c.unlock()
	if !c.isMounted() {
		return
	}
	c.tracker.kill()
	c.tracker = nil
}

func (c *container) lock() {
	c.mu.Lock()
}

func (c *container) unlock() {
	c.mu.Unlock()
}

func (c *container) isMounted() bool {
	return c.tracker != nil
}

func (c *container) update(content gox.Comp) {
}

type tracker struct {
	thread   *shredder.Thread
	children common.Set[*container]
}

func (t *tracker) kill() {
	for child := range t.children.Iter() {
		child.unmount()
	}
}

/*
func (d *door) Job(ctx context.Context) gox.Job {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.container.unmount()
	return d.container
}

var _ gox.Provider = &door{}

func (d *door) Update(content gox.Comp) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.content = content
	if d.container == nil {
		return
	}

} */
