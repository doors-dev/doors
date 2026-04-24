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
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/shredder"
)

func trackerRoot(r *root) (*tracker, core.Core) {
	sh := &shredder.ValveFrame{}
	t := &tracker{
		id:             r.NewID(),
		root:           r,
		cancel:         func() {},
		ctx:            r.runtime.Context(),
		innerCallGuard: sh,
		outerCallGuard: sh,
	}
	t.cinema = beam.NewCinema(nil, t)
	core := core.NewCore(r.inst, t)
	t.contentCtx = context.WithValue(r.runtime.Context(), ctex.KeyCore, core)
	return t, core
}

func trackerShutdown(prev *tracker) {
	prev.clean(false)
	prev.container.clean()
}

func trackerRemove(prev *tracker, task *userTask) {
	prev.clean(false)
	prev.container.clean()
	prev.outerCallGuard.Submit(prev.parent.ctx, prev.root.runtime, func(b bool) {
		if !b {
			task.Cancel()
			return
		}
		prev.root.inst.Call(&call{
			ctx:     prev.parent.ctx,
			kind:    callReplace,
			id:      prev.id,
			payload: emptyPayload{},
			task:    task,
		})
	})
}

func trackerInherit(n *node, prev *tracker, preserveFrame bool) *tracker {
	ctx, cancel := context.WithCancel(prev.parent.ctx)
	t := &tracker{
		id:             prev.id,
		node:           n,
		root:           prev.root,
		parent:         prev.parent,
		ctx:            ctx,
		cancel:         cancel,
		innerCallGuard: &shredder.ValveFrame{},
		outerCallGuard: prev.outerCallGuard,
	}
	t.cinema = beam.NewCinema(t.parent.cinema, t)
	if preserveFrame {
		t.container = prev.container
		t.container.update(t)
	} else {
		t.container = newContainerTracker(t)
		prev.container.clean()
	}
	prev.clean(false)
	core := core.NewCore(t.root.inst, t)
	t.contentCtx = context.WithValue(ctx, ctex.KeyCore, core)
	t.parent.addChild(t)
	return t
}

func trackerCreate(n *node, p *pipe) *tracker {
	ctx, cancel := context.WithCancel(p.tracker.ctx)
	t := &tracker{
		id:             p.tracker.root.NewID(),
		node:           n,
		root:           p.tracker.root,
		parent:         p.tracker,
		ctx:            ctx,
		cancel:         cancel,
		innerCallGuard: p.callGuard,
		outerCallGuard: p.callGuard,
	}
	t.cinema = beam.NewCinema(t.parent.cinema, t)
	t.container = newContainerTracker(t)
	core := core.NewCore(t.root.inst, t)
	t.contentCtx = context.WithValue(ctx, ctex.KeyCore, core)
	t.parent.addChild(t)
	return t
}

type tracker struct {
	id             uint64
	mu             sync.Mutex
	node           *node
	root           *root
	parent         *tracker
	thread         shredder.ReadWriteThread
	cinema         beam.Cinema
	ctx            context.Context
	contentCtx     context.Context
	cancel         context.CancelFunc
	outerCallGuard *shredder.ValveFrame
	innerCallGuard *shredder.ValveFrame
	container      *containerTracker
	hooks          common.Set[uint64]
	children       common.Set[*tracker]
}

func (t *tracker) UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	frames := ctex.GetFrames(ctx)
	callFrame := shredder.Join(true, frames.Call(), t.innerCallGuard)
	defer callFrame.Release()
	callFrame.Run(ctx, t.root.runtime, func(b bool) {
		if !b {
			if onCancel != nil {
				onCancel()
			}
			return
		}
		t.inst().UserCall(ctx, check, action, onResult, onCancel, params)
	})
}

func (t *tracker) isEmpty() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.children) == 0
}

func (t *tracker) inst() Instance {
	return t.root.inst
}

func (t *tracker) Runtime() shredder.Runtime {
	return t.root.runtime
}

func (t *tracker) ID() uint64 {
	return t.id
}

func (t *tracker) addChild(child *tracker) {
	t.mu.Lock()
	if t.ctx.Err() != nil {
		t.mu.Unlock()
		child.clean(true)
		return
	}
	defer t.mu.Unlock()
	if t.children == nil {
		t.children = common.NewSet[*tracker]()
	}
	t.children.Add(child)
}

func (t *tracker) removeChild(child *tracker) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return
	}
	t.children.Remove(child)
}

func (t *tracker) clean(cascade bool) {
	t.cancel()
	if !cascade && t.parent != nil {
		t.parent.removeChild(t)
	}
	if cascade {
		t.container.clean()
		t.node.unmountedSelf()
	}
	t.mu.Lock()
	hooks := t.hooks
	children := t.children
	t.hooks = nil
	t.children = nil
	t.mu.Unlock()
	for child := range children {
		child.clean(true)
	}
	for hook := range hooks {
		t.root.cancelHook(hook)
	}
}

func (t *tracker) ReadFrame() shredder.Frame {
	return t.thread.Read()
}

func (t *tracker) containerCinemaFrame() shredder.AnyFrame {
	if t.container == nil {
		return shredder.FreeFrame{}
	}
	return t.container.cinema.ReadFrame()
}

func (t *tracker) writeFrame() shredder.Frame {
	write := t.thread.Write()
	frame := shredder.Join(false, write, t.cinema.ReadFrame(), t.containerCinemaFrame())
	write.Release()
	return frame
}

func (t *tracker) isCanceled() bool {
	return t.ctx.Err() != nil
}

func (t *tracker) Context() context.Context {
	return t.contentCtx
}

func (t *tracker) Cinema() beam.Cinema {
	return t.cinema
}

func (t *tracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return core.Hook{}, false
	}
	if t.hooks == nil {
		t.hooks = common.NewSet[uint64]()
	}
	h := newHook(t.root.NewID(), t, onTrigger, onCancel)
	t.hooks.Add(h.id)
	t.root.addHook(h)
	return core.Hook{
		HookID: h.id,
		Cancel: h.cancel,
	}, true
}

func (t *tracker) removeHook(id uint64) {
	t.root.removeHook(id)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return
	}
	t.hooks.Remove(id)
}

func (t *tracker) Reload(ctx context.Context) {
	if t.node == nil {
		return
	}
	t.node.reload(ctx)
}

func (t *tracker) XReload(ctx context.Context) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XReload")
	if t.node == nil {
		ch := make(chan error, 1)
		ch <- errors.New("root door cannot be reloaded")
		close(ch)
		return ch
	}
	return t.node.reload(ctx)
}

func (t *tracker) RootCore() core.Core {
	return t.root.core
}

func newContainerTracker(t *tracker) *containerTracker {
	ctx, cancel := context.WithCancel(t.parent.ctx)
	ft := &containerTracker{
		tracker: t,
		cancel:  cancel,
	}
	ft.cinema = beam.NewCinema(t.parent.Cinema(), ft)
	core := core.NewCore(t.root.inst, ft)
	ft.ctx = context.WithValue(ctx, ctex.KeyCore, core)
	return ft
}

type containerTracker struct {
	mu      sync.Mutex
	tracker *tracker
	cinema  beam.Cinema
	hooks   common.Set[uint64]
	ctx     context.Context
	cancel  context.CancelFunc
}

func (t *containerTracker) UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	t.tracker.UserCall(ctx, check, action, onResult, onCancel, params)
}

func (t *containerTracker) getTracker() *tracker {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tracker
}

func (t *containerTracker) inst() Instance {
	return t.getTracker().inst()
}

func (t *containerTracker) Runtime() shredder.Runtime {
	return t.getTracker().Runtime()
}

func (t *containerTracker) ID() uint64 {
	return t.getTracker().ID()
}

func (t *containerTracker) removeHook(id uint64) {
	t.getTracker().root.removeHook(id)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.hooks == nil {
		return
	}
	t.hooks.Remove(id)

}

func (f *containerTracker) update(t *tracker) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tracker = t
}

func (t *containerTracker) Context() context.Context {
	return t.ctx
}

func (t *containerTracker) ReadFrame() shredder.Frame {
	return t.tracker.ReadFrame()
}

func (t *containerTracker) clean() {
	t.cancel()
	t.mu.Lock()
	hooks := t.hooks
	t.hooks = nil
	t.mu.Unlock()
	t.cinema.Cancel()
	for hook := range hooks {
		t.tracker.root.cancelHook(hook)
	}
}

func (t *containerTracker) Cinema() beam.Cinema {
	return t.cinema
}

func (t *containerTracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return core.Hook{}, false
	}
	if t.hooks == nil {
		t.hooks = common.NewSet[uint64]()
	}
	h := newHook(t.tracker.root.NewID(), t, onTrigger, onCancel)
	t.hooks.Add(h.id)
	t.tracker.root.addHook(h)
	return core.Hook{
		HookID: h.id,
		Cancel: h.cancel,
	}, true
}

func (t *containerTracker) Reload(ctx context.Context) {
	t.getTracker().Reload(ctx)
}

func (t *containerTracker) RootCore() core.Core {
	return t.getTracker().RootCore()
}

func (t *containerTracker) XReload(ctx context.Context) <-chan error {
	return t.getTracker().XReload(ctx)
}
