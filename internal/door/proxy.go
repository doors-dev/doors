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
	done
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
		var err error
		r.state, err = r.init(job)
		if r.state != check {
			r.takoverFrame.Activate()
		}
		return err
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
			r.view.content = headLess{el: r.view.elem}
		}
		r.state = closing
		return r.Send(job)
	case closing:
		close, ok := job.(*gox.JobHeadClose)
		if ok && close.ID == r.headId {
			r.state = done
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
	case done:
		return errors.New("unexpected content after door is completed")
	default:
		panic("door: invalid state")
	}
}

func (r *proxyComponent) init(job gox.Job) (proxyRendererState, error) {
	node, isNode := job.(*node)
	if isNode {
		r.view.attrs = nil
		r.view.tag = ""
		r.view.content = r.view.elem
		open, close := r.view.headFrame(r.parentCtx, r.doorId, r.cursor.NewID())
		if err := r.cursor.Send(open); err != nil {
			return done, err
		}
		if err := r.cursor.Send(node); err != nil {
			return done, err
		}
		if err := r.cursor.Send(close); err != nil {
			return done, err
		}
		return done, nil
	}
	openJob, ok := job.(*gox.JobHeadOpen)
	if !ok {
		return done, errors.New("door: expected container")
	}
	var state proxyRendererState
	var buffered *gox.JobHeadOpen
	switch openJob.Kind {
	case gox.KindVoid:
		return done, errors.New("door: void tag can't be a door")
	case gox.KindContainer:
		gox.Release(openJob)
		r.view.attrs = nil
		r.view.tag = ""
	case gox.KindRegular:
		if openJob.Tag == "d0-r" || strings.EqualFold(openJob.Tag, "script") || strings.EqualFold(openJob.Tag, "style") || openJob.Attrs.Get("data-d0r").IsSet() {
			r.wrapOver = true
			r.view.attrs = nil
			r.view.tag = ""
			r.view.content = r.view.elem
			state = closing
			buffered = openJob
		} else {
			defer gox.Release(openJob)
			r.view.attrs = openJob.Attrs.Clone()
			r.view.tag = openJob.Tag
			if r.view.content == nil {
				r.view.content = headLess{el: r.view.elem}
				state = closing
			} else {
				state = check
			}
		}
	}
	r.headId = openJob.ID
	open, close := r.view.headFrame(r.parentCtx, r.doorId, r.cursor.NewID())
	r.close = close
	if err := r.cursor.Send(open); err != nil {
		return done, err
	}
	if buffered != nil {
		if err := r.cursor.Send(buffered); err != nil {
			return done, nil
		}
	}
	return state, nil
}

type headLess struct {
	el     gox.Elem
	headId uint64
	cur    gox.Cursor
}

func (h headLess) Main() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		h.cur = cur
		return h.el.Print(cur.Context(), &h)
	})
}

func (h *headLess) Send(j gox.Job) error {
	if h.headId == 0 {
		open, ok := j.(*gox.JobHeadOpen)
		if !ok {
			return errors.New("door: expected container")
		}
		h.headId = open.ID
		gox.Release(open)
		return nil
	}
	close, ok := j.(*gox.JobHeadClose)
	if ok && close.ID == h.headId {
		gox.Release(close)
		return nil
	}
	return h.cur.Send(j)
}
