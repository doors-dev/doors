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

func (p *pointerEventHook) init(event string, ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {

	(&eventAttr[PointerEvent]{
		capture: &front.PointerCapture{
			Event:           event,
			StopPropagation: p.StopPropagation,
			PreventDefault:  p.PreventDefault,
			ExactTarget:     p.ExactTarget,
		},
		inst:      inst,
		door:      n,
		scope:     p.Scope,
		ctx:       ctx,
		before:    p.Before,
		onError:   p.OnError,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
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

func (p AClick) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, p)
}

func (c AClick) Attr() AttrInit {
	return c
}

func (c AClick) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("click", ctx, n, inst, attrs)
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

func (c APointerDown) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerDown) Attr() AttrInit {
	return c
}

func (c APointerDown) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerdown", ctx, n, inst, attrs)
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

func (c APointerUp) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerUp) Attr() AttrInit {
	return c
}

func (c APointerUp) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerup", ctx, n, inst, attrs)
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

func (c APointerMove) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerMove) Attr() AttrInit {
	return c
}

func (c APointerMove) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointermove", ctx, n, inst, attrs)
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

func (c APointerOver) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerOver) Attr() AttrInit {
	return c
}

func (c APointerOver) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerover", ctx, n, inst, attrs)
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

func (c APointerOut) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerOut) Attr() AttrInit {
	return c
}

func (c APointerOut) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerout", ctx, n, inst, attrs)
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

func (c APointerEnter) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerEnter) Attr() AttrInit {
	return c
}

func (c APointerEnter) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerenter", ctx, n, inst, attrs)
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

func (c APointerLeave) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerLeave) Attr() AttrInit {
	return c
}

func (c APointerLeave) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerleave", ctx, n, inst, attrs)
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

func (c APointerCancel) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c APointerCancel) Attr() AttrInit {
	return c
}

func (c APointerCancel) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointercancel", ctx, n, inst, attrs)
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

func (c AGotPointerCapture) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c AGotPointerCapture) Attr() AttrInit {
	return c
}

func (c AGotPointerCapture) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("gotpointercapture", ctx, n, inst, attrs)
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

func (c ALostPointerCapture) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, c)
}

func (c ALostPointerCapture) Attr() AttrInit {
	return c
}

func (c ALostPointerCapture) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("lostpointercapture", ctx, n, inst, attrs)
}
