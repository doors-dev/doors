package door2

import (
	"context"
	"errors"
	"io"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type nodeKind int

const (
	unmountedNode nodeKind = iota
	replacedNode
	updatedNode
	editorNode
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
	case editorNode, proxyNode:
		n.shread = &sh.Shread{}
		renderGuard := n.shread.Guard()
		prev.takoverFrame.Run(nil, nil, func(bool) {
			defer renderGuard.Release()
			switch n.kind {
			case editorNode:
				n.jobTakover(prev)
			case proxyNode:
				n.proxyTakeover(prev)
			}
		})
	default:
		prev.takoverFrame.Run(nil, nil, func(bool) {
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
	case updatedNode, editorNode, proxyNode:
		n.view.content = prev.view.content
		prev.kill(remove)
	}
}

func (n *node) jobTakover(prev *node) {
	n.view = prev.view
	switch prev.kind {
	case updatedNode, editorNode:
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
	pipe.tracker = parent
	pipe.frame = sh.Join(parentFrame, renderFrame)
	defer pipe.frame.Release()
	printer := common.NewBufferPrinter()
	pipe.SendTo(printer)
	pipe.frame.Submit(parent.ctx, parent.root.runtime(), func(ok bool) {
		defer pipe.close()
		if !ok {
			return
		}
		cur := gox.NewCursor(parent.ctx, pipe)
		cur.Any(n.view.content)
	})
	finalFrame := shread.Frame()
	defer finalFrame.Release()
	updateFrame := sh.Join(finalFrame, prev.communicationFrame)
	defer updateFrame.Release()
	updateFrame.Run(parent.ctx, parent.root.runtime(), func(ok bool) {
		if !ok {
			return
		}
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
	trackerShread := &sh.Shread{}
	trackerRenderFrame := trackerShread.Frame()
	defer trackerRenderFrame.Release()
	n.communicationFrame = prev.communicationFrame
	n.tracker = newTrackerFrom(prev.tracker, trackerShread)
	n.tracker.parent.addChild(n)
	n.view.attrs = prev.view.attrs
	n.view.tag = prev.view.tag
	if n.view.content == nil {
		n.communicationFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(bool) {
			// push empty
		})
		return
	}
	renderShread := sh.Shread{}
	renderFrame := renderShread.Frame()
	defer renderFrame.Release()
	pipe := newPipe()
	pipe.tracker = n.tracker
	pipe.frame = sh.Join(trackerRenderFrame, renderFrame)
	defer pipe.frame.Release()
	printer := common.NewBufferPrinter()
	pipe.SendTo(printer)
	pipe.frame.Submit(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
		defer pipe.close()
		if !ok {
			return
		}
		cur := gox.NewCursor(n.tracker.ctx, pipe)
		if comp, ok := n.view.content.(gox.Comp); ok {
			comp.Main()(cur)
		} else {
			cur.Any(n.view.content)
		}
	})
	finalFrame := renderShread.Frame()
	defer finalFrame.Release()
	updateFrame := sh.Join(finalFrame, n.communicationFrame)
	defer updateFrame.Release()
	updateFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
		// push replace
	})
}

func (n *node) render(parentRenderer *pipe) {
	renderFrame := n.shread.Frame()
	defer renderFrame.Release()
	parent := parentRenderer.tracker
	pipe := parentRenderer.branch()
	pipe.frame = sh.Join(parentRenderer.frame, renderFrame)
	pipe.frame.Release()
	pipe.frame.Run(parent.ctx, parent.root.runtime(), func(ok bool) {
		defer pipe.close()
		;
		if n.kind == replacedNode {
			cur := gox.NewCursor(parent.ctx, pipe)
			cur.Any(n.view.content)
			return
		}
		n.communicationFrame = &sh.ValveFrame{}
		n.tracker = newTracker(parent, n.shread)
		parent.addChild(n)
		pipe.tracker = n.tracker
		cur := gox.NewCursor(n.tracker.ctx, pipe)
		switch n.kind {
		case editorNode:
			n.takoverFrame.Activate()
			open, close := n.view.headFrame(parent.ctx, n.tracker.id, cur.NewID())
			cur.Send(open)
			cur.Any(n.view.content)
			cur.Send(close)
		case proxyNode:
			proxy := newProxyRenderer(n.tracker.id, cur, n.view, parent.ctx)
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
		if !n.door.node.CompareAndSwap(n, nil) {
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
