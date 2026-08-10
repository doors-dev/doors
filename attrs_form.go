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
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
	"github.com/go-playground/form/v4"
)

// ARawSubmit handles the browser submit event on the form it is attached to
// and gives On raw multipart access.
//
// The native submission is always prevented and the form is sent as
// multipart/form-data. Return false from On to keep the handler active, true
// to remove it. A nil On still installs the hook and still sends the request.
//
// Prefer [ASubmit] when the form decodes into a Go type.
type ARawSubmit struct {
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
	// RequestTimeout overrides [Conf].RequestTimeout for this form's
	// requests. Optional.
	RequestTimeout time.Duration
	// On handles the submitted form on the server. Return false to keep the
	// handler active, true to remove it. Optional.
	On func(context.Context, RequestRawForm) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (s ARawSubmit) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(s, cur, elem)
}

func (s ARawSubmit) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(s.handle(core), nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, front.FormCapture{}, front.Hook{
		OnError:  intoActions(ctx, actionsOrNil(s.OnError)),
		Before:   intoActions(ctx, actionsOrNil(s.Before)),
		Scope:    scopesOrNil(core, s.Scope),
		Indicate: indicatorsOrNil(s.Indicator),
		Timeout:  s.RequestTimeout,
		Hook:     hook,
	})
	return nil
}

func (s *ARawSubmit) handle(core core.Core) func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		if s.On == nil {
			r.Body.Close()
			return false
		}
		return s.On(ctx, &request{
			w:     w,
			r:     r,
			ctx:   ctx,
			limit: core.App().Conf().ServerRequestBodyLimit,
		})
	}
}

var formDecoder *form.Decoder

func init() {
	formDecoder = form.NewDecoder()
}

// ASubmit handles the browser submit event on the form it is attached to and
// decodes the submitted values into T.
//
// The native submission is always prevented and the form is sent as
// multipart/form-data. Field names come from form struct tags, or the Go field
// name when the tag is absent. A parse or decode failure skips On and runs
// OnError. File parts are not decoded; use [ARawSubmit] for uploads.
//
// Return false from On to keep the handler active, true to remove it. A nil On
// still installs the hook and still sends the request.
type ASubmit[T any] struct {
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
	// RequestTimeout overrides [Conf].RequestTimeout for this form's
	// requests. Optional.
	RequestTimeout time.Duration
	// On handles the decoded form on the server. Return false to keep the
	// handler active, true to remove it. Optional.
	On func(context.Context, RequestForm[T]) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (s ASubmit[V]) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(s, cur, elem)
}

func (s ASubmit[V]) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(s.handle(core), nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, front.FormCapture{}, front.Hook{
		OnError:  intoActions(ctx, actionsOrNil(s.OnError)),
		Before:   intoActions(ctx, actionsOrNil(s.Before)),
		Scope:    scopesOrNil(core, s.Scope),
		Indicate: indicatorsOrNil(s.Indicator),
		Timeout:  s.RequestTimeout,
		Hook:     hook,
	})
	return nil
}

func (s *ASubmit[V]) handle(core core.Core) func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		if s.On == nil {
			r.Body.Close()
			return false
		}
		limit := core.App().Conf().ServerRequestBodyLimit
		r.Body = http.MaxBytesReader(w, r.Body, int64(limit))
		err := r.ParseMultipartForm(int64(limit))
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Multipart form parsing error"))
			return false
		}
		var v V
		err = formDecoder.Decode(&v, r.Form)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Form decoding error"))
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
}

// RequestChange is the request handle passed to [AChange] handlers.
type RequestChange = RequestEvent[ChangeEvent]

// AChange handles the browser change event on the element it is attached to,
// which fires on a committed value: on blur for text inputs, immediately for
// checkboxes and selects.
//
// Attach it as an attribute to one or more elements. Each event sends one
// request carrying the target's current value, subject to Scope. Return false
// from On to keep the handler active, true to remove it. A nil On still
// installs the hook and still sends the request, so omit the attr rather than
// nil-ing On.
type AChange struct {
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
	// On handles the event on the server. Return false to keep the handler
	// active, true to remove it. Optional.
	On func(context.Context, RequestChange) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
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

// RequestInput is the request handle passed to [AInput] handlers.
type RequestInput = RequestEvent[InputEvent]

// AInput handles the browser input event on the element it is attached to.
// Unlike [AChange], it fires on every edit rather than on the committed
// value. The handler contract follows [AChange].
type AInput struct {
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
	// On handles the event on the server. Return false to keep the handler
	// active, true to remove it. Optional.
	On func(context.Context, RequestInput) bool
	// ExcludeValue omits the input value from the payload, leaving Name,
	// Value, Number, Date, Selected, and Checked unset. Optional.
	ExcludeValue bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
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
