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

// Mod is the required state of a keyboard modifier in a Key match.
type Mod uint8

const (
	// ModAny matches regardless of whether the modifier is held.
	ModAny Mod = iota
	// ModOn requires the modifier to be held.
	ModOn
	// ModOff requires the modifier to not be held.
	ModOff
)

// Key describes a single keyboard shortcut to match against a keyboard event.
// A keyboard hook fires when the event matches any of the entries in Keys.
type Key struct {
	// Key is the event.key to match. An empty string matches any key.
	Key string
	// CtrlMod is the required ctrl modifier state.
	CtrlMod Mod
	// ShiftMod is the required shift modifier state.
	ShiftMod Mod
	// AltMod is the required alt modifier state.
	AltMod Mod
	// MetaMod is the required meta (cmd/win) modifier state.
	MetaMod Mod
}

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
	// Filters by event.key and modifier state. The hook fires only when the
	// event matches at least one entry.
	// Optional.
	Keys []Key
	// Filters by event.key.
	//
	// Deprecated: use Keys. Each entry matches its key regardless of modifier
	// state.
	// Optional.
	Filter []string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Backend event handler.
	// Receives a typed RequestEvent[KeyboardEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestKeyboard) bool
	// Actions to run on error.
	// Optional.
	OnError Actions
	// Actions to run before the hook request.
	// Optional.
	Before Actions
}

func (k *keyEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	keys := make([]front.KeyMatch, 0, len(k.Keys)+len(k.Filter))
	for _, key := range k.Keys {
		keys = append(keys, front.KeyMatch{
			Key:   key.Key,
			Ctrl:  uint8(key.CtrlMod),
			Shift: uint8(key.ShiftMod),
			Alt:   uint8(key.AltMod),
			Meta:  uint8(key.MetaMod),
		})
	}
	for _, filter := range k.Filter {
		keys = append(keys, front.KeyMatch{Key: filter})
	}
	return eventAttr[KeyboardEvent]{
		capture: front.KeyboardEventCapture{
			Event:           event,
			Keys:            keys,
			PreventDefault:  k.PreventDefault,
			StopPropagation: k.StopPropagation,
			ExactTarget:     k.ExactTarget,
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
	// Filters by event.key and modifier state. The hook fires only when the
	// event matches at least one entry.
	// Optional.
	Keys []Key
	// Filters by event.key.
	//
	// Deprecated: use Keys. Each entry matches its key regardless of modifier
	// state.
	// Optional.
	Filter []string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Backend event handler.
	// Receives a typed RequestEvent[KeyboardEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestKeyboard) bool
	// Actions to run on error.
	// Optional.
	OnError Actions
	// Actions to run before the hook request.
	// Optional.
	Before Actions
}

func (k AKeyDown) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(k, cur, elem)
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
	// Filters by event.key and modifier state. The hook fires only when the
	// event matches at least one entry.
	// Optional.
	Keys []Key
	// Filters by event.key.
	//
	// Deprecated: use Keys. Each entry matches its key regardless of modifier
	// state.
	// Optional.
	Filter []string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running.
	// Optional.
	Indicator Indicators
	// Backend event handler.
	// Receives a typed RequestEvent[KeyboardEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RequestKeyboard) bool
	// Actions to run on error.
	// Optional.
	OnError Actions
	// Actions to run before the hook request.
	// Optional.
	Before Actions
}

func (k AKeyUp) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(k, cur, elem)
}

func (k AKeyUp) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*keyEventHook)(&k).apply("keyup", ctx, attrs)
}
