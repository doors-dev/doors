package door2

import (
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

func newPipe() *pipe {
	return &pipe{
		Queue: sh.NewQueue(),
	}
}

type pipe struct {
	sh.Queue
	tracker *tracker
	frame   sh.Frame
}

func (p *pipe) Send(job gox.Job) error {
	switch job := job.(type) {
	case *node:
		job.render(p.tracker, p)
	case core.JobCore:
		job.Apply(p.tracker.core)
	case *gox.JobComp:
		newPipe := newPipe()
		newPipe.tracker = p.tracker
		newPipe.frame = p.frame
		p.Put(newPipe)
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		p.frame.Run(p.tracker.root.spawner, func() {
			defer newPipe.Close()
			comp.Main().Print(ctx, newPipe)
		})
	default:
		p.Put(job)
	}
	return nil
}
