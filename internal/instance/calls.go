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

	"github.com/doors-dev/doors/internal/front/actions"
)

func (c Instance) UserCall(ctx context.Context, check func() bool, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	logger := c.Logger()
	call := &call{
		ctx:      ctx,
		action:   action,
		check:    check,
		onResult: onResult,
		onCancel: onCancel,
		params:   params,
		logger:   logger,
	}
	c.solitaire.Call(call)
}

type checkCall struct {
	check    func() bool
	action   actions.Action
	onResult func(json.RawMessage, error)
	onCancel func()
	params   actions.CallParams
	logger   *slog.Logger
}

func (c *checkCall) Params() actions.CallParams {
	return c.params
}

func (c *checkCall) Action() (actions.Action, bool) {
	if !c.check() {
		return nil, false
	}
	return c.action, true
}

func (C *checkCall) Payload() ([]byte, actions.PayloadType) {
	return nil, actions.PayloadNone
}

func (c checkCall) Cancel() {
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *checkCall) Result(r json.RawMessage, err error) {
	if err != nil {
		c.logger.Error("Call failed", "action", c.action.Log(), "error", err)
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

type call struct {
	ctx      context.Context
	check    func() bool
	action   actions.Action
	onResult func(json.RawMessage, error)
	onCancel func()
	params   actions.CallParams
	logger   *slog.Logger
}

func (c *call) Params() actions.CallParams {
	return c.params
}

func (c *call) canceled() bool {
	if c.check != nil {
		return !c.check()
	}
	return c.ctx.Err() != nil
}

func (c *call) Action() (actions.Action, bool) {
	if c.canceled() {
		return nil, false
	}
	return c.action, true
}

func (c call) Cancel() {
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}

func (c *call) Result(r json.RawMessage, err error) {
	if err != nil {
		c.logger.Error("Call failed", "action", c.action.Log(), "error", err)
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
