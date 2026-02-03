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
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/ctex"
)

type Done = bool

const (
	hookActive int32 = iota
	hookDone
	hookCanceled
	hookErrored
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
		tracker:     t,
	}
}

func (h *hook) cancel() {
	state := h.state.Swap(hookCanceled)
	if state != hookActive && state != hookErrored {
		return
	}
	ch := h.wait()
	defer close(ch)
	state = h.state.Load()
	if state != hookCanceled && state != hookErrored {
		return
	}
	if h.cancelFunc == nil {
		return
	}
	h.tracker.root.runtime().SafeCtxFun(h.tracker.ctx, h.cancelFunc)
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
	slog.Info("HOOK_DOOR")
	ch := h.wait()
	if h.state.Load() != hookActive {
		close(ch)
		return false, false
	}
	ctx := ctex.WgInsert(h.tracker.ctx)
	ctx = ctex.SetBlockingCtx(ctx)
	done, err := h.tracker.root.runtime().SafeHook(ctx, w, r, h.triggerFunc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.state.Store(hookErrored)
	} else if done {
		h.state.Store(hookDone)
	}
	close(ch)
	if err != nil {
		ctex.WgWait(ctx)
	}
	return done, true
}
