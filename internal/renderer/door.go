package renderer

/*
import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type door struct {
	state atomic.Value
}

func (n *door) update(ctx context.Context, content any) {
	newState := &updateState{
		ctx:     ctx,
		content: content,
	}
	n.takover(newState)
}

func (n *door) takover(newState state) {
	value := n.state.Load()
	var prevState state
	if value != nil {
		prevState = value.(state)
	} else {
		u := updateState{
			ctx:     context.Background(),
			content: nil,
		}
		u.finishInit()
		u.finishTakeover()
		prevState = &u
	}
	prevState.afterInit(func(bool) {
		newState.takeover(prevState)
	})
}


func (s *proxyState) render() {

}




type tracker struct {
	id uint64
}


import (
	"context"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type nodeMode int

const (
	unmounted nodeMode = iota
	static
	render
	dynamic
)

type node struct {
	mode            nodeMode
	initializeGuard sh.Valve
	tracker         *tracker
	content         any
}

func (n *node) takeover(prev *node) {
	if prev == nil {
		n.dispatchTakeover(prev)
		return
	}
	prev.initializeGuard.Put(func(bool) {
		n.dispatchTakeover(prev)
	})
}

func (n *node) dispatchTakeover(prev *node) {
	defer n.initializeGuard.Open()
	switch n.mode {
	case static:
		n.staticTakeover(prev)
	case dynamic:
		n.dynamicTakeover(prev)
	case render:
	default:
		panic("unsupported node mode in takover")
	}
}

func (n *node) renderTakeover(prev *node) {
	if prev != nil && n.content == nil {
		switch true {
		case prev.mode == static:
			return
		case prev.content != nil && prev.mode == render:
			n.content = prev.content
		case prev.content != nil && (prev.mode == dynamic || prev.mode == unmounted):
			n.content = gox.Elem(func(ctx context.Context, cur gox.Cursor) (err error) {
				if err = cur.InitContainer(); err != nil {
					return
				}
				if err = cur.Any(prev.content); err != nil {
					return
				}
				err = cur.Close()
				return
			})
		}
	}
}

func (n *node) dynamicTakeover(prev *node) {
	if prev == nil || prev.mode == static || prev.mode == unmounted {
		n.mode = unmounted
		return
	}
	n.tracker = &tracker{
		id: prev.tracker.id,
	}
	prev.tracker.release()
	n.tracker.releaseGuard.Open()
	prev.tracker.takoverGuard.Put(func(bool) {
		n.tracker.takoverGuard.Open()
	})

}

func (n *node) staticTakeover(prev *node) {
	if prev == nil || prev.mode == static || prev.mode == unmounted {
		return
	}
	send := sh.Valve{}
	prev.tracker.takoverGuard.Put(func(bool) {
		send.Open()
	})
	id := prev.tracker.id
	prev.tracker.release()
	panic("todo")
}

type head struct {
	attrs *gox.Attrs
	tag   string
}

type tracker struct {
	id           uint64
	head         *head
	takoverGuard sh.Valve
	releaseGuard sh.Valve
}

func (t *tracker) release() {
}

/*

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type door struct {
	node atomic.Pointer[node]
}

func (d *door) update(content any) {
	new := &node{
		mode:    dynamic,
		content: content,
	}
	prev := d.node.Swap(new)
	new.submit(prev)
}

func (d *door) Job(ctx context.Context) gox.Job {
	new := &node{
		mode: render,
	}
	prev := d.node.Swap(new)
	return nil
}

type renderJob struct {
}

func (d *door) Proxy(ctx context.Context, cur gox.Cursor, elem gox.Elem) error {

}


var _ gox.Provider = &door{}
var _ gox.Proxy = &door{}

type nodeMode int

const (
	unmounted nodeMode = iota
	render
	dynamic
	static
)

type node struct {
	mode         nodeMode
	content      any
	tracker      *tracker
	trackerReady common.Valve
}

type tracker struct {
	id            uint64
	root          *Root
	ctx           context.Context
	cancel        context.CancelFunc
	suspendReady  common.Valve
	takeoverReady common.Valve
	sendReady     common.Valve
	thread        *sh.Thread
	children      common.Set[*node]
}

func (t *tracker) suspend() {
	t.suspendReady.Put(func() {
		if t.ctx.Err() != nil {
			return
		}
		t.cancel()
		t.thread.Kill(nil)
		for child := range t.children {
			child.suspend()
		}
	})
}

func (n *node) submit(prev *node) {
	switch n.mode {
	case render:
		n.submitRender(prev)
	case dynamic:
		n.submitDynamic(prev)
	case static:
		n.submitStatic(prev)
	}
}

func (n *node) submitRender(prev *node) {
	if prev != nil {
		n.content = prev.content
	}
	if prev.mode == static {
		n.mode = static
		// render static
		return
	}
	if prev.mode == dynamic || prev.mode == render {
		prev.trackerReady.Put(func() {
			prev.tracker.takeoverReady.Put(func() {
				prev.remove()
			})
		})
	}
}

func (n *node) render(r *Root, p *pipe, renderThread *sh.Thread) {
	n.tracker = &tracker{
		id:     r.newId(),
		thread: r.newThread(),
		root:   r,
	}
	n.trackerReady.Open()
	sh.Run(func(t *sh.Thread) {
		sh.Run(func(t *sh.Thread) {
			p.put(gox.NewJobFunc(nil, func(w io.Writer) error {
				n.tracker.takeoverReady.Open()
				n.tracker.suspendReady.Open()
				return nil
			}))
		}, t.Ws())
	}, renderThread.R(), n.tracker.thread.Wi())
}

func (n *node) submitStatic(prev *node) {
	if prev == nil || prev.mode == static || prev.mode == unmounted {
		return
	}
	prev.trackerReady.Put(func() {
		prev.tracker.suspend()
		thread := prev.tracker.thread.Spawner().NewThead()
		sendReady := common.Valve{}
		prev.tracker.takeoverReady.Put(func() {
			sendReady.Open()
		})
		sh.Run(func(t *sh.Thread) {
		}, thread.W())
	})
}

func (n *node) remove() {

}

func (n *node) submitDynamic(prev *node) {
	if prev == nil || prev.mode == static || prev.mode == unmounted {
		n.mode = unmounted
		return
	}
	prev.trackerReady.Put(func() {
		prev.tracker.suspend()
		n.tracker = &tracker{
			id:     prev.tracker.id,
			thread: prev.tracker.thread.Spawner().NewThead(),
			root:   prev.tracker.root,
		}
		n.trackerReady.Open()
		prev.tracker.takeoverReady.Put(func() {
			n.tracker.takeoverReady.Open()
			n.tracker.sendReady.Open()
		})
		n.tracker.suspendReady.Open()
		sh.Run(func(t *sh.Thread) {
			if t == nil {
				return
			}
			sh.Run(func(t *sh.Thread) {
				if t == nil {
					return
				}
				n.tracker.sendReady.Put(func() {
					// send
				})
			}, t.Ws())
		}, n.tracker.thread.W())
	})
}

func (n *node) suspend() {
}

/*
func (n *node) waitGuard() bool {
	if n == nil {
		return true
	}
	select {
	case <-n.guard:
		return true
	case <-n.ctx.Done():
		return false
	}
}

func (n *node) defuseGuard() {
	close(n.guard)
}

func (n *node) into(new *node) {
	if n == nil {
		return
	}
} */

/*
import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
)

type nodeState struct {
	content any
}

type node struct {
	mu      sync.Mutex
	static  bool
	soul    *soul
	content any
}

func (n *node) update(content any) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.static = false
	n.content = content
	if !n.soul.isValid() {
		return
	}
	n.soul.update(content)
}

func (n *node) remove() {
	n.replace(nil)
}

func (n *node) replace(content any) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.static = true
	n.content = content
	if !n.soul.isValid() {
		return
	}
	n.soul.replace(content)
}

func (n *node) render(r *Root) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.static {
		// render static one
		return
	}
	if n.soul.isValid() {
		n.soul.remove()
	}
	n.soul = &soul{
		id: r.newId(),
		th: r.newThread(),
	}
	n.soul.render()
}

type soulState int

const (
	pending soulState = iota
	ready
	suspended
)

type operation struct {
	replace bool
	content any
}

type soul struct {
	mu       sync.Mutex
	state    soulState
	buffered *operation
	id       uint64
	th       *sh.Thread
	parent   *soul
	children common.Set[*soul]
}

func (s *soul) replace(content any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isValid() {
		panic("Must not call replace on invalid soul")
	}
	o := operation{
		content: content,
		replace: true,
	}
	if !s.isReady() {
		s.buffered = &o
		return
	}
	s.apply(o)
}

func (s *soul) update(content any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isValid() {
		panic("Must not call update on invalid soul")
	}
	o := operation{
		content: content,
	}
	if !s.isReady() {
		s.buffered = &o
		return
	}
	s.apply(o)
}

func (s *soul) apply(o operation) {
	s.th.Kill(nil)
	s.th = s.th.Spawner().NewThead()
}

type readyJob struct {
	soul *soul
}

func (s *readyJob) Context() context.Context {
	return nil
}

func (s *readyJob) Output(io.Writer) error {
	return nil
}

func (s *soul) render(renderThread *sh.Thread, p *) {
	sh.Run(func(t *sh.Thread) {
		//
		sh.Run(func(t *sh.Thread) {
			s.mu.Lock()
			defer s.mu.Unlock()
			s.state = ready
			pending := s.buffered
			if pending == nil {
				return
			}
			s.apply(*pending)
		}, t.Ws())
	}, renderThread.R(), s.th.Wi())
}

func (s *soul) isValid() bool {
	return s == nil || s.state == suspended
}

func (s *soul) isReady() bool {
	return s.state == ready
}

func (s *soul) suspend() {

}

/*
import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type doorState int

const (
	dynamic doorState = iota
	rendering
	mounted
	static
)

type Door struct {
	mu        sync.Mutex
	state     doorState
	id        uint64
	unmounted bool
	th        *sh.Thread
}

func (d *Door) render(root *Root, parent *sh.Thread) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch d.state {
	case dynamic:
		d.remove()
	case rendering:

	}
	if d.state != static {

	}
	d.state = rendering
	d.th = root.newThread()
	sh.Run(func(t *sh.Thread) {
		sh.Run(func(t *sh.Thread) {
			d.mu.Lock()
			defer d.mu.Unlock()
		}, t.Ws())
	}, parent.R(), d.th.Wi())

}

func (d *Door) update(content any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.content = content
	d.static = false
	if d.node == nil {
		return
	}
	d.node.update(content)
}

func (d *Door) replace(content any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.content = content
	d.static = true
	if d.node == nil {
		return
	}
	d.node.replace(content)
}

func (d *Door) remove() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.node == nil {
		return
	}
	d.node.remove()
	d.node = nil
	d.static = true
}

type node struct {
	mu        sync.Mutex
	id        uint64
	unmounted bool
	th        *sh.Thread
}

func (d *node) exists() bool {
	if d == nil {
		return false
	}
	return true
}

func (d *node) render(root *Root, content any) {

}

func (c *node) remove() {
	c.unmounted = true
}

func (c *node) update(content any) {
}

func (c *node) replace(content any) {
}

func (n *node) unmount() {
	n.mu.Lock()
	defer n.mu.Unlock()
}

/*

type Door struct {
	mu      sync.Mutex
	n       atomic.Pointer[Node]
	node    *Node
	content any
}

var _ gox.JobProvider = &Door{}
var _ gox.ProxyProvider = &Door{}

func (d *Door) Update(content any) {
	node := &Node{
		state:   dynamic,
		content: content,
		lock:    make(chan struct{}),
	}
	old := d.n.Swap(node)
	node.transform(old)
}

func (d *Door) Replace(content any) {
	node := &Node{
		state:   static,
		content: content,
		lock:    make(chan struct{}),
	}
	old := d.n.Swap(node)
	node.transform(old)
}

func (d *Door) Job(ctx context.Context) gox.Job {
	node := &Node{
		state: toJob,
		lock:  make(chan struct{}),
	}
	old := d.n.Swap(node)
	node.transform(old)
	return node.job()
}

func (d *Door) Proxy(ctx context.Context, p gox.Printer) gox.Proxy {
	node := &Node{
		state: toProxy,
		lock:  make(chan struct{}),
	}
	old := d.n.Swap(node)
	node.transform(old)
	return node.proxy()
}

type job struct {
	ctx     context.Context
	node    *Node
	content any
}

func (j *job) Context() context.Context {
	return j.ctx
}

func (j *job) Output(w io.Writer) error {
	panic("unimplemented")
}

type nodeState int

const (
	dynamic nodeState = iota
	static
	toJob
	toProxy
)

type Node struct {
	state   nodeState
	lock    chan struct{}
	root    *Root
	id      uint64
	ctx     context.Context
	cancel  context.CancelFunc
	content any
	// children  common.Set[*Node]
}

func (n *Node) render(root *Root, thread *sh.Thread) *pipe {

}

func (n *Node) proxy() gox.Proxy {
	panic("unimplemented")
}

func (n *Node) job() gox.Job {
	panic("unimplemented")
}

func (n *Node) transform(old *Node) {
	if old != nil {
		<-old.lock
		if old.state == toJob || old.state == toProxy {
			panic("impossible, protected by lock")
		}
	}
	if n.state == toProxy {
		if old != nil {
			old.remove()
		}
		n.state = dynamic
		return
	}
	defer close(n.lock)
	if n.state == toJob {
		if old == nil {
			n.state = dynamic
			return
		}
		if old.state == static {
			n.content = old.content
			n.state = static
			return
		}
		old.remove()
		n.content = old.content
		n.state = dynamic
		return
	}
	if n.state == static {
		if old == nil || old.state == static {
			return
		}
		n.id = old.id
		n.root = old.root
		old.unmount()
		n.replace()
		return
	}
	if n.state == dynamic {
		if old == nil || old.state == static {
			return
		}
		n.id = old.id
		n.root = old.root
		old.unmount()
		n.update()
		return
	}
}

func (n *Node) update() {

}

func (n *Node) replace() {

}

func (n *Node) remove() {

}

func (n *Node) update(content any) {
}

func (n *Node) unmount() {
	if n.ctx.Err() != nil {
		return
	}
	n.cancel()
}

type paren interface {
	context() context.Context
}

func (n *Node) render(root *Root, parentThread *sh.Thread) *pipe {
	if n.unmounted {
		return nil
	}
	thread := root.newThread()

} */
