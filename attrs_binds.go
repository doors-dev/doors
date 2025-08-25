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
	"io"
	"net/http"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/door"
)

type AHook[I any, O any] struct {
	On        func(ctx context.Context, r RHook[I]) (O, bool)
	Name      string
	Scope     []Scope
	Indicator []Indicator
}

func (h AHook[I, O]) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, h)
}

func (h AHook[I, O]) Attr() AttrInit {
	return h
}

func (h AHook[I, O]) Init(ctx context.Context, n door.Core, inst instance.Core, attr *front.Attrs) {
	if h.On == nil {
		println("Hook withoud handler")
		return
	}
	entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
		Trigger: h.handle,
	})
	if !ok {
		return
	}
	attr.SetHook(h.Name, &front.Hook{
		Scope:     front.IntoScopeSet(inst, h.Scope),
		Indicate:  front.IntoIndicate(h.Indicator),
		HookEntry: entry,
	})
}

func (h *AHook[I, O]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	var input I
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&input)
	r.Body.Close()
	if err != nil {
		println(err.Error())
		w.WriteHeader(400)
		return false
	}
	output, done := h.On(ctx, &formHookRequest[I]{
		data: &input,
		request: request{
			w:   w,
			r:   r,
			ctx: ctx,
		},
	})
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&output)
	if err != nil {
		println(err.Error())
		w.WriteHeader(500)
	}
	return done

}

type ARawHook struct {
	Name      string
	On        func(ctx context.Context, r RRawHook) bool
	Scope     []Scope
	Indicator []Indicator
}

func (h ARawHook) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, h)
}

func (h ARawHook) Attr() AttrInit {
	return h
}

func (h ARawHook) Init(ctx context.Context, n door.Core, inst instance.Core, attr *front.Attrs) {
	if h.On == nil {
		println("Hook withoud handler")
		return
	}
	entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
		Trigger: h.handle,
	})
	if !ok {
		return
	}
	attr.SetHook(h.Name, &front.Hook{
		Scope:     front.IntoScopeSet(inst, h.Scope),
		Indicate:  front.IntoIndicate(h.Indicator),
		HookEntry: entry,
	})
}

func (h *ARawHook) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return h.On(ctx, &request{
		r:   r,
		w:   w,
		ctx: ctx,
	})
}

type AData struct {
	Name  string
	Value any
}

func (a AData) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, a)
}

func (a AData) Attr() AttrInit {
	return a
}

func (a AData) Init(_ context.Context, n door.Core, _ instance.Core, attr *front.Attrs) {
	attr.SetData(a.Name, a.Value)
}

type ADataMap map[string]any

func (dm ADataMap) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, dm)
}

func (dm ADataMap) Attr() AttrInit {
	return dm
}

func (dm ADataMap) Init(_ context.Context, n door.Core, _ instance.Core, attr *front.Attrs) {
	for name := range dm {
		attr.SetData(name, dm[name])
	}
}
