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
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/gox"
)

var _ Attr = (*Emitter)(nil)

// Emitter dispatches synthetic DOM events to the elements it is attached to.
//
// Attach it as an attribute to one or more elements, then pass an event
// method's [Action] to [Call], [XCall], or any action-accepting API. The
// result is the number of captures the event fired; any capture error fails
// the call.
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

// Click returns an [Action] emitting a click event.
func (e *Emitter) Click(emit PointerEmit) Action {
	return e.emit("click", "pointer", emit)
}

// PointerDown returns an [Action] emitting a pointerdown event.
func (e *Emitter) PointerDown(emit PointerEmit) Action {
	return e.emit("pointerdown", "pointer", emit)
}

// PointerUp returns an [Action] emitting a pointerup event.
func (e *Emitter) PointerUp(emit PointerEmit) Action {
	return e.emit("pointerup", "pointer", emit)
}

// PointerMove returns an [Action] emitting a pointermove event.
func (e *Emitter) PointerMove(emit PointerEmit) Action {
	return e.emit("pointermove", "pointer", emit)
}

// PointerOver returns an [Action] emitting a pointerover event.
func (e *Emitter) PointerOver(emit PointerEmit) Action {
	return e.emit("pointerover", "pointer", emit)
}

// PointerOut returns an [Action] emitting a pointerout event.
func (e *Emitter) PointerOut(emit PointerEmit) Action {
	return e.emit("pointerout", "pointer", emit)
}

// PointerEnter returns an [Action] emitting a pointerenter event.
func (e *Emitter) PointerEnter(emit PointerEmit) Action {
	return e.emit("pointerenter", "pointer", emit)
}

// PointerLeave returns an [Action] emitting a pointerleave event.
func (e *Emitter) PointerLeave(emit PointerEmit) Action {
	return e.emit("pointerleave", "pointer", emit)
}

// PointerCancel returns an [Action] emitting a pointercancel event.
func (e *Emitter) PointerCancel(emit PointerEmit) Action {
	return e.emit("pointercancel", "pointer", emit)
}

// GotPointerCapture returns an [Action] emitting a gotpointercapture event.
func (e *Emitter) GotPointerCapture(emit PointerEmit) Action {
	return e.emit("gotpointercapture", "pointer", emit)
}

// LostPointerCapture returns an [Action] emitting a lostpointercapture event.
func (e *Emitter) LostPointerCapture(emit PointerEmit) Action {
	return e.emit("lostpointercapture", "pointer", emit)
}

// KeyDown returns an [Action] emitting a keydown event.
func (e *Emitter) KeyDown(emit KeyboardEmit) Action {
	return e.emit("keydown", "keyboard", emit)
}

// KeyUp returns an [Action] emitting a keyup event.
func (e *Emitter) KeyUp(emit KeyboardEmit) Action {
	return e.emit("keyup", "keyboard", emit)
}

// Focus returns an [Action] emitting a focus event.
func (e *Emitter) Focus(emit FocusEmit) Action {
	return e.emit("focus", "focus", emit)
}

// Blur returns an [Action] emitting a blur event.
func (e *Emitter) Blur(emit FocusEmit) Action {
	return e.emit("blur", "focus", emit)
}

// FocusIn returns an [Action] emitting a focusin event.
func (e *Emitter) FocusIn(emit FocusEmit) Action {
	return e.emit("focusin", "focus_io", emit)
}

// FocusOut returns an [Action] emitting a focusout event.
func (e *Emitter) FocusOut(emit FocusEmit) Action {
	return e.emit("focusout", "focus_io", emit)
}

// Input returns an [Action] emitting an input event.
func (e *Emitter) Input(emit InputEmit) Action {
	return e.emit("input", "input", emit)
}

// Change returns an [Action] emitting a change event.
func (e *Emitter) Change(emit ChangeEmit) Action {
	return e.emit("change", "change", emit)
}

// Submit returns an [Action] emitting a submit event.
func (e *Emitter) Submit(emit SubmitEmit) Action {
	return e.emit("submit", "submit", emit)
}

func (e *Emitter) emit(kind string, capture string, data any) Action {
	return actionFunc(func(ctx context.Context, core core.Core, gz bool) (action.Action, action.CallParams, error) {
		payload, err := action.IntoPayload(data, gz)
		if err != nil {
			return nil, action.CallParams{}, err
		}
		act := action.EmitEvent{
			EmitterID: e.id.ID(core),
			Type:      kind,
			Capture:   capture,
			Payload:   payload,
		}
		return act, action.CallParams{}, nil
	})
}
