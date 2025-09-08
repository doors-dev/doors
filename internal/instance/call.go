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
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/front/action"
)

func (c *core[M]) SimpleCall(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc {
	if ctx.Err() != nil {
		if onCancel != nil {
			onCancel()
		}
		return func() {}
	}
	done := ctxwg.Add(ctx)
	ctx, cancel := context.WithCancel(context.Background())
	call := &SimpleCall{
		ctx:      ctx,
		done:     done,
		action:   action,
		onResult: onResult,
		onCancel: onCancel,
		params:   params,
	}
	c.Call(call)
	return cancel
}

type SimpleCall struct {
	ctx      context.Context
	action   action.Action
	done     func()
	onResult func(json.RawMessage, error)
	onCancel func()
	params   action.CallParams
}

func (c *SimpleCall) Params() action.CallParams {
	return c.params
}

func (c *SimpleCall) Action() (action.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	return c.action, true
}

func (C *SimpleCall) Payload() common.Writable {
	return common.WritableNone{}
}

func (c SimpleCall) Cancel() {
	defer c.done()
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *SimpleCall) Result(r json.RawMessage, err error) {
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

func (c *SimpleCall) Clean() {}
