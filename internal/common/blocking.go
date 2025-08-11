package common

import (
	"context"
	"log/slog"
)


func IsBlockingCtx(ctx context.Context) bool {
    blocking, ok := ctx.Value(BlockingCtxKey).(bool)
    if !ok {
        return false
    }
    return blocking
}

func SetBlockingCtx(ctx context.Context) context.Context {
    return context.WithValue(ctx, BlockingCtxKey, true)
}

func ClearBlockingCtx(ctx context.Context) context.Context {
    if !IsBlockingCtx(ctx) {
        return ctx
    }
    return context.WithValue(ctx, BlockingCtxKey, false)
}

func LogBlockingWarning(ctx context.Context, entity string, operation string) {
	if !IsBlockingCtx(ctx) {
		slog.Warn("Extended "+entity+" operation "+operation+" is used in non blocking context. Receiving from channel could lead to DEADLOCK under extreme conditions, please refer to documentation")
	}
}
