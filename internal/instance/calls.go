// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package instance

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/front/action"
)

func (c *Instance[M]) CallCtx(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc {
	if ctx.Err() != nil {
		if onCancel != nil {
			onCancel()
		}
		return func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	call := &ctxCall{
		ctx:      ctx,
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

func (C *checkCall) Payload() ([]byte, action.PayloadType) {
	return nil, action.PayloadNone
}

func (c checkCall) Cancel() {
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *checkCall) Result(r json.RawMessage, err error) {
	if err != nil {
		slog.Error("Call failed", "action", c.action.Log(), "error", err)
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

type ctxCall struct {
	ctx      context.Context
	action   action.Action
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

func (c ctxCall) Cancel() {
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}

func (c *ctxCall) Result(r json.RawMessage, err error) {
	if err != nil {
		slog.Error("Call failed", "action", c.action.Log(), "error", err)
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
