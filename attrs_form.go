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
	"errors"
	"net/http"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
	"github.com/go-playground/form/v4"
)

// ARawSubmit handles a form submission with raw multipart access.
type ARawSubmit struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend form handler.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestRawForm) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (s ARawSubmit) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(s, cur, elem)
}

func (s ARawSubmit) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(s.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, front.FormCapture{}, front.Hook{
		OnError:  intoActions(ctx, s.OnError),
		Before:   intoActions(ctx, s.Before),
		Scope:    front.IntoScopeSet(core, s.Scope),
		Indicate: front.IntoIndicate(s.Indicator),
		Hook:     hook,
	})
	return nil
}

func (s *ARawSubmit) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	done := s.On(ctx, &request{
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
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend form handler.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestForm[T]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (s ASubmit[V]) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(s, cur, elem)
}

func (s ASubmit[V]) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(s.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, front.FormCapture{}, front.Hook{
		OnError:  intoActions(ctx, s.OnError),
		Before:   intoActions(ctx, s.Before),
		Scope:    front.IntoScopeSet(core, s.Scope),
		Indicate: front.IntoIndicate(s.Indicator),
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
	return s.On(ctx, &formHookRequest[V]{
		data: &v,
		request: request{
			w:   w,
			r:   r,
			ctx: ctx,
		},
	})
}

// ChangeEvent is the payload sent to [AChange] handlers.
type ChangeEvent = front.ChangeEvent

// RequestChange is the typed request passed to [AChange] handlers.
type RequestChange = RequestEvent[ChangeEvent]

// AChange handles the browser `change` event.
//
// Use it for committed values such as blur-triggered input changes or select
// changes.
type AChange struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend event handler.
	// Receives a typed REvent[ChangeEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestChange) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p AChange) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
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

// InputEvent is the payload sent to [AInput] handlers.
type InputEvent = front.InputEvent

// RequestInput is the typed request passed to [AInput] handlers.
type RequestInput = RequestEvent[InputEvent]

// AInput handles the browser `input` event.
//
// Use it for live updates while the user is still editing a value.
type AInput struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend event handler.
	// Receives a typed REvent[InputEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestInput) bool
	// If true, does not include value in event
	// Optional.
	ExcludeValue bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p AInput) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
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
