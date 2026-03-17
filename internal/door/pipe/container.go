package pipe

import (
	"context"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type ProxyContainer struct {
	Tag   string
	Attrs gox.Attrs
}

func (p ProxyContainer) Apply(pipe Pipe, containerCtx context.Context, doorID uint64, parentID uint64) {
	headID := pipe.cursor.NewID()
	var openJob *gox.JobHeadOpen
	var closeJob *gox.JobHeadClose
	if p.Tag == "" {
		attrs := gox.NewAttrs()
		front.AttrsSetDoor(attrs, doorID, true)
		front.AttrsSetParent(attrs, parentID)
		openJob = gox.NewJobHeadOpen(containerCtx, headID, gox.KindRegular, "d0-r", attrs)
		closeJob = gox.NewJobHeadClose(containerCtx, headID, gox.KindRegular, "d0-r")
	} else {
		attrs := p.Attrs.Clone()
		front.AttrsSetDoor(attrs, doorID, false)
		front.AttrsSetParent(attrs, parentID)
		openJob = gox.NewJobHeadOpen(containerCtx, headID, gox.KindRegular, p.Tag, attrs)
		closeJob = gox.NewJobHeadClose(containerCtx, headID, gox.KindRegular, p.Tag)
	}
	pipe.unshift(openJob)
	pipe.push(closeJob)
}
