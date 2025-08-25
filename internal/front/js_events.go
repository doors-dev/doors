// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

import "time"

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

type FocusEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}


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
type InputEvent struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Data      string
	Date      *time.Time `json:"date"`
	Value     string   `json:"value"`
	Number    *float64 `json:"number"`
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
