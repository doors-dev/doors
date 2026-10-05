package door

import (
	"errors"
	"strings"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type nodePrinter struct {
	pipe        *pipe
	skipContent bool
	ready       bool
	open        *gox.JobOpen
	close       *gox.JobClose
}

func (r *nodePrinter) submitContainer() error {
	if r.open != nil && r.close == nil {
		return errors.New("door container tag was not closed")
	}
	var openJob *gox.JobOpen
	var closeJob *gox.JobClose
	if r.open != nil && r.open.Kind == gox.KindContainer {
		gox.Release(r.open)
		gox.Release(r.close)
		r.open = nil
		r.close = nil
	}
	if r.open == nil {
		attrs := gox.NewAttrs()
		front.AttrsSetDoor(attrs, r.pipe.id(), true)
		front.AttrsSetParent(attrs, r.pipe.parentID())
		openJob = gox.NewJobOpen(r.pipe.outerContext(), 0, gox.KindRegular, "d0-r", attrs)
		closeJob = gox.NewJobClose(r.pipe.outerContext(), 0, gox.KindRegular, "d0-r")
	} else {
		r.open.Ctx = r.pipe.outerContext()
		r.close.Ctx = r.pipe.outerContext()
		front.AttrsSetDoor(r.open.Attrs, r.pipe.id(), false)
		front.AttrsSetParent(r.open.Attrs, r.pipe.parentID())
		openJob = r.open
		closeJob = r.close
		r.open = nil
		r.close = nil
	}
	if err := r.pipe.presend(openJob); err != nil {
		return err
	}
	if err := r.pipe.Send(closeJob); err != nil {
		return err
	}
	return nil
}

func (r *nodePrinter) pipeSend(job gox.Job) error {
	if r.skipContent {
		if rel, ok := job.(gox.Releaser); ok {
			gox.Release(rel)
		}
		return nil
	}
	return r.pipe.Send(job)
}

func (r *nodePrinter) pipePresend(job *gox.JobOpen) error {
	if r.skipContent {
		gox.Release(job)
		return nil
	}
	return r.pipe.presend(job)
}

func (r *nodePrinter) Send(job gox.Job) error {
	if !r.ready {
		return r.init(job)
	}
	if r.open == nil {
		return r.pipeSend(job)
	}
	if r.close != nil {
		openJob := r.open
		closeJob := r.close
		r.open = nil
		r.close = nil
		if err := r.pipePresend(openJob); err != nil {
			return err
		}
		if err := r.Send(closeJob); err != nil {
			return err
		}
		return r.pipeSend(job)
	}
	if closeJob, ok := job.(*gox.JobClose); ok {
		if closeJob.ID == r.open.ID {
			r.close = closeJob
			return nil
		}
	}
	return r.pipeSend(job)
}

func (r *nodePrinter) init(job gox.Job) error {
	switch job := job.(type) {
	case *gox.JobOpen:
		r.ready = true
		return r.initOpenJob(job)
	default:
		r.ready = true
		return r.pipeSend(job)
	}
}

func (r *nodePrinter) initOpenJob(openJob *gox.JobOpen) error {
	switch openJob.Kind {
	case gox.KindRegular:
		if strings.EqualFold(openJob.Tag, "head") {
			return errors.New("door does not support <head> as a container")
		}
		if strings.EqualFold(openJob.Tag, "title") {
			return r.pipeSend(openJob)
		}
		if strings.EqualFold(openJob.Tag, "script") {
			return r.pipeSend(openJob)
		}
		if strings.EqualFold(openJob.Tag, "style") {
			return r.pipeSend(openJob)
		}
		if openJob.Tag == "d0-r" {
			return r.pipeSend(openJob)
		}
		if openJob.Attrs.Has("data-d0c") {
			return r.pipeSend(openJob)
		}
		if openJob.Attrs.Has("data-d0r") {
			return r.pipeSend(openJob)
		}
		if openJob.Tag == "" {
			return r.pipeSend(openJob)
		}
		r.open = openJob
		return nil
	case gox.KindContainer:
		r.open = openJob
		return nil
	case gox.KindVoid:
		return r.pipeSend(openJob)
	default:
		panic("unknown gox head kind")
	}
}
