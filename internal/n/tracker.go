package n

import (
	"context"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type tracker struct {
	ctx       context.Context
	cancel    context.CancelFunc
	outer     *outerTracker
	rw        shredder.ReadWriteThread
	renderCtx context.Context
}

func (t *tracker) clean() {
	t.cancel()
}

func newOuterTracker1(prev *outerTracker) *outerTracker {
	ctx, cancel := context.WithCancel(prev.parent.ctx)
	tracker := &outerTracker{
		id:        prev.id,
		root:      prev.root,
		parent:    prev.parent,
		ctx:       ctx,
		cancel:    cancel,
		callGuard: prev.callGuard,
	}
	tracker.renderCtx = context.WithValue(tracker.ctx, common.KeyCore, tracker)
	return tracker
}

func newOuterTracker2(parent *tracker, callGuard *shredder.ValveFrame) *outerTracker {
	ctx, cancel := context.WithCancel(parent.ctx)
	tracker := &outerTracker{
		id:        parent.outer.root.ID(),
		root:      parent.outer.root,
		parent:    parent,
		ctx:       ctx,
		cancel:    cancel,
		callGuard: callGuard,
	}
	tracker.renderCtx = context.WithValue(tracker.ctx, common.KeyCore, tracker)
	return tracker
}

type outerTracker struct {
	id             uint64
	root           *root
	parent         *tracker
	ctx            context.Context
	cancel         context.CancelFunc
	callGuard      *shredder.ValveFrame
	outerGuard     shredder.ValveFrame
	renderCtx      context.Context
	placementError error
}

func (t *outerTracker) clean() {
	t.cancel()
}

func (t *outerTracker) newTracker() *tracker {
	ctx, cancel := context.WithCancel(t.ctx)
	tracker := &tracker{
		ctx:    ctx,
		cancel: cancel,
		outer:  t,
	}
	tracker.renderCtx = context.WithValue(tracker.ctx, common.KeyCore, tracker)
	return tracker
}
