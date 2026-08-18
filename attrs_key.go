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

// Mod is the required state of a keyboard modifier in a [Key] match.
type Mod uint8

// Modifier states for [Key].
const (
	// ModAny matches regardless of whether the modifier is held.
	ModAny Mod = iota
	// ModOn requires the modifier to be held.
	ModOn
	// ModOff requires the modifier to not be held.
	ModOff
)

// Key is one key and modifier combination matched against a keyboard event.
type Key struct {
	// Key is the event.key value to match. An empty string matches any key.
	Key string
	// CtrlMod is the required ctrl state. Default: [ModAny].
	CtrlMod Mod
	// ShiftMod is the required shift state. Default: [ModAny].
	ShiftMod Mod
	// AltMod is the required alt state. Default: [ModAny].
	AltMod Mod
	// MetaMod is the required meta state. Default: [ModAny].
	MetaMod Mod
}

// RequestKeyboard is the request handle passed to [AKeyDown] and [AKeyUp]
// handlers.
type RequestKeyboard = RequestEvent[KeyboardEvent]

type keyEventHook struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
	// ExactTarget limits the handler to events whose target is the element
	// itself. Optional.
	ExactTarget bool
	// Keys filters by key and modifier state; the handler fires when the event
	// matches any entry. Optional; without entries every event fires.
	Keys []Key
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// On handles the event on the server. Return false to keep the handler
	// active, true to remove it. Optional.
	On func(context.Context, RequestKeyboard) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
}

func (k *keyEventHook) apply(event string, ctx context.Context, attrs gox.Attrs) error {
	keys := make([]front.KeyMatch, 0, len(k.Keys))
	for _, key := range k.Keys {
		keys = append(keys, front.KeyMatch{
			Key:   key.Key,
			Ctrl:  uint8(key.CtrlMod),
			Shift: uint8(key.ShiftMod),
			Alt:   uint8(key.AltMod),
			Meta:  uint8(key.MetaMod),
		})
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

// AKeyDown handles the browser keydown event on the element it is attached to.
//
// Attach it as an attribute to one or more elements. Each matching event sends
// one request, subject to Scope. Return false from On to keep the handler
// active, true to remove it. A nil On still installs the hook and still sends
// the request, so omit the attr rather than nil-ing On.
type AKeyDown struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
	// ExactTarget limits the handler to events whose target is the element
	// itself. Optional.
	ExactTarget bool
	// Keys limits the handler to events matching at least one entry.
	// PreventDefault and StopPropagation apply only to matching
	// events. Optional; without entries every keydown fires.
	Keys []Key
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// On handles the event on the server. Return false to keep the handler
	// active, true to remove it. Optional.
	On func(context.Context, RequestKeyboard) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
}

func (k AKeyDown) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(k, cur, elem)
}

func (k AKeyDown) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*keyEventHook)(&k).apply("keydown", ctx, attrs)
}

// AKeyUp handles the browser keyup event on the element it is attached to.
// Fields and handler contract follow [AKeyDown].
type AKeyUp struct {
	// StopPropagation stops the event from bubbling up the DOM. Optional.
	StopPropagation bool
	// PreventDefault suppresses the browser's default action. Optional.
	PreventDefault bool
	// ExactTarget limits the handler to events whose target is the element
	// itself. Optional.
	ExactTarget bool
	// Keys limits the handler to events matching at least one entry.
	// PreventDefault and StopPropagation apply only to matching
	// events. Optional; without entries every keyup fires.
	Keys []Key
	// Scope controls how the request is scheduled. Optional; unscoped requests
	// are sent as soon as the event fires.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// On handles the event on the server. Return false to keep the handler
	// active, true to remove it. Optional.
	On func(context.Context, RequestKeyboard) bool
	// OnError lists client-side actions to run when the request
	// fails. Optional.
	OnError Actions
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
}

func (k AKeyUp) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(k, cur, elem)
}

func (k AKeyUp) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	return (*keyEventHook)(&k).apply("keyup", ctx, attrs)
}
