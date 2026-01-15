package door

import (
	"context"
	"fmt"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type state interface {
	takeover(prev state)
	whenTakoverReady(f func(bool))
	afterInit(f func(bool))
	initFinished()
}

type stateTakover sh.Valv

func (s *stateTakover) readyForTakeover() {
	(*sh.Valv)(s).Open()
}

func (s *stateTakover) whenTakoverReady(f func(bool)) {
	(*sh.Valv)(s).Put(f)
}

type stateInit sh.Valv

func (s *stateInit) initFinished() {
	(*sh.Valv)(s).Open()
}

func (s *stateInit) afterInit(f func(bool)) {
	(*sh.Valv)(s).Put(f)
}

type stateRender sh.Valv

func (s *stateInit) readyToRender() {
	(*sh.Valv)(s).Open()
}

func (s *stateInit) whenReadyToRender(f func(bool)) {
	(*sh.Valv)(s).Put(f)
}

type doorHead struct {
	attrs gox.Attrs
	tag   string
}

func (h *doorHead) frame(ctx context.Context, doorId uint64, headId uint64) (*gox.JobHeadOpen, *gox.JobHeadClose) {
	if h.tag == "" {
		attrs := gox.NewAttrs(ctx)
		attrs.Get("id").Set(fmt.Sprintf("d00r/%d", doorId))
		openJob := gox.NewJobHeadOpen(
			headId,
			gox.KindRegular,
			"d0-0r",
			ctx,
			attrs,
		)
		closeJob := gox.NewJobHeadClose(headId, gox.KindRegular, "d0-0r", ctx)
		return openJob, closeJob
	}
	attrs := h.attrs.Clone()
	attrs.Get("data-d00r").Set(fmt.Sprintf("%d", doorId))
	openJob := gox.NewJobHeadOpen(
		headId,
		gox.KindRegular,
		h.tag,
		ctx,
		attrs,
	)
	closeJob := gox.NewJobHeadClose(headId, gox.KindRegular, h.tag, ctx)
	return openJob, closeJob
}
