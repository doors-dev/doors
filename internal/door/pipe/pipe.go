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

package pipe

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type Pipe = *pipe

type Buffer = *deque.Deque[any]

type Branch = *atomic.Value

func NewPipe(
	ctx context.Context,
	runtime shredder.Runtime,
	renderFrame shredder.Frame,
	finalFrame *shredder.ValveFrame,
) Pipe {
	p := &pipe{
		runtime:     runtime,
		buffer:      &deque.Deque[any]{},
		renderFrame: renderFrame,
		finalFrame:  finalFrame,
	}
	p.printFront = printer.NewResourcePrinter(pushFrontPrinter{p.buffer})
	p.printBack = printer.NewResourcePrinter(pushBackPrinter{p.buffer})
	p.cursor = gox.NewCursor(ctx, p)
	return p
}

type pipe struct {
	runtime     shredder.Runtime
	buffer      Buffer
	renderFrame shredder.Frame
	finalFrame  *shredder.ValveFrame
	printFront  gox.Printer
	printBack   gox.Printer
	cursor      gox.Cursor
	err         error
}

func (p Pipe) Runtime() shredder.Runtime {
	return p.runtime
}

func (p Pipe) Context() context.Context {
	return p.cursor.Context()
}

func (p Pipe) RenderFrame() shredder.Frame {
	return p.renderFrame
}

func (p Pipe) FinalFrame() *shredder.ValveFrame {
	return p.finalFrame
}

func (p Pipe) Branch() Branch {
	v := &atomic.Value{}
	p.buffer.PushBack(v)
	return v
}

func (p Pipe) RenderProxy(el gox.Elem) (ProxyContainer, bool) {
	printer := &proxyPrinter{
		pip: p,
	}
	cursor := gox.NewCursor(p.cursor.Context(), printer)
	if err := el(cursor); err != nil {
		p.err = err
		return ProxyContainer{}, false
	}
	cont, err := printer.finalize()
	if err != nil {
		p.err = err
		return ProxyContainer{}, false
	}
	return cont, true
}

func (p Pipe) RenderContent(content any) {
	if comp, ok := content.(gox.Comp); ok {
		p.RenderComp(comp)
	} else {
		if err := p.cursor.Any(content); err != nil {
			p.err = err
		}
	}
}

func (p Pipe) RenderComp(comp gox.Comp) {
	el := comp.Main()
	if el == nil {
		return
	}
	if err := el(p.cursor); err != nil {
		p.err = err
	}
}

func (p *pipe) Payload(disableGzip bool) Payload {
	jobs, err := p.Collect()
	if err != nil {
		return NewError(err)
	}
	pr := printer.NewPayloadPrinter(disableGzip)
	err = jobs.Print(pr)
	if err != nil {
		pr.Release()
		return NewError(err)
	}
	pr.Finalize()
	return pr
}

func (p *pipe) Collect() (Stack, error) {
	if p.err != nil {
		return nil, p.err
	}
	j := stack([]*deque.Deque[any]{p.buffer})
	p.buffer = nil
	return &j, nil
}

func (p *pipe) Submit(b Branch) error {
	if p.err != nil {
		b.Store(p.err)
		return p.err
	}
	b.Store(p.buffer)
	p.buffer = nil
	return nil
}

func (p Pipe) IsEmpty() bool {
	return p.buffer.Len() == 0
}

type Renderable interface {
	Render(p *pipe)
}

func (p Pipe) FrameSubmit(f func(bool)) {
	p.renderFrame.Submit(p.Context(), p.runtime, f)
}

func (p *pipe) Send(j gox.Job) error {
	switch v := j.(type) {
	case *gox.JobHeadOpen:
		if err := v.Attrs.ApplyMods(v.Ctx, v.Tag); err != nil {
			return err
		}
		return p.printBack.Send(j)
	case Renderable:
		v.Render(p)
	case *gox.JobComp:
		comp := v.Comp
		ctx := v.Ctx
		gox.Release(v)
		branch := p.Branch()
		pip := NewPipe(ctx, p.runtime, p.renderFrame, p.finalFrame)
		defer pip.Submit(branch)
		pip.RenderComp(comp)
	default:
		return p.printBack.Send(j)
	}
	return nil
}
func (p *pipe) push(j gox.Job) {
	if v, ok := j.(*gox.JobHeadOpen); ok {
		if err := v.Attrs.ApplyMods(v.Ctx, v.Tag); err != nil {
			p.err = err
		}
	}
	if err := p.printBack.Send(j); err != nil {
		p.err = err
	}
}

func (p *pipe) unshift(j gox.Job) {
	if v, ok := j.(*gox.JobHeadOpen); ok {
		if err := v.Attrs.ApplyMods(v.Ctx, v.Tag); err != nil {
			p.err = err
		}
	}
	if err := p.printFront.Send(j); err != nil {
		p.err = err
	}
}

type pushFrontPrinter struct {
	buf Buffer
}

func (p pushFrontPrinter) Send(j gox.Job) error {
	p.buf.PushFront(j)
	return nil
}

type pushBackPrinter struct {
	buf Buffer
}

func (p pushBackPrinter) Send(j gox.Job) error {
	p.buf.PushBack(j)
	return nil
}
