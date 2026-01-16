package node

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

const (
	initGuard = iota + 1
	updateGuard
	renderGuard
	unmountGuard
)

type door struct {
	node atomic.Pointer[node]
}

func (d *door) Job(ctx context.Context) gox.Job {
	next := &node{
		ctx:  ctx,
		door: d,
		kind: jobNode,
	}
	d.takover(next)
	return next
}

func (d *door) Proxy(ctx context.Context, cur gox.Cursor, elem gox.Elem) error {
	next := &node{
		ctx:  ctx,
		door: d,
		kind: proxyNode,
		view: &view{
			elem: &elem,
		},
	}
	d.takover(next)
}

func (d *door) takover(next *node) {
	prev := d.node.Swap(next)
	if prev == nil {
		prev = &node{
			door: d,
			kind: unmountedNode,
		}
		prev.guard.Open(initGuard)
	}
	prev.guard.Run(initGuard, func() {
		next.takeover(prev)
	})
}

type nodeKind int

const (
	unmountedNode nodeKind = iota
	replacedNode
	updatedNode
	jobNode
	proxyNode
)

type node struct {
	ctx     context.Context
	door    *door
	kind    nodeKind
	guard   sh.Guard
	shread  sh.Shread
	tracker *tracker
	view    *view
}

func (n *node) takeover(prev *node) {
	switch n.kind {
	case unmountedNode:
		panic("door: unmounted node can't takeover")
	case jobNode:
		n.jobTakeover(prev)
	case proxyNode:
		n.proxyTakeover(prev)
	case updatedNode:
		n.updatedTakeover(prev)
	case replacedNode:
		n.replacedTakeover(prev)
	}
}

func (n *node) proxyTakeover(prev *node) {
	renderFrame := n.shread.Frame()
	panic("unimplemented")
}

func (n *node) updatedTakeover(prev *node) {
	defer n.guard.Open(initGuard, unmountGuard)
	switch prev.kind {
	case unmountedNode:
		n.kind = unmountedNode
	case replacedNode:
		n.kind = unmountedNode
	case updatedNode, proxyNode, jobNode:
		prev.guard.Run(nil, updateGuard, func() {
			n.guard.Open(updateGuard)
		})
		renderFrame := n.shread.Frame()
		sendFrame := n.shread.Frame()
		trackerFrame := n.shread.Frame()
		prev.unmount(false)
		n.view.attrs = prev.view.attrs
		n.view.tag = prev.view.tag
		n.tracker = newTrackerFrom(prev.tracker, trackerFrame)
		n.tracker.parent.addChild(n)
		pipe := newPipe()
		pipe.parent = n.tracker
		pipe.frame = renderFrame
		renderFrame.Run(nil, func() {
			defer renderFrame.Release()
			defer pipe.close()
			cur := gox.NewCursor(n.tracker.getContext(), pipe)
			cur.Any(n.view.content)
		})
		shredder2.Run(func(t *shredder2.Thread) {
			pipe.thread = t
			shredder2.Run(func(t *shredder2.Thread) {
				prev.guard.Run(updateGuard, func() {
					// deploy update ->
				})
			}, t.Ws())
		}, n.tracker.thread.W())
	}
}

func (n *node) replacedTakeover(prev *node) {
	defer n.guard.Open(initGuard)
	switch prev.kind {
	case replacedNode, unmountedNode:
		return
	default:
		prev.unmount(false)
		id := prev.tracker.id
		thread := prev.tracker.root.newThread()
		shredder2.Run(func(t *shredder2.Thread) {
			pipe := newPipe()
			pipe.thread = t
			pipe.parent = prev.tracker.parent
			cur := gox.NewCursor(prev.tracker.parent.getContext(), pipe)
			cur.Any(n.view.content)
		}, thread.W())
		shredder2.Run(func(t *shredder2.Thread) {
			prev.guard.Run(updateGuard, func() {
				// deploy replace ->
			})
		}, thread.W())
	}
}

func (n *node) jobTakeover(prev *node) {
	switch prev.kind {
	case replacedNode:
		n.view = prev.view
		n.kind = replacedNode
		n.guard.Open(initGuard, renderGuard)
	case proxyNode:
		n.view = prev.view
		n.kind = proxyNode
		prev.unmount(true)
		n.guard.Open(renderGuard)
	}
}

func (n *node) render(parent parent, parentPipe *pipe) error {
	pipe := newPipe()
	parentPipe.put(pipe)
	renderFrame := sh.Join(n.shread.Frame(), parentPipe.frame)
	trackerFrame := n.shread.Frame()
	renderFrame.Run(nil, func() {
		if n.kind == replacedNode {
			defer pipe.close()
			cur := gox.NewCursor(parent.getContext(), pipe)
			cur.Any(n.view.content)
			return
		}
		if n.kind != jobNode && n.kind != proxyNode {
			panic("wrong node to render")
		}
		n.tracker = newTracker(parent, trackerFrame)
	})
	return nil
}

func (c *node) unmount(remove bool) {
	c.guard.Run(unmountGuard, func() {
		c.door.node.CompareAndSwap(c, &node{
			door: c.door,
			kind: unmountedNode,
			view: c.view,
		})
	})
}

func (n *node) Context() context.Context {
	return n.ctx
}

func (n *node) Output(io.Writer) error {
	return errors.New("door: used outside render pipeline")
}

type parent interface {
	getContext() context.Context
	getRoot() *root
	addChild(child *node)
}

type root struct {
}

func (r *root) newId() uint64 {
	panic("unimpl")
}

func newTrackerFrom(old *tracker, frame sh.Frame) *tracker {
	panic("unimpl")
}

func newTracker(parent parent, frame sh.Frame) *tracker {
	panic("unimpl")
}

type tracker struct {
	id       uint64
	ctx      context.Context
	frame    sh.Frame
	cancel   context.CancelFunc
	root     *root
	parent   parent
	mu       sync.Mutex
	children common.Set[*node]
}

func (t *tracker) getContext() context.Context {
	return t.ctx
}

func (t *tracker) getRoot() *root {
	return t.root
}

var _ parent = &tracker{}

func (t *tracker) removeChild(child *node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.children.Remove(child)
}

func (t *tracker) addChild(child *node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.children.Add(child)
}

type view struct {
	tag     string
	attrs   gox.Attrs
	content any
	elem    *gox.Elem
}
