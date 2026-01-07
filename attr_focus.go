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
	"io"

	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
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

func (p *focusIOEventHook) init(event string, ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[FocusEvent]{
		capture: &front.FocusIOCapture{
			Event:           event,
			StopPropagation: p.StopPropagation,
			ExactTarget:     p.ExactTarget,
		},
		door:      n,
		ctx:       ctx,
		inst:      inst,
		onError:   p.OnError,
		scope:     p.Scope,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
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

func (p *focusEventHook) init(event string, ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[FocusEvent]{
		capture: &front.FocusCapture{
			Event: event,
		},
		door:      n,
		ctx:       ctx,
		inst:      inst,
		onError:   p.OnError,
		before:    p.Before,
		scope:     p.Scope,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
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

func (f AFocus) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, f)
}

func (f AFocus) Attr() AttrInit {
	return f
}

func (f AFocus) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(&f)
	p.init("focus", ctx, n, inst, attrs)
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

func (b ABlur) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, b)
}

func (b ABlur) Attr() AttrInit {
	return b
}

func (b ABlur) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(&b)
	p.init("blur", ctx, n, inst, attrs)
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

func (f AFocusIn) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, f)
}

func (f AFocusIn) Attr() AttrInit {
	return f
}

func (f AFocusIn) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusIOEventHook)(&f)
	p.init("focusin", ctx, n, inst, attrs)
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

func (f AFocusOut) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, f)
}

func (f AFocusOut) Attr() AttrInit {
	return f
}

func (f AFocusOut) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusIOEventHook)(&f)
	p.init("focusout", ctx, n, inst, attrs)
}
