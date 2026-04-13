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

// AHook exposes a named Go handler to the browser through `$hook(name, ...)`.
//
// Input data is unmarshaled from JSON into type T.
// The returned value is encoded back to JSON.
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
	On func(ctx context.Context, r RequestHook[T]) (any, bool)
}

func (h AHook[T]) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(h, cur, elem)
}

func (h AHook[T]) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
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
		slog.Error("Hook decoding error", "error", err)
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
		slog.Error("Hook output encoding error", "error", err)
		w.WriteHeader(500)
	}
	return done

}

// ARawHook exposes a named Go handler to the browser through `$hook(name, ...)`
// without JSON decoding or encoding.
//
// Use it for streaming, custom protocols, or file uploads.
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
	On func(ctx context.Context, r RequestRawHook) bool
}

func (h ARawHook) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(h, cur, elem)
}

func (h ARawHook) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
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

// AData exposes Value to browser code through `$data(name)`.
//
// `$data(...)` returns strings and JSON-backed values directly. For `[]byte`,
// it returns a promise that resolves to an `ArrayBuffer`.
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

func (a AData) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	front.AttrsSetData(attrs, a.Name, a.Value)
	return nil
}
