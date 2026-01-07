// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package front

import (
	"encoding/json"

	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front/action"
)

type Hook struct {
	Before   action.Actions
	OnError  action.Actions
	Scope    []*ScopeSet
	Indicate []*Indicate
	*door.HookEntry
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	a := []any{h.DoorId, h.HookId, h.Scope, h.Indicate, h.Before, h.OnError}
	return json.Marshal(a)
}
