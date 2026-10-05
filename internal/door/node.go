package door

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/printer"
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
	caller    common.Caller
	door      *Door
	state     atomic.Int32
	initGuard shredder.ValveFrame
	mode      nodeMode
	tracker   *tracker
	outer     any
	inner     any
	static    struct {
		id        uint64
		callGuard *shredder.ValveFrame
		user      context.Context
	}
}

func (n *node) runtimeContext() context.Context {
	return n.tracker.outer.root.runtime().Context()
}

func (n *node) isMounted() bool {
	return n.mode != modeStatic && n.tracker != nil && n.tracker.outer.parent.ctx.Err() == nil
}

func (n *node) newPrinter() (*printer.PayloadPrinter, bool) {
	switch n.mode {
	case modeInner, modeStatic:
		return n.tracker.newPrinter()
	default:
		return n.tracker.outer.newPrinter()
	}
}

func (n *node) printerMiddleware() func(gox.Printer) gox.Printer {
	return n.tracker.outer.root.inst.Session().App().PrinterMiddleware()
}

func (n *node) logger() *slog.Logger {
	return n.tracker.outer.root.inst.Logger()
}

func (n *node) call(call actions.Call) {
	n.tracker.outer.root.inst.Call(call)
}

func (n *node) callKind() callKind {
	switch n.mode {
	case modeOuter:
		return callReplace
	case modeInner:
		return callUpdate
	case modeBlend:
		return callReplace
	case modeStatic:
		return callReplace
	default:
		panic("unknown node mode")
	}
}

func (n *node) callContext() context.Context {
	switch n.mode {
	case modeOuter, modeBlend:
		return n.tracker.outer.ctx
	case modeInner, modeStatic:
		return n.tracker.ctx
	default:
		panic("unknown node mode")
	}
}

func (n *node) syncRender(pip *pipe) error {
	switch n.mode {
	case modeOuter:
		return n.renderOuter(pip)
	case modeInner:
		return n.renderInner(pip)
	case modeBlend:
		return n.renderBlend(pip)
	case modeStatic:
		return n.renderStatic(pip)
	default:
		panic("unknown node mode")
	}
}

func (n *node) callID() uint64 {
	switch n.mode {
	case modeOuter, modeInner, modeBlend:
		return n.tracker.outer.id
	case modeStatic:
		return n.static.id
	default:
		panic("unknown node mode")
	}
}

func (n *node) onSyncError(err error) {
	logError(n.tracker.ctx, n.logger(), err, n.caller)
	switch n.mode {
	case modeOuter, modeBlend:
		n.tracker.outer.outerError = err
		n.tracker.outer.clean()
		n.tracker.clean()
	case modeInner:
		n.tracker.clean()
	case modeStatic:
	default:
		panic("unknown node mode")
	}
}

func (n *node) syncRenderFrame(task *userTask, threadFrame shredder.ReleaseFrame, writeFrame shredder.ReleaseFrame) shredder.ReleaseFrame {
	switch n.mode {
	case modeOuter, modeInner, modeBlend:
		return shredder.JoinRelease(n.runtimeContext(), threadFrame, writeFrame, n.tracker.cinemaFrame(), task.RenderFrame())
	case modeStatic:
		return shredder.JoinRelease(n.runtimeContext(), threadFrame, writeFrame, n.tracker.cinemaFrame())
	default:
		panic("unknown node mode")
	}
}

func (n *node) sync(task *userTask) {
	thread := shredder.Thread{}
	writeFrame := n.tracker.rw.Write()
	renderFrame := n.syncRenderFrame(task, thread.Frame(), writeFrame)
	defer renderFrame.Release()
	innerCallGuard := new(shredder.ValveFrame)
	pip := newPipe(
		n.tracker,
		common.GetDequeBuffer(),
		renderFrame,
		innerCallGuard,
	)
	var err error
	pip.renderFrame.Submit(n.tracker.ctx, pip.runtime(), func(b bool) {
		if !b {
			return
		}
		err = n.syncRender(pip)
	})
	callFrame := shredder.JoinRelease(n.runtimeContext(), thread.Frame(), writeFrame, n.tracker.outer.callGuard, task.CallFrame())
	defer callFrame.Release()
	callFrame.Run(n.tracker.ctx, pip.runtime(), func(b bool) {
		defer innerCallGuard.Activate()
		if n.mode != modeInner {
			defer n.tracker.outer.outerGuard.Activate()
		}
		if !b {
			pip.Release()
			task.Cancel()
			return
		}
		var payload printer.Payload
		if err == nil {
			p, ok := n.newPrinter()
			if !ok {
				pip.Release()
				task.Cancel()
				return
			}
			if !p.Lock() {
				pip.Release()
				task.Cancel()
				return
			}
			payload, err = pip.Render(p, n.printerMiddleware())
			p.Free()
		}
		if n.tracker.ctx.Err() != nil {
			if payload != nil {
				payload.Release()
			}
			pip.Release()
			task.Cancel()
			return
		}
		if err != nil {
			if payload != nil {
				payload.Release()
			}
			pip.Release()
			task.Report(err)
			n.onSyncError(err)
			return
		}
		n.call(&call{
			ctx:     n.callContext(),
			kind:    n.callKind(),
			task:    task,
			id:      n.callID(),
			payload: payload,
			logger:  n.logger(),
		})
	})
}

func (n *node) placeRender(parentPipe *pipe, buffer *deque.Deque[any]) {
	if n.mode == modeStatic {
		panic("unexpected mode")
	}
	thread := shredder.Thread{}
	renderFrame := shredder.JoinRelease(n.runtimeContext(), parentPipe.renderFrame, thread.Frame(), n.tracker.rw.Write(), n.tracker.cinemaFrame())
	defer renderFrame.Release()
	pip := newPipe(
		n.tracker,
		buffer,
		renderFrame,
		n.tracker.outer.callGuard,
	)
	var err error
	renderFrame.Submit(n.tracker.outer.ctx, pip.runtime(), func(b bool) {
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
	finalFrame := shredder.JoinRelease(n.runtimeContext(), parentPipe.renderFrame, thread.Frame())
	defer finalFrame.Release()
	finalFrame.Run(parentPipe.tracker.ctx, parentPipe.runtime(), func(b bool) {
		defer n.tracker.outer.placeGuard.Activate()
		if !b {
			return
		}
		if err == nil {
			return
		}
		n.tracker.outer.clean()
		n.tracker.clean()
		n.tracker.outer.placementError = err
		n.tracker.outer.outerError = err
		pip.error(err, n.caller)
	})
}

func (n *node) placeRenderStatic(parentPipe *pipe, buffer *deque.Deque[any]) {
	if n.mode != modeStatic {
		panic("unexpected mode")
	}
	parentPipe.renderFrame.Submit(parentPipe.tracker.renderCtx, parentPipe.runtime(), func(b bool) {
		if !b {
			return
		}
		pipe := newPipe(
			parentPipe.tracker,
			buffer,
			parentPipe.renderFrame,
			parentPipe.callGuard,
		)
		cur := gox.NewCursor(common.NewRenderCtx(parentPipe.tracker.renderCtx, n.static.user), pipe)
		if err := cur.Any(n.outer); err != nil {
			pipe.error(err, n.caller)
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
	placeFrame := shredder.JoinRelease(n.runtimeContext(), n.tracker.outer.placeGuard, thread.Frame())
	placeFrame.Run(nil, n.tracker.Runtime(), func(b bool) {
		if n.tracker.outer.placementError != nil {
			return
		}
		n.tracker.outer.clean()
		n.tracker.clean()
	})
	placeFrame.Release()
	callFrame := shredder.JoinRelease(n.runtimeContext(), n.tracker.outer.callGuard, thread.Frame())
	defer callFrame.Release()
	callFrame.Run(n.tracker.outer.parent.ctx, n.tracker.Runtime(), func(b bool) {
		if !b {
			return
		}
		if n.tracker.outer.placementError != nil {
			return
		}
		n.call(&call{
			ctx:     n.tracker.outer.parent.ctx,
			kind:    callReplace,
			id:      n.tracker.outer.id,
			payload: emptyPayload{},
			logger:  n.logger(),
		})
	})
}

func (n *node) renderStatic(pip *pipe) (err error) {
	cur := gox.NewCursor(common.NewRenderCtx(n.tracker.renderCtx, n.static.user), pip)
	return cur.Any(n.outer)
}
