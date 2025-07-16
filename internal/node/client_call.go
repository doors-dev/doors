package node

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
)

type ClientCall struct {
	Name    string
	Arg     common.JsonWritabeRaw
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
	core      *core
	hookEntry *HookEntry
	ctx       context.Context
}

func (cc *clientCall) kill() {
	cc.cancelCall(errors.New("context killed"))
}

func (cc *clientCall) makeDone() bool {
	if cc.done.Swap(true) {
		return false
	}
	cc.core.removeClientCall(cc)
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
	return
}

func (j *clientCall) Name() string {
	return "call"
}

func (j *clientCall) Arg() common.JsonWritable {
	hook := common.JsonWritableAny{nil}
	if j.hookEntry != nil {
		hook = common.JsonWritableAny{j.hookEntry.HookId}
	}
	return common.JsonWritables([]common.JsonWritable{common.JsonWritableAny{j.call.Name}, j.call.Arg, common.JsonWritableAny{j.core.id}, hook})
}

func (k *clientCall) Payload() (common.Writable, bool) {
	return nil, false
}

func (cc *clientCall) OnResult(err error) {
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

func (cc *clientCall) OnWriteErr() bool {
	return !cc.done.Load()
}

func (cc *clientCall) Call() (Call, bool) {
	return cc, !cc.done.Load()
}
