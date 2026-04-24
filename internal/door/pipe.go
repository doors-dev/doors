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

	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

func newPipe(
	tracker *tracker,
	buffer *deque.Deque[any],
	renderFrame shredder.Frame,
	callGuard *shredder.ValveFrame,
) *pipe {
	p := &pipe{
		tracker:     tracker,
		buffer:      buffer,
		renderFrame: renderFrame,
		callGuard:   callGuard,
	}
	p.printFront = printer.NewResourcePrinter((*pushFrontPrinter)(p.buffer))
	p.printBack = printer.NewResourcePrinter((*pushBackPrinter)(p.buffer))
	return p
}

type pipe struct {
	tracker     *tracker
	buffer      *deque.Deque[any]
	renderFrame shredder.Frame
	callGuard   *shredder.ValveFrame
	printFront  gox.Printer
	printBack   gox.Printer
}

func (p *pipe) isEmpty() bool {
	return p.buffer.Len() == 0
}

func (p *pipe) Collect() Stack {
	stack := Stack([]*deque.Deque[any]{p.buffer})
	p.buffer = nil
	return stack
}

func (p *pipe) Render(disableGzip bool) (printer.Payload, error) {
	stack := p.Collect()
	pr := printer.NewPayloadPrinter(disableGzip)
	err := stack.Print(pr)
	if err != nil {
		pr.Release()
		return nil, err
	}
	pr.Finalize()
	return pr, nil
}

func (p *pipe) error(err error) {
	p.buffer.Clear()
	p.buffer.PushBack(gox.NewJobComp(context.Background(), newError(err)))
}

func (p *pipe) branch() *deque.Deque[any] {
	buffer := new(deque.Deque[any])
	p.buffer.PushBack(buffer)
	return buffer
}

func (p *pipe) Submit(f func(cur gox.Cursor) error) {
	pip := newPipe(
		p.tracker,
		p.branch(),
		p.renderFrame,
		p.callGuard,
	)
	pip.renderFrame.Submit(p.tracker.ctx, p.tracker.Runtime(), func(b bool) {
		if !b {
			return
		}
		cur := gox.NewCursor(pip.tracker.Context(), pip)
		if err := f(cur); err != nil {
			pip.error(err)
		}
	})
}

func (p *pipe) presend(open *gox.JobHeadOpen) error {
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
	switch j := j.(type) {
	case renderer:
		j.Render(p)
		return nil
	case *gox.JobHeadOpen:
		if err := j.Attrs.ApplyMods(j.Ctx, j.Tag); err != nil {
			return err
		}
		return p.printBack.Send(j)
	case *gox.JobComp:
		ctx := j.Ctx
		comp := j.Comp
		gox.Release(j)
		el := comp.Main()
		if el == nil {
			return nil
		}
		cur := gox.NewCursor(ctx, p)
		return el(cur)
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
	for item := range next.IterPopFront() {
		switch item := item.(type) {
		case *deque.Deque[any]:
			p.push(item)
			goto cycle
		case gox.Job:
			if err := pr.Send(item); err != nil {
				return err
			}
		default:
			panic("unknown item type in the render buffer")
		}
	}
	p.pop()
	goto cycle
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
	(*p)[len(*p)-1] = nil
	*p = (*p)[:len(*p)-1]
}

func EmptyPayload() printer.Payload {
	return emptyPayload{}
}

type emptyPayload struct{}

func (e emptyPayload) Payload() action.Payload {
	return action.NewText("")
}

func (e emptyPayload) Release() {}

type pushFrontPrinter deque.Deque[any]

func (p *pushFrontPrinter) buf() *deque.Deque[any] {
	return (*deque.Deque[any])(p)
}

func (p *pushFrontPrinter) Send(j gox.Job) error {
	p.buf().PushFront(j)
	return nil
}

type pushBackPrinter deque.Deque[any]

func (p *pushBackPrinter) buf() *deque.Deque[any] {
	return (*deque.Deque[any])(p)
}

func (p *pushBackPrinter) Send(j gox.Job) error {
	p.buf().PushBack(j)
	return nil
}
