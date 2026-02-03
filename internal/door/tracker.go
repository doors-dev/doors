package door

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

func newRootTracker(r *root) *tracker {
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
	return t
}

func newTrackerFrom(prev *tracker) *tracker {
	root := prev.root
	t := &tracker{
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

func newTracker(parent *tracker) *tracker {
	t := &tracker{
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
	id                   uint64
	root                 *root
	parent               *tracker
	thread               shredder.Thread
	readFrame            atomic.Value
	ctx                  context.Context
	cancel               context.CancelFunc
	children             common.Set[*node]
	cinema               beam.Cinema
	hooks                map[uint64]*hook
	core                 core.Core
	childDoorHooksCancel map[uint64][]context.CancelFunc
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

func (t *tracker) registerChildDoorHook(childId uint64, onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	hook, ok := t.registerHook(onTrigger, onCancel)
	if !ok {
		return hook, false
	}
	if t.childDoorHooksCancel == nil {
		t.childDoorHooksCancel = make(map[uint64][]context.CancelFunc)
	}
	t.childDoorHooksCancel[childId] = append(t.childDoorHooksCancel[childId], hook.Cancel)
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
	defer t.mu.Unlock()
	if hook, ok := t.hooks[hookID]; ok {
		hook.cancel()
		delete(t.hooks, hookID)
	}
}

func (t *tracker) trigger(id uint64, w http.ResponseWriter, r *http.Request) bool {
	t.mu.Lock()
	hook, ok := t.hooks[id]
	t.mu.Unlock()
	if !ok {
		return false
	}
	done, ok := hook.trigger(w, r)
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
	defer t.mu.Unlock()
	for _, hook := range t.hooks {
		hook.cancel()
	}
	clear(t.hooks)
	for child := range t.children {
		child.kill(cascade)
	}
	t.children.Clear()
}

func (t *tracker) removeChild(n *node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.children.Remove(n)
	if t.childDoorHooksCancel == nil {
		return
	}
	cancels, ok := t.childDoorHooksCancel[n.tracker.id]
	if !ok {
		return
	}
	for _, cancel := range cancels {
		cancel()
	}
	delete(t.childDoorHooksCancel, n.tracker.id)
}

func (t *tracker) replaceChild(prev *node, next *node) {
	if prev.tracker.id != next.tracker.id {
		panic("wrong replace child operation")
	}
	t.mu.Lock()
	if t.isKilled() {
		t.mu.Unlock()
		next.kill(cascade)
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
		n.kill(cascade)
		return
	}
	if t.children == nil {
		t.children = common.NewSet[*node]()
	}
	t.children.Add(n)
	t.mu.Unlock()
}

func (t *tracker) NewFrame() shredder.Frame {
	f := t.readFrame.Load().(shredder.Frame)
	return shredder.Join(false, f)
}

func (t *tracker) newRenderFrame() shredder.Frame {
	frame := t.thread.Frame()
	prev := t.readFrame.Swap(t.thread.Frame())
	if prev != nil {
		prev := prev.(shredder.Frame)
		prev.Release()
	}
	return frame
}

type childDoorCore struct {
	tracker *tracker
	id      uint64
}

func (h childDoorCore) Cinema() beam.Cinema {
	return h.tracker.Cinema()
}

func (h childDoorCore) ID() uint64 {
	return h.tracker.id
}

func (h childDoorCore) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	return h.tracker.registerChildDoorHook(h.id, onTrigger, onCancel)
}

var _ core.Door = &childDoorCore{}
