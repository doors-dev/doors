package door

import (
	"context"
	"errors"
	"io"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type jobState struct {
	ctx          context.Context
	head         *doorHead
	dynamic      bool
	content      any
	proxyElement *gox.Elem
	tracker      *tracker
	stateInit
	stateRender
	stateTakover
}

func (s *jobState) takeover(prev state) {
	defer s.readyToRender()
	switch prev := prev.(type) {
	case *replaceState:
		s.content = prev.content
	case *updateState:
		s.content = prev.content
		s.dynamic = true
		s.head = prev.head
		if s.head == nil {
			s.head = &doorHead{}
		}
	case *proxyState:
		s.proxyElement = &prev.element
		s.content = prev.content
		s.dynamic = true
	case *jobState:
		s.content = prev.content
		s.proxyElement = prev.proxyElement
		s.head = prev.head
		s.dynamic = prev.dynamic
	}
	if !s.dynamic {
		s.readyForTakeover()
		s.initFinished()
	}
}

func (s *jobState) render(parent parent, p *pipe, th *sh.Thread) {
	newPipe := newPipe()
	p.put(newPipe)
	s.whenReadyToRender(func(bool) {
		if !s.dynamic {
			defer newPipe.close()
			if s.content != nil {
				cur := gox.NewCursor(context.Background(), newPipe)
				cur.Any(s.content)
			}
			return
		}
		s.tracker = newTracker(parent)
		if s.proxyElement == nil {
			s.initFinished()
		}
		sh.Run(func(t *sh.Thread) {
			newPipe.thread = t
			defer newPipe.close()
			if s.proxyElement != nil {
				s.renderProxy(newPipe)
			} else {
				s.renderContent(newPipe)
			}
		}, th.R(), s.tracker.th.Wi())
	})
}

func (s *jobState) renderContent(pipe *pipe) {
	cur := gox.NewCursor(s.tracker.ctx, pipe)
	open, close := s.head.frame(s.tracker.ctx, s.tracker.id, cur.NewId())
	cur.Job(open)
	cur.Any(s.content)
	cur.Job(close)
	cur.Func(func(io.Writer) error {
		s.readyForTakeover()
		return nil
	})
}

func (s *jobState) renderProxy(pipe *pipe) {
	pipeCur := gox.NewCursor(s.tracker.ctx, pipe)
	renderer := newProxyRender(s.tracker, pipeCur, s.content, func(head *doorHead, contentUsed bool) {
		s.head = head
		if !contentUsed {
			s.content = nil
		}
		s.initFinished()
	})
	(*s.proxyElement)(gox.NewCursor(s.tracker.ctx, renderer))
	pipeCur.Func(func(io.Writer) error {
		s.readyForTakeover()
		return nil
	})
}

func (s *jobState) Context() context.Context {
	return s.ctx
}

func (s *jobState) Output(io.Writer) error {
	return errors.New("door: used outside render pipeline")
}
