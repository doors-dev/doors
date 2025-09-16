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
	"log/slog"
	"net/http"

	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
)

// AHook binds a backend handler to a named client-side hook, allowing
// JavaScript code to call Go functions via $d.hook(name, ...).
//
// Input data is unmarshaled from JSON into type T.
// Output data is marshaled to JSON from any.
//
// Generic parameters:
//   - T: input data type, sent from the client
type AHook[T any] struct {
	// Name of the hook to call from JavaScript via $d.hook(name, ...).
	// Required.
	Name string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Backend handler for the hook.
	// Receives typed input (T, unmarshaled from JSON) through RHook,
	// and returns any output which will be marshaled to JSON.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(ctx context.Context, r RHook[T]) (any, bool)
}

func (h AHook[T]) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, h)
}

func (h AHook[T]) Attr() AttrInit {
	return h
}

func (h AHook[T]) Init(ctx context.Context, n door.Core, inst instance.Core, attr *front.Attrs) {
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

func (h *AHook[T]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	var input T
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&input)
	r.Body.Close()
	if err != nil {
		println(err.Error())
		w.WriteHeader(400)
		return false
	}
	output, done := h.On(ctx, &formHookRequest[T]{
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
		slog.Error("Hook output encoding error", slog.String("json_error", err.Error()))
		println(err.Error())
		w.WriteHeader(500)
	}
	return done

}

// ARawHook binds a backend handler to a named client-side hook, allowing
// JavaScript code to call Go functions via $d.hook(name, ...).
//
// Unlike AHook, ARawHook does not perform JSON unmarshaling or marshaling.
// Instead, it gives full access to the raw request body and multipart form data,
// useful for streaming, custom parsing, or file uploads.
type ARawHook struct {
	// Name of the hook to call from JavaScript via $d.hook(name, ...).
	// Required.
	Name string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Backend handler for the hook.
	// Provides raw access via RRawHook (body reader, multipart parser).
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(ctx context.Context, r RRawHook) bool
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

// AData exposes server-provided data to JavaScript via $d.data(name).
//
// The Value is marshaled to JSON and made available for client-side access.
// This is useful for passing initial state, configuration, or constants
// directly into the client runtime.
type AData struct {
	// Name of the data entry to read via JavaScript with $d.data(name).
	// Required.
	Name string
	// Value to expose to the client. Marshaled to JSON.
	// Required.
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
