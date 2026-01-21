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
	"github.com/doors-dev/doors/internal/common/ctxwg"
)

type Done = bool

type HookEntry struct {
	DoorID  uint64
	HookID  uint64
	tracker *tracker
}

func (h HookEntry) Cancel() {
	h.tracker.cancelHook(h.HookID)
}

type Hook struct {
	Trigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	Cancel  func(ctx context.Context)
}

func (s *Hook) trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done {
	return s.Trigger(ctx, w, r)
}

func (s *Hook) cancel(ctx context.Context) {
	if s.Cancel != nil {
		s.Cancel(ctx)
	}
}

const (
	hookActive int32 = iota
	hookDone
	hookCanceled
)

type hook struct {
	hook    Hook
	state   atomic.Int32
	ch      atomic.Pointer[chan struct{}]
	done    atomic.Bool
	tracker *tracker
}

func newHook(h Hook, t *tracker) *hook {
	return &hook{
		hook: h,
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
	h.hook.cancel(h.tracker.ctx)
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
	ctx := ctxwg.Insert(h.tracker.getContext())
	ctx = common.SetBlockingCtx(ctx)
	done, err := common.CatchValue(func() bool {
		return h.hook.trigger(ctx, w, r)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		defer h.tracker.getRoot().onPanic(err)
	}
	if done {
		h.state.Store(hookDone)
	}
	close(ch)
	if err != nil {
		ctxwg.Wait(ctx)
	}
	return done, true
}
