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

// KeyboardEvent is the payload sent to keyboard event handlers.
type KeyboardEvent = front.KeyboardEvent

// RequestKeyboard is the typed request passed to keyboard event handlers.
type RequestKeyboard = RequestEvent[KeyboardEvent]

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
	On func(context.Context, RequestKeyboard) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
	// Actions to run before the hook request.
	// Optional.
	Before []Action
}

func (k *keyEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	return eventAttr[KeyboardEvent]{
		capture: front.KeyboardEventCapture{
			Event:           event,
			Filter:          k.Filter,
			PreventDefault:  k.PreventDefault,
			StopPropagation: k.StopPropagation,
		},
		before:    k.Before,
		scope:     k.Scope,
		onError:   k.OnError,
		indicator: k.Indicator,
		on:        k.On,
	}.apply(ctx, attrs)
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
	On func(context.Context, RequestKeyboard) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
	// Actions to run before the hook request.
	// Optional.
	Before []Action
}

func (k AKeyDown) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(k, cur, elem)
}

func (k AKeyDown) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*keyEventHook)(&k).apply("keydown", ctx, attrs)
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
	On func(context.Context, RequestKeyboard) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
	// Actions to run before the hook request.
	// Optional.
	Before []Action
}

func (k AKeyUp) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(k, cur, elem)
}

func (k AKeyUp) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*keyEventHook)(&k).apply("keyup", ctx, attrs)
}
