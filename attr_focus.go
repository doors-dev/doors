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

type FocusEvent = front.FocusEvent

type focusIOEventHook struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
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
	// Receives a typed REvent[FocusEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p *focusIOEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	return eventAttr[FocusEvent]{
		capture: &front.FocusIOCapture{
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
	// Receives a typed REvent[FocusEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p *focusEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	return eventAttr[FocusEvent]{
		capture: &front.FocusCapture{
			Event: event,
		},
		scope:     p.Scope,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}.apply(ctx, attrs)
}


// AFocus prepares a focus event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AFocus struct {
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
	// Receives a typed REvent[FocusEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (f AFocus) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(f, cur, elem)
}

func (f AFocus) Apply(ctx context.Context, attrs gox.Attrs) error {
	return (*focusEventHook)(&f).apply("focus", ctx, attrs)
}


// ABlur prepares a blur event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type ABlur struct {
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
	// Receives a typed REvent[FocusEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (b ABlur) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(b, cur, elem)
}

func (b ABlur) Apply(ctx context.Context, attrs gox.Attrs) error {
	return (*focusEventHook)(&b).apply("blur", ctx, attrs)
}


// AFocusIn prepares a focusin event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AFocusIn struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
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
	// Receives a typed REvent[FocusEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (f AFocusIn) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(f, cur, elem)
}

func (f AFocusIn) Apply(ctx context.Context, attrs gox.Attrs) error {
	return (*focusIOEventHook)(&f).apply("focusin", ctx, attrs)
}


// AFocusOut prepares a focusout event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AFocusOut struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
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
	// Receives a typed REvent[FocusEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (f AFocusOut) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(f, cur, elem)
}

func (f AFocusOut) Apply(ctx context.Context, attrs gox.Attrs) error {
	return (*focusIOEventHook)(&f).apply("focusout", ctx, attrs)
}

