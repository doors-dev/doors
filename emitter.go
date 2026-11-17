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
	"encoding/json"
	"errors"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/gox"
)

var _ Attr = (*Emitter)(nil)

// Emitter dispatches synthetic DOM events to the elements it is attached to.
//
// Attach it as an attribute to one or more elements, then dispatch an event
// method's action with [Call]. Unset event init fields use the browser
// defaults. The events bubble, so event attrs on ancestor elements run too.
//
// Several emitters can share an element; each dispatches its own event.
//
// Into captures an [EmitReport] of the hook requests the emitted events
// triggered, including hooks of events they trigger, such as a form submit
// after a click on a submit button (see [ActionInto]). The action completes
// only after every triggered hook request has finished; a failed or canceled
// one is reported and does not fail the call. The call fails when the emitter
// has no attached elements.
//
// If any attached element is a disabled form control or inside an inert
// subtree, nothing is dispatched and the call fails with [ErrExecution].
type Emitter struct {
	id front.AutoID
}

// EmitReport is the outcome of the hook requests an [Emitter] event triggered.
type EmitReport struct {
	// OK is the number of hook requests that completed.
	OK int
	// Canceled is the number of hook requests that were canceled, for example
	// by a scope or by a filter such as ExactTarget or Key.
	Canceled int
	// Failed holds an error for each hook request that failed.
	Failed []error
}

func (r *EmitReport) UnmarshalJSON(data []byte) error {
	var raw struct {
		OK       int      `json:"ok"`
		Canceled int      `json:"canceled"`
		Failed   []string `json:"failed"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = EmitReport{OK: raw.OK, Canceled: raw.Canceled}
	for _, f := range raw.Failed {
		r.Failed = append(r.Failed, errors.New(f))
	}
	return nil
}

func (e *Emitter) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(e, cur, elem)
}

func (e *Emitter) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	front.AttrsAppendEmitter(attrs, e.id.ID(core))
	front.AttrsSetParent(attrs, core.Door().ID())
	return nil
}

// Click returns an action emitting a click event.
func (e *Emitter) Click(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("click", "pointer", emit)
}

// PointerDown returns an action emitting a pointerdown event.
func (e *Emitter) PointerDown(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointerdown", "pointer", emit)
}

// PointerUp returns an action emitting a pointerup event.
func (e *Emitter) PointerUp(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointerup", "pointer", emit)
}

// PointerMove returns an action emitting a pointermove event.
func (e *Emitter) PointerMove(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointermove", "pointer", emit)
}

// PointerOver returns an action emitting a pointerover event.
func (e *Emitter) PointerOver(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointerover", "pointer", emit)
}

// PointerOut returns an action emitting a pointerout event.
func (e *Emitter) PointerOut(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointerout", "pointer", emit)
}

// PointerEnter returns an action emitting a pointerenter event.
func (e *Emitter) PointerEnter(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointerenter", "pointer", emit)
}

// PointerLeave returns an action emitting a pointerleave event.
func (e *Emitter) PointerLeave(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointerleave", "pointer", emit)
}

// PointerCancel returns an action emitting a pointercancel event.
func (e *Emitter) PointerCancel(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("pointercancel", "pointer", emit)
}

// GotPointerCapture returns an action emitting a gotpointercapture event.
func (e *Emitter) GotPointerCapture(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("gotpointercapture", "pointer", emit)
}

// LostPointerCapture returns an action emitting a lostpointercapture event.
func (e *Emitter) LostPointerCapture(emit PointerEmit) ActionInto[EmitReport] {
	return e.emit("lostpointercapture", "pointer", emit)
}

// KeyDown returns an action emitting a keydown event.
func (e *Emitter) KeyDown(emit KeyboardEmit) ActionInto[EmitReport] {
	return e.emit("keydown", "keyboard", emit)
}

// KeyUp returns an action emitting a keyup event.
func (e *Emitter) KeyUp(emit KeyboardEmit) ActionInto[EmitReport] {
	return e.emit("keyup", "keyboard", emit)
}

// Focus returns an action emitting a focus event.
func (e *Emitter) Focus(emit FocusEmit) ActionInto[EmitReport] {
	return e.emit("focus", "focus", emit)
}

// Blur returns an action emitting a blur event.
func (e *Emitter) Blur(emit FocusEmit) ActionInto[EmitReport] {
	return e.emit("blur", "focus", emit)
}

// FocusIn returns an action emitting a focusin event.
func (e *Emitter) FocusIn(emit FocusEmit) ActionInto[EmitReport] {
	return e.emit("focusin", "focus_io", emit)
}

// FocusOut returns an action emitting a focusout event.
func (e *Emitter) FocusOut(emit FocusEmit) ActionInto[EmitReport] {
	return e.emit("focusout", "focus_io", emit)
}

// Input returns an action emitting an input event.
func (e *Emitter) Input(emit InputEmit) ActionInto[EmitReport] {
	return e.emit("input", "input", emit)
}

// Change returns an action emitting a change event.
func (e *Emitter) Change(emit ChangeEmit) ActionInto[EmitReport] {
	return e.emit("change", "change", emit)
}

// Submit returns an action emitting a submit event.
func (e *Emitter) Submit(emit SubmitEmit) ActionInto[EmitReport] {
	return e.emit("submit", "submit", emit)
}

func (e *Emitter) emit(kind string, capture string, data any) ActionInto[EmitReport] {
	return actionIntoFunc[EmitReport](func(ctx context.Context, core core.Core, gz bool) (action, error) {
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
