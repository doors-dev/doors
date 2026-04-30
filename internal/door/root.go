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
	"sync"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type Instance interface {
	Call(call action.Call)
	core.Instance
}

type Root = *root

func NewRoot(inst Instance) Root {
	r := &root{
		inst: inst,
	}
	r.tracker, r.core = trackerRoot(r)
	return r
}

type root struct {
	core    core.Core
	tracker *tracker
	hooks   sync.Map
	inst    Instance
}

func (r Root) Kill() {
	r.tracker.clean(false, shredder.FreeFrame{})
}

func (r Root) ID() uint64 {
	return r.tracker.id
}

func (r Root) instance() Instance {
	return r.inst
}

func (r *root) cancelHook(id uint64) {
	entry, ok := r.hooks.Load(id)
	if !ok {
		return
	}
	hook := entry.(*hook)
	hook.cancel()
}

func (r *root) runtime() shredder.Runtime {
	return r.inst.Runtime()
}

func (r *root) addHook(h *hook) {
	r.hooks.Store(h.id, h)
}

func (r *root) removeHook(id uint64) {
	r.hooks.Delete(id)
}

func (r *root) TriggerHook(id uint64, w http.ResponseWriter, rq *http.Request, track uint64) bool {
	entry, ok := r.hooks.Load(id)
	if !ok {
		return false
	}
	hook := entry.(*hook)
	return hook.trigger(w, rq, track)
}

func (r Root) IsStatic() bool {
	if !r.tracker.isEmpty() {
		return false
	}
	if !r.tracker.cinema.IsEmpty() {
		return false
	}
	static := true
	r.hooks.Range(func(_, _ any) bool {
		static = false
		return false
	})
	return static
}

func (r Root) Render(requestCtx context.Context, comp gox.Comp) (Stack, error) {
	thread := shredder.Thread{}
	renderFrame := shredder.Join(true, thread.Frame(), r.tracker.writeFrame())
	pipe := newPipe(
		r.tracker,
		new(deque.Deque[any]),
		renderFrame,
		r.tracker.innerCallGuard,
	)
	ch := make(chan struct{})
	var err error
	renderFrame.Submit(r.tracker.ctx, r.runtime(), func(b bool) {
		cur := gox.NewCursor(r.tracker.Context(), pipe)
		err = cur.Comp(comp)
	})
	renderFrame.Release()
	thread.Frame().Run(r.tracker.ctx, r.runtime(), func(b bool) {
		r.tracker.innerCallGuard.Activate()
		close(ch)
	})
	select {
	case <-ch:
		if err != nil {
			return nil, err
		}
		return pipe.Collect(), nil
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	}
}
