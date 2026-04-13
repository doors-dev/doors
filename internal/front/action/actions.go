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

package action

import (
	"time"
)

type LocationReload struct{}

func (a LocationReload) Log() string {
	return "location_reload"
}
func (a LocationReload) Invocation() Invocation {
	return Invocation{
		name: "location_reload",
		arg:  []any{},
	}
}

type LocationReplace struct {
	URL    string
	Origin bool
}

func (a LocationReplace) Log() string {
	return "location_replace"
}
func (a LocationReplace) Invocation() Invocation {
	return Invocation{
		name: "location_replace",
		arg:  []any{a.URL, a.Origin},
	}
}

type Scroll struct {
	Selector string
	Options  any
}

func (a Scroll) Log() string {
	return "scroll"
}
func (a Scroll) Invocation() Invocation {
	return Invocation{
		name: "scroll",
		arg:  []any{a.Selector, a.Options},
	}
}

type LocationAssign struct {
	URL    string
	Origin bool
}

func (a LocationAssign) Log() string {
	return "location_assign"
}
func (a LocationAssign) Invocation() Invocation {
	return Invocation{
		name: "location_assign",
		arg:  []any{a.URL, a.Origin},
	}
}

type Emit struct {
	Name    string
	DoorID  uint64
	Payload Payload
}

func (a Emit) Log() string {
	return "emit: " + a.Name
}
func (a Emit) Invocation() Invocation {
	return Invocation{
		name:    "emit",
		arg:     []any{a.Name, a.DoorID},
		payload: a.Payload,
	}
}

type DynaSet struct {
	ID    uint64
	Value string
}

func (a DynaSet) Log() string {
	return "dyna_set"
}
func (a DynaSet) Invocation() Invocation {
	return Invocation{
		name: "dyna_set",
		arg:  []any{a.ID, a.Value},
	}
}

type DynaRemove struct {
	ID uint64
}

func (a DynaRemove) Log() string {
	return "dyna_remove"
}
func (a DynaRemove) Invocation() Invocation {
	return Invocation{
		name: "dyna_remove",
		arg:  []any{a.ID},
	}
}

type SetPath struct {
	Path    string
	Replace bool
}

func (a SetPath) Log() string {
	return "set_path"
}
func (a SetPath) Invocation() Invocation {
	return Invocation{
		name: "set_path",
		arg:  []any{a.Path, a.Replace},
	}
}

type DoorReplace struct {
	ID      uint64
	Payload Payload
}

func (a DoorReplace) Log() string {
	return "door_replace"
}
func (a DoorReplace) Invocation() Invocation {
	return Invocation{
		name:    "door_replace",
		arg:     []any{a.ID},
		payload: a.Payload,
	}
}

type DoorUpdate struct {
	ID      uint64
	Payload Payload
}

func (a DoorUpdate) Log() string {
	return "door_update"
}
func (a DoorUpdate) Invocation() Invocation {
	return Invocation{
		name:    "door_update",
		arg:     []any{a.ID},
		payload: a.Payload,
	}
}

type Indicate struct {
	Duration time.Duration
	Indicate any
}

func (a Indicate) Log() string {
	return "indicate"
}
func (a Indicate) Invocation() Invocation {
	return Invocation{
		name: "indicate",
		arg:  []any{a.Duration.Milliseconds(), a.Indicate},
	}
}

type ReportHook struct {
	HookId uint64
}

func (a ReportHook) Log() string {
	return "report hook"
}
func (a ReportHook) Invocation() Invocation {
	return Invocation{
		name: "report_hook",
		arg:  []any{a.HookId},
	}
}

type UpdateTitle struct {
	Content string
	Attrs   map[string]string
}

func (u UpdateTitle) Log() string {
	return "update_title"
}

func (u UpdateTitle) Invocation() Invocation {
	return Invocation{
		name: "update_title",
		arg:  []any{u.Content, u.Attrs},
	}
}

type UpdateMeta struct {
	Name     string
	Property bool
	Attrs    map[string]string
}

func (u UpdateMeta) Log() string {
	return "update_meta"
}

func (u UpdateMeta) Invocation() Invocation {
	return Invocation{
		name: "update_meta",
		arg:  []any{u.Name, u.Property, u.Attrs},
	}
}

type Test struct {
	Arg any
}

func (a Test) Log() string {
	return "test"
}
func (a Test) Invocation() Invocation {
	return Invocation{
		name: "test",
		arg:  []any{a.Arg},
	}
}
