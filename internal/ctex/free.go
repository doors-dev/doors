// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package ctex

import (
	"context"
	"log/slog"
	"time"
)

func NewFreeContext(ctx context.Context, runtime context.Context) context.Context {
	ctx = FrameRemove(ctx)
	return freeContext{ctx, runtime}
}

type freeContext struct {
	ctx     context.Context
	runtime context.Context
}

func (f freeContext) Deadline() (deadline time.Time, ok bool) {
	return f.runtime.Deadline()
}

func (f freeContext) Done() <-chan struct{} {
	return f.runtime.Done()
}

func (f freeContext) Err() error {
	return f.runtime.Err()
}

func (f freeContext) Value(key any) any {
	return f.ctx.Value(key)
}

var _ context.Context = freeContext{}

type core interface {
	RootCore()
}

func IsFreeCtx(ctx context.Context) bool {
	_, ok := ctx.(freeContext)
	return ok
}

func ClearFreeCtx(ctx context.Context) context.Context {
	fc, ok := ctx.(freeContext)
	if !ok {
		return ctx
	}
	return ClearFreeCtx(fc.ctx)
}

func LogFreeWarning(ctx context.Context, entity string, operation string) {
	if !IsFreeCtx(ctx) {
		slog.Warn(
			"extended operation is used in non-free context. Receiving from channel could lead to DEADLOCK under extreme conditions, please refer to documentation",
			"entity",
			entity,
			"operation",
			operation,
		)
	}
}
