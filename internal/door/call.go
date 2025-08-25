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
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
)

type doorCall struct {
	ctx     context.Context
	name    string
	ch      chan error
	arg     any
	payload common.Writable
	done    ctxwg.Done
}

func (n *doorCall) stale() {
	n.Result(errors.New("stale"))
}

func (n *doorCall) Result(err error) {
	if err != nil {
		slog.Error("Door call failed", slog.String("call_name", n.name), slog.String("error", err.Error()))
	}
	n.ch <- err
	close(n.ch)
	if n.payload != nil {
		n.payload.Destroy()
	}
	n.done()
}

func (n *doorCall) Data() *common.CallData {
	if n.ctx.Err() != nil {
		n.stale()
		return nil
	}
	return &common.CallData{
		Name:    n.name,
		Arg:     n.arg,
		Payload: n.payload,
	}
}
