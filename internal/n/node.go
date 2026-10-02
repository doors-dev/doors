package n

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type nodeMode int

const (
	modeOuter nodeMode = iota
	modeInner
	modeBlend
	modeStatic
)

const (
	nodeInit int32 = iota
	nodeFlushed
	nodeSuperseded
)

type node struct {
	door       *Door
	state      atomic.Int32
	initGuard  shredder.ValveFrame
	placeGuard *shredder.ValveFrame
	mode       nodeMode
	tracker    *tracker
	outer      any
	inner      any
	static     struct {
		id        uint64
		callGuard *shredder.ValveFrame
	}
}

func (n *node) runtime() shredder.Runtime {
	return n.tracker.outer.root.runtime()
}

func (n *node) runtimeContext() context.Context {
	return n.tracker.outer.root.runtime().Context()
}

func (n *node) isMounted() bool {
	return n.mode != modeStatic && n.tracker != nil
}

func (n *node) sync(task *userTask) {
	if n.mode == modeStatic {
		panic("unexpected mode")
	}
}

func (n *node) syncStatic(task *userTask) {
	if n.mode != modeStatic {
		panic("unexpected mode")
	}
}

func (n *node) render(parentPipe *pipe, buffer *deque.Deque[any]) {
	if n.mode == modeStatic {
		panic("unexpected mode")
	}
	thread := shredder.Thread{}
	renderFrame := shredder.Join(n.runtimeContext(), true, parentPipe.renderFrame, thread.Frame(), n.tracker.rw.Write())
	defer renderFrame.Release()
	pip := newPipe(
		n.tracker,
		buffer,
		renderFrame,
		n.tracker.outer.callGuard,
	)
	var err error
	renderFrame.Submit(n.tracker.outer.ctx, n.runtime(), func(b bool) {
		if !b {
			return
		}
		switch n.mode {
		case modeOuter:
			err = n.renderOuter(pip)
		case modeInner:
			err = n.renderInnerOuter(pip)
		case modeBlend:
			err = n.renderBlend(pip)
		default:
			panic("unexpected node mode")
		}
	})
	finalFrame := shredder.Join(n.runtimeContext(), true, parentPipe.renderFrame, thread.Frame())
	defer finalFrame.Release()
	finalFrame.Run(parentPipe.tracker.ctx, n.runtime(), func(b bool) {
		defer n.placeGuard.Activate()
		if !b {
			return
		}
		if err == nil {
			return
		}
		n.tracker.outer.clean()
		n.tracker.clean()
		n.tracker.outer.placementError = err
		pip.error(err)
	})
}

func (n *node) renderStatic(parentPipe *pipe, buffer *deque.Deque[any]) {
	if n.mode != modeStatic {
		panic("unexpected mode")
	}
	parentPipe.renderFrame.Submit(parentPipe.tracker.renderCtx, n.runtime(), func(b bool) {
		if !b {
			return
		}
		pipe := newPipe(
			parentPipe.tracker,
			buffer,
			parentPipe.renderFrame,
			parentPipe.callGuard,
		)
		cur := gox.NewCursor(parentPipe.tracker.renderCtx, pipe)
		if err := cur.Any(n.outer); err != nil {
			pipe.error(err)
		}
	})
}

func (n *node) renderOuter(pip *pipe) (err error) {
	printer := &nodePrinter{
		pipe: pip,
	}
	if n.outer != nil {
		cur := gox.NewCursor(pip.innerContext(), printer)
		err = cur.Any(n.outer)
	}
	if err != nil {
		return err
	}
	return printer.submitContainer()
}

func (n *node) renderInnerOuter(pip *pipe) (err error) {
	printer := &nodePrinter{
		pipe:        pip,
		skipContent: true,
	}
	if n.outer != nil {
		cur := gox.NewCursor(pip.innerContext(), printer)
		err = cur.Any(n.outer)
	}
	if err != nil {
		return err
	}
	if n.inner != nil {
		err = n.renderInner(pip)
	}
	if err != nil {
		return err
	}
	return printer.submitContainer()
}

func (n *node) renderBlend(pip *pipe) (err error) {
	printer := &nodePrinter{
		pipe: pip,
	}
	if n.outer != nil {
		cur := gox.NewCursor(pip.innerContext(), printer)
		err = cur.Any(n.outer)
	}
	if err != nil {
		return err
	}
	if pip.isEmpty() && n.inner != nil {
		err = n.renderInner(pip)
	}
	if err != nil {
		return err
	}
	return printer.submitContainer()
}

func (n *node) renderInner(pip *pipe) (err error) {
	cur := gox.NewCursor(pip.innerContext(), pip)
	return cur.Any(n.inner)
}

func (n *node) scheduleRemoval() {
	var thread shredder.Thread
	placeFrame := shredder.Join(n.runtimeContext(), true, n.placeGuard, thread.Frame())
	placeFrame.Run(nil, n.runtime(), func(b bool) {
		n.tracker.outer.clean()
		n.tracker.clean()
	})
	placeFrame.Release()
	callFrame := shredder.Join(n.runtimeContext(), true, n.tracker.outer.callGuard, thread.Frame())
	callFrame.Run(n.tracker.outer.parent.ctx, n.runtime(), func(b bool) {
		if !b {
			return
		}
		// remove call
	})
}
