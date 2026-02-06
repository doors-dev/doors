package door

import (
	"context"
	"errors"
	"fmt"
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
		err := r.view.elem.Print(cur.Context(), r)
		if err != nil {
			r.takoverFrame.Activate()
			return err
		}
		if r.state == done {
			return nil
		}
		if r.state != initial {
			return fmt.Errorf("door [%d]: invalid state %d", r.doorId, r.state)
		}
		id := r.cursor.NewID()
		if err := r.Send(gox.NewJobHeadOpen(id, gox.KindContainer, "", cur.Context(), gox.NewAttrs())); err != nil {
			return err
		}
		if err := r.Send(gox.NewJobHeadClose(id, gox.KindContainer, "", cur.Context())); err != nil {
			return err
		}
		return nil
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
			var err error
			if comp, ok := r.view.content.(gox.Comp); ok {
				err = comp.Main()(r.cursor)
			} else {
				err = r.cursor.Any(r.view.content)
			}
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
	openJob, isOpen := job.(*gox.JobHeadOpen)
	switch true {
	case isNode:
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
	case !isOpen:
		return done, errors.New("door: expected container")
	case openJob.Kind == gox.KindVoid:
		return done, errors.New("door: void tag can't be a door")
	case openJob.Kind == gox.KindRegular && (openJob.Tag == "d0-r" || strings.EqualFold(openJob.Tag, "script") || strings.EqualFold(openJob.Tag, "style") || openJob.Attrs.Get("data-d0r").IsSet()):
		r.wrapOver = true
		r.view.attrs = nil
		r.view.tag = ""
		r.view.content = r.view.elem
		r.headId = openJob.ID
		open, close := r.view.headFrame(r.parentCtx, r.doorId, r.cursor.NewID())
		r.close = close
		if err := r.cursor.Send(open); err != nil {
			return done, err
		}
		if err := r.cursor.Send(openJob); err != nil {
			return done, err
		}
		return closing, nil
	case openJob.Kind == gox.KindContainer:
		r.view.attrs = nil
		r.view.tag = ""
	case openJob.Kind == gox.KindRegular:
		r.view.attrs = openJob.Attrs.Clone()
		r.view.tag = openJob.Tag
	}
	defer gox.Release(openJob)
	r.headId = openJob.ID
	state := check
	if r.view.content == nil {
		state = closing
		r.view.content = headLess{el: r.view.elem}
	}
	open, close := r.view.headFrame(r.parentCtx, r.doorId, r.cursor.NewID())
	r.close = close
	if err := r.cursor.Send(open); err != nil {
		return done, err
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
