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

package door

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/printer"
)

type reportHook uint64

func (c reportHook) Params() actions.CallParams {
	return actions.CallParams{}
}

func (c reportHook) Action() (actions.Action, bool) {
	return actions.ReportHook{HookId: uint64(c)}, true
}

func (c reportHook) Cancel() {}

func (c reportHook) Result(r json.RawMessage, err error) {}

type callKind int

const (
	callReplace callKind = iota
	callUpdate
	callFreeze
)

type call struct {
	ctx     context.Context
	task    *userTask
	kind    callKind
	id      uint64
	payload printer.Payload
	logger  *slog.Logger
}

func (n *call) Cancel() {
	n.payload.Release()
	n.send(context.Canceled)
}

func (n *call) Result(_ json.RawMessage, err error) {
	n.payload.Release()
	if err != nil {
		n.logger.Error("door rendering call failed", "error", err)
	}
	n.send(err)
}

func (n *call) send(err error) {
	n.task.Report(err)
}

func (c *call) Action() (actions.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	payload := c.payload.Payload()
	switch c.kind {
	case callReplace:
		return actions.DoorReplace{
			ID:      c.id,
			Payload: payload,
		}, true
	case callUpdate:
		return actions.DoorUpdate{
			ID:      c.id,
			Payload: payload,
		}, true
	case callFreeze:
		return actions.DoorFreeze{
			ID: c.id,
		}, true
	default:
		panic("unsupported door call type")
	}
}

func (c *call) Params() actions.CallParams {
	return actions.CallParams{}
}
