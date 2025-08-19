package door

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
	RegisterHook(uint64, uint64, *DoorHook)
	NewId() uint64
	Call(common.Call)
}

type doorMode int

const (
	dynamic doorMode = iota
	static
	removed
)

type Door struct {
	Tag       string
	A     templ.Attributes
	mu        sync.Mutex
	parent    *tracker
	container *container
	content   templ.Component
	mode      doorMode
}

func (n *Door) registerHook(container *container, tracker *tracker, ctx context.Context, h Hook) (*HookEntry, bool) {
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
		DoorId: n.container.id,
		HookId: hookId,
		inst:   n.container.inst,
	}, true
}

func (n *Door) suspend(parent *tracker) {
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

func (n *Door) reload(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx, "Door reload")
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

func (n *Door) clear(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx, "Door clear")
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
func (n *Door) update(ctx context.Context, content templ.Component) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx, "Door update")
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

func (n *Door) remove(ctx context.Context) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx, "Door remove")
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

func (n *Door) replace(ctx context.Context, content templ.Component) <-chan error {
	n.mu.Lock()
	defer n.mu.Unlock()
	ch, ok := common.ResultChannel(ctx, "Door replace")
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

func (n *Door) Render(ctx context.Context, w io.Writer) error {
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
	n.parent = parentCtx.Value(common.DoorCtxKey).(*tracker)
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
		door:         n,
	}
	return n.container.render(thread, rm, w, n.Tag, n.A, n.content)
}
