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
	"log/slog"
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

func (j joinedAttrs) Apply(ctx context.Context, attrs gox.Attrs) error {
	attrs.Inherit(j.attrs)
	return nil
}

// Attr is a doors attribute modifier.
type Attr interface {
	gox.AttrMod
	gox.Proxy
}

func AJoin(ctx context.Context, a ...Attr) Attr {
	 attrs := gox.NewAttrs(ctx)
	 for _, mod := range a {
	 	attrs.AddMod(mod)
	 }
	 attrs.ApplyMods()
	 return joinedAttrs{attrs: attrs}
}

type eventAttr[E any] struct {
	capture   front.Capture
	onError   []Action
	before    []Action
	scope     []Scope
	indicator []Indicator
	on        func(context.Context, REvent[E]) bool
}

func (p eventAttr[E]) apply(ctx context.Context, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(p.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, p.capture, front.Hook{
		OnError:  intoActions(ctx, p.onError),
		Before:   intoActions(ctx, p.before),
		Scope:    front.IntoScopeSet(core, p.scope),
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
	mod  gox.AttrMod
	cur  gox.Cursor
	elem gox.Elem
}

func (m *proxyAttrModPrinter) Send(job gox.Job) error {
	if m.mod == nil {
		return m.cur.Job(job)
	}
	mod := m.mod
	m.mod = nil
	open, ok := job.(*gox.JobHeadOpen)
	if !ok || open.Kind == gox.KindContainer || open.Tag == "d0-0r" || open.Attrs.Get("data-d00r").IsSet() {
		slog.Error("Can't attach attribute modifer on non-tag or door")
	} else {
		open.Attrs.AddMod(mod)
	}
	return m.Send(job)
}

func proxyAddAttrMod(mod gox.AttrMod, cur gox.Cursor, elem gox.Elem) error {
	printer := &proxyAttrModPrinter{
		mod:  mod,
		cur:  cur,
		elem: elem,
	}
	return elem.Print(cur.Context(), printer)
}
