// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

type Capture interface {
	Name() string
	Listen() string
}

type PointerCapture struct {
	Event           string `json:"-"`
	PreventDefault  bool   `json:"preventDefault"`
	StopPropagation bool   `json:"stopPropagation"`
	ExactTarget     bool   `json:"exactTarget"`
}

func (pc *PointerCapture) Name() string {
	return "pointer"
}

func (pc *PointerCapture) Listen() string {
	return pc.Event
}

type KeyboardEventCapture struct {
	Event           string `json:"-"`
	PreventDefault  bool   `json:"preventDefault"`
	StopPropagation bool   `json:"stopPropagation"`
	ExactTarget     bool   `json:"exactTarget"`
}

func (c *KeyboardEventCapture) Name() string {
	return "keyboard"
}

func (c *KeyboardEventCapture) Listen() string {
	return c.Event
}

type FormCapture struct {
}

func (c *FormCapture) Name() string {
	return "submit"
}

func (c *FormCapture) Listen() string {
	return "submit"
}

type LinkCapture struct {
	StopPropagation bool `json:"stopPropagation"`
}

func (c *LinkCapture) Name() string {
	return "link"
}

func (c *LinkCapture) Listen() string {
	return "click"
}

type FocusCapture struct {
	Event string `json:"-"`
}

func (c *FocusCapture) Name() string {
	return "focus"
}

func (c *FocusCapture) Listen() string {
	return c.Event
}

type FocusIOCapture struct {
	Event           string `json:"-"`
	StopPropagation bool   `json:"stopPropagation"`
	ExactTarget     bool   `json:"exactTarget"`
}

func (c *FocusIOCapture) Name() string {
	return "focus_io"
}

func (c *FocusIOCapture) Listen() string {
	return c.Event
}

type ChangeCapture struct {
}

func (c *ChangeCapture) Name() string {
	return "change"
}

func (c *ChangeCapture) Listen() string {
	return "change"
}

type InputCapture struct {
	ExcludeValue bool `json:"excludeValue"`
}

func (c *InputCapture) Name() string {
	return "input"
}

func (c *InputCapture) Listen() string {
	return "input"
}
