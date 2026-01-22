// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package door2

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
)

type Done = bool

const (
	hookActive int32 = iota
	hookDone
	hookCanceled
)

type hook struct {
	triggerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	cancelFunc  func(ctx context.Context)
	state       atomic.Int32
	ch          atomic.Pointer[chan struct{}]
	done        atomic.Bool
	tracker     *tracker
}

func newHook(t *tracker, triggerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done, cancelFunc func(ctx context.Context)) *hook {
	return &hook{
		triggerFunc: triggerFunc,
		cancelFunc:  cancelFunc,
		tracker: t,
	}
}

func (h *hook) cancel() {
	prev := h.state.Swap(hookCanceled)
	if prev != hookActive {
		return
	}
	ch := h.wait()
	defer close(ch)
	if h.state.Load() != hookCanceled {
		return
	}
	if h.cancelFunc == nil {
		return
	}
	h.cancelFunc(h.tracker.ctx)
}

func (h *hook) wait() chan struct{} {
	ch := make(chan struct{})
	prevCh := h.ch.Swap(&ch)
	if prevCh != nil {
		<-*prevCh
	}
	return ch
}

func (h *hook) trigger(w http.ResponseWriter, r *http.Request) (Done, bool) {
	ch := h.wait()
	if h.state.Load() != hookActive {
		close(ch)
		return false, false
	}
	ctx := ctex.WgInsert(h.tracker.ctx)
	ctx = common.SetBlockingCtx(ctx)
	done, err := common.CatchValue(func() bool {
		return h.triggerFunc(ctx, w, r)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		defer h.tracker.root.onPanic(err)
	}
	if done {
		h.state.Store(hookDone)
	}
	close(ch)
	if err != nil {
		ctex.WgWait(ctx)
	}
	return done, true
}
