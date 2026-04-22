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

// FocusEvent is the payload sent to focus event handlers.
type FocusEvent = front.FocusEvent

// RequestFocus is the typed request passed to focus event handlers.
type RequestFocus = RequestEvent[FocusEvent]

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
	On func(context.Context, RequestFocus) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
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
	On func(context.Context, RequestEvent[FocusEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
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
	On func(context.Context, RequestFocus) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (f AFocus) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(f, cur, elem)
}

func (f AFocus) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
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
	On func(context.Context, RequestFocus) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (b ABlur) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(b, cur, elem)
}

func (b ABlur) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
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
	On func(context.Context, RequestFocus) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (f AFocusIn) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(f, cur, elem)
}

func (f AFocusIn) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
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
	On func(context.Context, RequestFocus) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (f AFocusOut) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(f, cur, elem)
}

func (f AFocusOut) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*focusIOEventHook)(&f).apply("focusout", ctx, attrs)
}
