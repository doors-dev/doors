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
	reportCh           chan error
	done               func()
	kind               nodeKind
	door               *Door
	takoverFrame       sh.ValveFrame
	renderShread       sh.Shread
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
		renderGuard := n.renderShread.Guard()
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
			case updatedNode:
				n.updatedTakeover(prev)
			case unmountedNode:
				n.replaceTakeover(prev)
			case replacedNode:
				n.replaceTakeover(prev)
			}
		})
	}
}

func (n *node) updatedTakeover(prev *node) {
	switch prev.kind {
	case unmountedNode:
		n.kind = unmountedNode
		n.accept()
		return
	case replacedNode:
		n.kind = unmountedNode
		n.accept()
		return
	}
	prev.kill(unmount)
	n.communicationFrame = prev.communicationFrame
	n.tracker = newTrackerFrom(prev.tracker)
	n.tracker.parent.addChild(n)
	if n.view == nil {
		n.view = prev.view
	} else {
		n.view.attrs = prev.view.attrs
		n.view.tag = prev.view.tag
	}
	if n.view.content == nil {
		n.communicationFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
			defer n.done()
			if !ok {
				n.report(context.Canceled)
				return
			}
			// push empty
		})
		return
	}
	rootFrame := sh.ValveFrame{}
	printer := common.NewBufferPrinter()
	pipe := newPipe(&rootFrame)
	pipe.printer = printer
	pipe.tracker = n.tracker
	pipe.renderFrame = sh.Join(n.renderShread.Frame(), n.tracker.newRenderFrame())
	pipe.submit(func(ok bool) {
		defer pipe.close()
		if !ok {
			return
		}
		cur := gox.NewCursor(n.tracker.ctx, pipe)
		n.view.renderContent(cur)
	})
	updateFrame := sh.Join(n.renderShread.Frame(), n.communicationFrame)
	updateFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
		defer rootFrame.Activate()
		defer n.done()
		if !ok {
			n.report(context.Canceled)
			return
		}
	})
	updateFrame.Release()
}

func (n *node) replaceTakeover(prev *node) {
	switch prev.kind {
	case replacedNode, unmountedNode:
		n.accept()
		return
	}
	prev.kill(unmount)
	parent := prev.tracker.parent
	if n.view.content == nil {
		prev.communicationFrame.Run(parent.ctx, parent.root.runtime(), func(ok bool) {
			defer n.done()
			if !ok {
				n.report(context.Canceled)
				return
			}
			// push replace
		})
		return
	}
	rootFrame := sh.ValveFrame{}
	printer := common.NewBufferPrinter()
	pipe := newPipe(&rootFrame)
	pipe.printer = printer
	pipe.tracker = parent
	pipe.renderFrame = sh.Join(parent.newRenderFrame(), n.renderShread.Frame())
	pipe.submit(func(ok bool) {
		defer pipe.close()
		if !ok {
			return
		}
		cur := gox.NewCursor(parent.ctx, pipe)
		n.view.renderContent(cur)
	})
	replaceFrame := sh.Join(n.renderShread.Frame(), n.communicationFrame)
	replaceFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
		defer rootFrame.Activate()
		defer n.done()
		if !ok {
			n.report(context.Canceled)
			return
		}
		// push replace
	})
	replaceFrame.Release()
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

func (n *node) render(parentPipe *pipe) {
	pipe := parentPipe.branch()
	parent := parentPipe.tracker
	frame := sh.Join(parentPipe.renderFrame, n.renderShread.Frame())
	defer frame.Release()
	frame.Run(parent.ctx, parent.root.runtime(), func(ok bool) {
		if !ok {
			defer n.takoverFrame.Activate()
			if n.kind == replacedNode {
				return
			}
			n.kind = unmountedNode
			return
		}
		if n.kind == replacedNode {
			cur := gox.NewCursor(parent.ctx, pipe)
			cur.Any(n.view.content)
			return
		}
		n.communicationFrame = pipe.rootFrame
		n.tracker = newTracker(parent)
		n.tracker.parent.addChild(n)
		pipe.tracker = n.tracker
		pipe.renderFrame = sh.Join(parentPipe.renderFrame, n.tracker.newRenderFrame())
		pipe.submit(func(ok bool) {
			defer pipe.close()
			if !ok {
				defer n.takoverFrame.Activate()
				n.kind = unmountedNode
				n.tracker.kill()
				return
			}
			cur := gox.NewCursor(n.tracker.ctx, pipe)
			switch n.kind {
			case editorNode:
				n.takoverFrame.Activate()
				open, close := n.view.headFrame(parent.ctx, n.tracker.id, cur.NewID())
				cur.Send(open)
				n.view.renderContent(cur)
				cur.Send(close)
			case proxyNode:
				proxy := newProxyComponent(n.tracker.id, n.view, parent.ctx, &n.takoverFrame)
				proxy.Main()(cur)
			}
		})
	})
}

func (n *node) accept() {
	n.done()
	close(n.reportCh)
}

func (n *node) report(err error) {
	n.reportCh <- err
	close(n.reportCh)
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
		n.communicationFrame.Run(n.tracker.parent.ctx, n.tracker.root.runtime(), func(ok bool) {
			if !ok {
				return
			}
		})
		n.tracker.kill()
	}
}
