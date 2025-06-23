package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type Core interface {
	Id() uint64
	Kill()
	RegisterAttrHook(ctx context.Context, h *AttrHook) (*HookEntry, bool)
	RegisterCallHook(ctx context.Context, name string, arg common.WritableRaw, hook *CallHook) (TryCancel, bool)
}

type instance interface {
	Setup(Core, *Cinema, context.Context)
	Thread() *shredder.Thread
	CancelHooks(uint64, error)
	CancelHook(uint64, uint64, error)
	RegisterHook(uint64, uint64, *NodeHook)
	NewId() uint64

	Call(Caller)
}

const noId uint64 = ^uint64(0)

func newCore(node *Node, ctx context.Context, id uint64) *core {
	inst, ok := ctx.Value(common.InstanceCtxKey).(instance)
	if !ok {
		panic(errors.New("Node rendered outside doors context"))
	}
	parent, _ := ctx.Value(common.NodeCtxKey).(*core)
	parentCinema, _ := ctx.Value(common.CinemaCtxKey).(*Cinema)
	ctx = context.WithValue(ctx, common.RenderMapCtxKey, nil)
	ctx = context.WithValue(ctx, common.ThreadCtxKey, nil)
	if id == noId {
		id = inst.NewId()
	}
	thread := inst.Thread()
	cinema := newCinema(parentCinema, inst, thread)
	return &core{
		node:      node,
		id:        id,
		instance:  inst,
		parent:    parent,
		thread:    thread,
		cinema:    cinema,
		children:  common.NewSet[*Node](),
		parentCtx: ctx,
		jsCallsMu: sync.Mutex{},
		jsCalls:   common.NewSet[*jsCall](),
	}
}

type core struct {
	node      *Node
	id        uint64
	instance  instance
	parent    *core
	thread    *shredder.Thread
	cinema    *Cinema
	children  common.Set[*Node]
	parentCtx context.Context
	cancel    context.CancelFunc
	jsCallsMu sync.Mutex
	jsCalls   common.Set[*jsCall]
}

func (c *core) removeJsCall(j *jsCall) {
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		c.jsCalls.Remove(j)
	})
}

func (c *core) isRoot() bool {
	return c.parent == nil
}

type TryCancel = func() bool

func (c *core) RegisterCallHook(ctx context.Context, name string, arg common.WritableRaw, hook *CallHook) (TryCancel, bool) {
	entry, ok := c.node.addHook(ctx, hook)
	if !ok {
		return nil, false
	}
	call := &jsCall{
		name:      name,
		arg:       arg,
		core:      c,
		hookEntry: entry,
	}
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			call.kill()
		}
		c.jsCalls.Add(call)
	})
	c.instance.Call(call)
	return call.cancel, true

}

func (c *core) RegisterAttrHook(ctx context.Context, h *AttrHook) (*HookEntry, bool) {
	return c.node.addHook(ctx, h)
}

func (c *core) Id() uint64 {
	return c.id
}

func (c *core) Kill() {
	c.node.suspend()
}

func (c *core) kill(init bool) {
	killed := c.thread.Kill(func() {
		c.cinema.kill(init)
		if c.cancel != nil {
			c.cancel()
		}
		for child := range c.children.Iter() {
			child.suspend()
		}
		for call := range c.jsCalls.Iter() {
			call.kill()
		}
	})
	if !killed {
		return
	}
	c.instance.CancelHooks(c.id, errors.New("element removed"))
}

func (c *core) addChild(child *Node) {
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			child.suspend()
			return
		}
		c.children.Add(child)
	})
}

func (c *core) removeChild(child *Node) {
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		c.children.Remove(child)
	})
}

func (c *core) renderUpdateCall(content templ.Component, ch chan<- Call, commitId uint) {
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			close(ch)
			return
		}
		rm := common.NewRenderMap()
		rw := rm.Writer(c.id)
		var err error
		t.Write(func(t *shredder.Thread) {
			if t == nil || content == nil {
				return
			}
			ctx := context.WithValue(c.parentCtx, common.ThreadCtxKey, t)
			ctx = context.WithValue(ctx, common.NodeCtxKey, c)
			ctx = context.WithValue(ctx, common.RenderMapCtxKey, rm)
			ctx = context.WithValue(ctx, common.CinemaCtxKey, c.cinema)
			ctx, c.cancel = context.WithCancel(ctx)
			err = content.Render(ctx, rw)
		})
		t.Write(func(t *shredder.Thread) {
			if t == nil {
				close(ch)
				return
			}
			if err != nil {
				rw.SubmitError(err)
			} else {
				rw.Submit()
			}
			ch <- &commitCall{
				name: "node_update",
				args: []common.Writable{common.WritableAny{c.id}, &common.WritableRenderMap{
					Rm:    rm,
					Index: c.id,
				}},
				id:   commitId,
				node: c.node,
			}
			close(ch)
		})
	})
}

func (c *core) renderReplaceCall(content templ.Component, ch chan<- Call, commitId uint) {
	thread := c.instance.Thread()
	rm := common.NewRenderMap()
	rw := rm.Writer(c.id)
	var err error
	thread.Write(func(t *shredder.Thread) {
		if content == nil {
			return
		}
		ctx := context.WithValue(c.parentCtx, common.ThreadCtxKey, t)
		ctx = context.WithValue(ctx, common.RenderMapCtxKey, rm)
		err = content.Render(ctx, rw)
	})
	thread.Write(func(t *shredder.Thread) {
		if err != nil {
			rw.SubmitError(err)
		} else {
			rw.Submit()
		}
		ch <- &commitCall{
			name: "node_replace",
			args: []common.Writable{common.WritableAny{c.id}, &common.WritableRenderMap{
				Rm:    rm,
				Index: c.id,
			}},
			id:   commitId,
			node: c.node,
		}
		close(ch)
	})
}
func (c *core) renderRemoveCall(ch chan<- Call, commitId uint) {
	ch <- &commitCall{
		name: "node_remove",
		args: []common.Writable{common.WritableAny{c.id}},
		id:   commitId,
		node: c.node,
	}
	close(ch)
}

func (c *core) render(ctx context.Context, w io.Writer, children templ.Component, content templ.Component, commit *commit) error {
	rm := ctx.Value(common.RenderMapCtxKey).(*common.RenderMap)
	thread := ctx.Value(common.ThreadCtxKey).(*shredder.Thread)
	rw := rm.Writer(c.id)
	if w != nil {
		err := rw.Holdplace(w)
		if err != nil {
			return err
		}
	}
	thread.Read(func(t *shredder.Thread) {
		if t == nil {
			rw.SubmitEmpty()
			return
		}
		var contentRendered bool
		var err error
		t.Write(func(t *shredder.Thread) {
			if t == nil {
				return
			}
			ctx = context.WithValue(ctx, common.NodeCtxKey, c)
			ctx = context.WithValue(ctx, common.CinemaCtxKey, c.cinema)
			if c.isRoot() {
				c.instance.Setup(c, c.cinema, ctx)
			}
			ctx = context.WithValue(ctx, common.ThreadCtxKey, t)
			ctx, c.cancel = context.WithCancel(ctx)
			contentRendered, err = c.writeRender(ctx, rw, children, content)
		})
		t.Write(func(t *shredder.Thread) {
			if t == nil {
				rw.SubmitEmpty()
				return
			}
			if err != nil {
				rw.SubmitError(err)
			} else {
				rw.Submit()
			}
			if commit == nil {
				return
			}
			if contentRendered {
				commit.result(err)
			} else {
				commit.owerwrite()
			}
		})

	}, shredder.W(c.thread))
	return nil
}

func (c *core) writeRender(ctx context.Context, rw *common.RenderWriter, children templ.Component, content templ.Component) (bool, error) {
	var err error
	if !c.isRoot() {
		_, err := rw.Write(fmt.Appendf(nil, "<do-or id =\"d00r/%d\">", c.id))
		if err != nil {
			return true, err
		}
	}
	before := rw.Len()
	err = children.Render(ctx, rw)
	if err != nil {
		return false, err
	}
	renderContent := (rw.Len()-before == 0)
	if renderContent && content != nil {
		err = content.Render(ctx, rw)
		if err != nil {
			rw.SubmitError(err)
			return true, err
		}
	}
	if !c.isRoot() {
		_, err = rw.Write([]byte("</do-or>"))
	}
	return renderContent, err
}
