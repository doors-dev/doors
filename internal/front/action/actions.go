// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package action

import "time"

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
	Smooth   bool
}

func (a Scroll) Log() string {
	return "scroll"
}
func (a Scroll) Invocation() Invocation {
	return Invocation{
		name: "scroll",
		arg:  []any{a.Selector, a.Smooth},
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
	Name   string
	Arg    any
	DoorID uint64
}

func (a Emit) Log() string {
	return "emit: " + a.Name
}
func (a Emit) Invocation() Invocation {
	return Invocation{
		name: "emit",
		arg:  []any{a.Name, a.Arg, a.DoorID},
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
	ID uint64
}

func (a DoorReplace) Log() string {
	return "door_replace"
}
func (a DoorReplace) Invocation() Invocation {
	return Invocation{
		name: "door_replace",
		arg:  []any{a.ID},
	}
}

type DoorUpdate struct {
	ID uint64
}

func (a DoorUpdate) Log() string {
	return "door_update"
}
func (a DoorUpdate) Invocation() Invocation {
	return Invocation{
		name: "door_update",
		arg:  []any{a.ID},
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
