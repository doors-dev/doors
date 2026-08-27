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
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
)

func (c Instance) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	call := &call{
		ctx:      ctx,
		onResult: onResult,
		onCancel: onCancel,
		params:   params,
		logger:   c.Logger(),
	}
	call.action.Store(&action)
	call.stop = context.AfterFunc(ctx, func() {
		call.action.Store(nil)
	})
	c.solitaire.Call(call)
}

func (c Instance) UserCallCheck(check func() bool, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	call := &call{
		ctx:      context.Background(),
		check:    check,
		onResult: onResult,
		onCancel: onCancel,
		params:   params,
		logger:   c.Logger(),
	}
	call.action.Store(&action)
	c.solitaire.Call(call)
}

type call struct {
	ctx      context.Context
	check    func() bool
	action   atomic.Pointer[actions.Action]
	stop     func() bool
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

func (c *call) finish() *actions.Action {
	if c.stop != nil {
		c.stop()
	}
	return c.action.Swap(nil)
}

func (c *call) Action() (actions.Action, func(), bool) {
	if c.canceled() {
		return nil, nil, false
	}
	action := c.action.Load()
	if action == nil {
		return nil, nil, false
	}
	return *action, func() {}, true
}

func (c *call) Cancel() {
	c.finish()
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}

func (c *call) Result(r json.RawMessage, err error) {
	action := c.finish()
	if err != nil {
		if action != nil {
			c.logger.Error("Call failed", "action", (*action).Log(), "error", err)
		} else {
			c.logger.Error("Call failed", "error", err)
		}
	}
	if c.onResult == nil {
		return
	}
	if err != nil && !errors.Is(err, common.ErrTerminated) {
		err = fmt.Errorf("%w: %w", common.ErrExecution, err)
	}
	c.onResult(r, err)
}
