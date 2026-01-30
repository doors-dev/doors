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
)

func IsBlockingCtx(ctx context.Context) bool {
	blocking, ok := ctx.Value(KeyBlocking).(bool)
	if !ok {
		return false
	}
	return blocking
}

func SetBlockingCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, KeyBlocking, true)
}

func ClearBlockingCtx(ctx context.Context) context.Context {
	if !IsBlockingCtx(ctx) {
		return ctx
	}
	return context.WithValue(ctx, KeyBlocking, false)
}

func LogBlockingWarning(ctx context.Context, entity string, operation string) {
	if !IsBlockingCtx(ctx) {
		slog.Warn("Extended " + entity + " operation " + operation + " is used in non blocking context. Receiving from channel could lead to DEADLOCK under extreme conditions, please refer to documentation")
	}
}
