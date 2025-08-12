package common

import (
	"context"
	"log/slog"
)
type ctxKey int

const (
    InstanceCtxKey ctxKey = iota
    NodeCtxKey
    ThreadCtxKey
    RenderMapCtxKey
    BlockingCtxKey
    AdaptersCtxKey
    SessionStoreCtxKey
    InstanceStoreCtxKey
    ParentCtxKey
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

