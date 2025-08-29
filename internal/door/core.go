// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package door

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
}

type container struct {
	id           uint64
	inst         instance
	parentCtx    context.Context
	parentCinema *Cinema
	tracker      *tracker
	door         *Door
}

func (c *container) suspend() {
	c.tracker.suspend(false)
}

func (c *container) render(thread *shredder.Thread, rm *common.RenderMap, w io.Writer, tag string, attrs templ.Attributes, content templ.Component) error {
	if attrs == nil {
		attrs = make(templ.Attributes, 2)
	}

	attrs["id"] = fmt.Sprintf("d00r/%d", c.id)
	if tag != "" {
		attrs["data-d00r"] = true
		tag = templ.EscapeString(tag)
	} else {
		tag = "d0-0r"
	}

	tag = templ.EscapeString(tag)
	rw, _ := rm.Writer(c.id)
	err := rw.Holdplace(w)
	if err != nil {
		return err
	}
	tracker, parentCtx := c.newTacker()
	thread.Read(func(t *shredder.Thread) {
		if t == nil || parentCtx.Err() != nil {
			rw.SubmitEmpty()
			return
		}
		var err error
		t.Write(func(t *shredder.Thread) {
			if t == nil || parentCtx.Err() != nil {
				return
			}
			ctx := context.WithValue(parentCtx, common.CtxKeyParent, parentCtx)
			ctx = context.WithValue(ctx, common.CtxKeyRenderMap, rm)
			ctx = context.WithValue(ctx, common.CtxKeyThread, t)
			_, err = fmt.Fprintf(rw, "<%s", tag)
			if err != nil {
				return
			}
			err = templ.RenderAttributes(ctx, rw, attrs)
			if err != nil {
				return
			}
			_, err = rw.Write([]byte{'>'})
			if err != nil {
				return
			}
			if content != nil {
				err = content.Render(ctx, rw)
				if err != nil {
					return
				}
			}
			_, err = fmt.Fprintf(rw, "</%s>", tag)
		})
		t.Write(func(t *shredder.Thread) {
			if t == nil || parentCtx.Err() != nil {
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

func (c *container) replace(userCtx context.Context, content templ.Component, ch chan error) {
	c.tracker.suspend(true)

	thread := c.inst.Thread()
	rm := common.NewRenderMap()
	rw, _ := rm.Writer(c.id)

	call := &doorCall{
		ctx:  c.parentCtx,
		name: "door_replace",
		arg:  c.id,
		ch:   ch,
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
		ctx := context.WithValue(c.parentCtx, common.CtxKeyParent, c.parentCtx)
		ctx = context.WithValue(ctx, common.CtxKeyRenderMap, rm)
		ctx = context.WithValue(ctx, common.CtxKeyThread, t)
		if content != nil {
			err = content.Render(ctx, rw)
		}
	})

	thread.Write(func(t *shredder.Thread) {
		if t == nil || c.parentCtx.Err() != nil {
			call.Cancel()
			return
		}
		if err != nil {
			call.Result(nil, err)
			return
		}
		rw.Submit()
		c.inst.Call(call)
	})
}

func (c *container) update(userCtx context.Context, content templ.Component, ch chan error) {
	if content == nil {
		content = templ.NopComponent
	}
	c.tracker.suspend(true)

	tracker, parentCtx := c.newTacker()
	rm := common.NewRenderMap()
	rw, _ := rm.Writer(c.id)

	call := &doorCall{
		ctx:  parentCtx,
		name: "door_update",
		arg:  c.id,
		ch:   ch,
		done: ctxwg.Add(userCtx),
		payload: &common.WritableRenderMap{
			Rm:    rm,
			Index: c.id,
		},
	}

	tracker.thread.Write(func(t *shredder.Thread) {
		if t == nil || parentCtx.Err() != nil {
			call.Cancel()
			return
		}

		var err error

		t.Write(func(t *shredder.Thread) {
			if t == nil || parentCtx.Err() != nil {
				return
			}
			ctx := context.WithValue(parentCtx, common.CtxKeyParent, parentCtx)
			ctx = context.WithValue(ctx, common.CtxKeyRenderMap, rm)
			ctx = context.WithValue(ctx, common.CtxKeyThread, t)
			if content != nil {
				err = content.Render(ctx, rw)
			}
		})
		t.Write(func(t *shredder.Thread) {
			if t == nil || parentCtx.Err() != nil {
				call.Cancel()
				return
			}
			if err != nil {
				call.Result(nil, err)
				return
			}
			rw.Submit()
			c.inst.Call(call)
		})
	})
	c.tracker = tracker
}

func (n *container) remove(userCtx context.Context, ch chan error) {
	n.replace(userCtx, nil, ch)
}

func (n *container) clear(userCtx context.Context, ch chan error) {
	n.update(userCtx, nil, ch)
}

func (c *container) newTacker() (*tracker, context.Context) {
	thread := c.inst.Thread()
	ctx, cancel := context.WithCancel(c.parentCtx)
	cinema := newCinema(c.parentCinema, c.inst, thread, c.id)
	t := &tracker{
		cinema:      cinema,
		children:    common.NewSet[*Door](),
		thread:      thread,
		cancel:      cancel,
		container:   c,
	}
	ctx = context.WithValue(ctx, common.CtxKeyDoor, t)
	return t, ctx
}

func (c *container) instance() instance {
	return c.inst
}

func (c *container) registerHook(tracker *tracker, ctx context.Context, h Hook) (*HookEntry, bool) {
	return c.door.registerHook(c, tracker, ctx, h)
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
	children    common.Set[*Door]
	thread      *shredder.Thread
	cancel      context.CancelFunc
	container   trackerContainer
}

func (c *tracker) Cinema() *Cinema {
	return c.cinema
}

func (c *tracker) Id() uint64 {
	return c.container.getId()
}


func (c *tracker) RegisterAttrHook(ctx context.Context, h *AttrHook) (*HookEntry, bool) {
	return c.container.registerHook(c, ctx, h)
}


func (c *tracker) addChild(door *Door) {
	if c == nil {
		return
	}
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			door.suspend(c)
			return
		}
		c.children.Add(door)
	})
}

func (c *tracker) removeChild(door *Door) {
	if c == nil {
		return
	}
	c.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		c.children.Remove(door)
	})
}

func (t *tracker) suspend(init bool) {
	t.cancel()
	t.thread.Kill(func() {
		t.cinema.kill(init)
		for child := range t.children.Iter() {
			child.suspend(t)
		}
	})
}
