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
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type PointerEvent = front.PointerEvent

type pointerEventHook struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
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

// AClick prepares a click event hook for DOM elements,
// configuring propagation, scheduling, indicators, and handlers.
type AClick struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p AClick) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p AClick) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("click", ctx, attrs)
}

// APointerDown prepares a pointer down event hook for DOM elements,
// configuring propagation, scheduling, indicators, and handlers.
type APointerDown struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerDown) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerDown) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerdown", ctx, attrs)
}

// APointerUp prepares a pointer up event hook for DOM elements,
// configuring propagation, scheduling, indicators, and handlers.
type APointerUp struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerUp) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerUp) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerup", ctx, attrs)
}

// APointerMove prepares a pointer move event hook for DOM elements,
// configuring propagation, scheduling, indicators, and handlers.
type APointerMove struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerMove) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerMove) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointermove", ctx, attrs)
}

// APointerOver prepares a pointer over event hook for DOM elements,
// configuring propagation, scheduling, indicators, and handlers.
type APointerOver struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerOver) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerOver) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerover", ctx, attrs)
}

// APointerOut prepares a pointer out event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type APointerOut struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerOut) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerOut) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerout", ctx, attrs)
}

// APointerEnter prepares a pointer enter event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type APointerEnter struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerEnter) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerEnter) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerenter", ctx, attrs)
}

// APointerLeave prepares a pointer leave event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type APointerLeave struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerLeave) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerLeave) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointerleave", ctx, attrs)
}

// APointerCancel prepares a pointer cancel event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type APointerCancel struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p APointerCancel) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p APointerCancel) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("pointercancel", ctx, attrs)
}

// AGotPointerCapture prepares a gotpointercapture event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AGotPointerCapture struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p AGotPointerCapture) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p AGotPointerCapture) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("gotpointercapture", ctx, attrs)
}

// ALostPointerCapture prepares a lostpointercapture event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type ALostPointerCapture struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
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
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p ALostPointerCapture) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(p, cur, elem)
}

func (p ALostPointerCapture) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*pointerEventHook)(&p).apply("lostpointercapture", ctx, attrs)
}
