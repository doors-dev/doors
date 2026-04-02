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

	"github.com/doors-dev/doors/internal/door/pipe"
	"github.com/doors-dev/doors/internal/front/action"
)

type reportHook uint64

func (c reportHook) Params() action.CallParams {
	return action.CallParams{}
}

func (c reportHook) Action() (action.Action, bool) {
	return action.ReportHook{HookId: uint64(c)}, true
}

func (c reportHook) Cancel() {}

func (c reportHook) Result(r json.RawMessage, err error) {}

type callKind int

const (
	callReplace callKind = iota
	callUpdate
)

type call struct {
	ctx     context.Context
	task    *taskNode
	kind    callKind
	id      uint64
	payload pipe.Payload
}

func (n *call) Cancel() {
	n.payload.Release()
	n.send(context.Canceled)
}

func (n *call) Result(_ json.RawMessage, err error) {
	n.payload.Release()
	if err != nil {
		slog.Error("door rendering call failed", "error", err)
	}
	n.send(err)
}

func (n *call) send(err error) {
	if n.task == nil {
		return
	}
	n.task.Report(err)
}

func (c *call) Action() (action.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	payload := c.payload.Payload()
	switch c.kind {
	case callReplace:
		return action.DoorReplace{
			ID:      c.id,
			Payload: payload,
		}, true
	case callUpdate:
		return action.DoorUpdate{
			ID:      c.id,
			Payload: payload,
		}, true
	default:
		panic("unsupported door call type")
	}
}

func (c *call) Params() action.CallParams {
	return action.CallParams{}
}
