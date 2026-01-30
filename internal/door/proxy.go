package door

import (
	"context"
	"errors"
	"strings"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type proxyRendererState int

const (
	initial proxyRendererState = iota
	check
	closing
)

func newProxyComponent(doorId uint64, view *view, parentCtx context.Context, takoverFrame *shredder.ValveFrame) *proxyComponent {
	return &proxyComponent{
		doorId:       doorId,
		view:         view,
		parentCtx:    parentCtx,
		takoverFrame: takoverFrame,
	}
}

type proxyComponent struct {
	state        proxyRendererState
	wrapOver     bool
	cursor       gox.Cursor
	doorId       uint64
	view         *view
	headId       uint64
	close        *gox.JobHeadClose
	parentCtx    context.Context
	takoverFrame *shredder.ValveFrame
}

func (r *proxyComponent) Main() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		r.cursor = cur
		return r.view.elem.Print(cur.Context(), r)
	})
}

func (r *proxyComponent) Send(job gox.Job) error {
	switch r.state {
	case initial:
		err := r.init(job)
		if err != nil {
			return err
		}
		if r.view.content != nil {
			r.state = check
		} else {
			r.takoverFrame.Activate()
			r.state = closing
		}
		return nil
	case check:
		defer r.takoverFrame.Activate()
		close, ok := job.(*gox.JobHeadClose)
		if ok {
			if close.ID != r.headId {
				return errors.New("door: invalid close")
			}
			err := r.view.renderContent(r.cursor)
			if err != nil {
				return err
			}
		} else {
			r.view.content = nil
		}
		r.state = closing
		return r.Send(job)
	case closing:
		close, ok := job.(*gox.JobHeadClose)
		if ok && close.ID == r.headId {
			if r.wrapOver {
				err := r.cursor.Send(close)
				if err != nil {
					return err
				}
			} else {
				gox.Release(close)
			}
			return r.cursor.Send(r.close)
		}
		return r.cursor.Send(job)
	default:
		panic("door: invalid state")
	}
}

func (r *proxyComponent) init(job gox.Job) error {
	openJob, ok := job.(*gox.JobHeadOpen)
	if !ok {
		return errors.New("door: expected container")
	}
	switch openJob.Kind {
	case gox.KindVoid:
		return errors.New("door: void tag can't be a door")
	case gox.KindContainer:
		gox.Release(openJob)
		r.view.attrs = nil
		r.view.tag = ""
	case gox.KindRegular:
		if openJob.Tag == "d0-0r" || strings.EqualFold(openJob.Tag, "script") || strings.EqualFold(openJob.Tag, "style") || openJob.Attrs.Get("data-d00r").IsSet() {
			r.wrapOver = true
			r.view.attrs = nil
			r.view.tag = ""
			if err := r.cursor.Send(openJob); err != nil {
				return err
			}
		} else {
			gox.Release(openJob)
			r.view.attrs = openJob.Attrs.Clone()
			r.view.tag = openJob.Tag
		}
	}
	r.headId = openJob.ID
	open, close := r.view.headFrame(r.parentCtx, r.doorId, r.cursor.NewID())
	r.close = close
	return r.cursor.Send(open)
}
