// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

import (
	"encoding/json"
	"github.com/doors-dev/doors/internal/door"
)


type Hook struct {
	Error    []*ErrorAction
	Scope    []*ScopeSet
	Indicate []*Indicate
	*door.HookEntry
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	a := []any{h.DoorId, h.HookId, h.Scope, h.Indicate, h.Error}
	return json.Marshal(a)
}
