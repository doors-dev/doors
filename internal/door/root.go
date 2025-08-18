package door

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

func NewRoot(ctx context.Context, inst instance) *Root {
	thread := inst.Thread()
	id := inst.NewId()
	ctx, cancel := context.WithCancel(ctx)
	t := &tracker{
		cinema:      newCinema(nil, inst, thread, id),
		children:    common.NewSet[*Door](),
		thread:      thread,
		cancel:      cancel,
		clientCalls: common.NewSet[*clientCall](),
	}
	r := &Root{
		id:      id,
		inst:    inst,
		ctx:     context.WithValue(ctx, common.DoorCtxKey, t),
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
	parentCtx := context.WithValue(r.ctx, common.ParentCtxKey, r.ctx)
	r.tracker.thread.Write(func(t *shredder.Thread) {
		if t == nil {
			close(ch)
			return
		}

		rm := common.NewRenderMap()
		rw, _ := rm.Writer(r.id)

		var err error
		t.Write(func(t *shredder.Thread) {
			if t == nil {
				close(ch)
				return
			}

			ctx := context.WithValue(parentCtx, common.RenderMapCtxKey, rm)
			ctx = context.WithValue(ctx, common.ThreadCtxKey, t)
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

			ch <- &RootRender{
				id:  r.id,
				rm:  rm,
				err: err,
			}
			close(ch)
		})
	})
	return ch
}
