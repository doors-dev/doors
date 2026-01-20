package door2

import (
	"context"
	"errors"
	"io"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type nodeKind int

const (
	unmountedNode nodeKind = iota
	replacedNode
	updatedNode
	jobNode
	proxyNode
)

type node struct {
	ctx                context.Context
	kind               nodeKind
	door               *Door
	takoverFrame       sh.ValveFrame
	shread             *sh.Shread
	communicationFrame *sh.ValveFrame
	tracker            *tracker
	view               *view
}

func (n *node) Context() context.Context {
	return n.ctx
}

func (n *node) Output(io.Writer) error {
	return errors.New("door: used outside render pipeline")
}

func (n *node) takover(prev *node) {
	switch n.kind {
	case jobNode, proxyNode:
		n.shread = &sh.Shread{}
		renderGuard := n.shread.Guard()
		prev.takoverFrame.Run(nil, func() {
			defer renderGuard.Release()
			switch n.kind {
			case jobNode:
				n.jobTakover(prev)
			case proxyNode:
				n.proxyTakeover(prev)
			}
		})
	default:
		prev.takoverFrame.Run(nil, func() {
			defer n.takoverFrame.Activate()
			switch n.kind {
			case unmountedNode:
				n.replaceTakeover(prev)
			case updatedNode:
				n.updatedTakeover(prev)
			case replacedNode:
				n.replaceTakeover(prev)
			}
		})
	}
}

func (n *node) proxyTakeover(prev *node) {
	switch prev.kind {
	case updatedNode, jobNode, proxyNode:
		n.view.content = prev.view.content
		prev.kill(remove)
	}
}

func (n *node) jobTakover(prev *node) {
	n.view = prev.view
	switch prev.kind {
	case updatedNode, jobNode:
		prev.kill(remove)
	case replacedNode:
		n.kind = replacedNode
		n.takoverFrame.Activate()
	case proxyNode:
		prev.kill(remove)
		n.kind = proxyNode
	}
}

func (n *node) replaceTakeover(prev *node) {
	switch prev.kind {
	case replacedNode, unmountedNode:
		return
	}
	prev.kill(unmount)
	if n.view.content == nil {
		prev.communicationFrame.Run(nil, func() {
			// push remove
		})
		return
	}
	shread := sh.Shread{}
	parent := prev.tracker.parent
	parentFrame := parent.newRenderFrame()
	defer parentFrame.Release()
	renderFrame := shread.Frame()
	defer renderFrame.Release()
	pipe := newPipe()
	pipe.parent = parent
	pipe.frame = sh.Join(parentFrame, renderFrame)
	defer pipe.frame.Release()
	pipe.frame.Run(parent.getRoot().Spawner(), func() {
		defer pipe.Close()
		cur := gox.NewCursor(parent.getContext(), pipe)
		cur.Any(n.view.content)
	})
	finalFrame := shread.Frame()
	defer finalFrame.Release()
	updateFrame := sh.Join(finalFrame, prev.communicationFrame)
	defer updateFrame.Release()
	updateFrame.Run(nil, func() {
		// push replace
	})
}

func (n *node) updatedTakeover(prev *node) {
	switch prev.kind {
	case unmountedNode:
		n.kind = unmountedNode
		return
	case replacedNode:
		n.kind = unmountedNode
		return
	}
	prev.kill(unmount)
	prev.tracker.parent.addChild(n)
	trackerShread := &sh.Shread{}
	trackerRenderFrame := trackerShread.Frame()
	defer trackerRenderFrame.Release()
	n.communicationFrame = prev.communicationFrame
	n.tracker = newTrackerFrom(prev.tracker, trackerShread)
	n.view.attrs = prev.view.attrs
	n.view.tag = prev.view.tag
	if n.view.content == nil {
		n.communicationFrame.Run(nil, func() {
			// push empty
		})
		return
	}

	renderShread := sh.Shread{}
	renderFrame := renderShread.Frame()
	defer renderFrame.Release()
	pipe := newPipe()
	pipe.parent = n.tracker
	pipe.frame = sh.Join(trackerRenderFrame, renderFrame)
	defer pipe.frame.Release()
	pipe.frame.Run(n.tracker.parent.getRoot().Spawner(), func() {
		defer pipe.Close()
		cur := gox.NewCursor(n.tracker.getContext(), pipe)
		cur.Any(n.view.content)
	})
	finalFrame := renderShread.Frame()
	defer finalFrame.Release()
	updateFrame := sh.Join(finalFrame, n.communicationFrame)
	defer updateFrame.Release()
	updateFrame.Run(nil, func() {
		// push replace
	})
}

func (n *node) render(parent parent, parentPipe *pipe) {
	renderFrame := n.shread.Frame()
	defer renderFrame.Release()
	pipe := newPipe()
	parentPipe.Put(pipe)
	pipe.frame = sh.Join(parentPipe.frame, renderFrame)
	pipe.frame.Release()
	pipe.frame.Run(parent.getRoot().Spawner(), func() {
		defer pipe.Close()
		if n.kind == replacedNode {
			pipe.parent = parent
			cur := gox.NewCursor(parent.getContext(), pipe)
			cur.Any(n.view.content)
			return
		}
		n.communicationFrame = &sh.ValveFrame{}
		n.tracker = newTracker(parent, n.shread)
		parent.addChild(n)
		pipe.parent = n.tracker
		cur := gox.NewCursor(n.tracker.getContext(), pipe)
		cur.Func(func(io.Writer) error {
			n.communicationFrame.Activate()
			return nil
		})
		switch n.kind {
		case jobNode:
			n.takoverFrame.Activate()
			open, close := n.view.headFrame(parent.getContext(), n.tracker.id, cur.NewId())
			cur.Job(open)
			cur.Any(n.view.content)
			cur.Job(close)
		case proxyNode:
			proxy := newProxyRenderer(n.tracker.id, cur, n.view, parent.getContext())
			proxy.render()
			proxy.InitFrame().Run(nil, func() {
				n.takoverFrame.Activate()
			})
		default:
			panic("door: wrong node kind to render")
		}
	})
}

type killKind int

const (
	cascade killKind = iota
	unmount
	remove
)

func (n *node) kill(kind killKind) {
	if n.kind == unmountedNode || n.kind == replacedNode {
		panic("door: unmounted/replaced node can't be killed")
	}
	switch kind {
	case cascade:
		if !n.door.node.CompareAndSwap(n, nil){
			return
		}
		n.tracker.kill()
	case unmount:
		n.tracker.parent.removeChild(n)
		n.tracker.kill()
	case remove:
		n.tracker.parent.removeChild(n)
		n.tracker.kill()
		n.communicationFrame.Run(nil, func() {
			// push update
		})
	}
}
