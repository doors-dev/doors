// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance2

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
)


func (c *Instance[M]) CallCtx(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc {
	if ctx.Err() != nil {
		if onCancel != nil {
			onCancel()
		}
		return func() {}
	}
	done := ctex.WgAdd(ctx)
	ctx, cancel := context.WithCancel(context.Background())
	call := &ctxCall{
		ctx:      ctx,
		done:     done,
		action:   action,
		onResult: onResult,
		onCancel: onCancel,
		params:   params,
	}
	c.solitaire.Call(call)
	return cancel
}

func (c *Instance[M]) CallCheck(check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	call := &checkCall{
		check:    check,
		action:   action,
		onResult: onResult,
		onCancel: onCancel,
		params:   params,
	}
	c.solitaire.Call(call)
	return 
}

type checkCall struct {
	check    func() bool
	action   action.Action
	onResult func(json.RawMessage, error)
	onCancel func()
	params   action.CallParams
}

func (c *checkCall) Params() action.CallParams {
	return c.params
}

func (c *checkCall) Action() (action.Action, bool) {
	if !c.check() {
		return nil, false
	}
	return c.action, true
}

func (C *checkCall) Payload() common.Writable {
	return common.WritableNone{}
}

func (c checkCall) Cancel() {
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *checkCall) Result(r json.RawMessage, err error) {
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

func (c *checkCall) Clean() {}

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

func (c *ctxCall) Clean() {}

type reportHook uint64

func (c reportHook) Params() action.CallParams {
	return action.CallParams{}
}

func (c reportHook) Action() (action.Action, bool) {
	return &action.ReportHook{HookId: uint64(c)}, true
}

func (C reportHook) Payload() common.Writable {
	return common.WritableNone{}
}

func (c reportHook) Cancel()                             {}
func (c reportHook) Result(r json.RawMessage, err error) {}
func (c reportHook) Clean()                              {}
