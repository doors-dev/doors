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
	"errors"
	"strings"

	"github.com/doors-dev/doors/internal/front"
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

type node struct {
	guard   shredder.ValveFrame
	door    *Door
	mode    nodeMode
	tracker *tracker
	outer   gox.Elem
	content any
}

func (n *node) reload(ctx context.Context) <-chan error {
	return n.door.reloadSelf(ctx, n)
}

func (n *node) unmountedSelf() {
	n.door.unmountedSelf(n)
}

func (n *node) onErr(err error) {
	if !n.isMounted() {
		return
	}
	trackerShutdown(n.tracker)
	n.unmountedSelf()
}

func (n *node) isMounted() bool {
	return n.mode != modeStatic && n.tracker != nil
}

func (n *node) sync(task *userTask) {
	thread := shredder.Thread{}
	ownerTracker := n.tracker
	callGuard := n.tracker.innerCallGuard
	if n.mode == modeStatic {
		ownerTracker = n.tracker.parent
		callGuard = &shredder.ValveFrame{}
	}
	renderFrame := shredder.Join(true, thread.Frame(), ownerTracker.writeFrame(), task.RenderFrame())
	defer renderFrame.Release()
	pip := newPipe(
		ownerTracker,
		new(deque.Deque[any]),
		renderFrame,
		callGuard,
	)
	var err error
	var callKind callKind
	pip.renderFrame.Submit(ownerTracker.ctx, ownerTracker.root.runtime, func(b bool) {
		if !b {
			return
		}
		switch n.mode {
		case modeOuter:
			callKind = callReplace
			err = n.renderOuter(pip)
		case modeInner:
			callKind = callUpdate
			err = n.renderInner(pip)
		case modeBlend:
			callKind = callReplace
			err = n.renderBlend(pip)
		case modeStatic:
			callKind = callReplace
			err = n.renderStatic(pip)
		default:
			panic("unknown node mode")
		}
	})
	callFrame := shredder.Join(true, thread.Frame(), n.tracker.outerCallGuard, task.CallFrame())
	defer callFrame.Release()
	callFrame.Run(ownerTracker.ctx, ownerTracker.root.runtime, func(b bool) {
		defer callGuard.Activate()
		if !b {
			task.Cancel()
			return
		}
		var payload printer.Payload
		if err == nil {
			payload, err = pip.Render(ownerTracker.root.inst.Conf().ServerDisableGzip)
		}
		callCtx := ownerTracker.ctx
		if err != nil {
			n.onErr(err)
			task.Report(err)
			payload = newError(err)
			callCtx = n.tracker.parent.ctx
		}
		ownerTracker.root.inst.Call(&call{
			ctx:     callCtx,
			kind:    callKind,
			id:      n.tracker.id,
			task:    task,
			payload: payload,
		})
	})
}

func (n *node) render(parentPipe *pipe, buffer *deque.Deque[any]) {
	thread := shredder.Thread{}
	ownerTracker := parentPipe.tracker
	renderFrame := shredder.Join(true, parentPipe.renderFrame, thread.Frame())
	if n.isMounted() {
		ownerTracker = n.tracker
		renderFrame = shredder.Join(true, renderFrame, n.tracker.writeFrame())
	}
	defer renderFrame.Release()
	pip := newPipe(
		ownerTracker,
		buffer,
		renderFrame,
		parentPipe.callGuard,
	)
	var err error
	pip.renderFrame.Submit(parentPipe.tracker.ctx, ownerTracker.root.runtime, func(b bool) {
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
		case modeStatic:
			err = n.renderStatic(pip)
		default:
			panic("unknown node mode")
		}
	})
	finalFrame := shredder.Join(true, parentPipe.renderFrame, thread.Frame())
	defer finalFrame.Release()
	finalFrame.Run(parentPipe.tracker.ctx, ownerTracker.root.runtime, func(b bool) {
		if !b {
			return
		}
		if err == nil {
			return
		}
		pip.error(err)
		n.onErr(err)
	})
}

func (n *node) renderStatic(pip *pipe) (err error) {
	cur := gox.NewCursor(pip.tracker.Context(), pip)
	return cur.Any(n.content)
}

func (n *node) renderBlend(pip *pipe) (err error) {
	printer := &nodePrinter{
		pipe: pip,
	}
	cur := gox.NewCursor(n.tracker.Context(), printer)
	err = n.outer(cur)
	if err != nil {
		return err
	}
	if pip.isEmpty() && n.content != nil && !n.tracker.isCanceled() {
		err = n.renderInner(pip)
	}
	if err != nil {
		return err
	}
	return printer.submitContainer()
}

func (n *node) renderOuter(pip *pipe) (err error) {
	printer := &nodePrinter{
		pipe: pip,
	}
	if n.outer != nil {
		cur := gox.NewCursor(n.tracker.Context(), printer)
		err = n.outer(cur)
	}
	if err != nil {
		return err
	}
	return printer.submitContainer()
}

func (n *node) renderInner(pip *pipe) (err error) {
	cur := gox.NewCursor(n.tracker.Context(), pip)
	return cur.Any(n.content)
}

func (n *node) renderInnerOuter(pip *pipe) (err error) {
	printer := &nodePrinter{
		pipe:        pip,
		skipContent: true,
	}
	if n.outer != nil {
		cur := gox.NewCursor(n.tracker.Context(), printer)
		err = n.outer(cur)
	}
	if err != nil {
		return err
	}
	if n.content != nil {
		err = n.renderInner(pip)
	}
	if err != nil {
		return err
	}
	return printer.submitContainer()
}

type nodePrinter struct {
	pipe        *pipe
	skipContent bool
	ready       bool
	open        *gox.JobHeadOpen
	close       *gox.JobHeadClose
}

func (r *nodePrinter) submitContainer() error {
	if r.open != nil && r.close == nil {
		return errors.New("door container tag was not closed")
	}
	ctx := r.pipe.tracker.container.Context()
	var openJob *gox.JobHeadOpen
	var closeJob *gox.JobHeadClose
	if r.open != nil && r.open.Kind == gox.KindContainer {
		gox.Release(r.open)
		gox.Release(r.close)
		r.open = nil
		r.close = nil
	}
	if r.open == nil {
		attrs := gox.NewAttrs()
		front.AttrsSetDoor(attrs, r.pipe.tracker.id, true)
		front.AttrsSetParent(attrs, r.pipe.tracker.parent.id)
		openJob = gox.NewJobHeadOpen(ctx, 0, gox.KindRegular, "d0-r", attrs)
		closeJob = gox.NewJobHeadClose(ctx, 0, gox.KindRegular, "d0-r")
	} else {
		r.open.Ctx = ctx
		r.close.Ctx = ctx
		front.AttrsSetDoor(r.open.Attrs, r.pipe.tracker.id, false)
		front.AttrsSetParent(r.open.Attrs, r.pipe.tracker.parent.id)
		openJob = r.open
		closeJob = r.close
		r.open = nil
		r.close = nil
	}
	if err := r.pipe.presend(openJob); err != nil {
		return err
	}
	if err := r.pipe.Send(closeJob); err != nil {
		return err
	}
	return nil
}

func (r *nodePrinter) pipeSend(job gox.Job) error {
	if r.skipContent {
		if rel, ok := job.(gox.Releaser); ok {
			gox.Release(rel)
		}
		return nil
	}
	return r.pipe.Send(job)
}

func (r *nodePrinter) pipePresend(job *gox.JobHeadOpen) error {
	if r.skipContent {
		gox.Release(job)
		return nil
	}
	return r.pipe.presend(job)
}

func (r *nodePrinter) Send(job gox.Job) error {
	if !r.ready {
		r.ready = true
		return r.init(job)
	}
	if r.open == nil {
		return r.pipeSend(job)
	}
	if r.close != nil {
		openJob := r.open
		closeJob := r.close
		r.open = nil
		r.close = nil
		if err := r.pipePresend(openJob); err != nil {
			return err
		}
		if err := r.Send(closeJob); err != nil {
			return err
		}
		return r.pipeSend(job)
	}
	if closeJob, ok := job.(*gox.JobHeadClose); ok {
		if closeJob.ID == r.open.ID {
			r.close = closeJob
			return nil
		}
	}
	return r.pipeSend(job)
}

func (r *nodePrinter) init(job gox.Job) error {
	openJob, isOpen := job.(*gox.JobHeadOpen)
	if !isOpen {
		return r.pipeSend(job)
	}
	if openJob.Kind == gox.KindVoid {
		return r.pipeSend(job)
	}
	if openJob.Kind == gox.KindRegular {
		if strings.EqualFold(openJob.Tag, "head") {
			return errors.New("door does not support <head> as a container")
		}
		if strings.EqualFold(openJob.Tag, "title") {
			return r.pipeSend(job)
		}
		if strings.EqualFold(openJob.Tag, "script") {
			return r.pipeSend(job)
		}
		if strings.EqualFold(openJob.Tag, "style") {
			return r.pipeSend(job)
		}
		if openJob.Tag == "d0-r" {
			return r.pipeSend(job)
		}
		if openJob.Attrs.Has("data-d0c") {
			return r.pipeSend(job)
		}
		if openJob.Attrs.Has("data-d0r") {
			return r.pipeSend(job)
		}
		if openJob.Tag == "" {
			return r.pipeSend(job)
		}
	}
	r.open = openJob
	return nil
}
