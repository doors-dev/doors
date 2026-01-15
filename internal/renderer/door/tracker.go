package door

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
)

type TrackerKey struct{}

type parent interface {
	root() *Root
	context() context.Context
	thread() *sh.Thread
}

func newTrackerFrom(old *tracker) *tracker {
	ctx, cancel := context.WithCancel(context.Background())
	t := tracker{
		id:     old.id,
		rt:     old.rt,
		th:     old.rt.newThread(),
		parent: old.parent,
		cancel: cancel,
	}
	t.ctx = context.WithValue(ctx, TrackerKey{}, t)
	return &t
}

func newTracker(parent parent) *tracker {
	ctx, cancel := context.WithCancel(context.Background())
	root := parent.root()
	t := tracker{
		rt:     root,
		id:     root.newId(),
		th:     root.newThread(),
		parent: parent,
		cancel: cancel,
	}
	t.ctx = context.WithValue(ctx, TrackerKey{}, t)
	return &t
}

type tracker struct {
	id      uint64
	rt      *Root
	parent  parent
	th      *sh.Thread
	ctx     context.Context
	cancel  context.CancelFunc
	children common.Set[*tracker]
}

func (x *tracker) addChildren(children []*tracker) {
	sh.Run(func(t *sh.Thread) {
		for _, child := range children {
			x.children.Add(child)
		}
	}, x.th.W())
}

func (x *tracker) root() *Root {
	return x.rt
}

func (x *tracker) context() context.Context {
	return x.ctx
}

func (x *tracker) thread() *sh.Thread {
	return x.th
}
