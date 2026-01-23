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

// AHook binds a backend handler to a named client-side hook, allowing
// JavaScript code to call Go functions via $hook(name, ...).
//
// Input data is unmarshaled from JSON into type T.
// Output data is marshaled to JSON from any.
//
// Generic parameters:
//   - T: input data type, sent from the client
type AHook[T any] struct {
	// Name of the hook to call from JavaScript via $hook(name, ...).
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

func (h AHook[T]) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(h, cur, elem)
}

func (h AHook[T]) Apply(ctx context.Context, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(h.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsSetHook(attrs, h.Name, front.Hook{
		Scope:    front.IntoScopeSet(core, h.Scope),
		Indicate: front.IntoIndicate(h.Indicator),
		Hook:     hook,
	})
	return nil
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
// JavaScript code to call Go functions via $hook(name, ...).
//
// Unlike AHook, ARawHook does not perform JSON unmarshaling or marshaling.
// Instead, it gives full access to the raw request body and multipart form data,
// useful for streaming, custom parsing, or file uploads.
type ARawHook struct {
	// Name of the hook to call from JavaScript via $hook(name, ...).
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

func (h ARawHook) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(h, cur, elem)
}

func (h ARawHook) Apply(ctx context.Context, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(h.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsSetHook(attrs, h.Name, front.Hook{
		Scope:    front.IntoScopeSet(core, h.Scope),
		Indicate: front.IntoIndicate(h.Indicator),
		Hook:     hook,
	})
	return nil
}

func (h *ARawHook) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return h.On(ctx, &request{
		r:   r,
		w:   w,
		ctx: ctx,
	})
}

// AData exposes server-provided data to JavaScript via $data(name).
//
// The Value is marshaled to JSON and made available for client-side access.
// This is useful for passing initial state, configuration, or constants
// directly into the client runtime.
type AData struct {
	// Name of the data entry to read via JavaScript with $data(name).
	// Required.
	Name string
	// Value to expose to the client. Marshaled to JSON.
	// Required.
	Value any
}

func (a AData) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(a, cur, elem)
}

func (a AData) Apply(ctx context.Context, attrs gox.Attrs) error {
	front.AttrsSetData(attrs, a.Name, a.Value)
	return nil
}

type ADataMap map[string]any

func (dm ADataMap) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(dm, cur, elem)
}

func (dm ADataMap) Apply(ctx context.Context, attrs gox.Attrs) error {
	for name, value := range dm {
		front.AttrsSetData(attrs, name, value)
	}
	return nil
}
