// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package ctxwg

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

type ctxKeyWgType struct{}

var ctxKeyWg = ctxKeyWgType{}

type atomicWg = *atomic.Pointer[sync.WaitGroup]

func Insert(ctx context.Context) context.Context {
	awg := &atomic.Pointer[sync.WaitGroup]{}
	awg.Store(&sync.WaitGroup{})
	return context.WithValue(ctx, ctxKeyWg, awg)
}

func Infect(source context.Context, target context.Context) context.Context {
	awg, ok := source.Value(ctxKeyWg).(atomicWg)
	if !ok {
		return target
	}
	return context.WithValue(target, ctxKeyWg, awg)
}

func Wait(ctx context.Context) {
	awg, ok := ctx.Value(ctxKeyWg).(atomicWg)
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

func Add(ctx context.Context) Done {
	awg, ok := ctx.Value(ctxKeyWg).(atomicWg)
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
}
