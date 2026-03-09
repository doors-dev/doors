package pipe

import (
	"context"
	"fmt"

	"github.com/doors-dev/gox"
)

type ProxyContainer struct {
	Tag   string
	Attrs gox.Attrs
}

func (p ProxyContainer) Apply(pipe Pipe, containerCtx context.Context, doorID uint64) {
	headID := pipe.cursor.NewID()
	var openJob *gox.JobHeadOpen
	var closeJob *gox.JobHeadClose
	if p.Tag == "" {
		attrs := gox.NewAttrs()
		attrs.Get("id").Set(fmt.Sprintf("d0r%d", doorID))
		openJob = gox.NewJobHeadOpen(headID, gox.KindRegular, "d0-r", containerCtx, attrs)
		closeJob = gox.NewJobHeadClose(headID, gox.KindRegular, "d0-r", containerCtx)
	} else {
		attrs := p.Attrs.Clone()
		attrs.Get("data-d0r").Set(fmt.Sprintf("%d", doorID))
		openJob = gox.NewJobHeadOpen(headID, gox.KindRegular, p.Tag, containerCtx, attrs)
		closeJob = gox.NewJobHeadClose(headID, gox.KindRegular, p.Tag, containerCtx)
	}
	pipe.unshift(openJob)
	pipe.push(closeJob)
}
