// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package door

import (
	"context"
	"fmt"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type renderNode struct {
	pipe   *pipe
	buffer *deque.Deque[any]
	ctx    context.Context
	caller common.Caller
}

func (n renderNode) apply(_ shredder.Frame, next *node, prev *node) {
	next.mode = prev.mode
	next.caller = n.caller
	next.outer = prev.outer
	next.inner = prev.inner
	if next.mode == modeStatic {
		next.static.user = common.UserCtx(n.ctx, n.pipe.innerContext())
		next.placeRenderStatic(n.pipe, n.buffer)
		return
	}
	if prev.isMounted() {
		prev.scheduleRemoval(n.pipe.callGuard)
	}
	outer := newOuterTrackerCreate(n.pipe, n.ctx)
	next.tracker = outer.newTracker(next)
	next.placeRender(n.pipe, n.buffer)
}

type nodeProxy struct {
	el     gox.Elem
	pipe   *pipe
	buffer *deque.Deque[any]
	ctx    context.Context
	caller common.Caller
}

func (n nodeProxy) apply(_ shredder.Frame, next *node, prev *node) {
	next.mode = modeBlend
	next.caller = n.caller
	next.outer = n.el
	next.inner = prev.inner
	if prev.isMounted() {
		prev.scheduleRemoval(n.pipe.callGuard)
	}
	outer := newOuterTrackerCreate(n.pipe, n.ctx)
	next.tracker = outer.newTracker(next)
	next.placeRender(n.pipe, n.buffer)
}

type innerNode struct {
	*userTask
	inner   any
	defered bool
}

func (n innerNode) apply(initFrame shredder.Frame, next *node, prev *node) {
	next.mode = modeInner
	next.caller = n.caller
	if prev.mode != modeStatic {
		next.outer = prev.outer
	}
	next.inner = n.inner
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	next.tracker = prev.tracker.outer.newTracker(next)
	var thread shredder.Thread
	if n.defered {
		n.userTask.Detach()
		if prev.mode == modeInner {
			prevTracker, prevInnerClean := prev.tracker, prev.deferedInnerClean
			next.deferedInnerClean = common.Once(func() {
				prevTracker.clean()
				if prevInnerClean != nil {
					prevInnerClean()
				}
			})
			next.deferedOuterClean = prev.deferedOuterClean
		} else {
			prevTracker, prevInnerClean, prevOuterClean := prev.tracker, prev.deferedInnerClean, prev.deferedOuterClean
			next.deferedOuterClean = common.Once(func() {
				prevTracker.clean()
				if prevInnerClean != nil {
					prevInnerClean()
				}
				if prevOuterClean != nil {
					prevOuterClean()
				}
			})
		}
		write := next.tracker.rw.Write()
		syncFrame := shredder.JoinRelease(prev.runtimeContext(), write, prev.tracker.outer.callGuard, prev.tracker.rw.Read())
		defer syncFrame.Release()
		syncFrame.Run(next.tracker.ctx, prev.tracker.Runtime(), func(b bool) {
			if next.tracker.outer.placementError != nil {
				next.tracker.clean()
				n.Report(fmt.Errorf("placement error: %w", next.tracker.outer.placementError))
				next.cleanDefered()
				return
			}
			if next.tracker.outer.outerError != nil {
				next.tracker.clean()
				n.Report(fmt.Errorf("outer error: %w", next.tracker.outer.outerError))
				next.cleanDefered()
				return
			}
			if !b {
				n.Cancel()
				next.cleanDefered()
				return
			}
			next.sync(n.userTask, write)
		})
		return
	}
	next.deferedOuterClean = prev.deferedOuterClean
	if prev.mode == modeInner {
		placeFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.placeGuard, thread.Frame())
		placeFrame.Run(nil, prev.tracker.Runtime(), func(b bool) {
			prev.tracker.clean()
			if prev.deferedInnerClean != nil {
				prev.deferedInnerClean()
			}
		})
		placeFrame.Release()
	}
	outerFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.outerGuard, thread.Frame())
	outerFrame.Run(nil, prev.tracker.Runtime(), func(bool) {
		if prev.mode != modeInner {
			prev.tracker.clean()
		}
		if prev.deferedOuterClean != nil {
			prev.deferedOuterClean()
		}
	})
	outerFrame.Release()
	write := next.tracker.rw.Write()
	syncFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, thread.Frame(), write)
	defer syncFrame.Release()
	syncFrame.Run(next.tracker.ctx, prev.tracker.Runtime(), func(b bool) {
		if next.tracker.outer.placementError != nil {
			next.tracker.clean()
			n.Report(fmt.Errorf("placement error: %w", next.tracker.outer.placementError))
			return
		}
		if next.tracker.outer.outerError != nil {
			next.tracker.clean()
			n.Report(fmt.Errorf("outer error: %w", next.tracker.outer.outerError))
			return
		}
		if !b {
			n.Cancel()
			return
		}
		next.sync(n.userTask, write)
	})
}

type outerNode struct {
	*userTask
	outer   any
	defered bool
}

func (n outerNode) apply(initFrame shredder.Frame, next *node, prev *node) {
	next.mode = modeOuter
	next.caller = n.caller
	next.outer = n.outer
	next.inner = nil
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	outer := newOuterTrackerInherit(prev.tracker.outer)
	next.tracker = outer.newTracker(next)
	if n.defered {
		n.userTask.Detach()
		prevTracker, prevInnerClean, prevOuterClean := prev.tracker, prev.deferedInnerClean, prev.deferedOuterClean
		next.deferedOuterClean = common.Once(func() {
			prevTracker.outer.clean()
			prevTracker.clean()
			if prevInnerClean != nil {
				prevInnerClean()
			}
			if prevOuterClean != nil {
				prevOuterClean()
			}
		})
		var thread shredder.Thread
		placeFrame := shredder.JoinRelease(prev.runtimeContext(), prev.tracker.outer.placeGuard, thread.Frame())
		placeFrame.Run(nil, prev.tracker.Runtime(), func(bool) {
			next.tracker.outer.placementError = prev.tracker.outer.placementError
		})
		placeFrame.Release()
		write := next.tracker.rw.Write()
		syncFrame := shredder.JoinRelease(prev.runtimeContext(), write, prev.tracker.outer.callGuard, prev.tracker.rw.Read(), thread.Frame())
		defer syncFrame.Release()
		syncFrame.Run(next.tracker.outer.ctx, prev.tracker.Runtime(), func(b bool) {
			if !b {
				defer next.tracker.outer.outerGuard.Activate()
				n.Cancel()
				next.cleanDefered()
				return
			}
			if next.tracker.outer.placementError != nil {
				defer next.tracker.outer.outerGuard.Activate()
				next.tracker.outer.clean()
				next.tracker.clean()
				n.Report(fmt.Errorf("placement error: %w", next.tracker.outer.placementError))
				next.cleanDefered()
				return
			}
			next.sync(n.userTask, write)
		})
		return
	}
	var thread shredder.Thread
	placeFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.placeGuard, thread.Frame())
	placeFrame.Run(nil, prev.tracker.Runtime(), func(bool) {
		next.tracker.outer.placementError = prev.tracker.outer.placementError
		prev.tracker.outer.clean()
		prev.tracker.clean()
		prev.cleanDefered()
	})
	placeFrame.Release()
	write := next.tracker.rw.Write()
	syncFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, thread.Frame(), write)
	defer syncFrame.Release()
	syncFrame.Run(next.tracker.outer.ctx, prev.tracker.Runtime(), func(b bool) {
		if !b {
			defer next.tracker.outer.outerGuard.Activate()
			n.Cancel()
			return
		}
		if next.tracker.outer.placementError != nil {
			defer next.tracker.outer.outerGuard.Activate()
			next.tracker.outer.clean()
			next.tracker.clean()
			n.Report(fmt.Errorf("placement error: %w", next.tracker.outer.placementError))
			return
		}
		next.sync(n.userTask, write)
	})
}

type staticNode struct {
	*userTask
	outer   any
	defered bool
}

func (n staticNode) apply(initFrame shredder.Frame, next *node, prev *node) {
	next.mode = modeStatic
	next.caller = n.caller
	next.outer = n.outer
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	next.tracker = prev.tracker.outer.parent
	next.static.id = prev.tracker.outer.id
	next.static.user = prev.tracker.outer.user
	if n.defered {
		n.userTask.Detach()
		prevTracker, prevInnerClean, prevOuterClean := prev.tracker, prev.deferedInnerClean, prev.deferedOuterClean
		next.deferedOuterClean = common.Once(func() {
			prevTracker.outer.clean()
			prevTracker.clean()
			if prevInnerClean != nil {
				prevInnerClean()
			}
			if prevOuterClean != nil {
				prevOuterClean()
			}
		})
		syncFrame := shredder.JoinRelease(prev.runtimeContext(), prev.tracker.outer.callGuard, prev.tracker.rw.Read())
		defer syncFrame.Release()
		syncFrame.Run(next.tracker.ctx, prev.tracker.Runtime(), func(b bool) {
			if !b {
				n.Cancel()
				next.cleanDefered()
				return
			}
			if prev.tracker.outer.placementError != nil {
				n.Report(fmt.Errorf("placement error: %w", prev.tracker.outer.placementError))
				next.cleanDefered()
				return
			}
			next.sync(n.userTask, nil)
		})
		return
	}
	var thread shredder.Thread
	placeFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.placeGuard, thread.Frame())
	placeFrame.Run(nil, prev.tracker.Runtime(), func(bool) {
		prev.tracker.outer.clean()
		prev.tracker.clean()
		prev.cleanDefered()
	})
	placeFrame.Release()
	syncFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, thread.Frame())
	defer syncFrame.Release()
	syncFrame.Run(next.tracker.ctx, prev.tracker.Runtime(), func(b bool) {
		if !b {
			n.Cancel()
			return
		}
		if prev.tracker.outer.placementError != nil {
			n.Report(fmt.Errorf("placement error: %w", prev.tracker.outer.placementError))
			return
		}
		next.sync(n.userTask, nil)
	})
}

type nodeReload struct {
	*userTask
}

func (n nodeReload) apply(initFrame shredder.Frame, next *node, prev *node) {
	next.mode = prev.mode
	next.caller = n.caller
	next.outer = prev.outer
	next.inner = prev.inner
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	if prev.mode != modeInner {
		outer := newOuterTrackerInherit(prev.tracker.outer)
		next.tracker = outer.newTracker(next)
		var thread shredder.Thread
		placeFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.placeGuard, thread.Frame())
		placeFrame.Run(nil, prev.tracker.Runtime(), func(bool) {
			next.tracker.outer.placementError = prev.tracker.outer.placementError
			prev.tracker.outer.clean()
			prev.tracker.clean()
			prev.cleanDefered()
		})
		placeFrame.Release()
		write := next.tracker.rw.Write()
		syncFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, thread.Frame(), write)
		defer syncFrame.Release()
		syncFrame.Run(next.tracker.outer.ctx, prev.tracker.Runtime(), func(b bool) {
			if !b {
				defer next.tracker.outer.outerGuard.Activate()
				n.Cancel()
				return
			}
			if next.tracker.outer.placementError != nil {
				defer next.tracker.outer.outerGuard.Activate()
				next.tracker.outer.clean()
				next.tracker.clean()
				n.Report(fmt.Errorf("placement error: %w", next.tracker.outer.placementError))
				return
			}
			next.sync(n.userTask, write)
		})
		return
	}
	next.tracker = prev.tracker.outer.newTracker(next)
	next.deferedOuterClean = prev.deferedOuterClean
	var thread shredder.Thread
	placeFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.placeGuard, thread.Frame())
	placeFrame.Run(nil, prev.tracker.Runtime(), func(b bool) {
		prev.tracker.clean()
		if prev.deferedInnerClean != nil {
			prev.deferedInnerClean()
		}
	})
	placeFrame.Release()
	outerFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.outerGuard, thread.Frame())
	outerFrame.Run(nil, prev.tracker.Runtime(), func(bool) {
		if prev.deferedOuterClean != nil {
			prev.deferedOuterClean()
		}
	})
	outerFrame.Release()
	write := next.tracker.rw.Write()
	syncFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, thread.Frame(), write)
	defer syncFrame.Release()
	syncFrame.Run(next.tracker.ctx, prev.tracker.Runtime(), func(b bool) {
		if next.tracker.outer.placementError != nil {
			next.tracker.clean()
			n.Report(fmt.Errorf("placement error: %w", next.tracker.outer.placementError))
			return
		}
		if next.tracker.outer.outerError != nil {
			next.tracker.clean()
			n.Report(fmt.Errorf("outer error: %w", next.tracker.outer.outerError))
			return
		}
		if !b {
			n.Cancel()
			return
		}
		next.sync(n.userTask, write)
	})
}

type unmountNode struct {
	*userTask
}

func (n unmountNode) apply(initFrame shredder.Frame, next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.inner = prev.inner
	if !prev.isMounted() {
		n.userTask.Accept()
		return
	}
	var thread shredder.Thread
	placeFrame := shredder.JoinRelease(prev.runtimeContext(), initFrame, prev.tracker.outer.placeGuard, thread.Frame())
	placeFrame.Run(nil, prev.tracker.Runtime(), func(b bool) {
		prev.tracker.outer.clean()
		prev.tracker.clean()
		prev.cleanDefered()
	})
	placeFrame.Release()
	callFrame := shredder.Join(prev.runtimeContext(), prev.tracker.outer.callGuard, n.CallFrame(), thread.Frame())
	defer callFrame.Release()
	callFrame.Run(prev.tracker.outer.parent.ctx, prev.tracker.Runtime(), func(b bool) {
		if !b {
			n.Cancel()
			return
		}
		if prev.tracker.outer.placementError != nil {
			n.Report(fmt.Errorf("placement error: %w", prev.tracker.outer.placementError))
			return
		}
		n.Scheduled()
		prev.call(&call{
			ctx:     prev.tracker.outer.parent.ctx,
			kind:    callReplace,
			id:      prev.tracker.outer.id,
			payload: emptyPayload{},
			task:    n.userTask,
			logger:  prev.logger(),
		})
	})
}

var _ nodeTask = renderNode{}

type nodeTask interface {
	apply(initFrame shredder.Frame, next *node, prev *node)
}

func newUserTask(ctx context.Context) (*userTask, <-chan error) {
	ch := make(chan error, 2)
	return &userTask{ch: &ch, ctx: ctx, frames: ctex.GetFrames(ctx), caller: common.CaptureCaller()}, ch
}

type userTask struct {
	ch     *chan error
	ctx    context.Context
	frames ctex.Frames
	caller common.Caller
}

func (t *userTask) Detach() {
	t.frames = ctex.Frames{}
}

func (t *userTask) InitFrame() shredder.ReleaseFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.InitFrame(t.ctx)
}

func (t *userTask) CallFrame() shredder.Frame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Call()
}

func (t *userTask) RenderFrame() shredder.Frame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return shredder.Join(t.ctx, t.frames.Render())
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
