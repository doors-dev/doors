// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type joinedAttrs struct {
	attrs gox.Attrs
}

func (j joinedAttrs) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(j, cur, elem)
}

func (j joinedAttrs) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	attrs.Inherit(j.attrs)
	return nil
}

// Attr is a Doors attribute modifier that can be attached directly to an
// element or applied through a proxy component.
type Attr interface {
	gox.Modify
	gox.Proxy
}

// A combines one or more [Attr] values into a single modifier.
//
// Example:
//
//	attrs := doors.A(ctx,
//		doors.AClick{On: onClick},
//		doors.AData{Name: "user", Value: user},
//	)
func A(ctx context.Context, a ...Attr) Attr {
	attrs := gox.NewAttrs()
	for _, mod := range a {
		attrs.AddMod(mod)
	}
	attrs.ApplyMods(ctx, "")
	return joinedAttrs{attrs: attrs}
}

type eventAttr[E any] struct {
	capture   front.Capture
	onError   []Action
	before    []Action
	scope     []Scope
	indicator []Indicator
	on        func(context.Context, RequestEvent[E]) bool
}

func (p eventAttr[E]) apply(ctx context.Context, attrs gox.Attrs) error {
	c := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := c.RegisterHook(p.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, p.capture, front.Hook{
		OnError:  intoActions(ctx, p.onError),
		Before:   intoActions(ctx, p.before),
		Scope:    front.IntoScopeSet(c, p.scope),
		Indicate: front.IntoIndicate(p.indicator),
		Hook:     hook,
	})
	return nil
}

func (p *eventAttr[E]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	var e E
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&e)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(400)
		return false
	}
	return p.on(ctx, &eventRequest[E]{
		request: request{
			r:   r,
			w:   w,
			ctx: ctx,
		},
		e: &e,
	})
}

type proxyAttrModPrinter struct {
	mod gox.Modify
	cur gox.Cursor
}

func (m *proxyAttrModPrinter) Send(job gox.Job) error {
	if m.mod == nil {
		return m.cur.Send(job)
	}
	comp, ok := job.(*gox.JobComp)
	if ok {
		mod := m.mod
		m.mod = nil
		return m.printComp(mod, comp)
	}
	open, ok := job.(*gox.JobHeadOpen)
	if ok {
		if open.Kind == gox.KindContainer {
			return m.cur.Send(open)
		}
		mod := m.mod
		m.mod = nil
		return m.printHead(mod, open)
	}
	m.mod = nil
	return errors.New("Can't attach attribute modifer - unexpected job type")
}

func (m *proxyAttrModPrinter) printHead(mod gox.Modify, job *gox.JobHeadOpen) error {
	if job.Tag == "d0-r" {
		return errors.New("Can't attach attribute modifer on door container")
	}
	job.Attrs.AddMod(mod)
	return m.cur.Send(job)
}

func (m *proxyAttrModPrinter) printComp(mod gox.Modify, job *gox.JobComp) error {
	ctx := job.Ctx
	comp := job.Comp
	return m.cur.CompCtx(ctx, gox.Elem(func(cur gox.Cursor) error {
		elem := comp.Main()
		if elem == nil {
			return nil
		}
		printer := &proxyAttrModPrinter{
			mod: mod,
			cur: cur,
		}
		return elem.Print(cur.Context(), printer)
	}))
}

func proxyAddAttrMod(mod gox.Modify, cur gox.Cursor, elem gox.Elem) error {
	printer := &proxyAttrModPrinter{
		mod: mod,
		cur: cur,
	}
	return elem.Print(cur.Context(), printer)
}
