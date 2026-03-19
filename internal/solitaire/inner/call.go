// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package inner

import (
	"encoding/json"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/front/action"
)

type Call struct {
	Call     action.Call
	Params   action.CallParams
	reported atomic.Bool
}

func (p *Call) Written() {
	if !p.Params.Optimistic {
		return
	}
	p.Result([]byte("null"), nil)
}

func (c *Call) Action() (action.Action, bool) {
	return c.Call.Action()
}

func (c *Call) Cancel() {
	if c.reported.Swap(true) {
		return
	}
	c.Call.Cancel()
}

func (c *Call) Result(ok json.RawMessage, err error) {
	if c.reported.Swap(true) {
		return
	}
	c.Call.Result(ok, err)
}
