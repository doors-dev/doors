// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"context"
	"encoding/json"
	"net/http"

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
