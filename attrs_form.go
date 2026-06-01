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
	"errors"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
	"github.com/go-playground/form/v4"
)

// ARawSubmit handles a form submission with raw multipart access.
type ARawSubmit struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Actions to run before the hook request.
	// Optional.
	Before Actions
	// Backend form handler.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestRawForm) bool
	// Actions to run on error.
	// Optional.
	OnError Actions
}

func (s ARawSubmit) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(s, cur, elem)
}

func (s ARawSubmit) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(s.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, front.FormCapture{}, front.Hook{
		OnError:  intoActions(ctx, actionsOrNil(s.OnError)),
		Before:   intoActions(ctx, actionsOrNil(s.Before)),
		Scope:    scopesOrNil(core, s.Scope),
		Indicate: indicatorsOrNil(s.Indicator),
		Hook:     hook,
	})
	return nil
}

func (s *ARawSubmit) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	done := s.On(ctx, request{
		w:   w,
		r:   r,
		ctx: ctx,
	})
	return done
}

var formDecoder *form.Decoder

func init() {
	formDecoder = form.NewDecoder()
}

// ASubmit handles a form submission by decoding it into T with
// go-playground/form.
type ASubmit[T any] struct {
	// MaxMemory sets the maximum number of bytes to parse into memory.
	// It is passed to ParseMultipartForm.
	// Defaults to 8 MB if zero.
	MaxMemory int
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Actions to run before the hook request.
	// Optional.
	Before Actions
	// Backend form handler.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestForm[T]) bool
	// Actions to run on error.
	// Optional.
	OnError Actions
}

func (s ASubmit[V]) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(s, cur, elem)
}

func (s ASubmit[V]) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(s.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, front.FormCapture{}, front.Hook{
		OnError:  intoActions(ctx, actionsOrNil(s.OnError)),
		Before:   intoActions(ctx, actionsOrNil(s.Before)),
		Scope:    scopesOrNil(core, s.Scope),
		Indicate: indicatorsOrNil(s.Indicator),
		Hook:     hook,
	})
	return nil
}

const defaultMaxMemory = 8 << 20

func (s *ASubmit[V]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	maxMemory := defaultMaxMemory
	if s.MaxMemory > 0 {
		maxMemory = s.MaxMemory
	}
	err := r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		w.Write([]byte("Multipart form parsing error"))
		w.WriteHeader(400)
		return false
	}
	var v V
	err = formDecoder.Decode(&v, r.Form)
	if err != nil {
		w.Write([]byte("Form decoding error"))
		w.WriteHeader(400)
		return false
	}
	return s.On(ctx, formHookRequest[V]{
		data: &v,
		request: request{
			w:   w,
			r:   r,
			ctx: ctx,
		},
	})
}

// RequestChange is the typed request passed to [AChange] handlers.
type RequestChange = RequestEvent[ChangeEvent]

// AChange handles the browser `change` event.
//
// Use it for committed values such as blur-triggered input changes or select
// changes.
type AChange struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Actions to run before the hook request.
	// Optional.
	Before Actions
	// Backend event handler.
	// Receives a typed RequestEvent[ChangeEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestChange) bool
	// Actions to run on error.
	// Optional.
	OnError Actions
}

func (p AChange) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p AChange) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return eventAttr[ChangeEvent]{
		capture:   front.ChangeCapture{},
		scope:     p.Scope,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}.apply(ctx, attrs)
}

// RequestInput is the typed request passed to [AInput] handlers.
type RequestInput = RequestEvent[InputEvent]

// AInput handles the browser `input` event.
//
// Use it for live updates while the user is still editing a value.
type AInput struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Actions to run before the hook request.
	// Optional.
	Before Actions
	// Backend event handler.
	// Receives a typed RequestEvent[InputEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestInput) bool
	// If true, does not include value in event
	// Optional.
	ExcludeValue bool
	// Actions to run on error.
	// Optional.
	OnError Actions
}

func (p AInput) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p AInput) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return eventAttr[InputEvent]{
		capture: front.InputCapture{
			ExcludeValue: p.ExcludeValue,
		},
		scope:     p.Scope,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}.apply(ctx, attrs)
}
