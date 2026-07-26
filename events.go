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
	"time"
)

// Rect is a rectangle in CSS pixels: origin X, Y and size Width, Height.
type Rect struct {
	// X is the horizontal origin.
	X float64 `json:"x"`
	// Y is the vertical origin.
	Y float64 `json:"y"`
	// Width is the horizontal size.
	Width float64 `json:"width"`
	// Height is the vertical size.
	Height float64 `json:"height"`
}

// PointerEvent is the payload of a browser pointer event.
//
// Geometry arrives as rectangles; the OffsetX, ClientX, PageX and ScreenX
// methods and their Y counterparts derive coordinates from them.
type PointerEvent struct {
	// Type is the DOM event type, such as click or pointerdown.
	Type string `json:"type"`
	// PointerID identifies the pointer that produced the event.
	PointerID int `json:"pointerId"`
	// Pressure is the normalized pointer pressure, from 0 to 1.
	Pressure float64 `json:"pressure"`
	// TangentialPressure is the normalized barrel pressure, from -1 to 1.
	TangentialPressure float64 `json:"tangentialPressure"`
	// TiltX is the pointer tilt along the X axis, in degrees from -90 to 90.
	TiltX float64 `json:"tiltX"`
	// TiltY is the pointer tilt along the Y axis, in degrees from -90 to 90.
	TiltY float64 `json:"tiltY"`
	// Twist is the clockwise rotation of the pointer, in degrees.
	Twist float64 `json:"twist"`
	// Buttons is the bitmask of the buttons currently held down.
	Buttons int `json:"buttons"`
	// Button is the button whose state changed, or -1 when none did.
	Button int `json:"button"`
	// PointerType is the device kind, such as mouse, pen, or touch.
	PointerType string `json:"pointerType"`
	// IsPrimary reports whether this is the primary pointer of its type.
	IsPrimary bool `json:"isPrimary"`
	// Pointer is the pointer offset inside the event target as X, Y and the
	// contact geometry as Width, Height.
	Pointer Rect `json:"pointer"`
	// Target is the viewport-relative bounding rectangle of the element the
	// attr is attached to.
	Target Rect `json:"target"`
	// Page is the document scroll offset as X, Y and the viewport size as
	// Width, Height.
	Page Rect `json:"page"`
	// Screen is the viewport offset on the screen as X, Y and the browser
	// window's outer size as Width, Height.
	Screen Rect `json:"screen"`
	// Timestamp is the client clock reading when the event was captured.
	Timestamp time.Time `json:"timestamp"`
}

// OffsetX returns the horizontal offset relative to the target element's
// padding edge.
func (e PointerEvent) OffsetX() float64 {
	return e.Pointer.X
}

// OffsetY returns the vertical offset relative to the target element's
// padding edge.
func (e PointerEvent) OffsetY() float64 {
	return e.Pointer.Y
}

// ClientX returns the horizontal coordinate relative to the viewport.
func (e PointerEvent) ClientX() float64 {
	return e.Pointer.X + e.Target.X
}

// ClientY returns the vertical coordinate relative to the viewport.
func (e PointerEvent) ClientY() float64 {
	return e.Pointer.Y + e.Target.Y
}

// PageX returns the horizontal coordinate relative to the document.
func (e PointerEvent) PageX() float64 {
	return e.ClientX() + e.Page.X
}

// PageY returns the vertical coordinate relative to the document.
func (e PointerEvent) PageY() float64 {
	return e.ClientY() + e.Page.Y
}

// ScreenX returns the horizontal coordinate relative to the screen.
func (e PointerEvent) ScreenX() float64 {
	return e.ClientX() + e.Screen.X
}

// ScreenY returns the vertical coordinate relative to the screen.
func (e PointerEvent) ScreenY() float64 {
	return e.ClientY() + e.Screen.Y
}

// PointerEmit is the event init for the synthetic pointer events [Emitter]
// dispatches.
type PointerEmit struct {
	// PointerID is the pointer identifier to report. Optional.
	PointerID int `json:"pointerId,omitempty"`
	// Buttons is the bitmask of buttons to report as held down. Optional.
	Buttons int `json:"buttons,omitempty"`
	// Button is the button to report as changed. Optional.
	Button int `json:"button,omitempty"`
	// PointerType is the device kind to report, such as mouse or pen. Optional.
	PointerType string `json:"pointerType,omitempty"`
	// IsPrimary reports the pointer as the primary one of its type. Optional.
	IsPrimary bool `json:"isPrimary,omitempty"`
}

// KeyboardEvent is the payload of a browser keyboard event.
type KeyboardEvent struct {
	// Type is the DOM event type: keydown or keyup.
	Type string `json:"type"`
	// Key is the value of the key pressed, such as a or Enter.
	Key string `json:"key"`
	// Code is the physical key code, such as KeyA.
	Code string `json:"code"`
	// Repeat reports whether the key was already down and is auto-repeating.
	Repeat bool `json:"repeat"`
	// CtrlKey reports whether Control was held down.
	CtrlKey bool `json:"ctrlKey"`
	// ShiftKey reports whether Shift was held down.
	ShiftKey bool `json:"shiftKey"`
	// AltKey reports whether Alt was held down.
	AltKey bool `json:"altKey"`
	// MetaKey reports whether Meta was held down.
	MetaKey bool `json:"metaKey"`
	// Timestamp is the client clock reading when the event was captured.
	Timestamp time.Time `json:"timestamp"`
}

// KeyboardEmit is the event init for the synthetic keyboard events [Emitter]
// dispatches. An event that matches no entry of [AKeyDown.Keys] or
// [AKeyUp.Keys] on the target attr sends no request.
type KeyboardEmit struct {
	// Key is the key value to report. Optional.
	Key string `json:"key,omitempty"`
	// Code is the physical key code to report. Optional.
	Code string `json:"code,omitempty"`
	// Repeat reports the key as auto-repeating. Optional.
	Repeat bool `json:"repeat,omitempty"`
	// CtrlKey reports Control as held down. Optional.
	CtrlKey bool `json:"ctrlKey,omitempty"`
	// ShiftKey reports Shift as held down. Optional.
	ShiftKey bool `json:"shiftKey,omitempty"`
	// AltKey reports Alt as held down. Optional.
	AltKey bool `json:"altKey,omitempty"`
	// MetaKey reports Meta as held down. Optional.
	MetaKey bool `json:"metaKey,omitempty"`
}

// FocusEvent is the payload of a browser focus event.
type FocusEvent struct {
	// Type is the DOM event type: focus, blur, focusin, or focusout.
	Type string `json:"type"`
	// Timestamp is the client clock reading when the event was captured.
	Timestamp time.Time `json:"timestamp"`
}

// FocusEmit is the event init for the synthetic focus events [Emitter]
// dispatches. Dispatching one runs the handlers without moving focus.
type FocusEmit struct{}

// ChangeEvent is the payload of a browser change event.
//
// The value fields are read from the element the event targets: Value always,
// Number, Date, Selected, and Checked when that element provides them.
type ChangeEvent struct {
	// Type is the DOM event type: change.
	Type string `json:"type"`
	// Name is the name attribute of the target element, empty if it has none.
	Name string `json:"name"`
	// Value is the current value of the target element.
	Value string `json:"value"`
	// Number is the value as a number, nil when it is not numeric.
	Number *float64 `json:"number"`
	// Date is the value as a date in UTC, nil when it is not a date.
	Date *time.Time `json:"date"`
	// Selected lists the values of the selected options, empty when the target
	// is not a select.
	Selected []string `json:"selected"`
	// Checked is the checked state of a checkbox or radio, false otherwise.
	Checked bool `json:"checked"`
	// Timestamp is the client clock reading when the event was captured.
	Timestamp time.Time `json:"timestamp"`
}

// ChangeEmit is the event init for the synthetic change events [Emitter]
// dispatches. Values are read from the event target, so attach the Emitter to
// the input or select element.
type ChangeEmit struct{}

// InputEvent is the payload of a browser input event.
//
// The value fields are read from the element the event targets. When
// [AInput.ExcludeValue] is set, only Type, Data, and Timestamp are filled.
type InputEvent struct {
	// Type is the DOM event type: input.
	Type string `json:"type"`
	// Name is the name attribute of the target element, empty if it has none.
	Name string `json:"name"`
	// Data is the text inserted by the edit, empty when the edit inserts none.
	Data string
	// Date is the value as a date in UTC, nil when it is not a date.
	Date *time.Time `json:"date"`
	// Value is the current value of the target element.
	Value string `json:"value"`
	// Number is the value as a number, nil when it is not numeric.
	Number *float64 `json:"number"`
	// Selected lists the values of the selected options, empty when the target
	// is not a select.
	Selected []string `json:"selected"`
	// Checked is the checked state of a checkbox or radio, false otherwise.
	Checked bool `json:"checked"`
	// Timestamp is the client clock reading when the event was captured.
	Timestamp time.Time `json:"timestamp"`
}

// InputEmit is the event init for the synthetic input events [Emitter]
// dispatches. Values are read from the event target, so attach the Emitter to
// the input or select element.
type InputEmit struct {
	// Data is the inserted text to report. Optional.
	Data string `json:"data,omitempty"`
}

// SubmitEmit is the event init for the synthetic submit events [Emitter]
// dispatches. The form data is read from the event target, so attach the
// Emitter to the form element.
type SubmitEmit struct{}
