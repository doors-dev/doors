package door2

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam2"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
)

type parent interface {
	getRoot() *Root
	getContext() context.Context
	Cinema() beam2.Cinema
	newRenderFrame() sh.Frame
	addChild(t *node) bool
	removeChild(t *node)
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
	t.readFrame.Store(shread.Frame())
	t.cinema = beam2.NewCinema(t.parent.Cinema(), t, root.Spawner())
	core := newCore(t.cinema)
	ctx := context.WithValue(prev.parent.getContext(), common.CtxKeyCore, core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	root.addTracker(t)
	return t
}

func newTracker(p parent, shread *sh.Shread) *tracker {
	root := p.getRoot()
	t := &tracker{
		root:     root,
		id:       root.newId(),
		parent:   p,
		shread:   shread,
		children: common.NewSet[*node](),
	}
	t.readFrame.Store(shread.Frame())
	t.cinema = beam2.NewCinema(t.parent.Cinema(), t, root.Spawner())
	core := newCore(t.cinema)
	ctx := context.WithValue(p.getContext(), common.CtxKeyCore, core)
	t.ctx, t.cancel = context.WithCancel(ctx)
	root.addTracker(t)
	return t
}

type tracker struct {
	mu        sync.Mutex
	id        uint64
	root      *Root
	parent    parent
	shread    *sh.Shread
	readFrame atomic.Value
	ctx       context.Context
	cancel    context.CancelFunc
	children  common.Set[*node]
	cinema    beam2.Cinema
	hooks     map[uint64]*hook
}

func (t *tracker) Cinema() beam2.Cinema {
	return t.cinema
}

func (t *tracker) registerHook(h Hook) HookEntry {
	t.mu.Lock()
	defer t.mu.Unlock()
	hook := newHook(h, t)
	id := t.root.newId()
	t.hooks[id] = hook
	return HookEntry{
		DoorID:  t.id,
		HookID:  id,
		tracker: t,
	}
}

func (t *tracker) cancelHook(hookId uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if hook, ok := t.hooks[hookId]; ok {
		hook.cancel()
		delete(t.hooks, hookId)
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
	t.root.Spawner().Spawn(func() {
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

func (t *tracker) getContext() context.Context {
	return t.ctx
}

func (t *tracker) getRoot() *Root {
	return t.root
}
