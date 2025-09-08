// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package door

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/front/action"
)

type doorCall struct {
	ctx        context.Context
	ch         chan error
	action     action.Action
	payload    common.Writable
	done       ctxwg.Done
	optimistic bool
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
	n.done()
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
	return action.CallParams{
		Optimistic: c.optimistic,
	}
}
