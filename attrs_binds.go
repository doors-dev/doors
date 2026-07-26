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
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

// AHook exposes a Go handler to browser code as $hook(name, ...).
//
// Attach it to the script element that calls $hook or $fetch. Input is decoded
// from JSON into T, and the value On returns is encoded back to JSON as the
// call result.
//
// Prefer [ARawHook] when the handler needs the raw body, multipart parsing, or
// control over the response.
type AHook[T any] struct {
	// Name is the hook name used from JavaScript as $hook(name, ...). Required.
	Name string
	// Scope controls how calls to this hook are scheduled on the client.
	// Optional; without it every call is sent immediately.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while a call is in flight.
	// A hook call has no event element, so only the Query and QueryAll variants
	// apply. Optional.
	Indicator Indicators
	// On handles the call and returns the value sent back to the caller. Return
	// false to keep the hook active, true to remove it. Optional; a nil On
	// answers with null.
	On func(ctx context.Context, r RequestHook[T]) (any, bool)
}

func (h AHook[T]) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(h, cur, elem)
}

func (h AHook[T]) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(h.handle(core), nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsSetHook(attrs, h.Name, front.Hook{
		Scope:    scopesOrNil(core, h.Scope),
		Indicate: indicatorsOrNil(h.Indicator),
		Hook:     hook,
	})
	return nil
}

func (h *AHook[T]) handle(core core.Core) func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, int64(core.App().Conf().ServerRequestBodyLimit)))
		var input T
		err := dec.Decode(&input)
		r.Body.Close()
		if err != nil {
			common.Logger(ctx).Error("Hook decoding error", "error", err)
			w.WriteHeader(400)
			return false
		}
		var output any
		var done bool
		if h.On != nil {
			output, done = h.On(ctx, &formHookRequest[T]{
				data: &input,
				request: request{
					w:   w,
					r:   r,
					ctx: ctx,
				},
			})
		}
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		err = enc.Encode(&output)
		if err != nil {
			common.Logger(ctx).Error("Hook output encoding error", "error", err)
			w.WriteHeader(500)
		}
		return done
	}

}

// ARawHook exposes a Go handler to browser code as $hook(name, ...) without
// JSON decoding or encoding.
//
// Attach it to the script element that calls $hook or $fetch. The handler owns
// the request body and the response, which suits streaming, uploads, and
// custom formats. Prefer [AHook] for plain JSON input and output.
type ARawHook struct {
	// Name is the hook name used from JavaScript as $hook(name, ...). Required.
	Name string
	// Scope controls how calls to this hook are scheduled on the client.
	// Optional; without it every call is sent immediately.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while a call is in flight.
	// A hook call has no event element, so only the Query and QueryAll variants
	// apply. Optional.
	Indicator Indicators
	// On handles the call with raw access to the request body and the response.
	// Return false to keep the hook active, true to remove it. Optional; a nil
	// On answers with an empty response.
	On func(ctx context.Context, r RequestRawHook) bool
}

func (h ARawHook) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(h, cur, elem)
}

func (h ARawHook) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(h.handle(core), nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsSetHook(attrs, h.Name, front.Hook{
		Scope:    scopesOrNil(core, h.Scope),
		Indicate: indicatorsOrNil(h.Indicator),
		Hook:     hook,
	})
	return nil
}

func (h *ARawHook) handle(core core.Core) func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		if h.On == nil {
			r.Body.Close()
			return false
		}
		return h.On(ctx, &request{
			r:     r,
			w:     w,
			ctx:   ctx,
			limit: core.App().Conf().ServerRequestBodyLimit,
		})
	}
}

// AData exposes a value to browser code as $data(name).
//
// Attach it to the script element that reads the value. A []byte value arrives
// as a promise resolving to an ArrayBuffer; other values arrive directly.
type AData struct {
	// Name is the entry name used from JavaScript as $data(name). Required.
	Name string
	// Value is the exposed value: a string is sent as text, []byte as binary,
	// anything else as JSON. Required.
	Value any
}

func (a AData) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(a, cur, elem)
}

func (a AData) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	front.AttrsSetData(attrs, a.Name, a.Value)
	return nil
}
