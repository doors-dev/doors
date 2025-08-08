package node

import (
	"context"
	"io"
	"sync"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type instance interface {
	OnPanic(error)
	Thread() *shredder.Thread
	CancelHooks(uint64, error)
	CancelHook(uint64, uint64, error)
	RegisterHook(uint64, uint64, *NodeHook)
	NewId() uint64
	Call(common.Call)
}

type nodeMode int

const (
	dynamic nodeMode = iota
	static
	removed
)

type Node struct {
	mu        sync.Mutex
	parent    *tracker
	container *container
	content   templ.Component
	mode      nodeMode
}

func (n *Node) registerHook(container *container, tracker *tracker, ctx context.Context, h Hook) (*HookEntry, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if ctx.Err() != nil || n.container == nil {
		return nil, false
	}
	if container.tracker != tracker {
		return nil, false
	}
	if n.container != container {
		return nil, false
	}
	hookId := n.container.inst.NewId()
	hook := newHook(common.ClearBlockingCtx(ctx), h, n.container.inst)
	n.container.inst.RegisterHook(n.container.id, hookId, hook)
	return &HookEntry{
		NodeId: n.container.id,
		HookId: hookId,
		inst:   n.container.inst,
	}, true
}

func (n *Node) suspend(parent *tracker) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.parent != parent {
		return
	}
	if n.container == nil {
		return
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.container.suspend()
	n.container = nil
}

func (n *Node) reload(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.container == nil {
		return closedCh
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	return n.container.update(ctx, n.content)
}

func (n *Node) update(ctx context.Context, content templ.Component) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.content = content
	if n.container == nil {
		n.mode = dynamic
		return closedCh
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	return n.container.update(ctx, content)
}

func (n *Node) remove(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.mode = removed
	if n.container == nil {
		return closedCh
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.parent.removeChild(n)
	container := n.container
	n.container = nil
	return container.remove(ctx)
}

func (n *Node) replace(ctx context.Context, content templ.Component) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.mode = static
	n.content = content
	if n.container == nil {
		return closedCh
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.parent.removeChild(n)
	container := n.container
	n.container = nil
	return container.replace(ctx, content)
}

func (n *Node) Render(ctx context.Context, w io.Writer) error {
	n.mu.Lock()
	if n.container != nil {
		n.parent.removeChild(n)
		n.container.remove(ctx)
		n.container = nil
	}
	ctx, children, hasChildren := common.GetChildren(ctx)
	if hasChildren {
		n.content = children
		n.mode = dynamic
	}
	if n.mode == removed {
		n.mu.Unlock()
		return nil
	}
	if n.mode == static {
		n.mu.Unlock()
		if n.content == nil {
			return nil
		}
		return n.content.Render(ctx, w)
	}
	defer n.mu.Unlock()
	n.parent = ctx.Value(common.NodeCtxKey).(*tracker)
	if n.parent != nil {
		n.parent.addChild(n)
	}
	inst := ctx.Value(common.InstanceCtxKey).(instance)
	thread := ctx.Value(common.ThreadCtxKey).(*shredder.Thread)
	rm := ctx.Value(common.RenderMapCtxKey).(*common.RenderMap)
	parentCtx := context.WithValue(ctx, common.RenderMapCtxKey, nil)
	parentCtx = context.WithValue(parentCtx, common.ThreadCtxKey, nil)
	var parentCinema *Cinema
	if n.parent != nil {
		parentCinema = n.parent.cinema
	}
	n.container = &container{
		id:           inst.NewId(),
		inst:         inst,
		parentCtx:    ctx,
		parentCinema: parentCinema,
		node:         n,
	}
	return n.container.render(thread, rm, w, n.content)
}
