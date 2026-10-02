package n

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/shredder"
)

type Door struct {
	node atomic.Pointer[node]
}

func (d *Door) render(ctx context.Context, p *pipe) {
	d.schedule(ctx, renderNode{
		pipe:   p,
		buffer: p.branch(),
		ctx:    ctx,
	}, p.renderFrame)
}

func (d *Door) schedule(ctx context.Context, task nodeTask, externalFrame shredder.Frame) {
	next := &node{
		door: d,
	}
	prev := d.node.Swap(next)
	if prev == nil {
		prev = &node{
			door:       d,
			mode:       modeOuter,
			placeGuard: shredder.FreeFrame{},
		}
		prev.initGuard.Activate()
	}
	initFrame := shredder.Join(ctx, true, &prev.initGuard, externalFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.initGuard.Activate()
		task.apply(initFrame, next, prev)
	})
}
