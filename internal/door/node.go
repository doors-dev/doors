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

	"github.com/doors-dev/doors/internal/common"
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
	static  *staticTracker
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
	if n.static != nil {
		n.static.cancel()
	}
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
		n.static = newStaticTracker(ownerTracker, callGuard)
	}
	renderFrame := shredder.Join(ownerTracker.Context(), true, thread.Frame(), ownerTracker.writeFrame(ownerTracker.Context()), task.RenderFrame())
	defer renderFrame.Release()
	pip := newPipe(
		ownerTracker,
		common.GetDequeBuffer(),
		renderFrame,
		callGuard,
	)
	var err error
	var callKind callKind
	pip.renderFrame.Submit(ownerTracker.ctx, ownerTracker.root.runtime(), func(b bool) {
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
			err = n.renderStatic(pip, n.static.Context())
		default:
			panic("unknown node mode")
		}
	})
	callFrame := shredder.Join(ownerTracker.Context(), true, thread.Frame(), n.tracker.outerCallGuard, task.CallFrame())
	defer callFrame.Release()
	callFrame.Run(ownerTracker.ctx, ownerTracker.root.runtime(), func(b bool) {
		defer callGuard.Activate()
		if !b {
			pip.Release()
			task.Cancel()
			return
		}
		var payload printer.Payload
		if err == nil {
			app := ownerTracker.Instance().Session().App()
			payload, err = pip.Render(app.Conf().ServerDisableGzip, app.PrinterMiddleware())
		}
		logger := ownerTracker.root.inst.Logger()
		callCtx := ownerTracker.ctx
		if err != nil {
			n.onErr(err)
			task.Report(err)
			payload = newError(err, logger)
			callCtx = n.tracker.parent.ctx
		} else {
			task.Scheduled()
		}
		ownerTracker.root.inst.Call(&call{
			ctx:     callCtx,
			kind:    callKind,
			id:      n.tracker.id,
			task:    task,
			payload: payload,
			logger:  logger,
		})
	})
}

func (n *node) render(parentPipe *pipe, buffer *deque.Deque[any]) {
	thread := shredder.Thread{}
	ownerTracker := parentPipe.tracker
	renderFrame := shredder.Join(parentPipe.tracker.Context(), true, parentPipe.renderFrame, thread.Frame())
	if n.isMounted() {
		ownerTracker = n.tracker
		renderFrame = shredder.Join(ownerTracker.Context(), true, renderFrame, n.tracker.writeFrame(ownerTracker.Context()))
	}
	defer renderFrame.Release()
	pip := newPipe(
		ownerTracker,
		buffer,
		renderFrame,
		parentPipe.callGuard,
	)
	var err error
	pip.renderFrame.Submit(parentPipe.tracker.ctx, ownerTracker.root.runtime(), func(b bool) {
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
			err = n.renderStatic(pip, pip.tracker.Context())
		default:
			panic("unknown node mode")
		}
	})
	finalFrame := shredder.Join(parentPipe.tracker.Context(), true, parentPipe.renderFrame, thread.Frame())
	defer finalFrame.Release()
	finalFrame.Run(parentPipe.tracker.ctx, ownerTracker.root.runtime(), func(b bool) {
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

func (n *node) renderStatic(pip *pipe, ctx context.Context) (err error) {
	cur := gox.NewCursor(ctx, pip)
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
	switch job := job.(type) {
	case *gox.JobComp:
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		el := comp.Main()
		if el == nil {
			r.ready = true
			return nil
		}
		cur := gox.NewCursor(ctx, r)
		return el(cur)
	case *gox.JobHeadOpen:
		r.ready = true
		return r.initOpenJob(job)
	default:
		r.ready = true
		return r.pipeSend(job)
	}
}

func (r *nodePrinter) initOpenJob(openJob *gox.JobHeadOpen) error {
	switch openJob.Kind {
	case gox.KindRegular:
		if strings.EqualFold(openJob.Tag, "head") {
			return errors.New("door does not support <head> as a container")
		}
		if strings.EqualFold(openJob.Tag, "title") {
			return r.pipeSend(openJob)
		}
		if strings.EqualFold(openJob.Tag, "script") {
			return r.pipeSend(openJob)
		}
		if strings.EqualFold(openJob.Tag, "style") {
			return r.pipeSend(openJob)
		}
		if openJob.Tag == "d0-r" {
			return r.pipeSend(openJob)
		}
		if openJob.Attrs.Has("data-d0c") {
			return r.pipeSend(openJob)
		}
		if openJob.Attrs.Has("data-d0r") {
			return r.pipeSend(openJob)
		}
		if openJob.Tag == "" {
			return r.pipeSend(openJob)
		}
		r.open = openJob
		return nil
	case gox.KindContainer:
		r.open = openJob
		return nil
	case gox.KindVoid:
		return r.pipeSend(openJob)
	default:
		panic("unknown gox head kind")
	}
}
