// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package ctxwg

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

type ctxWgKeyType struct{}

var ctxWgKey = ctxWgKeyType{}

type atomicWg = *atomic.Pointer[sync.WaitGroup]

func Insert(ctx context.Context) context.Context {
	awg := &atomic.Pointer[sync.WaitGroup]{}
	awg.Store(&sync.WaitGroup{})
	return context.WithValue(ctx, ctxWgKey, awg)
}

func Infect(source context.Context, target context.Context) context.Context {
	awg, ok := source.Value(ctxWgKey).(atomicWg)
	if !ok {
		return target
	}
	return context.WithValue(target, ctxWgKey, awg)
}

func Wait(ctx context.Context) {
	awg, ok := ctx.Value(ctxWgKey).(atomicWg)
	if !ok {
		log.Fatal("Must have")
	}
	wg := awg.Swap(nil)
	if wg == nil {
		log.Fatal("Must have")
	}
	wg.Wait()
}

type Done = func()

var none = func() {}

func Add(ctx context.Context) Done {
	awg, ok := ctx.Value(ctxWgKey).(atomicWg)
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
