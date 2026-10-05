package n

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/ctex"
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

func (d *Door) schedule(ctx context.Context, task nodeTask, externalFrame shredder.ReleaseFrame) {
	next := &node{
		door: d,
	}
	prev := d.node.Swap(next)
	if prev == nil {
		prev = &node{
			door: d,
			mode: modeOuter,
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

func (d *Door) reloadSelf(ctx context.Context, prev *node) <-chan error {
	ctex.LogCanceled(ctx, "Door reload")
	userTask, ch := newUserTask(ctx)
	frame := userTask.InitFrame()
	if !d.atomicSchedule(ctx, prev, nodeReload{userTask: userTask}, frame) {
		frame.Release()
		userTask.Cancel()
	}
	return ch
}

func (d *Door) atomicSchedule(ctx context.Context, prev *node, task nodeTask, externalFrame shredder.ReleaseFrame) bool {
	next := &node{
		door: d,
	}
	if !d.node.CompareAndSwap(prev, next) {
		return false
	}
	initFrame := shredder.Join(ctx, true, &prev.initGuard, externalFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.initGuard.Activate()
		task.apply(initFrame, next, prev)
	})
	return true
}
