package door2

import (
	"context"
	"errors"
	"strings"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type proxyRendererState int

const (
	initial proxyRendererState = iota
	check
	closing
)

func newProxyRenderer(doorId uint64, cursor gox.Cursor, view *view, parentCtx context.Context) *proxyRenderer {
	return &proxyRenderer{
		doorId: doorId,
		cursor: cursor,
		view:   view,
		parentCtx: parentCtx,
	}
}

type proxyRenderer struct {
	state     proxyRendererState
	wrapOver  bool
	cursor    gox.Cursor
	doorId    uint64
	view      *view
	headId    uint64
	close     *gox.JobHeadClose
	initReady sh.ValveFrame
	parentCtx context.Context
}

func (r *proxyRenderer) render() error {
	return r.view.elem.Print(r.cursor.Context(), r)
}

func (r *proxyRenderer) InitFrame() sh.SimpleFrame {
	return &r.initReady
}

func (r *proxyRenderer) Send(job gox.Job) error {
	switch r.state {
	case initial:
		err := r.init(job)
		if err != nil {
			return err
		}
		if r.view.content != nil {
			r.state = check
		} else {
			r.initReady.Activate()
			r.state = closing
		}
		return nil
	case check:
		defer r.initReady.Activate()
		close, ok := job.(*gox.JobHeadClose)
		if ok {
			if close.Id != r.headId {
				return errors.New("door: invalid close")
			}
			err := r.cursor.Any(r.view.content)
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
		if ok && close.Id == r.headId {
			if r.wrapOver {
				err := r.cursor.Job(close)
				if err != nil {
					return err
				}
			} else {
				gox.Release(close)
			}
			return r.cursor.Job(r.close)
		}
		return r.cursor.Job(job)
	default:
		panic("door: invalid state")
	}
}

func (r *proxyRenderer) init(job gox.Job) error {
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
			if err := r.cursor.Job(openJob); err != nil {
				return err
			}
		} else {
			gox.Release(openJob)
			r.view.attrs = openJob.Attrs.Clone()
			r.view.tag = openJob.Tag
		}
	}
	r.headId = openJob.Id
	open, close := r.view.headFrame(r.parentCtx, r.doorId, r.cursor.NewId())
	r.close = close
	return r.cursor.Job(open)
}
