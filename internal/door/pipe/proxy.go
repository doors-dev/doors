package pipe

import (
	"errors"
	"strings"

	"github.com/doors-dev/gox"
)

type proxyPrinter struct {
	wrapper wrapper
	pip     *pipe
}

func (r *proxyPrinter) finalize() (ProxyContainer, error) {
	if r.wrapper == nil {
		r.wrapper = wrapperContainer{}
	}
	if cont, ok := r.wrapper.(wrapperContainer); ok {
		return cont.newContainer(), nil
	}
	wrapper := r.wrapper.(*wrapperHead)
	if !wrapper.isClosed() {
		return ProxyContainer{}, errors.New("non-closed head")
	}
	return wrapper.newContainer(), nil
}

func (r *proxyPrinter) Send(job gox.Job) error {
	if r.wrapper == nil {
		return r.init(job)
	}
	wrapper, ok := r.wrapper.(*wrapperHead)
	if !ok {
		return r.pip.Send(job)
	}
	if wrapper.isClosed() {
		wrapper.dispose(r.pip)
		r.wrapper = wrapperContainer{}
		return r.pip.Send(job)
	}
	if wrapper.tryClose(job) {
		return nil
	}
	return r.pip.Send(job)
}

func (r *proxyPrinter) init(job gox.Job) error {
	openJob, isOpen := job.(*gox.JobHeadOpen)
	if !isOpen {
		r.wrapper = wrapperContainer{}
		return r.pip.Send(job)
	}
	if openJob.Kind == gox.KindVoid {
		r.wrapper = wrapperContainer{}
		return r.pip.Send(job)
	}
	if openJob.Kind == gox.KindRegular {
		if strings.EqualFold(openJob.Tag, "script") {
			r.wrapper = wrapperContainer{}
			return r.pip.Send(job)
		}
		if strings.EqualFold(openJob.Tag, "style") {
			r.wrapper = wrapperContainer{}
			return r.pip.Send(job)
		}
		if openJob.Tag == "d0-r" {
			r.wrapper = wrapperContainer{}
			return r.pip.Send(job)
		}
		if openJob.Attrs.Has("data-d0r") {
			r.wrapper = wrapperContainer{}
			return r.pip.Send(job)
		}
		if openJob.Tag == "" {
			r.wrapper = wrapperContainer{}
			return r.pip.Send(job)
		}
	}
	r.wrapper = &wrapperHead{open: openJob}
	return nil
}

type wrapperContainer struct{}

func (w wrapperContainer) newContainer() ProxyContainer {
	return ProxyContainer{}
}

type wrapper interface {
	newContainer() ProxyContainer
}

type wrapperHead struct {
	open  *gox.JobHeadOpen
	close *gox.JobHeadClose
}

func (w *wrapperHead) newContainer() ProxyContainer {
	if !w.isClosed() {
		panic("wrapper must be closed to newContainer")
	}
	if w.open.Kind == gox.KindContainer {
		gox.Release(w.open)
		gox.Release(w.close)
		w.open = nil
		w.close = nil
		return ProxyContainer{}
	}
	tag := w.open.Tag
	attrs := w.open.Attrs.Clone()
	gox.Release(w.open)
	gox.Release(w.close)
	w.open = nil
	w.close = nil
	return ProxyContainer{
		Tag:   tag,
		Attrs: attrs,
	}
}

func (w *wrapperHead) dispose(pip *pipe) {
	pip.unshift(w.open)
	pip.push(w.close)
	w.close = nil
	w.open = nil
}

func (w *wrapperHead) tryClose(job gox.Job) bool {
	closeJob, isClose := job.(*gox.JobHeadClose)
	if !isClose {
		return false
	}
	if closeJob.ID != w.open.ID {
		return false
	}
	w.close = closeJob
	return true
}

func (w *wrapperHead) isClosed() bool {
	return w.close != nil
}
