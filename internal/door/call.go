// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package door

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
)

type doorCall struct {
	ctx        context.Context
	ch         chan error
	action     action.Action
	payload    common.Writable
}

func (n *doorCall) Clean() {
	if n.payload != nil {
		n.payload.Destroy()
	}
}

func (n *doorCall) Cancel() {
	n.send(context.Canceled)
}

func (n *doorCall) Result(_ json.RawMessage, err error) {
	if err != nil {
		slog.Error("door rendering error", slog.String("error", err.Error()))
	}
	n.send(err)
}

func (n *doorCall) send(err error) {
	n.ch <- err
	close(n.ch)
}

func (c *doorCall) Action() (action.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	return c.action, true
}

func (c *doorCall) Payload() common.Writable {
	return c.payload
}

func (c *doorCall) Params() action.CallParams {
	return action.CallParams{}
}
