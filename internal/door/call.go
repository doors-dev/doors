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


type seqCall struct {
	check    func() bool
	action   action.Action
	onResult func(json.RawMessage, error)
	onCancel func()
	params   action.CallParams
}

func (c *seqCall) Params() action.CallParams {
	return c.params
}

func (c *seqCall) Action() (action.Action, bool) {
	if !c.check() {
		return nil, false
	}
	return c.action, true
}

func (C *seqCall) Payload() common.Writable {
	return common.WritableNone{}
}

func (c seqCall) Cancel() {
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *seqCall) Result(r json.RawMessage, err error) {
	if err != nil {
		slog.Error("Call failed", slog.String("action", c.action.Log()), slog.String("error", err.Error()))
	}
	if c.onResult == nil {
		return
	}
	if err != nil {
		c.onResult(r, errors.Join(errors.New("execution error"), err))
		return
	}
	c.onResult(r, err)
}

func (c *seqCall) Clean() {}


type ctxCall struct {
	ctx      context.Context
	action   action.Action
	done     func()
	onResult func(json.RawMessage, error)
	onCancel func()
	params   action.CallParams
}

func (c *ctxCall) Params() action.CallParams {
	return c.params
}

func (c *ctxCall) Action() (action.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	return c.action, true
}

func (C *ctxCall) Payload() common.Writable {
	return common.WritableNone{}
}

func (c ctxCall) Cancel() {
	defer c.done()
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *ctxCall) Result(r json.RawMessage, err error) {
	if err != nil {
		slog.Error("Call failed", slog.String("action", c.action.Log()), slog.String("error", err.Error()))
	}
	defer c.done()
	if c.onResult == nil {
		return
	}
	if err != nil {
		c.onResult(r, errors.Join(errors.New("execution error"), err))
		return
	}
	c.onResult(r, err)
}
