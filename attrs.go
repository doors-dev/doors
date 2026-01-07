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
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
)

type AttrInit = front.Attr

type Attr interface {
	Attr() AttrInit
	templ.Component
}

// Attrs is a renderable and spreadable set of HTML attributes.
type Attrs = front.Attrs

// A builds an attribute set from the given attributes.
// It can be spread in templ tags or rendered inline before an element.
func A(ctx context.Context, a ...Attr) *Attrs {
	ar := make([]front.Attr, len(a))
	for i, attr := range a {
		ar[i] = attr.Attr()
	}
	return front.A(ctx, ar...)
}

func AClass(class ...string) AOne {
	return AOne{"class", strings.Join(class, " ")}
}

type AOne [2]string

func (a AOne) Init(_ context.Context, _ door.Core, _ instance.Core, attrs *front.Attrs) {
	attrs.Set(a[0], a[1])
}

func (a AOne) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, a)
}

func (a AOne) Attr() AttrInit {
	return a
}

type AMap map[string]any

func (a AMap) Init(_ context.Context, _ door.Core, _ instance.Core, attrs *front.Attrs) {
	attrs.SetRaw(templ.Attributes(a))
}

func (a AMap) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, a)
}

func (a AMap) Attr() AttrInit {
	return a
}

type eventAttr[E any] struct {
	door      door.Core
	ctx       context.Context
	capture   front.Capture
	onError   []Action
	before    []Action
	scope     []Scope
	indicator []Indicator
	inst      instance.Core
	on        func(context.Context, REvent[E]) bool
}

func (p *eventAttr[E]) init(attrs *front.Attrs) {
	entry, ok := p.door.RegisterAttrHook(p.ctx, &door.AttrHook{
		Trigger: p.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(p.capture, &front.Hook{
		OnError:   intoActions(p.ctx, p.onError),
		Before:    intoActions(p.ctx, p.before),
		Scope:     front.IntoScopeSet(p.inst, p.scope),
		Indicate:  front.IntoIndicate(p.indicator),
		HookEntry: entry,
	})
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
