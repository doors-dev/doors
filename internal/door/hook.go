// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	ctx, frame := ctex.AfterFrameInsert(h.tracker.ctx)
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
