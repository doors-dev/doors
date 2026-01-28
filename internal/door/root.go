// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package door

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/doors/internal/shredder"
)

func NewRoot(ctx context.Context, inst instance) *Root {
	thread := inst.Thread()
	id := inst.NewId()
	t := &tracker{
		cinema:   newCinema(nil, inst, thread, id),
		children: common.NewSet[*Door](),
		thread:   thread,
		cancel: func() {}
	}
	r := &Root{
		id:      id,
		inst:    inst,
		ctx:     context.WithValue(ctx, common.CtxKeyDoor, t),
		tracker: t,
	}
	t.container = r
	return r
}

type Root struct {
	id      uint64
	tracker *tracker
	inst    instance
	ctx     context.Context
}

func (r *Root) Go(f func()) {
	sh.Go(func() {
		err := common.Catch(f)
		if err != nil {
			r.inst.OnPanic(err)
		}
	})
}

func (r *Root) Ctx() context.Context {
	return r.ctx
}

func (r *Root) Cinema() *Cinema {
	return r.tracker.cinema
}

func (r *Root) Kill() {
	r.tracker.suspend(false)
}

func (r *Root) getId() uint64 {
	return r.id
}

func (r *Root) instance() instance {
	return r.inst
}

func (r *Root) registerHook(tracker *tracker, ctx context.Context, h Hook) (*HookEntry, bool) {
	if ctx.Err() != nil {
		return nil, false
	}
	hookId := r.inst.NewId()
	hook := newHook(common.ClearBlockingCtx(ctx), h, r.inst)
	r.inst.RegisterHook(r.id, hookId, hook)
	return &HookEntry{
		DoorId: r.id,
		HookId: hookId,
		inst:   r.inst,
	}, true
}

type RootRender struct {
	id  uint64
	rm  *common.RenderMap
	err error
}

func (r *RootRender) Err() error {
	return r.err
}

func (r *RootRender) InitImportMap(c *common.CSPCollector) {
	r.rm.InitImportMap(c)
}

func (r *RootRender) Write(w io.Writer) error {
	err := r.rm.Render(w, r.id)
	r.rm.Destroy()
	if r.err != nil {
		return r.err
	}
	return err
}

func (r *Root) Render(content templ.Component) <-chan *RootRender {
	ch := make(chan *RootRender, 1)
	parentCtx := context.WithValue(r.ctx, common.CtxKeyParent, r.ctx)
	shredder.Run(func(t *shredder.Thread) {
		if t == nil {
			close(ch)
			return
		}

		rm := common.NewRenderMap()
		rw, _ := rm.Writer(r.id)

		var err error
		shredder.Run(func(t *shredder.Thread) {
			if t == nil {
				close(ch)
				return
			}

			ctx := context.WithValue(parentCtx, common.CtxKeyRenderMap, rm)
			ctx = context.WithValue(ctx, common.CtxKeyThread, t)
			err = content.Render(ctx, rw)
		}, shredder.W(t))
		shredder.Run(func(t *shredder.Thread) {
			if t == nil {
				close(ch)
				return
			}

			if err != nil {
				rw.SubmitError(err)
			} else {
				rw.Submit()
			}

			ch <- &RootRender{
				id:  r.id,
				rm:  rm,
				err: err,
			}
			close(ch)
		}, shredder.W(t))
	}, shredder.W(r.tracker.thread))
	return ch
}
