// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
	p.cursor = gox.NewCursor(ctx, p)
	return p
}

type pipe struct {
	runtime     shredder.Runtime
	buffer      Buffer
	renderFrame shredder.Frame
	finalFrame  *shredder.ValveFrame
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
	if err := el.Print(p.cursor.Context(), printer); err != nil {
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
		if el := comp.Main(); el != nil {
			if err := el(p.cursor); err != nil {
				p.err = err
			}
		}
	} else {
		if err := p.cursor.Any(content); err != nil {
			p.err = err
		}
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

func (p *pipe) Send(j gox.Job) error {
	switch v := j.(type) {
	case *gox.JobHeadOpen:
		if err := v.Attrs.ApplyMods(v.Ctx, v.Tag); err != nil {
			return err
		}
		p.buffer.PushBack(j)
	case Renderable:
		v.Render(p)
	case *gox.JobComp:
		comp := v.Comp
		ctx := v.Ctx
		gox.Release(v)
		branch := p.Branch()
		pip := NewPipe(ctx, p.runtime, p.renderFrame, p.finalFrame)
		p.renderFrame.Submit(ctx, p.runtime, func(b bool) {
			if !b {
				return
			}
			defer pip.Submit(branch)
			pip.RenderContent(comp)
		})
	default:
		p.buffer.PushBack(j)
	}
	return nil
}

func (p *pipe) push(j gox.Job) {
	if v, ok := j.(*gox.JobHeadOpen); ok {
		if err := v.Attrs.ApplyMods(v.Ctx, v.Tag); err != nil {
			p.err = err
		}
	}
	p.buffer.PushBack(j)
}

func (p *pipe) unshift(j gox.Job) {
	if v, ok := j.(*gox.JobHeadOpen); ok {
		if err := v.Attrs.ApplyMods(v.Ctx, v.Tag); err != nil {
			p.err = err
		}
	}
	p.buffer.PushFront(j)
}
