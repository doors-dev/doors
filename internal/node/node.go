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
	ctx = ctx.Value(common.ParentCtxKey).(context.Context)
	hookId := n.container.inst.NewId()
	hook := newHook(ctx, h, n.container.inst)
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
	ch, ok := common.ResultChannel(ctx,"Node reload")
	if !ok {
		return ch
	}
	if n.container == nil {
		close(ch)
		return ch
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.container.update(ctx, n.content, ch)
	return ch
}

func (n *Node) clear(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx,"Node clear")
	if !ok {
		return ch
	}
	n.content = nil
	if n.container == nil {
		n.mode = dynamic
		close(ch)
		return ch
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.container.clear(ctx, ch)
	return ch
}
func (n *Node) update(ctx context.Context, content templ.Component) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx,"Node update")
	if !ok {
		return ch
	}
	n.content = content
	if n.container == nil {
		n.mode = dynamic
		close(ch)
		return ch
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.container.update(ctx, content, ch)
	return ch
}

func (n *Node) remove(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx,"Node remove")
	if !ok {
		return ch
	}
	n.mode = removed
	if n.container == nil {
		close(ch)
		return nil
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.parent.removeChild(n)
	container := n.container
	n.container = nil
	container.remove(ctx, ch)
	return ch
}

func (n *Node) replace(ctx context.Context, content templ.Component) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx,"Node replace")
	if !ok {
		return ch
	}
	n.mode = static
	n.content = content
	if n.container == nil {
		close(ch)
		return ch
	}
	n.container.inst.CancelHooks(n.container.id, nil)
	n.parent.removeChild(n)
	container := n.container
	n.container = nil
	container.replace(ctx, content, ch)
	return ch
}

func (n *Node) Render(ctx context.Context, w io.Writer) error {
	n.mu.Lock()
	if n.container != nil {
		n.parent.removeChild(n)
		ch := make(chan error, 1)
		n.container.remove(ctx, ch)
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
	parentCtx := ctx.Value(common.ParentCtxKey).(context.Context)
	n.parent = parentCtx.Value(common.NodeCtxKey).(*tracker)
	if n.parent != nil {
		n.parent.addChild(n)
	}
	inst := parentCtx.Value(common.InstanceCtxKey).(instance)
	thread := ctx.Value(common.ThreadCtxKey).(*shredder.Thread)
	rm := ctx.Value(common.RenderMapCtxKey).(*common.RenderMap)
	var parentCinema *Cinema
	if n.parent != nil {
		parentCinema = n.parent.cinema
	}
	n.container = &container{
		id:           inst.NewId(),
		inst:         inst,
		parentCtx:    parentCtx,
		parentCinema: parentCinema,
		node:         n,
	}
	return n.container.render(thread, rm, w, n.content)
}
