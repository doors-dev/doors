package door

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
)

type parent interface {
	getRoot() *Root
	getContext() context.Context
	newWriteFrame() sh.Frame
	addChild(t *node) bool
	removeChild(t *node)
}

type trackerKey struct{}

func newTrackerFrom(prev *tracker, shread *sh.Shread) *tracker {
	t := &tracker{}
	t.root = prev.root
	t.parent = prev.parent
	t.id = prev.id
	ctx := context.WithValue(prev.parent.getContext(), trackerKey{}, t)
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.shread = shread
	t.readFrame.Store(shread.Frame())
	return t
}

func newTracker(p parent, shread *sh.Shread) *tracker {
	t := &tracker{}
	t.root = p.getRoot()
	t.parent = p
	t.id = t.root.newId()
	ctx := context.WithValue(p.getContext(), trackerKey{}, t)
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.shread = shread
	t.readFrame.Store(shread.Frame())
	return t
}

type tracker struct {
	id        uint64
	root      *Root
	parent    parent
	shread    *sh.Shread
	readFrame atomic.Value
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	children  common.Set[*node]
}

func (t *tracker) kill() {
	t.cancel()
	t.mu.Lock()
	defer t.mu.Unlock()
	for child := range t.children {
		child.kill(false)
	}
	t.children.Clear()
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

func (t *tracker) newReadFrame() sh.Frame {
	f := t.readFrame.Load().(sh.Frame)
	return sh.Join(f)
}

func (t *tracker) newWriteFrame() sh.Frame {
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
