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
	"bytes"
	"context"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

func newPipe(
	tracker *tracker,
	buffer *deque.Deque[any],
	renderFrame shredder.ReleaseFrame,
	callGuard *shredder.ValveFrame,
	document bool,
) *pipe {
	p := &pipe{
		tracker:     tracker,
		buffer:      buffer,
		renderFrame: renderFrame,
		callGuard:   callGuard,
		document:    document,
		ctx:         tracker.renderCtx,
		outerCtx:    tracker.outer.renderCtx,
	}
	if document {
		p.ctx = common.DocumentCtx(p.ctx)
		p.outerCtx = common.DocumentCtx(p.outerCtx)
	}
	p.printFront = printer.NewResourcePrinter((*pushFrontPrinter)(p.buffer))
	p.printBack = printer.NewResourcePrinter((*pushBackPrinter)(p.buffer))
	return p
}

type pipe struct {
	document    bool
	ctx         context.Context
	outerCtx    context.Context
	tracker     *tracker
	buffer      *deque.Deque[any]
	renderFrame shredder.ReleaseFrame
	callGuard   *shredder.ValveFrame
	printFront  gox.Printer
	printBack   gox.Printer
}

func (p *pipe) id() uint64 {
	return p.tracker.outer.id
}

func (p *pipe) parentID() uint64 {
	return p.tracker.outer.parent.outer.id
}

func (p *pipe) innerContext() context.Context {
	return p.ctx
}

func (p *pipe) outerContext() context.Context {
	return p.outerCtx
}

func (p *pipe) runtime() shredder.Runtime {
	return p.tracker.Runtime()
}

func (p *pipe) isEmpty() bool {
	return p.buffer.Len() == 0
}

func (p *pipe) Collect() Stack {
	stack := Stack([]*deque.Deque[any]{p.buffer})
	p.buffer = nil
	return stack
}

func (p *pipe) Release() {
	stack := p.Collect()
	stack.Release()
}

func (p *pipe) Render(pr *printer.PayloadPrinter, printerMiddleware func(next gox.Printer) gox.Printer) (printer.Payload, error) {
	stack := p.Collect()
	err := stack.Print(printerMiddleware(pr))
	if err != nil {
		pr.Release()
		return nil, err
	}
	pr.Finalize()
	return pr, nil
}

func (p *pipe) error(err error, caller common.Caller) {
	p.buffer.Clear()
	logError(p.tracker.ctx, common.Logger(p.tracker.ctx), err, caller)
}

func (p *pipe) branch() *deque.Deque[any] {
	buffer := common.GetDequeBuffer()
	p.buffer.PushBack(buffer)
	return buffer
}

func (p *pipe) fork(buffer *deque.Deque[any]) *pipe {
	pip := *p
	pip.buffer = buffer
	pip.printFront = printer.NewResourcePrinter((*pushFrontPrinter)(buffer))
	pip.printBack = printer.NewResourcePrinter((*pushBackPrinter)(buffer))
	return &pip
}

func (p *pipe) Submit(caller common.Caller, f func(cur gox.Cursor) error) {
	pip := p.fork(p.branch())
	pip.renderFrame.Submit(p.tracker.ctx, p.tracker.Runtime(), func(b bool) {
		if !b {
			return
		}
		cur := gox.NewCursor(pip.innerContext(), pip)
		if err := f(cur); err != nil {
			pip.error(err, caller)
		}
	})
}

func (p *pipe) presend(open *gox.JobOpen) error {
	if err := open.Attrs.ApplyMods(open.Ctx, open.Tag); err != nil {
		return err
	}
	return p.printFront.Send(open)
}

type Pipe = *pipe

type renderer interface {
	Render(p Pipe)
}

func (p *pipe) Send(j gox.Job) error {
	if j.Context().Err() != nil {
		return j.Context().Err()
	}
	switch j := j.(type) {
	case renderer:
		j.Render(p)
		return nil
	case *gox.JobOpen:
		if err := j.Attrs.ApplyMods(j.Ctx, j.Tag); err != nil {
			return err
		}
		return p.printBack.Send(j)
	case *gox.JobTempl:
		ctx := j.Ctx
		var buf bytes.Buffer
		if err := j.Output(&buf); err != nil {
			return err
		}
		return p.printBack.Send(gox.NewJobBytes(ctx, buf.Bytes()))
	default:
		return p.printBack.Send(j)
	}
}

type Stack []*deque.Deque[any]

func (p *Stack) Print(pr gox.Printer) error {
cycle:
	next := p.next()
	if next == nil {
		return nil
	}
	common.FreezeDequeBuffer(next)
	var err error
loop:
	for item := range next.IterPopFront() {
		switch item := item.(type) {
		case *deque.Deque[any]:
			p.push(item)
			goto cycle
		case gox.Job:
			if err = pr.Send(item); err != nil {
				break loop
			}
		default:
			panic("unknown item type in the render buffer")
		}
	}
	if err != nil {
		p.Release()
		return err
	}
	p.pop()
	goto cycle
}

func (p *Stack) Release() {
	for i, buffer := range *p {
		common.ReleaseDequeBuffer(buffer)
		(*p)[i] = nil
	}
	*p = nil
}

func (p Stack) next() *deque.Deque[any] {
	if len(p) == 0 {
		return nil
	}
	return p[len(p)-1]
}

func (p *Stack) push(buf *deque.Deque[any]) {
	*p = append(*p, buf)
}

func (p *Stack) pop() {
	index := len(*p) - 1
	buffer := (*p)[index]
	(*p)[index] = nil
	*p = (*p)[:index]
	common.PutDequeBuffer(buffer)
}

func EmptyPayload() printer.Payload {
	return emptyPayload{}
}

type emptyPayload struct{}

func (e emptyPayload) Payload() (actions.Payload, bool) {
	return actions.NewText(""), true
}

func (e emptyPayload) Release() {}

func (e emptyPayload) Free() {}

func (e emptyPayload) Lock() bool {
	return true
}

type pushFrontPrinter deque.Deque[any]

func (p *pushFrontPrinter) buf() *deque.Deque[any] {
	return (*deque.Deque[any])(p)
}

func (p *pushFrontPrinter) Send(j gox.Job) error {
	if j.Context().Err() != nil {
		return j.Context().Err()
	}
	p.buf().PushFront(j)
	return nil
}

type pushBackPrinter deque.Deque[any]

func (p *pushBackPrinter) buf() *deque.Deque[any] {
	return (*deque.Deque[any])(p)
}

func (p *pushBackPrinter) Send(j gox.Job) error {
	if j.Context().Err() != nil {
		return j.Context().Err()
	}
	p.buf().PushBack(j)
	return nil
}
