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
	"sync"

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
	once sync.Once
	id   uint64
}

func (e *Emitter) autoID(c core.Core) uint64 {
	e.once.Do(func() {
		e.id = c.Instance().NewID()
	})
	return e.id
}

func (e *Emitter) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(e, cur, elem)
}

func (e *Emitter) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	front.AttrsSetEmitter(attrs, e.autoID(core))
	front.AttrsSetParent(attrs, core.Door().ID())
	return nil
}

// Click returns an [Action] emitting a click event.
func (e *Emitter) Click(init PointerAction) Action {
	return e.emitAction("click", "pointer", init)
}

// PointerDown returns an [Action] emitting a pointerdown event.
func (e *Emitter) PointerDown(init PointerAction) Action {
	return e.emitAction("pointerdown", "pointer", init)
}

// PointerUp returns an [Action] emitting a pointerup event.
func (e *Emitter) PointerUp(init PointerAction) Action {
	return e.emitAction("pointerup", "pointer", init)
}

// PointerMove returns an [Action] emitting a pointermove event.
func (e *Emitter) PointerMove(init PointerAction) Action {
	return e.emitAction("pointermove", "pointer", init)
}

// PointerOver returns an [Action] emitting a pointerover event.
func (e *Emitter) PointerOver(init PointerAction) Action {
	return e.emitAction("pointerover", "pointer", init)
}

// PointerOut returns an [Action] emitting a pointerout event.
func (e *Emitter) PointerOut(init PointerAction) Action {
	return e.emitAction("pointerout", "pointer", init)
}

// PointerEnter returns an [Action] emitting a pointerenter event.
func (e *Emitter) PointerEnter(init PointerAction) Action {
	return e.emitAction("pointerenter", "pointer", init)
}

// PointerLeave returns an [Action] emitting a pointerleave event.
func (e *Emitter) PointerLeave(init PointerAction) Action {
	return e.emitAction("pointerleave", "pointer", init)
}

// PointerCancel returns an [Action] emitting a pointercancel event.
func (e *Emitter) PointerCancel(init PointerAction) Action {
	return e.emitAction("pointercancel", "pointer", init)
}

// GotPointerCapture returns an [Action] emitting a gotpointercapture event.
func (e *Emitter) GotPointerCapture(init PointerAction) Action {
	return e.emitAction("gotpointercapture", "pointer", init)
}

// LostPointerCapture returns an [Action] emitting a lostpointercapture event.
func (e *Emitter) LostPointerCapture(init PointerAction) Action {
	return e.emitAction("lostpointercapture", "pointer", init)
}

// KeyDown returns an [Action] emitting a keydown event.
func (e *Emitter) KeyDown(init KeyboardAction) Action {
	return e.emitAction("keydown", "keyboard", init)
}

// KeyUp returns an [Action] emitting a keyup event.
func (e *Emitter) KeyUp(init KeyboardAction) Action {
	return e.emitAction("keyup", "keyboard", init)
}

// Focus returns an [Action] emitting a focus event.
func (e *Emitter) Focus(init FocusAction) Action {
	return e.emitAction("focus", "focus", init)
}

// Blur returns an [Action] emitting a blur event.
func (e *Emitter) Blur(init FocusAction) Action {
	return e.emitAction("blur", "focus", init)
}

// FocusIn returns an [Action] emitting a focusin event.
func (e *Emitter) FocusIn(init FocusAction) Action {
	return e.emitAction("focusin", "focus_io", init)
}

// FocusOut returns an [Action] emitting a focusout event.
func (e *Emitter) FocusOut(init FocusAction) Action {
	return e.emitAction("focusout", "focus_io", init)
}

// Input returns an [Action] emitting an input event.
func (e *Emitter) Input(init InputAction) Action {
	return e.emitAction("input", "input", init)
}

// Change returns an [Action] emitting a change event.
func (e *Emitter) Change(init ChangeAction) Action {
	return e.emitAction("change", "change", init)
}

// Submit returns an [Action] emitting a submit event.
func (e *Emitter) Submit(init SubmitAction) Action {
	return e.emitAction("submit", "submit", init)
}

// Emit returns an [Action] emitting an arbitrary event with init.
//
// Known event types use the matching event constructor; others dispatch a
// CustomEvent with init as detail.
func (e *Emitter) Emit(event string, init map[string]any) Action {
	return e.emitAction(event, emitCaptures[event], init)
}

func (e *Emitter) emitAction(event string, capture string, data any) Action {
	return actionEmitEvent{
		emitter: e,
		event:   event,
		capture: capture,
		data:    data,
	}
}

var emitCaptures = map[string]string{
	"click":              "pointer",
	"pointerdown":        "pointer",
	"pointerup":          "pointer",
	"pointermove":        "pointer",
	"pointerover":        "pointer",
	"pointerout":         "pointer",
	"pointerenter":       "pointer",
	"pointerleave":       "pointer",
	"pointercancel":      "pointer",
	"gotpointercapture":  "pointer",
	"lostpointercapture": "pointer",
	"keydown":            "keyboard",
	"keyup":              "keyboard",
	"focus":              "focus",
	"blur":               "focus",
	"focusin":            "focus_io",
	"focusout":           "focus_io",
	"input":              "input",
	"change":             "change",
	"submit":             "submit",
}

type actionEmitEvent struct {
	emitter *Emitter
	event   string
	capture string
	data    any
}

func (ae actionEmitEvent) action(ctx context.Context, core core.Core, gz bool) (action.Action, action.CallParams, error) {
	payload, err := action.IntoPayload(ae.data, gz)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	act := action.EmitEvent{
		EmitterID: ae.emitter.autoID(core),
		Type:      ae.event,
		Capture:   ae.capture,
		Payload:   payload,
	}
	return act, action.CallParams{}, nil
}

var _ Action = actionEmitEvent{}
