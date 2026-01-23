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

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
)

type Hook struct {
	Before   action.Actions
	OnError  action.Actions
	Scope    []*ScopeSet
	Indicate []*Indicate
	core.Hook
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	a := []any{h.DoorID, h.HookID, h.Scope, h.Indicate, h.Before, h.OnError}
	return json.Marshal(a)
}
