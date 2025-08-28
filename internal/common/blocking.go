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


func IsBlockingCtx(ctx context.Context) bool {
    blocking, ok := ctx.Value(CtxKeyBlocking).(bool)
    if !ok {
        return false
    }
    return blocking
}

func SetBlockingCtx(ctx context.Context) context.Context {
    return context.WithValue(ctx, CtxKeyBlocking, true)
}

func ClearBlockingCtx(ctx context.Context) context.Context {
    if !IsBlockingCtx(ctx) {
        return ctx
    }
    return context.WithValue(ctx, CtxKeyBlocking, false)
}

func LogBlockingWarning(ctx context.Context, entity string, operation string) {
	if !IsBlockingCtx(ctx) {
		slog.Warn("Extended "+entity+" operation "+operation+" is used in non blocking context. Receiving from channel could lead to DEADLOCK under extreme conditions, please refer to documentation")
	}
}
