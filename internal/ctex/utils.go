package ctex

import (
	"context"
	"log/slog"
)

func LogCanceled(ctx context.Context, action string) {
	if ctx.Err() == nil {
		return
	}
	slog.Warn("Tried to perfrom " + action + " from the canceled context")
}
