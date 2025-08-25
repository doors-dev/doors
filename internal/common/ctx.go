// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import (
	"context"
	"log/slog"
)
type ctxKey int

const (
    InstanceCtxKey ctxKey = iota
    DoorCtxKey
    ThreadCtxKey
    RenderMapCtxKey
    BlockingCtxKey
    AdaptersCtxKey
    SessionStoreCtxKey
    InstanceStoreCtxKey
    ParentCtxKey
    AttrsCtxKey
)

func ResultChannel(ctx context.Context, action string) (chan error, bool) {
	ch := make(chan error, 1)
	if ctx.Err() != nil {
		ch <- ctx.Err()
		slog.Warn("Tried to perfrom "+action+" from canceled context")
		close(ch)
		return ch, false
	}
	return ch, true
}

