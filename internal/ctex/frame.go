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

/*
type atomicWg = *atomic.Pointer[sync.WaitGroup]

func WgInsert(ctx context.Context) context.Context {
	awg := &atomic.Pointer[sync.WaitGroup]{}
	awg.Store(&sync.WaitGroup{})
	return context.WithValue(ctx, keyWg, awg)
}

func WgInfect(source context.Context, target context.Context) context.Context {
	awg, ok := source.Value(keyWg).(atomicWg)
	if !ok {
		return target
	}
	return context.WithValue(target, keyWg, awg)
}

func WgWait(ctx context.Context) {
	awg, ok := ctx.Value(keyWg).(atomicWg)
	if !ok {
		log.Fatal("Must have")
	}
	wg := awg.Load()
	if wg == nil {
		log.Fatal("Must have")
	}
	wg.Wait()
	awg.Store(nil)
}

type Done = func()

var none = func() {}

func WgAdd(ctx context.Context) Done {
	awg, ok := ctx.Value(keyWg).(atomicWg)
	if !ok {
		return none
	}
	wg := awg.Load()
	if wg == nil {
		return none
	}
	wg.Add(1)
	return func() {
		wg.Done()
	}
} */
