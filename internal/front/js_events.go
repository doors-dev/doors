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

package front

import "time"

// PointerEvent mirrors the browser PointerEvent payload sent to Doors.
type PointerEvent struct {
	Type               string    `json:"type"`
	PointerID          int       `json:"pointerId"`
	Width              float64   `json:"width"`
	Height             float64   `json:"height"`
	Pressure           float64   `json:"pressure"`
	TangentialPressure float64   `json:"tangentialPressure"`
	TiltX              float64   `json:"tiltX"`
	TiltY              float64   `json:"tiltY"`
	Twist              float64   `json:"twist"`
	Buttons            int       `json:"buttons"`
	Button             int       `json:"button"`
	PointerType        string    `json:"pointerType"`
	IsPrimary          bool      `json:"isPrimary"`
	ClientX            float64   `json:"clientX"`
	ClientY            float64   `json:"clientY"`
	ScreenX            float64   `json:"screenX"`
	ScreenY            float64   `json:"screenY"`
	PageX              float64   `json:"pageX"`
	PageY              float64   `json:"pageY"`
	Timestamp          time.Time `json:"timestamp"`
}

// KeyboardEvent mirrors the browser KeyboardEvent payload sent to Doors.
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

// FocusEvent mirrors the browser FocusEvent payload sent to Doors.
type FocusEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

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

/*
Future support
type TouchEvent struct {
	Type      string    `json:"type"`
	Touches   []Touch   `json:"touches"`
	ChangedTouches []Touch `json:"changedTouches"`
	TargetTouches  []Touch `json:"targetTouches"`
	Timestamp time.Time `json:"timestamp"`
}

type Touch struct {
	Identifier int     `json:"identifier"`
	ClientX    float64 `json:"clientX"`
	ClientY    float64 `json:"clientY"`
	ScreenX    float64 `json:"screenX"`
	ScreenY    float64 `json:"screenY"`
	PageX      float64 `json:"pageX"`
	PageY      float64 `json:"pageY"`
}

type ClipboardEvent struct {
	Type      string    `json:"type"`
	ClipboardData string `json:"clipboardData"`
	Timestamp time.Time `json:"timestamp"`
}

type DragEvent struct {
	Type      string    `json:"type"`
	ClientX   float64   `json:"clientX"`
	ClientY   float64   `json:"clientY"`
	ScreenX   float64   `json:"screenX"`
	ScreenY   float64   `json:"screenY"`
	PageX     float64   `json:"pageX"`
	PageY     float64   `json:"pageY"`
	Timestamp time.Time `json:"timestamp"`
}
*/
