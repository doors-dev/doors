package door2

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam2"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/sh"
)

func newRootTracker(ctx context.Context, r *root) *tracker {
	t := &tracker{
		id:       r.NewID(),
		root:     r,
		parent:   nil,
		children: common.NewSet[*node](),
	}
	t.cinema = beam2.NewCinema(nil, t, r.spawner)
	core := core.NewCore(r.inst, t)
	ctx = context.WithValue(ctx, ctex.KeyCore, core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	return t
}

func newTrackerFrom(prev *tracker, shread *sh.Shread) *tracker {
	root := prev.root
	t := &tracker{
		root:     root,
		id:       prev.id,
		parent:   prev.parent,
		shread:   shread,
		children: common.NewSet[*node](),
	}
	t.initShread(shread)
	t.cinema = beam2.NewCinema(t.parent.cinema, t, root.spawner)
	t.core = core.NewCore(root.inst, t)
	ctx := context.WithValue(prev.parent.ctx, ctex.KeyCore, t.core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	root.addTracker(t)
	return t
}

func newTracker(parent *tracker, shread *sh.Shread) *tracker {
	t := &tracker{
		root:     parent.root,
		id:       parent.root.NewID(),
		parent:   parent,
		children: common.NewSet[*node](),
	}
	t.initShread(shread)
	t.cinema = beam2.NewCinema(t.parent.cinema, t, t.root.spawner)
	t.core = core.NewCore(t.root.inst, t)
	ctx := context.WithValue(parent.ctx, ctex.KeyCore, t.core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.root.addTracker(t)
	return t
}

var _ core.Door = &tracker{}
var _ beam2.Door = &tracker{}

type tracker struct {
	mu        sync.Mutex
	id        uint64
	root      *root
	parent    *tracker
	shread    *sh.Shread
	readFrame atomic.Value
	ctx       context.Context
	cancel    context.CancelFunc
	children  common.Set[*node]
	cinema    beam2.Cinema
	hooks     map[uint64]*hook
	core      core.Core
}

func (t *tracker) ID() uint64 {
	return t.id
}

func (t *tracker) initShread(shread *sh.Shread) {
	t.readFrame.Store(shread.Frame())
	t.shread = shread
}

func (t *tracker) Cinema() beam2.Cinema {
	return t.cinema
}

func (t *tracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	hook := newHook(t, onTrigger, onCancel)
	id := t.root.NewID()
	t.hooks[id] = hook
	return core.Hook{
		DoorID: t.id,
		HookID: id,
		Cancel: func() {
			t.cancelHook(id)
		},
	}, true

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
	t.mu.Lock()
	defer t.mu.Unlock()
	for child := range t.children {
		child.kill(cascade)
	}
	t.children.Clear()
	t.root.spawner.Spawn(func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		for _, hook := range t.hooks {
			hook.cancel()
		}
	})
}

func (t *tracker) removeChild(n *node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.children.Remove(n)
}

func (t *tracker) addChild(n *node) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return false
	}
	t.children.Add(n)
	return true
}

func (t *tracker) NewFrame() sh.Frame {
	f := t.readFrame.Load().(sh.Frame)
	return sh.Join(f)
}

func (t *tracker) newRenderFrame() sh.Frame {
	frame := t.shread.Frame()
	prev := t.readFrame.Swap(t.shread.Frame()).(sh.Frame)
	prev.Release()
	return frame
}
