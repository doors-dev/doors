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
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/shredder"
)

type Done = bool

type HookEntry struct {
	DoorId uint64
	HookId uint64
	inst   instance
}

func (h HookEntry) Cancel() {
	h.inst.CancelHook(h.DoorId, h.HookId, nil)
}

func (h HookEntry) cancel(err error) {
	h.inst.CancelHook(h.DoorId, h.HookId, err)
}


type AttrHook struct {
	Trigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	Cancel  func(ctx context.Context, err error)
}

func (s *AttrHook) trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done {
	return s.Trigger(ctx, w, r)
}

func (s *AttrHook) cancel(ctx context.Context, err error) {
	if s.Cancel != nil {
		s.Cancel(ctx, err)
	}
}

type Hook interface {
	trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	cancel(ctx context.Context, err error)
}

type DoorHook struct {
	hook    Hook
	mu      sync.Mutex
	counter uint
	done    bool
	ch      chan struct{}
	err     error
	ctx     context.Context
	op      shredder.OnPanic
}

func newHook(ctx context.Context, h Hook, op shredder.OnPanic) *DoorHook {
	return &DoorHook{
		hook: h,
		ctx:  ctx,
		op:   op,
	}
}

func (h *DoorHook) Cancel(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.done {
		return
	}
	h.done = true
	h.err = err
	if h.counter == 0 {
		h.hook.cancel(h.ctx, h.err)
	}
}

func (h *DoorHook) Trigger(w http.ResponseWriter, r *http.Request) (Done, bool) {
	h.mu.Lock()
	ch := make(chan struct{})
	prevCh := h.ch
	h.ch = ch
	h.mu.Unlock()
	if prevCh != nil {
		<-prevCh
	}
	h.mu.Lock()
	if h.done {
		h.mu.Unlock()
		close(ch)
		return false, false
	}
	h.counter += 1
	h.mu.Unlock()
	ctx := ctxwg.Insert(h.ctx)
	ctx = common.SetBlockingCtx(ctx)
	done, err := common.CatchValue(func() bool {
		return h.hook.trigger(ctx, w, r)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		defer h.op.OnPanic(err)
	}
	if done || err != nil {
		h.mu.Lock()
		h.done = true
		h.mu.Unlock()
	}
	done = done || err != nil
	close(ch)
	ctxwg.Wait(ctx)
	h.mu.Lock()
	h.counter -= 1
	last := !done && h.counter == 0 && h.done
	h.mu.Unlock()
	if last {
		h.hook.cancel(h.ctx, h.err)
	}
	return done, true
}
