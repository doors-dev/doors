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
	"fmt"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

func newRootTracker(r *root) (*tracker, core.Core) {
	t := &tracker{
		id:     r.NewID(),
		root:   r,
		parent: nil,
		cancel: r.runtime().Cancel,
	}
	t.cinema = beam.NewCinema(nil, t, r.runtime())
	core := core.NewCore(r.inst, t)
	t.ctx = context.WithValue(r.runtime().Context(), ctex.KeyCore, core)
	t.cancel = func() {}
	return t, core
}

func newTrackerFrom(prev *tracker, node *node) *tracker {
	root := prev.root
	t := &tracker{
		node:   node,
		root:   root,
		id:     prev.id,
		parent: prev.parent,
	}
	t.cinema = beam.NewCinema(t.parent.cinema, t, root.runtime())
	t.core = core.NewCore(root.inst, t)
	ctx := context.WithValue(prev.parent.ctx, ctex.KeyCore, t.core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	root.addTracker(t)
	return t
}

func newTracker(parent *tracker, node *node) *tracker {
	t := &tracker{
		node:   node,
		root:   parent.root,
		id:     parent.root.NewID(),
		parent: parent,
	}
	t.cinema = beam.NewCinema(t.parent.cinema, t, t.root.runtime())
	t.core = core.NewCore(t.root.inst, t)
	ctx := context.WithValue(parent.ctx, ctex.KeyCore, t.core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.root.addTracker(t)
	return t
}

var _ core.Door = &tracker{}
var _ beam.Door = &tracker{}

type tracker struct {
	mu                   sync.Mutex
	node                 *node
	id                   uint64
	root                 *root
	parent               *tracker
	thread               shredder.ReadWriteThread
	ctx                  context.Context
	cancel               context.CancelFunc
	children             common.Set[*node]
	cinema               beam.Cinema
	hooks                map[uint64]*hook
	core                 core.Core
	containerHooksCancel map[uint64][]context.CancelFunc
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
		ch <- errors.New("Root can't be reloaded")
		close(ch)
		return ch
	}
	return t.node.reload(ctx)
}

func (t *tracker) debug(tab string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("%s%d\n", tab, t.id)
	for node := range t.children {
		mu := node.entity.(nodeMount)
		mu.Tracker().debug(tab + "-")
	}
}

func (t *tracker) inst() Instance {
	return t.root.inst
}

func (t *tracker) RootCore() core.Core {
	return t.root.core
}

func (t *tracker) runtime() shredder.Runtime {
	return t.root.runtime()
}

func (t *tracker) parentContext() context.Context {
	if t.parent == nil {
		return t.root.runtime().Context()
	}
	return t.parent.ctx
}

func (t *tracker) containerContext() context.Context {
	parentCore := core.NewCore(t.inst(), containerCore{
		tracker:      t.parent,
		childTracker: t,
		id:           t.id,
	})
	return context.WithValue(t.parent.parentContext(), ctex.KeyCore, parentCore)
}

func (t *tracker) Context() context.Context {
	return t.ctx
}

func (t *tracker) ID() uint64 {
	return t.id
}

func (t *tracker) Cinema() beam.Cinema {
	return t.cinema
}

func (t *tracker) registerContainerHook(childId uint64, onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	hook, ok := t.registerHook(onTrigger, onCancel)
	if !ok {
		return hook, false
	}
	if t.containerHooksCancel == nil {
		t.containerHooksCancel = make(map[uint64][]context.CancelFunc)
	}
	t.containerHooksCancel[childId] = append(t.containerHooksCancel[childId], hook.Cancel)
	return hook, true
}

func (t *tracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.registerHook(onTrigger, onCancel)

}

func (t *tracker) registerHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	if t.isKilled() {
		return core.Hook{}, false
	}
	h := newHook(t, onTrigger, onCancel)
	id := t.root.NewID()
	if t.hooks == nil {
		t.hooks = make(map[uint64]*hook)
	}
	t.hooks[id] = h
	return core.Hook{
		DoorID: t.id,
		HookID: id,
		Cancel: func() {
			t.cancelHook(id)
		},
	}, true

}

func (t *tracker) isKilled() bool {
	return t.ctx.Err() != nil
}

func (t *tracker) cancelHook(hookID uint64) {
	t.mu.Lock()
	hook, ok := t.hooks[hookID]
	if !ok {
		t.mu.Unlock()
		return
	}
	delete(t.hooks, hookID)
	t.mu.Unlock()
	hook.cancel()
}

func (t *tracker) trigger(id uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	t.mu.Lock()
	hook, ok := t.hooks[id]
	t.mu.Unlock()
	if !ok {
		return false
	}
	done, ok := hook.trigger(w, r, track)
	if !ok {
		return false
	}
	if done {
		t.mu.Lock()
		delete(t.hooks, id)
		t.mu.Unlock()
	}
	return true
}

func (t *tracker) kill() {
	t.cancel()
	t.root.removeTracker(t)
	t.cinema.Cancel()
	t.mu.Lock()
	hooks := t.hooks
	t.hooks = nil
	children := t.children
	t.children = nil
	t.mu.Unlock()
	for _, hook := range hooks {
		hook.cancel()
	}
	for child := range children {
		child.killCascade()
	}
}

func (t *tracker) removeChild(n *node, id uint64) {
	t.mu.Lock()
	if !t.children.Remove(n) {
		defer t.mu.Unlock()
		return
	}
	if t.containerHooksCancel == nil {
		defer t.mu.Unlock()
		return
	}
	cancels, ok := t.containerHooksCancel[id]
	if !ok {
		defer t.mu.Unlock()
		return
	}
	delete(t.containerHooksCancel, id)
	t.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
}

func (t *tracker) replaceChild(prev *node, next *node) {
	t.mu.Lock()
	if t.isKilled() {
		t.mu.Unlock()
		next.killCascade()
		return
	}
	t.children.Remove(prev)
	t.children.Add(next)
	t.mu.Unlock()
}

func (t *tracker) addChild(n *node) {
	t.mu.Lock()
	if t.isKilled() {
		t.mu.Unlock()
		n.killCascade()
		return
	}
	if t.children == nil {
		t.children = common.NewSet[*node]()
	}
	t.children.Add(n)
	t.mu.Unlock()
}

func (t *tracker) ReadFrame() shredder.Frame {
	return t.thread.Read()
}

func (t *tracker) writeFrame() shredder.Frame {
	write := t.thread.Write()
	join := shredder.Join(false, write, t.cinema.ReadFrame())
	write.Release()
	return join
}

type containerCore struct {
	tracker      *tracker
	childTracker *tracker
	id           uint64
}

func (h containerCore) Reload(ctx context.Context) {
	h.childTracker.Reload(ctx)
}

func (h containerCore) XReload(ctx context.Context) <-chan error {
	return h.childTracker.XReload(ctx)
}

func (h containerCore) Cinema() beam.Cinema {
	return h.tracker.Cinema()
}

func (h containerCore) ID() uint64 {
	return h.tracker.id
}

func (h containerCore) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	return h.tracker.registerContainerHook(h.id, onTrigger, onCancel)
}

func (h containerCore) RootCore() core.Core {
	return h.tracker.RootCore()
}

var _ core.Door = &containerCore{}
