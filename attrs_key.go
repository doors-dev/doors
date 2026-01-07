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

type KeyboardEvent = front.KeyboardEvent

type keyEventHook struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
	// Filters by event.key if provided.
	// Optional.
	Filter []string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Backend event handler.
	// Receives a typed REvent[KeyboardEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[KeyboardEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
	// Actions to run before the hook request.
	// Optional.
	Before []Action
}

func (k *keyEventHook) init(event string, ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[KeyboardEvent]{
		door: n,
		ctx:  ctx,
		capture: &front.KeyboardEventCapture{
			Event:           event,
			Filter:          k.Filter,
			PreventDefault:  k.PreventDefault,
			StopPropagation: k.StopPropagation,
		},
		inst:      inst,
		before:    k.Before,
		scope:     k.Scope,
		onError:   k.OnError,
		indicator: k.Indicator,
		on:        k.On,
	}).init(attrs)
}

// AKeyDown prepares a key down event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AKeyDown struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
	// Filters by event.key if provided.
	// Optional.
	Filter []string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Backend event handler.
	// Receives a typed REvent[KeyboardEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[KeyboardEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
	// Actions to run before the hook request.
	// Optional.
	Before []Action
}

func (k AKeyDown) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, k)
}

func (k AKeyDown) Attr() AttrInit {
	return k
}

func (k AKeyDown) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(&k)
	p.init("keydown", ctx, n, inst, attrs)
}

// AKeyUp prepares a key up event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AKeyUp struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
	// Filters by event.key if provided.
	// Optional.
	Filter []string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Backend event handler.
	// Receives a typed REvent[KeyboardEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[KeyboardEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
	// Actions to run before the hook request.
	// Optional.
	Before []Action
}

func (k AKeyUp) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, k)
}

func (k AKeyUp) Attr() AttrInit {
	return k
}

func (k AKeyUp) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(&k)
	p.init("keyup", ctx, n, inst, attrs)
}
