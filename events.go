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

// Rect describes a rectangle in a coordinate space: origin (X, Y) and
// dimensions (W, H). Used by [PointerEvent] to carry target,
// page, and screen geometry needed to derive client, page, and screen
// coordinates from offset values.
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// PointerEvent is the browser pointer event payload sent to Doors.
//
// Pointer carries the pointer's offset position within the target element
// and the contact geometry size. Target, Page, and Screen carry the source
// geometry used to derive client, page, and screen coordinates via the
// helper methods below.
type PointerEvent struct {
	Type               string    `json:"type"`
	PointerID          int       `json:"pointerId"`
	Pressure           float64   `json:"pressure"`
	TangentialPressure float64   `json:"tangentialPressure"`
	TiltX              float64   `json:"tiltX"`
	TiltY              float64   `json:"tiltY"`
	Twist              float64   `json:"twist"`
	Buttons            int       `json:"buttons"`
	Button             int       `json:"button"`
	PointerType        string    `json:"pointerType"`
	IsPrimary          bool      `json:"isPrimary"`
	Pointer            Rect      `json:"pointer"`
	Target             Rect      `json:"target"`
	Page               Rect      `json:"page"`
	Screen             Rect      `json:"screen"`
	Timestamp          time.Time `json:"timestamp"`
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
//
//	clientX = offsetX + target.X
func (e PointerEvent) ClientX() float64 {
	return e.Pointer.X + e.Target.X
}

// ClientY returns the vertical coordinate relative to the viewport.
//
//	clientY = offsetY + target.Y
func (e PointerEvent) ClientY() float64 {
	return e.Pointer.Y + e.Target.Y
}

// PageX returns the horizontal coordinate relative to the document.
//
//	pageX = clientX + page.X  (page.X is window.scrollX)
func (e PointerEvent) PageX() float64 {
	return e.ClientX() + e.Page.X
}

// PageY returns the vertical coordinate relative to the document.
//
//	pageY = clientY + page.Y  (page.Y is window.scrollY)
func (e PointerEvent) PageY() float64 {
	return e.ClientY() + e.Page.Y
}

// ScreenX returns the horizontal coordinate relative to the screen.
//
//	screenX = clientX + screen.X  (screen.X is window.screenX)
func (e PointerEvent) ScreenX() float64 {
	return e.ClientX() + e.Screen.X
}

// ScreenY returns the vertical coordinate relative to the screen.
//
//	screenY = clientY + screen.Y  (screen.Y is window.screenY)
func (e PointerEvent) ScreenY() float64 {
	return e.ClientY() + e.Screen.Y
}

// PointerEmit is the synthetic pointer event init an [Emitter] sends to the client.
type PointerEmit struct {
	PointerID   int     `json:"pointerId,omitempty"`
	Buttons     int     `json:"buttons,omitempty"`
	Button      int     `json:"button,omitempty"`
	PointerType string  `json:"pointerType,omitempty"`
	IsPrimary   bool    `json:"isPrimary,omitempty"`
	ClientX     float64 `json:"clientX,omitempty"`
	ClientY     float64 `json:"clientY,omitempty"`
	CtrlKey     bool    `json:"ctrlKey,omitempty"`
	ShiftKey    bool    `json:"shiftKey,omitempty"`
	AltKey      bool    `json:"altKey,omitempty"`
	MetaKey     bool    `json:"metaKey,omitempty"`
}

// KeyboardEvent is the browser keyboard event payload sent to Doors.
type KeyboardEvent struct {
	Type      string    `json:"type"`
	Key       string    `json:"key"`
	Code      string    `json:"code"`
	Repeat    bool      `json:"repeat"`
	CtrlKey   bool      `json:"ctrlKey"`
	ShiftKey  bool      `json:"shiftKey"`
	AltKey    bool      `json:"altKey"`
	MetaKey   bool      `json:"metaKey"`
	Timestamp time.Time `json:"timestamp"`
}

// KeyboardEmit is the synthetic keyboard event init an [Emitter] sends to the client.
type KeyboardEmit struct {
	Key      string `json:"key,omitempty"`
	Code     string `json:"code,omitempty"`
	Repeat   bool   `json:"repeat,omitempty"`
	CtrlKey  bool   `json:"ctrlKey,omitempty"`
	ShiftKey bool   `json:"shiftKey,omitempty"`
	AltKey   bool   `json:"altKey,omitempty"`
	MetaKey  bool   `json:"metaKey,omitempty"`
}

// FocusEvent is the browser focus event payload sent to Doors.
type FocusEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// FocusEmit is the synthetic focus event init an [Emitter] sends to the client.
type FocusEmit struct{}

// ChangeEvent describes a committed form value change.
type ChangeEvent struct {
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Value     string     `json:"value"`
	Number    *float64   `json:"number"`
	Date      *time.Time `json:"date"`
	Selected  []string   `json:"selected"`
	Checked   bool       `json:"checked"`
	Timestamp time.Time  `json:"timestamp"`
}

// ChangeEmit is the synthetic change event init an [Emitter] sends to the client.
type ChangeEmit struct{}

// InputEvent describes a live form value edit.
type InputEvent struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Data      string
	Date      *time.Time `json:"date"`
	Value     string     `json:"value"`
	Number    *float64   `json:"number"`
	Selected  []string   `json:"selected"`
	Checked   bool       `json:"checked"`
	Timestamp time.Time  `json:"timestamp"`
}

// InputEmit is the synthetic input event init an [Emitter] sends to the client.
type InputEmit struct {
	Data      string `json:"data,omitempty"`
	InputType string `json:"inputType,omitempty"`
}

// SubmitEmit is the synthetic submit event init an [Emitter] sends to the client.
type SubmitEmit struct{}
