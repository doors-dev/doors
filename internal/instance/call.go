// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"encoding/json"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
)

type setPathCall struct {
	path     string
	replace  bool
	canceled atomic.Bool
}

func (t *setPathCall) cancel() {
	t.canceled.Store(true)
}

func (t *setPathCall) Data() *common.CallData {
	if t.canceled.Load() {
		return nil
	}
	return &common.CallData{
		Name:    "set_path",
		Arg:     t.arg(),
		Payload: common.WritableNone{},
	}
}

func (t *setPathCall) arg() []any {
	return []any{t.path, t.replace}
}

func (t *setPathCall) Result(json.RawMessage, error) {}
func (t *setPathCall) Cancel()                       {}

type LocatinReload struct {
}

func (l *LocatinReload) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_reload",
		Arg:     []any{},
		Payload: common.WritableNone{},
	}
}
func (t *LocatinReload) Result(json.RawMessage, error) {}
func (t *LocatinReload) Cancel()                       {}

type LocationReplace struct {
	Href   string
	Origin bool
}

func (l *LocationReplace) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_replace",
		Arg:     []any{l.Href, l.Origin},
		Payload: common.WritableNone{},
	}
}
func (t *LocationReplace) Result(json.RawMessage, error) {}
func (t *LocationReplace) Cancel()                       {}

type LocationAssign struct {
	Href   string
	Origin bool
}

func (l *LocationAssign) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_assign",
		Arg:     []any{l.Href, l.Origin},
		Payload: common.WritableNone{},
	}
}

func (t *LocationAssign) Result(json.RawMessage, error) {}
func (t *LocationAssign) Cancel()                       {}
