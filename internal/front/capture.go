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

type Capture interface {
	Name() string
	Listen() string
}

type PointerCapture struct {
	Event           string `json:"-"`
	PreventDefault  bool   `json:"pd"`
	StopPropagation bool   `json:"sp"`
	ExactTarget     bool   `json:"et"`
}

func (pc PointerCapture) Name() string {
	return "pointer"
}

func (pc PointerCapture) Listen() string {
	return pc.Event
}

type KeyboardEventCapture struct {
	Event           string   `json:"-"`
	Filter          []string `json:"fr"`
	PreventDefault  bool     `json:"pd"`
	StopPropagation bool     `json:"sp"`
	ExactTarget     bool     `json:"et"`
}

func (c KeyboardEventCapture) Name() string {
	return "keyboard"
}

func (c KeyboardEventCapture) Listen() string {
	return c.Event
}

type FormCapture struct {
}

func (c FormCapture) Name() string {
	return "submit"
}

func (c FormCapture) Listen() string {
	return "submit"
}

type LinkCapture struct {
	StopPropagation bool `json:"sp"`
}

func (c LinkCapture) Name() string {
	return "link"
}

func (c LinkCapture) Listen() string {
	return "click"
}

type FocusCapture struct {
	Event string `json:"-"`
}

func (c FocusCapture) Name() string {
	return "focus"
}

func (c FocusCapture) Listen() string {
	return c.Event
}

type FocusIOCapture struct {
	Event           string `json:"-"`
	StopPropagation bool   `json:"sp"`
	ExactTarget     bool   `json:"et"`
}

func (c FocusIOCapture) Name() string {
	return "focus_io"
}

func (c FocusIOCapture) Listen() string {
	return c.Event
}

type ChangeCapture struct {
}

func (c ChangeCapture) Name() string {
	return "change"
}

func (c ChangeCapture) Listen() string {
	return "change"
}

type InputCapture struct {
	ExcludeValue bool `json:"ev"`
}

func (c InputCapture) Name() string {
	return "input"
}

func (c InputCapture) Listen() string {
	return "input"
}
