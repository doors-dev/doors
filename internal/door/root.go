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

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type Instance interface {
	Conf() *common.SystemConf
	Call(call action.Call)
	core.Instance
}

type Root = *root

func NewRoot(inst Instance) Root {
	r := &root{
		inst:    inst,
		runtime: inst.Runtime(),
		prime:   common.NewPrime(),
	}
	r.tracker, r.core = trackerRoot(r)
	return r
}

type root struct {
	runtime shredder.Runtime
	core    core.Core
	tracker *tracker
	prime   common.Prime
	inst    Instance
	hooks   sync.Map
}

func (r Root) Kill() {
	r.tracker.clean(false)
}

func (r Root) ID() uint64 {
	return r.tracker.id
}

func (r Root) Context() context.Context {
	return r.tracker.Context()
}

func (r *root) NewID() uint64 {
	return r.prime.Gen()
}

func (r *root) cancelHook(id uint64) {
	entry, ok := r.hooks.Load(id)
	if !ok {
		return
	}
	hook := entry.(*hook)
	hook.cancel()
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

func (r Root) Render(requestCtx context.Context, comp gox.Comp, init func()) (Stack, error) {
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
	renderFrame.Submit(r.tracker.ctx, r.runtime, func(b bool) {
		init()
		cur := gox.NewCursor(r.tracker.Context(), pipe)
		err = cur.Comp(comp)
	})
	renderFrame.Release()
	thread.Frame().Run(r.tracker.ctx, r.runtime, func(b bool) {
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
