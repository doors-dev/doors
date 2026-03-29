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
	"fmt"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/door/pipe"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
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
		prime:   common.NewPrime(),
		tackers: make(map[uint64]*tracker),
	}
	r.tracker, r.core = newRootTracker(r)
	return r
}

type root struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	core    core.Core
	prime   *common.Prime
	inst    Instance
	tackers map[uint64]*tracker
	tracker *tracker
}

func (r Root) debug(messages ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	str := fmt.Sprint(messages...)
	println()
	println(str)
	r.tracker.debug("*")
}

func (r Root) IsStatic() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.tracker.children) != 0 {
		return false
	}
	if len(r.tracker.hooks) != 0 {
		return false
	}
	if !r.tracker.cinema.IsEmpty() {
		return false
	}
	return true
}

func (r *root) runtime() shredder.Runtime {
	return r.inst.Runtime()
}

func (r Root) ID() uint64 {
	return r.tracker.id
}

func (r Root) Context() context.Context {
	return r.tracker.ctx
}

func (r Root) Kill() {
	r.tracker.kill()
}

type Stack = pipe.Stack

func (r Root) Render(requestCtx context.Context, comp gox.Comp, init func()) (Stack, error) {
	thread := shredder.Thread{}
	mountFrame := &shredder.ValveFrame{}
	renderFrame := shredder.Join(true, thread.Frame(), r.tracker.newRenderFrame())
	pipe := pipe.NewPipe(
		r.tracker.ctx,
		r.tracker.runtime(),
		renderFrame,
		mountFrame,
	)
	ch := make(chan struct{})
	renderFrame.Submit(r.tracker.ctx, r.runtime(), func(b bool) {
		init()
		pipe.RenderContent(comp)
	})
	renderFrame.Release()
	thread.Frame().Run(r.tracker.ctx, r.runtime(), func(b bool) {
		mountFrame.Activate()
		close(ch)
	})
	select {
	case <-ch:
		return pipe.Collect()
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	}
}

func (i Root) TriggerHook(doorID uint64, hookId uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	var tracker *tracker
	if i.tracker.id == doorID {
		tracker = i.tracker
	} else {
		i.mu.Lock()
		var ok bool
		tracker, ok = i.tackers[doorID]
		i.mu.Unlock()
		if !ok {
			return false
		}
	}
	return tracker.trigger(hookId, w, r, track)
}

func (r *root) addTracker(t *tracker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tackers[t.id] = t
}

func (r *root) removeTracker(t *tracker) {
	if r.tracker.isKilled() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.tackers[t.id]
	if !ok || existing != t {
		return
	}
	delete(r.tackers, t.id)
}

func (r *root) NewID() uint64 {
	return r.prime.Gen()
}
