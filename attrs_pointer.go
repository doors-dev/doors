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

// RequestPointer is the request handle passed to pointer event attr handlers.
type RequestPointer = RequestEvent[PointerEvent]

type pointerEventHook struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p *pointerEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	return eventAttr[PointerEvent]{
		capture: &front.PointerCapture{
			Event:           event,
			StopPropagation: p.StopPropagation,
			PreventDefault:  p.PreventDefault,
			ExactTarget:     p.ExactTarget,
		},
		scope:     p.Scope,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}.apply(ctx, attrs)
}

// AClick handles the browser click event on the element it is attached to.
//
// Attach it as an attribute to one or more elements. Each event sends one
// request, subject to Scope. Return false from On to keep the handler active,
// true to remove it. A nil On still installs the hook and still sends the
// request, so omit the attr rather than nil-ing On.
type AClick struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p AClick) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p AClick) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("click", ctx, attrs)
}

// APointerDown handles the browser pointerdown event on the element it is
// attached to. Fields and handler contract follow [AClick].
type APointerDown struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerDown) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerDown) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerdown", ctx, attrs)
}

// APointerUp handles the browser pointerup event on the element it is attached
// to. Fields and handler contract follow [AClick].
type APointerUp struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerUp) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerUp) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerup", ctx, attrs)
}

// APointerMove handles the browser pointermove event on the element it is
// attached to. Fields and handler contract follow [AClick].
type APointerMove struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerMove) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerMove) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointermove", ctx, attrs)
}

// APointerOver handles the browser pointerover event on the element it is
// attached to. Fields and handler contract follow [AClick].
type APointerOver struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerOver) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerOver) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerover", ctx, attrs)
}

// APointerOut handles the browser pointerout event on the element it is
// attached to. Fields and handler contract follow [AClick].
type APointerOut struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerOut) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerOut) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerout", ctx, attrs)
}

// APointerEnter handles the browser pointerenter event on the element it is
// attached to. Unlike [APointerOver], pointerenter does not bubble. Fields and
// handler contract follow [AClick].
type APointerEnter struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerEnter) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerEnter) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerenter", ctx, attrs)
}

// APointerLeave handles the browser pointerleave event on the element it is
// attached to. Unlike [APointerOut], pointerleave does not bubble. Fields and
// handler contract follow [AClick].
type APointerLeave struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerLeave) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerLeave) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerleave", ctx, attrs)
}

// APointerCancel handles the browser pointercancel event on the element it is
// attached to. Fields and handler contract follow [AClick].
type APointerCancel struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p APointerCancel) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p APointerCancel) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointercancel", ctx, attrs)
}

// AGotPointerCapture handles the browser gotpointercapture event on the element
// it is attached to. Fields and handler contract follow [AClick].
type AGotPointerCapture struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p AGotPointerCapture) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p AGotPointerCapture) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("gotpointercapture", ctx, attrs)
}

// ALostPointerCapture handles the browser lostpointercapture event on the
// element it is attached to. Fields and handler contract follow [AClick].
type ALostPointerCapture struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
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
	On func(context.Context, RequestPointer) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
}

func (p ALostPointerCapture) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(p, cur, elem)
}

func (p ALostPointerCapture) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("lostpointercapture", ctx, attrs)
}
