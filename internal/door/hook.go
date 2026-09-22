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
	"github.com/doors-dev/doors/internal/shredder"
)

type Done = bool

const (
	hookActive int32 = iota
	hookDone
	hookCanceled
	hookErrored
)

type hook struct {
	id          uint64
	triggerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	state       atomic.Int32
	ch          atomic.Pointer[chan struct{}]
	tracker     hookTracker
	parallel    bool
	inflight    atomic.Int64
	once        atomic.Bool
}

type hookTracker interface {
	Runtime() shredder.Runtime
	inst() Instance
	removeHook(id uint64)
	Context() context.Context
}

func newHook(id uint64, tracker hookTracker, triggerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done, parallel bool) *hook {
	return &hook{
		id:          id,
		triggerFunc: triggerFunc,
		tracker:     tracker,
		parallel:    parallel,
	}
}

func (h *hook) cancel() {
	state := h.state.Swap(hookCanceled)
	if state != hookActive && state != hookErrored {
		return
	}
	if h.inflight.Load() == 0 && h.once.CompareAndSwap(false, true) {
		h.tracker.removeHook(h.id)
	}
}

func (h *hook) wait() chan struct{} {
	ch := make(chan struct{})
	prevCh := h.ch.Swap(&ch)
	if prevCh != nil {
		<-*prevCh
	}
	return ch
}

func (h *hook) trigger(w http.ResponseWriter, r *http.Request, track uint64) bool {
	if !h.parallel {
		ch := h.wait()
		defer close(ch)
	}
	if h.tracker.Context().Err() != nil {
		return false
	}
	h.inflight.Add(1)
	if h.state.Load() != hookActive {
		h.release()
		return false
	}
	ctx, frame := ctex.AfterFrameInsert(h.tracker.Context())
	defer frame.Activate()
	if track != 0 {
		frame.After().Run(nil, nil, func(b bool) {
			h.tracker.inst().Call(reportHook(track))
		})
	}
	done, err := h.tracker.Runtime().SafeHook(ctx, w, r, h.triggerFunc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.state.CompareAndSwap(hookActive, hookErrored)
	} else if done {
		h.state.CompareAndSwap(hookActive, hookDone)
	}
	h.release()
	return true
}

func (h *hook) release() {
	if h.inflight.Add(-1) != 0 {
		return
	}
	switch h.state.Load() {
	case hookCanceled, hookDone:
		if h.once.CompareAndSwap(false, true) {
			h.tracker.removeHook(h.id)
		}
	}
}
