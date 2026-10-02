package n

import (
	"context"
	"fmt"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type renderNode struct {
	pipe   *pipe
	buffer *deque.Deque[any]
	ctx    context.Context
}

func (n renderNode) apply(_ shredder.SimpleFrame, next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.inner = prev.inner
	if next.mode == modeStatic {
		next.renderStatic(n.pipe, n.buffer)
		return
	}
	if prev.isMounted() {
		prev.scheduleRemoval()
	}
	outer := newOuterTracker2(n.pipe.tracker, n.pipe.callGuard)
	next.tracker = outer.newTracker()
	next.placeGuard = &outer.outerGuard
	next.render(n.pipe, n.buffer)
}

type nodeProxy struct {
	el     gox.Elem
	pipe   *pipe
	buffer *deque.Deque[any]
	ctx    context.Context
}

func (n nodeProxy) apply(next *node, prev *node) {
	next.mode = modeBlend
	next.outer = n.el
	next.inner = prev.inner
	if prev.isMounted() {
		prev.scheduleRemoval()
	}
	outer := newOuterTracker2(n.pipe.tracker, n.pipe.callGuard)
	next.tracker = outer.newTracker()
	next.render(n.pipe, n.buffer)
}

type innerNode struct {
	*userTask
	inner any
}

func (n innerNode) apply(initFrame shredder.SimpleFrame, next *node, prev *node) {
	next.mode = modeInner
	next.outer = prev.outer
	next.inner = n.inner
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	next.tracker = prev.tracker.outer.newTracker()
	var thread shredder.Thread
	if prev.inner == modeInner {
		placeFrame := shredder.Join(prev.runtimeContext(), true, initFrame, prev.placeGuard, thread.Frame())
		placeFrame.Run(nil, prev.runtime(), func(b bool) {
			prev.tracker.clean()
		})
		placeFrame.Release()
	}
	outerFrame := shredder.Join(prev.runtimeContext(), true, initFrame, &prev.tracker.outer.outerGuard, thread.Frame())
	defer outerFrame.Release()
	outerFrame.Run(next.tracker.ctx, prev.runtime(), func(b bool) {
		if prev.inner != modeInner {
			prev.tracker.clean()
		}
		if !b {
			n.Cancel()
			return
		}
		if next.tracker.outer.placementError != nil {
			next.tracker.clean()
			n.Report(fmt.Errorf("placement error: %v", next.tracker.outer.placementError))
			return
		}
	})
}

type outerNode struct {
	*userTask
	inner any
}

func (n outerNode) apply(initFrame shredder.SimpleFrame, next *node, prev *node) {
	next.mode = modeOuter
	next.outer = prev.outer
	next.inner = nil
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	outer := newOuterTracker1(prev.tracker.outer)
	next.tracker = outer.newTracker()
	placeFrame := shredder.Join(prev.runtimeContext(), false, initFrame, prev.placeGuard)
	placeFrame.Run(next.tracker.outer.ctx, prev.runtime(), func(b bool) {
		next.tracker.outer.placementError = prev.tracker.outer.placementError
		prev.tracker.outer.clean()
		prev.tracker.clean()
		if !b {
			defer next.tracker.outer.outerGuard.Activate()
			n.Cancel()
			return
		}
		if next.tracker.outer.placementError != nil {
			defer next.tracker.outer.outerGuard.Activate()
			next.tracker.outer.clean()
			next.tracker.clean()
			n.Report(fmt.Errorf("placement error: %v", next.tracker.outer.placementError))
			return
		}
		next.sync(n.userTask)
	})
}

type staticNide struct {
	*userTask
	outer any
}

func (n staticNide) apply(initFrame shredder.SimpleFrame, next *node, prev *node) {
	next.mode = modeStatic
	next.outer = n.outer
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	next.tracker = prev.tracker.outer.parent
	next.static.id = prev.tracker.outer.id
	next.static.callGuard = new(shredder.ValveFrame)
	placeFrame := shredder.Join(prev.runtimeContext(), false, initFrame, prev.placeGuard)
	placeFrame.Run(next.tracker.ctx, prev.runtime(), func(b bool) {
		prev.tracker.outer.clean()
		prev.tracker.clean()
		if !b {
			n.Cancel()
			return
		}
		if prev.tracker.outer.placementError != nil {
			n.Report(fmt.Errorf("placement error: %v", next.tracker.outer.placementError))
			return
		}
		next.syncStatic(n.userTask)
	})
}

type nodeReload struct {
	*userTask
}

func (n nodeReload) apply(initFrame shredder.SimpleFrame, next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.inner = prev.inner
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	if prev.mode != modeInner {
		outer := newOuterTracker1(prev.tracker.outer)
		next.tracker = outer.newTracker()
		placeFrame := shredder.Join(prev.runtimeContext(), false, initFrame, prev.placeGuard)
		placeFrame.Run(next.tracker.outer.ctx, prev.runtime(), func(b bool) {
			next.tracker.outer.placementError = prev.tracker.outer.placementError
			prev.tracker.outer.clean()
			prev.tracker.clean()
			if !b {
				defer next.tracker.outer.outerGuard.Activate()
				n.Cancel()
				return
			}
			if next.tracker.outer.placementError != nil {
				defer next.tracker.outer.outerGuard.Activate()
				next.tracker.outer.clean()
				next.tracker.clean()
				n.Report(fmt.Errorf("placement error: %v", next.tracker.outer.placementError))
				return
			}
			next.sync(n.userTask)
		})
		return
	}
	next.tracker = prev.tracker.outer.newTracker()
	var thread shredder.Thread
	placeFrame := shredder.Join(prev.runtimeContext(), true, initFrame, prev.placeGuard, thread.Frame())
	placeFrame.Run(nil, prev.runtime(), func(b bool) {
		prev.tracker.clean()
	})
	placeFrame.Release()
	outerFrame := shredder.Join(prev.runtimeContext(), true, initFrame, &prev.tracker.outer.outerGuard, thread.Frame())
	defer outerFrame.Release()
	outerFrame.Run(next.tracker.ctx, prev.runtime(), func(b bool) {
		if !b {
			n.Cancel()
			return
		}
		if next.tracker.outer.placementError != nil {
			next.tracker.clean()
			n.Report(fmt.Errorf("placement error: %v", next.tracker.outer.placementError))
			return
		}
	})
}

type unmountNode struct {
	*userTask
}

func (t unmountNode) apply(initFrame shredder.SimpleFrame, next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.inner = prev.inner
	if !prev.isMounted() {
		t.userTask.Accept()
		return
	}
	var thread shredder.Thread
	placeFrame := shredder.Join(prev.runtimeContext(), true, initFrame, prev.placeGuard, thread.Frame())
	placeFrame.Run(nil, prev.runtime(), func(b bool) {
		prev.tracker.outer.clean()
		prev.tracker.clean()
	})
	placeFrame.Release()
	callFrame := shredder.Join(prev.runtimeContext(), false, prev.tracker.outer.callGuard, t.CallFrame(), thread.Frame())
	callFrame.Run(prev.tracker.outer.parent.ctx, prev.runtime(), func(b bool) {
		if !b {
			t.Cancel()
			return
		}
		// call remove
	})
}

var _ nodeTask = renderNode{}

type nodeTask interface {
	apply(initFrame shredder.SimpleFrame, next *node, prev *node)
}

func newUserTask(ctx context.Context) (*userTask, <-chan error) {
	ch := make(chan error, 2)
	return &userTask{ch: &ch, ctx: ctx, frames: ctex.GetFrames(ctx)}, ch
}

type userTask struct {
	ch     *chan error
	ctx    context.Context
	frames ctex.Frames
}

func (t *userTask) InitFrame() shredder.Frame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.InitFrame(t.ctx)
}

func (t *userTask) CallFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Call()
}

func (t *userTask) RenderFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Render()
}

func (t *userTask) Scheduled() {
	if t == nil {
		return
	}
	if t.ch == nil {
		return
	}
	*t.ch <- nil
}

func (t *userTask) Report(err error) {
	if t == nil {
		return
	}
	if t.ch == nil {
		return
	}
	*t.ch <- err
	close(*t.ch)
	t.ch = nil
}

func (t *userTask) Cancel() {
	if t == nil {
		return
	}
	t.Report(context.Canceled)
}

func (t *userTask) Accept() {
	if t == nil {
		return
	}
	if t.ch != nil {
		close(*t.ch)
		t.ch = nil
	}
}
