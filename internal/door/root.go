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
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Instance interface {
	Call(call actions.Call, group uint64, batch uint64, door uint64, depSeq uint64, depDoor uint64) uint64
	core.Instance
}

type Root = *root

func NewRoot(inst Instance) Root {
	r := &root{
		inst:  inst,
		hooks: make(map[uint64]*hook),
	}
	r.outer = newOuterTrackerRoot(r)
	r.tracker = r.outer.newTracker(nil)
	r.core = core.NewCore(r.tracker)
	return r
}

type root struct {
	core    core.Core
	outer   *outerTracker
	tracker *tracker
	mu      sync.Mutex
	hooks   map[uint64]*hook
	inst    Instance
}

func (r Root) Kill() {
	r.outer.clean()
}

func (r Root) ID() uint64 {
	return r.outer.id
}

func (r Root) instance() Instance {
	return r.inst
}

func (r *root) runtime() shredder.Runtime {
	return r.inst.Runtime()
}

func (r *root) cancelHook(id uint64) {
	r.mu.Lock()
	hook, ok := r.hooks[id]
	r.mu.Unlock()
	if !ok {
		return
	}
	hook.cancel()
}

func (r *root) addHook(h *hook, portable bool) {
	if portable {
		h.portable = common.RandId()
		session := r.inst.Session()
		pm := session.App().PathMaker()
		instanceID := r.inst.ID()
		prefix := pm.Portable(h.portable, "")
		id := h.id
		session.App().RegisterRewrite(h.portable, func(tail string) path.Rewrite {
			p, raw := pm.HookTail(instanceID, id, tail)
			return path.Rewrite{
				Path:    p,
				RawPath: raw,
				Context: func(ctx context.Context) context.Context {
					ctx = context.WithValue(ctx, common.KeySession, session)
					return context.WithValue(ctx, common.KeyHookPrefix, prefix)
				},
			}
		})
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[h.id] = h
}

func (r *root) removeHook(id uint64) {
	r.mu.Lock()
	h, ok := r.hooks[id]
	delete(r.hooks, id)
	r.mu.Unlock()
	if ok && h.portable != "" {
		r.inst.Session().App().UnregisterRewrite(h.portable)
	}
}

func (r *root) TriggerHook(id uint64, w http.ResponseWriter, rq *http.Request, track uint64) bool {
	r.mu.Lock()
	hook, ok := r.hooks[id]
	r.mu.Unlock()
	if !ok {
		return false
	}
	return hook.trigger(w, rq, track)
}

func (r Root) Render(requestCtx context.Context, comp gox.Comp) (Stack, error) {
	thread := shredder.Thread{}
	writeFrame := r.tracker.rw.Write()
	renderFrame := shredder.JoinRelease(r.tracker.Context(), thread.Frame(), writeFrame, r.tracker.cinemaFrame())
	passGuard := new(shredder.ValveFrame)
	pipe := newPipe(
		r.tracker,
		common.GetDequeBuffer(),
		renderFrame,
		passGuard,
		new(atomic.Uint64),
		true,
	)
	ch := make(chan struct{})
	callFrame := shredder.JoinRelease(r.tracker.Context(), thread.Frame(), writeFrame)
	var err error
	renderFrame.Submit(r.tracker.ctx, r.runtime(), func(b bool) {
		if !b {
			err = context.Canceled
			return
		}
		cur := gox.NewCursor(pipe.innerContext(), pipe)
		err = cur.Comp(comp)
	})
	callFrame.Run(r.tracker.ctx, r.runtime(), func(bool) {
		defer close(ch)
		defer passGuard.Activate()
		r.outer.outerGuard.Activate()
	})
	callFrame.Release()
	renderFrame.Release()
	select {
	case <-ch:
		if err != nil {
			logError(r.tracker.ctx, r.inst.Logger(), err, common.Caller{})
			return nil, err
		}
		if r.runtime().Context().Err() != nil {
			return nil, common.ErrTerminated
		}
		return pipe.Collect(), nil
	case <-requestCtx.Done():
		if errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			r.inst.Logger().Error("page rendering timed out", "error", requestCtx.Err())
		}
		return nil, requestCtx.Err()
	}
}
