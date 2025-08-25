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

// A constructs a set of HTML attributes.
//
// These attributes enable backend-connected interactivity — such as pointer events,
// data binding, and hook-based logic — by wiring frontend behavior to Go code via context.
//
// `A` is typically used inside HTML tags to attach event handlers.
//
// It should be passed within an attribute block and spread into the element using `...`.
//
// Example:
//
//	<button { doors.A(ctx, doors.AClick{
//	    On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
//	        log.Println("Clicked")
//	        return false
//	    },
//	})... }>
//	    Click Me
//	</button>
//
// Parameters:
//   - ctx: the current rendering context. It is used to bind interactive behavior
//     to the component’s lifecycle and scope.
//   - attrs: a list of special Attribute values (e.g., AClick, AHook, ABind).
//
// Returns:
//   - A templ.Attributes object that can be spread into a templ element.

type Attrs = front.Attrs

func A(ctx context.Context, a ...Attr) *Attrs {
	ar := make([]front.Attr, len(a))
	for i, attr := range a {
		ar[i] = attr.Attr()
	}
	return front.A(ctx, ar...)
}

/*
type ARaw templ.Attributes

func (s ARaw) Attr() AttrInit {
	return s
}

func (s ARaw) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s ARaw) Init(ctx context.Context, _ door.Core, _ instance.Core, attrs *front.Attrs) {
	if s == nil {
		return
	}
	attrs.SetRaw(templ.Attributes(s))
}*/

type eventAttr[E any] struct {
	door      door.Core
	ctx       context.Context
	capture   front.Capture
	onError   []OnError
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
		Error:     front.IntoErrorAction(p.onError),
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
