package door

import (
	"context"
	"errors"
	"io"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type proxyState struct {
	ctx           context.Context
	element       gox.Elem
	content any
	head          *doorHead
	tracker       *tracker
	stateInit
	stateRender
	stateTakover
}

func (s *proxyState) takeover(prev state) {
	defer s.readyToRender()
	switch prev := prev.(type) {
	case *updateState:
		s.content = prev.content
	case *jobState:
		s.content = prev.content
	}
}

func (s *proxyState) Context() context.Context {
	return s.ctx
}

func (s *proxyState) Output(w io.Writer) error {
	return errors.New("door: used outside render pipeline")
}

func (s *proxyState) render(parent parent, p *pipe, th *sh.Thread) *tracker {
	newPipe := newPipe()
	p.put(newPipe)
	tracker := newTracker(parent)
	s.whenReadyToRender(func(bool) {
		s.tracker = tracker
		sh.Run(func(t *sh.Thread) {
			newPipe.thread = t
			defer newPipe.close()
			pipeCur := gox.NewCursor(s.tracker.ctx, newPipe)
			renderer := newProxyRender(s.tracker, pipeCur, s.content, func(head *doorHead, contentUsed bool) {
				s.head = head
				if !contentUsed {
					s.content = nil
				}
				s.initFinished()
			})
			s.element(gox.NewCursor(s.tracker.ctx, renderer))
			pipeCur.Func(func(io.Writer) error {
				s.readyForTakeover()
				return nil
			})
		}, th.R(), s.tracker.th.Wi())
	})
	return tracker
}

type updateContent struct {
	content any
	onUse   func(bool)
}

func (b *updateContent) SetUsed(used bool) {
	if b.onUse == nil {
		return
	}
	b.onUse(used)
}

func newProxyRender(tracker *tracker, cur gox.Cursor, updateContent any, hook func(head *doorHead, contentUsed bool)) *proxyRender {
	return &proxyRender{
		tracker:       tracker,
		updateContent: updateContent,
		hook:          hook,
		cur:           cur,
	}
}

type proxyRender struct {
	hook          func(head *doorHead, contentUsed bool)
	updateContent any
	hookTriggered bool
	head          *doorHead
	tracker       *tracker
	cur           gox.Cursor
	done          bool
	closeJob      *gox.JobHeadClose
	wrapOver      bool
	id            uint64
}

func (s *proxyRender) setHead(head *doorHead) {
	s.head = head
	if s.updateContent == nil {
		s.hookTriggered = true
		s.hook(s.head, false)
	}
}

func (s *proxyRender) open(job gox.Job) error {
	if s.done {
		return errors.New("door: unexpected content after root head is closed")
	}
	openJob, ok := job.(*gox.JobHeadOpen)
	if !ok {
		return errors.New("door: expects head as a root content of element")
	}
	switch openJob.Kind {
	case gox.KindVoid:
		return errors.New("door: void head is not allowed as a root content of element")
	case gox.KindContainer:
		defer gox.Release(openJob)
		s.setHead(&doorHead{})
		s.id = openJob.Id
		open, close := s.head.frame(s.tracker.ctx, s.tracker.id, s.cur.NewId())
		s.closeJob = close
		return s.cur.Job(open)
	case gox.KindRegular:
		s.id = openJob.Id
		if openJob.Tag == "d0-0r" || openJob.Attrs.Get("data-d00r").IsSet() || openJob.Tag == "script" || openJob.Tag == "style" {
			s.wrapOver = true
			s.setHead(&doorHead{})
			open, close := s.head.frame(s.tracker.ctx, s.tracker.id, s.cur.NewId())
			s.closeJob = close
			if err := s.cur.Job(open); err != nil {
				return err
			}
			return s.cur.Job(openJob)
		} else {
			defer gox.Release(openJob)
			s.setHead(&doorHead{
				tag:   openJob.Tag,
				attrs: openJob.Attrs.Clone(),
			})
			open, close := s.head.frame(s.tracker.ctx, s.tracker.id, s.cur.NewId())
			s.closeJob = close
			return s.cur.Job(open)
		}
	default:
		return errors.New("door: unexpected head kind")
	}
}

func (s *proxyRender) close(gox *gox.JobHeadClose) error {
	s.done = true
	if s.wrapOver {
		if err := s.cur.Job(gox); err != nil {
			return err
		}
	}
	return s.cur.Job(s.closeJob)
}

func (s *proxyRender) Send(job gox.Job) error {
	if s.head == nil {
		return s.open(job)
	}
	closeJob, ok := job.(*gox.JobHeadClose)
	if !s.hookTriggered {
		s.hookTriggered = true
		if ok && closeJob.Id == s.id && !s.wrapOver {
			s.hook(s.head, true)
			if err := s.cur.Any(s.updateContent); err != nil {
				return err
			}
		} else {
			s.hook(s.head, false)
		}
	}
	if ok && closeJob.Id == s.id {
		return s.close(closeJob)
	}
	return s.cur.Job(job)
}
