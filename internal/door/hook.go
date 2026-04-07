// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package door

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/ctex"
)

type Done = bool

const (
	hookActive int32 = iota
	hookProgress
	hookDone
	hookCanceled
	hookErrored
)

type hook struct {
	triggerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	cancelFunc  func(ctx context.Context)
	state       atomic.Int32
	ch          atomic.Pointer[chan struct{}]
	tracker     *tracker
}

func newHook(t *tracker, triggerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done, cancelFunc func(ctx context.Context)) *hook {
	return &hook{
		triggerFunc: triggerFunc,
		cancelFunc:  cancelFunc,
		tracker:     t,
	}
}

func (h *hook) cancel() {
	state := h.state.Swap(hookCanceled)
	if state != hookActive && state != hookErrored {
		return
	}
	h.performCancel()
}

func (h *hook) wait() chan struct{} {
	ch := make(chan struct{})
	prevCh := h.ch.Swap(&ch)
	if prevCh != nil {
		<-*prevCh
	}
	return ch
}

func (h *hook) trigger(w http.ResponseWriter, r *http.Request, track uint64) (Done, bool) {
	ch := h.wait()
	if !h.state.CompareAndSwap(hookActive, hookProgress) {
		close(ch)
		return false, false
	}
	ctx, frame := ctex.FrameInsert(h.tracker.ctx)
	defer frame.Activate()
	if track != 0 {
		frame.RunAfter(nil, nil, func(b bool) {
			h.tracker.inst().Call(reportHook(track))
		})
	}
	done, err := h.tracker.root.runtime().SafeHook(ctx, w, r, h.triggerFunc)
	ok := false
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ok = h.state.CompareAndSwap(hookProgress, hookErrored)
	} else if done {
		ok = h.state.CompareAndSwap(hookProgress, hookDone)
	} else {
		ok = h.state.CompareAndSwap(hookProgress, hookActive)
	}
	if !ok {
		h.performCancel()
	}
	close(ch)
	return done, true
}

func (h *hook) performCancel() {
	if h.cancelFunc == nil {
		return
	}
	h.tracker.root.runtime().SafeCtxFun(h.tracker.ctx, h.cancelFunc)
}
