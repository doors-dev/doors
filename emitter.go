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

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/gox"
)

var _ Attr = (*Emitter)(nil)

// Emitter dispatches synthetic DOM events to the elements it is attached to.
//
// Attach it as an attribute to one or more elements, then pass an event
// method's action to [Call] or any action-accepting API. Each event method
// returns an [ActionInto]; Into captures the number of captures the event
// fired. Any capture error fails the call.
type Emitter struct {
	id front.AutoID
}

func (e *Emitter) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(e, cur, elem)
}

func (e *Emitter) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	front.AttrsSetEmitter(attrs, e.id.ID(core))
	front.AttrsSetParent(attrs, core.Door().ID())
	return nil
}

// Click returns an action emitting a click event.
func (e *Emitter) Click(emit PointerEmit) ActionInto[int] {
	return e.emit("click", "pointer", emit)
}

// PointerDown returns an action emitting a pointerdown event.
func (e *Emitter) PointerDown(emit PointerEmit) ActionInto[int] {
	return e.emit("pointerdown", "pointer", emit)
}

// PointerUp returns an action emitting a pointerup event.
func (e *Emitter) PointerUp(emit PointerEmit) ActionInto[int] {
	return e.emit("pointerup", "pointer", emit)
}

// PointerMove returns an action emitting a pointermove event.
func (e *Emitter) PointerMove(emit PointerEmit) ActionInto[int] {
	return e.emit("pointermove", "pointer", emit)
}

// PointerOver returns an action emitting a pointerover event.
func (e *Emitter) PointerOver(emit PointerEmit) ActionInto[int] {
	return e.emit("pointerover", "pointer", emit)
}

// PointerOut returns an action emitting a pointerout event.
func (e *Emitter) PointerOut(emit PointerEmit) ActionInto[int] {
	return e.emit("pointerout", "pointer", emit)
}

// PointerEnter returns an action emitting a pointerenter event.
func (e *Emitter) PointerEnter(emit PointerEmit) ActionInto[int] {
	return e.emit("pointerenter", "pointer", emit)
}

// PointerLeave returns an action emitting a pointerleave event.
func (e *Emitter) PointerLeave(emit PointerEmit) ActionInto[int] {
	return e.emit("pointerleave", "pointer", emit)
}

// PointerCancel returns an action emitting a pointercancel event.
func (e *Emitter) PointerCancel(emit PointerEmit) ActionInto[int] {
	return e.emit("pointercancel", "pointer", emit)
}

// GotPointerCapture returns an action emitting a gotpointercapture event.
func (e *Emitter) GotPointerCapture(emit PointerEmit) ActionInto[int] {
	return e.emit("gotpointercapture", "pointer", emit)
}

// LostPointerCapture returns an action emitting a lostpointercapture event.
func (e *Emitter) LostPointerCapture(emit PointerEmit) ActionInto[int] {
	return e.emit("lostpointercapture", "pointer", emit)
}

// KeyDown returns an action emitting a keydown event.
func (e *Emitter) KeyDown(emit KeyboardEmit) ActionInto[int] {
	return e.emit("keydown", "keyboard", emit)
}

// KeyUp returns an action emitting a keyup event.
func (e *Emitter) KeyUp(emit KeyboardEmit) ActionInto[int] {
	return e.emit("keyup", "keyboard", emit)
}

// Focus returns an action emitting a focus event.
func (e *Emitter) Focus(emit FocusEmit) ActionInto[int] {
	return e.emit("focus", "focus", emit)
}

// Blur returns an action emitting a blur event.
func (e *Emitter) Blur(emit FocusEmit) ActionInto[int] {
	return e.emit("blur", "focus", emit)
}

// FocusIn returns an action emitting a focusin event.
func (e *Emitter) FocusIn(emit FocusEmit) ActionInto[int] {
	return e.emit("focusin", "focus_io", emit)
}

// FocusOut returns an action emitting a focusout event.
func (e *Emitter) FocusOut(emit FocusEmit) ActionInto[int] {
	return e.emit("focusout", "focus_io", emit)
}

// Input returns an action emitting an input event.
func (e *Emitter) Input(emit InputEmit) ActionInto[int] {
	return e.emit("input", "input", emit)
}

// Change returns an action emitting a change event.
func (e *Emitter) Change(emit ChangeEmit) ActionInto[int] {
	return e.emit("change", "change", emit)
}

// Submit returns an action emitting a submit event.
func (e *Emitter) Submit(emit SubmitEmit) ActionInto[int] {
	return e.emit("submit", "submit", emit)
}

func (e *Emitter) emit(kind string, capture string, data any) ActionInto[int] {
	return actionIntoFunc[int](func(ctx context.Context, core core.Core, gz bool) (action, error) {
		payload, err := actions.IntoPayload(data, gz)
		if err != nil {
			return action{}, err
		}
		return action{
			action: actions.EmitEvent{
				EmitterID: e.id.ID(core),
				Type:      kind,
				Capture:   capture,
				Payload:   payload,
			},
		}, nil
	})
}
