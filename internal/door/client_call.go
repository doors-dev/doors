// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package door

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
)

type ClientCall struct {
	Name    string
	Arg     any
	Trigger func(ctx context.Context, w http.ResponseWriter, r *http.Request)
	Cancel  func(ctx context.Context, err error)
}

func (c *ClientCall) cancel(ctx context.Context, err error) {
	if c.Cancel == nil {
		return
	}
	c.Cancel(ctx, err)
}

type clientCall struct {
	done      atomic.Bool
	call      *ClientCall
	tracker   *tracker
	hookEntry *HookEntry
	ctx       context.Context
}

func (cc *clientCall) kill() {
	cc.cancelCall(nil)
}

func (cc *clientCall) makeDone() bool {
	if cc.done.Swap(true) {
		return false
	}
	cc.tracker.removeClientCall(cc)
	return true
}

func (cc *clientCall) trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done {
	if !cc.makeDone() {
		return true
	}
	cc.call.Trigger(ctx, w, r)
	return true
}

func (cc *clientCall) cancel(ctx context.Context, err error) {
	if !cc.makeDone() {
		return
	}
	cc.call.cancel(ctx, err)
}

func (cc *clientCall) cancelCall(err error) {
	if cc.hookEntry != nil {
		cc.hookEntry.cancel(err)
		return
	}
	if !cc.makeDone() {
		return
	}
	cc.call.cancel(cc.ctx, err)
}

func (cc *clientCall) Data() *common.CallData {
	if cc.done.Load() {
		return nil
	}
	return &common.CallData{
		Name:    "call",
		Arg:     cc.arg(),
		Payload: common.WritableNone{},
	}

}

func (cc *clientCall) arg() []any {
	var hook *uint64 = nil
	if cc.hookEntry != nil {
		hook = &cc.hookEntry.HookId
	}
	return []any{cc.call.Name, cc.call.Arg, cc.tracker.Id(), hook}
}

func (cc *clientCall) Result(err error) {
	if err != nil {
		slog.Error("Call failed", slog.String("call_name", cc.call.Name), slog.String("js_error", err.Error()))
		cc.cancelCall(err)
		return
	}
	if cc.hookEntry != nil {
		return
	}
	cc.makeDone()
}
