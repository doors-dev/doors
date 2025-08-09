package node

import (
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/shredder"
	"golang.org/x/net/context"
)

type Core interface {
	Id() uint64
	Cinema() *Cinema
	RegisterAttrHook(ctx context.Context, h *AttrHook) (*HookEntry, bool)
	RegisterClientCall(ctx context.Context, call *ClientCall) (func(), bool)
}

type container struct {
	id           uint64
	inst         instance
	parentCtx    context.Context
	parentCinema *Cinema
	tracker      *tracker
	node         *Node
}

func (c *container) suspend() {
	c.tracker.suspend(false)
}

func (c *container) render(thread *shredder.Thread, rm *common.RenderMap, w io.Writer, content templ.Component) error {
	rw, _ := rm.Writer(c.id)
	err := rw.Holdplace(w)
	if err != nil {
		return err
	}
	tracker, ctx := c.newTacker()
	thread.Read(func(t *shredder.Thread) {
		if t == nil || ctx.Err() != nil {
			rw.SubmitEmpty()
			return
		}
		var err error
		t.Write(func(t *shredder.Thread) {
			if t == nil || ctx.Err() != nil {
				return
			}
			ctx = context.WithValue(ctx, common.NodeCtxKey, tracker)
			ctx = context.WithValue(ctx, common.RenderMapCtxKey, rm)
			ctx = context.WithValue(ctx, common.ThreadCtxKey, t)
			_, err = rw.Write(fmt.Appendf(nil, "<do-or id =\"d00r/%d\">", c.id))
			if err != nil {
				return
			}
			if content != nil {
				err := content.Render(ctx, rw)
				if err != nil {
					return
				}
			}
			_, err = rw.Write([]byte("</do-or>"))
		})
		t.Write(func(t *shredder.Thread) {
			if t == nil || ctx.Err() != nil {
				rw.SubmitEmpty()
				return
			}
			if err != nil {
				rw.SubmitError(err)
			} else {
				rw.Submit()
			}
		})
	}, shredder.W(tracker.thread))
	c.tracker = tracker
	return nil
}

func (c *container) replace(userCtx context.Context, content templ.Component) <-chan error {
	c.tracker.suspend(true)

	thread := c.inst.Thread()
	rm := common.NewRenderMap()
	rw, _ := rm.Writer(c.id)

	call := &nodeCall{
		ctx:  c.parentCtx,
		name: "node_replace",
		arg:  c.id,
		ch:   make(chan error, 1),
		done: ctxwg.Add(userCtx),
		payload: &common.WritableRenderMap{
			Rm:    rm,
			Index: c.id,
		},
	}

	var err error

	thread.Write(func(t *shredder.Thread) {
		if t == nil || c.parentCtx.Err() != nil {
			return
		}
		ctx := context.WithValue(c.parentCtx, common.RenderMapCtxKey, rm)
		ctx = context.WithValue(ctx, common.ThreadCtxKey, t)
		if content != nil {
			err = content.Render(ctx, rw)
		}
	})

	thread.Write(func(t *shredder.Thread) {
		if t == nil || c.parentCtx.Err() != nil {
			call.stale()
			return
		}
		if err != nil {
			call.Result(err)
			return
		}
		rw.Submit()
		c.inst.Call(call)
	})

	return call.ch
}

func (c *container) update(userCtx context.Context, content templ.Component) <-chan error {
	if content == nil {
		content = templ.NopComponent
	}
	c.tracker.suspend(true)

	tracker, ctx := c.newTacker()
	rm := common.NewRenderMap()
	rw, _ := rm.Writer(c.id)

	call := &nodeCall{
		ctx:  ctx,
		name: "node_update",
		arg:  c.id,
		ch:   make(chan error, 1),
		done: ctxwg.Add(userCtx),
		payload: &common.WritableRenderMap{
			Rm:    rm,
			Index: c.id,
		},
	}

	tracker.thread.Write(func(t *shredder.Thread) {
		if t == nil || ctx.Err() != nil {
			call.stale()
			return
		}

		var err error

		t.Write(func(t *shredder.Thread) {
			if t == nil || ctx.Err() != nil {
				return
			}
			ctx := context.WithValue(ctx, common.NodeCtxKey, tracker)
			ctx = context.WithValue(ctx, common.RenderMapCtxKey, rm)
			ctx = context.WithValue(ctx, common.ThreadCtxKey, t)
			if content != nil {
				err = content.Render(ctx, rw)
			}
		})
		t.Write(func(t *shredder.Thread) {
			if t == nil || ctx.Err() != nil {
				call.stale()
				return
			}
			if err != nil {
				call.Result(err)
				return
			}
			rw.Submit()
			c.inst.Call(call)
		})
	})

	c.tracker = tracker

	return call.ch
}


func (n *container) remove(userCtx context.Context) <-chan error {
	return n.replace(userCtx, nil)
}

func (c *container) newTacker() (*tracker, context.Context) {
	thread := c.inst.Thread()
	ctx, cancel := context.WithCancel(c.parentCtx)
	cinema := newCinema(c.parentCinema, c.inst, thread, c.id)
	return &tracker{
		cinema:      cinema,
		children:    common.NewSet[*Node](),
		thread:      thread,
		cancel:      cancel,
		container:   c,
		clientCalls: common.NewSet[*clientCall](),
	}, ctx
}

func (c *container) instance() instance {
	return c.inst
}

func (c *container) registerHook(tracker *tracker, ctx context.Context, h Hook) (*HookEntry, bool) {
	return c.node.registerHook(c, tracker, ctx, h)
}

func (c *container) getId() uint64 {
	return c.id
}

type trackerContainer interface {
	registerHook(tracker *tracker, ctx context.Context, h Hook) (*HookEntry, bool)
	instance() instance
	getId() uint64
}

type tracker struct {
	cinema      *Cinema
	children    common.Set[*Node]
	thread      *shredder.Thread
	cancel      context.CancelFunc
	container   trackerContainer
	clientCalls common.Set[*clientCall]
}

func (c *tracker) Cinema() *Cinema {
	return c.cinema
}

func (c *tracker) Id() uint64 {
	return c.container.getId()
}

func (c *tracker) RegisterClientCall(ctx context.Context, call *ClientCall) (func(), bool) {
	cc := &clientCall{
		call:    call,
		tracker: c,
	}
	if call.Trigger != nil {
		entry, ok := c.container.registerHook(c, ctx, cc)
		if !ok {
			return nil, false
		}
		cc.hookEntry = entry
	}
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			cc.kill()
		}
		c.clientCalls.Add(cc)
	})
	c.container.instance().Call(cc)
	return func() {
		cc.cancelCall(nil)
	}, true
}

func (c *tracker) RegisterAttrHook(ctx context.Context, h *AttrHook) (*HookEntry, bool) {
	return c.container.registerHook(c, ctx, h)
}

func (c *tracker) removeClientCall(cc *clientCall) {
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		c.clientCalls.Remove(cc)
	})
}

func (c *tracker) addChild(node *Node) {
	if c == nil {
		return
	}
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			node.suspend(c)
			return
		}
		c.children.Add(node)
	})
}

func (c *tracker) removeChild(node *Node) {
	if c == nil {
		return
	}
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		c.children.Remove(node)
	})
}

func (t *tracker) suspend(init bool) {
	t.cancel()
	t.thread.Kill(func() {
		t.cinema.kill(init)
		for child := range t.children.Iter() {
			child.suspend(t)
		}
		for call := range t.clientCalls.Iter() {
			call.kill()
		}
	})
}
