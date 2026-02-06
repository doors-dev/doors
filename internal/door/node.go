package door

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
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
	takoverFrame       shredder.ValveFrame
	renderThread       shredder.Thread
	communicationFrame shredder.SimpleFrame
	tracker            *tracker
	view               *view
	killed             atomic.Bool
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
		renderGuard := n.renderThread.Guard()
		prev.takoverFrame.Run(nil, nil, func(bool) {
			defer renderGuard.Release()
			switch n.kind {
			case editorNode:
				n.editorTakover(prev)
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
				n.unmountTakeover(prev)
			case replacedNode:
				n.replaceTakeover(prev)
			}
		})
	}
}

func (n *node) updatedTakeover(prev *node) {
	switch prev.kind {
	case unmountedNode:
		n.view.attrs = prev.view.attrs
		n.view.tag = prev.view.tag
		n.kind = unmountedNode
		n.accept()
		return
	case replacedNode:
		n.kind = unmountedNode
		n.accept()
		return
	}
	prev.kill(unmountKill)
	n.communicationFrame = prev.communicationFrame
	n.tracker = newTrackerFrom(prev.tracker)
	n.tracker.parent.replaceChild(prev, n)
	if n.view == nil {
		n.view = prev.view
	} else {
		n.view.attrs = prev.view.attrs
		n.view.tag = prev.view.tag
	}
	if n.view.content == nil {
		n.communicationFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
			if !ok {
				n.cancel()
				return
			}
			n.tracker.root.inst.Call(&call{
				ctx:  n.tracker.ctx,
				id:   n.tracker.id,
				kind: callUpdate,
				ch:   n.reportCh,
			})
			n.done()
		})
		return
	}
	rootFrame := shredder.ValveFrame{}
	disableGzip := n.tracker.root.inst.Conf().ServerDisableGzip
	printer := newPrinter(disableGzip)
	pipe := newPipe(&rootFrame)
	pipe.printer = printer
	pipe.tracker = n.tracker
	pipe.renderFrame = shredder.Join(true, n.renderThread.Frame(), n.tracker.newRenderFrame())
	pipe.renderAny(n.tracker.ctx, n.view.content)
	updateFrame := shredder.Join(true, n.renderThread.Frame(), n.communicationFrame)
	updateFrame.Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
		defer rootFrame.Activate()
		if !ok {
			printer.release()
			n.cancel()
			return
		}
		printer.finalize()
		if err := pipe.Error(); err != nil {
			n.tracker.root.inst.Call(&call{
				ctx:     n.tracker.ctx,
				kind:    callUpdate,
				id:      n.tracker.id,
				payload: printer,
			})
			n.tracker.parent.removeChild(n)
			n.kill(errorKill)
			n.error(err)
			return
		}
		n.tracker.root.inst.Call(&call{
			ctx:     n.tracker.ctx,
			kind:    callUpdate,
			id:      n.tracker.id,
			ch:      n.reportCh,
			payload: printer,
		})
		n.done()
	})
	updateFrame.Release()
}

func (n *node) unmountTakeover(prev *node) {
	switch prev.kind {
	case replacedNode:
		n.accept()
		return
	case unmountedNode:
		n.view = prev.view
		n.accept()
		return
	}
	prev.kill(unmountKill)
	prev.tracker.removeChild(prev)
	n.view = prev.view
	id := prev.tracker.id
	ctx := prev.tracker.parentContext()
	prev.communicationFrame.Run(ctx, prev.tracker.root.runtime(), func(ok bool) {
		if !ok {
			n.cancel()
			return
		}
		prev.tracker.root.inst.Call(&call{
			ctx:  ctx,
			kind: callReplace,
			id:   id,
			ch:   n.reportCh,
		})
		n.done()
	})
}

func (n *node) replaceTakeover(prev *node) {
	switch prev.kind {
	case replacedNode, unmountedNode:
		n.accept()
		return
	}
	prev.kill(unmountKill)
	prev.tracker.parent.removeChild(prev)
	parent := prev.tracker.parent
	id := prev.tracker.id
	ctx := parent.ctx
	if n.view.content == nil {
		prev.communicationFrame.Run(parent.ctx, parent.root.runtime(), func(ok bool) {
			defer n.done()
			if !ok {
				n.cancel()
				return
			}
			prev.tracker.root.inst.Call(&call{
				ctx:     ctx,
				kind:    callReplace,
				id:      id,
				ch:      n.reportCh,
				payload: nil,
			})
		})
		return
	}
	rootFrame := shredder.ValveFrame{}
	disableGzip := prev.tracker.root.inst.Conf().ServerDisableGzip
	printer := newPrinter(disableGzip)
	pipe := newPipe(&rootFrame)
	pipe.printer = printer
	pipe.tracker = parent
	pipe.renderFrame = shredder.Join(true, parent.newRenderFrame(), n.renderThread.Frame())
	pipe.renderAny(parent.ctx, n.view.content)
	replaceFrame := shredder.Join(true, n.renderThread.Frame(), prev.communicationFrame)
	replaceFrame.Run(prev.tracker.parentContext(), prev.tracker.root.runtime(), func(ok bool) {
		defer rootFrame.Activate()
		if !ok {
			printer.release()
			n.cancel()
			return
		}
		printer.finalize()
		if err := pipe.Error(); err != nil {
			prev.tracker.root.inst.Call(&call{
				ctx:     parent.ctx,
				kind:    callReplace,
				id:      id,
				payload: printer,
			})
			n.error(err)
			return
		}
		prev.tracker.root.inst.Call(&call{
			ctx:     parent.ctx,
			kind:    callReplace,
			id:      id,
			ch:      n.reportCh,
			payload: printer,
		})
		n.done()
	})
	replaceFrame.Release()
}

func (n *node) proxyTakeover(prev *node) {
	n.view.content = prev.view.content
	switch prev.kind {
	case replacedNode:
	case updatedNode, editorNode, proxyNode:
		prev.kill(removeKill)
		prev.tracker.parent.removeChild(prev)
	}
}

func (n *node) editorTakover(prev *node) {
	n.view = prev.view
	switch prev.kind {
	case updatedNode, editorNode:
		prev.kill(removeKill)
		prev.tracker.parent.removeChild(prev)
	case replacedNode:
		n.kind = replacedNode
		n.takoverFrame.Activate()
	case proxyNode:
		prev.kill(removeKill)
		prev.tracker.parent.removeChild(prev)
		n.kind = proxyNode
	}
}

func (n *node) render(parentPipe *pipe) {
	pipe := parentPipe.branch()
	parent := parentPipe.tracker
	frame := shredder.Join(true, parentPipe.renderFrame, n.renderThread.Frame())
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
			pipe.renderAny(parent.ctx, n.view.content)
			return
		}
		n.communicationFrame = pipe.rootFrame
		n.tracker = newTracker(parent)
		n.tracker.parent.addChild(n)
		pipe.tracker = n.tracker
		renderThread := &shredder.Thread{}
		pipe.renderFrame = shredder.Join(true, parentPipe.renderFrame, n.tracker.newRenderFrame(), renderThread.Frame())
		parentCore := core.NewCore(n.tracker.root.inst, childDoorCore{
			tracker: parent,
			id:      n.tracker.id,
		})
		parentCtx := context.WithValue(parent.ctx, ctex.KeyCore, parentCore)
		switch n.kind {
		case editorNode:
			n.takoverFrame.Activate()
			pipe.renderView(parentCtx, n.view)
		case proxyNode:
			pipe.renderProxy(parentCtx, n.view, &n.takoverFrame)
		}
		renderThread.Frame().Run(n.tracker.ctx, n.tracker.root.runtime(), func(ok bool) {
			if !ok {
				return
			}
			if err := pipe.Error(); err != nil {
				n.tracker.parent.removeChild(n)
				n.kill(errorKill)
				n.error(err)
			}
		})
	})
}

func (n *node) error(err error) {
	n.reportCh <- err
	close(n.reportCh)
	n.done()
}

func (n *node) accept() {
	close(n.reportCh)
	n.done()
}

func (n *node) cancel() {
	n.error(context.Canceled)
}

type killKind int

const (
	cascadeKill killKind = iota
	unmountKill
	removeKill
	errorKill
)

func (n *node) kill(kind killKind) bool {
	if !n.killed.CompareAndSwap(false, true) {
		return false
	}
	if n.kind == unmountedNode || n.kind == replacedNode {
		panic("door: unmounted/replaced node can't be killed")
	}
	switch kind {
	case errorKill:
		n.tracker.kill()
	case cascadeKill:
		unmounted := &node{
			kind: unmountedNode,
			view: n.view,
		}
		unmounted.takoverFrame.Activate()
		n.door.node.CompareAndSwap(n, unmounted)
		n.tracker.kill()
	case unmountKill:
		n.tracker.kill()
	case removeKill:
		id := n.tracker.id
		ctx := n.tracker.parentContext()
		n.communicationFrame.Run(n.tracker.parentContext(), n.tracker.root.runtime(), func(ok bool) {
			if !ok {
				return
			}
			n.tracker.root.inst.Call(&call{
				ctx:  ctx,
				kind: callReplace,
				id:   id,
			})
		})
		n.tracker.kill()
	default:
		panic("door: invalid kill kind")
	}
	return true
}
