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

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

// RequestFocus is the request handle passed to focus event attr handlers.
type RequestFocus = RequestEvent[FocusEvent]

type focusIOEventHook struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// ExactTarget limits the handler to events whose target is the element
	// itself. Optional.
	ExactTarget bool
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
	On func(context.Context, RequestFocus) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p *focusIOEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	return eventAttr[FocusEvent]{
		capture: front.FocusIOCapture{
			Event:           event,
			StopPropagation: p.StopPropagation,
			ExactTarget:     p.ExactTarget,
		},
		scope:     p.Scope,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}.apply(ctx, attrs)
}

type focusEventHook struct {
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
	On func(context.Context, RequestEvent[FocusEvent]) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p *focusEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	return eventAttr[FocusEvent]{
		capture: front.FocusCapture{
			Event: event,
		},
		scope:     p.Scope,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}.apply(ctx, attrs)
}

// AFocus handles the browser focus event on the element it is attached to.
//
// Attach it as an attribute to one or more elements. Each event sends one
// request, subject to Scope. Return false from On to keep the handler active,
// true to remove it. A nil On still installs the hook and still sends the
// request, so omit the attr rather than nil-ing On.
//
// Unlike [AFocusIn], the focus event does not bubble, so there are no
// propagation fields.
type AFocus struct {
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
	On func(context.Context, RequestFocus) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (f AFocus) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(f, cur, elem)
}

func (f AFocus) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*focusEventHook)(&f).apply("focus", ctx, attrs)
}

// ABlur handles the browser blur event on the element it is attached to. Fields
// and handler contract follow [AFocus].
type ABlur struct {
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
	On func(context.Context, RequestFocus) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (b ABlur) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(b, cur, elem)
}

func (b ABlur) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*focusEventHook)(&b).apply("blur", ctx, attrs)
}

// AFocusIn handles the browser focusin event on the element it is attached to.
// Unlike focus, focusin bubbles; handler contract follows [AFocus].
type AFocusIn struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// ExactTarget limits the handler to events whose target is the element
	// itself. Optional.
	ExactTarget bool
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
	On func(context.Context, RequestFocus) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (f AFocusIn) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(f, cur, elem)
}

func (f AFocusIn) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*focusIOEventHook)(&f).apply("focusin", ctx, attrs)
}

// AFocusOut handles the browser focusout event on the element it is attached
// to. Unlike blur, focusout bubbles; handler contract follows [AFocus].
type AFocusOut struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// ExactTarget limits the handler to events whose target is the element
	// itself. Optional.
	ExactTarget bool
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
	On func(context.Context, RequestFocus) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (f AFocusOut) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(f, cur, elem)
}

func (f AFocusOut) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*focusIOEventHook)(&f).apply("focusout", ctx, attrs)
}
