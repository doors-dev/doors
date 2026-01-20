package door

import (
	"sync"

	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)


func newPipe() *pipe {
	return &pipe{}
}

type pipe struct {
	sh.Queue
	parent parent
	frame  sh.Frame
}


func (p *pipe) Send(job gox.Job) error {
	switch job := job.(type) {
	case *node:
		p.parent.addChild(job)
		job.render(p.parent, p)
	case *gox.JobComp:
		newPipe := newPipe()
		newPipe.parent = p.parent
		newPipe.thread = p.thread
		p.put(newPipe)
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		sh.Run(nil, func(ok bool) {
			comp.Main().Print(ctx, newPipe)
		}, p.shread)
		/*
			sh.Run(func(t *sh.Thread) {
				comp.Main().Print(ctx, newPipe)
			}, p.thread.R()) */
	default:
		p.put(job)
	}
	return nil
}

