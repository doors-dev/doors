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

	"github.com/doors-dev/doors/internal/shredder"
)

func FrameInsert(ctx context.Context) (context.Context, *shredder.AfterFrame) {
	sh := &shredder.AfterFrame{}
	return context.WithValue(ctx, keyFrame, sh), sh
}

func AfterFrame(ctx context.Context) (*shredder.AfterFrame, bool) {
	f, ok := ctx.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return nil, false
	}
	return f, true
}

func Frame(ctx context.Context) shredder.SimpleFrame {
	f, ok := ctx.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return shredder.FreeFrame{}
	}
	return f
}

func FrameInfect(source context.Context, target context.Context) context.Context {
	f, ok := source.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return target
	}
	return context.WithValue(target, keyFrame, f)
}

func FrameRemove(ctx context.Context) context.Context {
	_, ok := ctx.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return ctx
	}
	return context.WithValue(ctx, keyFrame, nil)
}
