package door

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Door struct {
	node atomic.Pointer[node]
}

func (d *Door) Proxy(cur gox.Cursor, elem gox.Elem) error {
	node := &node{
		ctx:  cur.Context(),
		door: d,
		entity: &proxyNode{
			elem: elem,
		},
	}
	d.takeover(node)
	return cur.Send(node)
}

func (d *Door) Edit(cur gox.Cursor) error {
	node := &node{
		ctx:    cur.Context(),
		door:   d,
		entity: &editorNode{},
	}
	d.takeover(node)
	return cur.Send(node)
}

func (d *Door) rebase(ctx context.Context, el gox.Elem) <-chan error {
	task, ch := newTaskNode(ctx)
	node := &node{
		ctx:  ctx,
		door: d,
		entity: &rebaseNode{
			taskNode: task,
			elem:     el,
		},
	}
	d.takeover(node)
	return ch
}

func (d *Door) unmount(ctx context.Context) <-chan error {
	task, ch := newTaskNode(ctx)
	node := &node{
		ctx:  ctx,
		door: d,
		entity: &unmountNode{
			taskNode: task,
			remove:   true,
		},
	}
	d.takeover(node)
	return ch
}

func (d *Door) update(ctx context.Context, content any) <-chan error {
	task, ch := newTaskNode(ctx)
	node := &node{
		ctx:  ctx,
		door: d,
		entity: &updateNode{
			taskNode: task,
			content:  content,
		},
	}
	d.takeover(node)
	return ch
}

func (d *Door) reload(ctx context.Context) <-chan error {
	task, ch := newTaskNode(ctx)
	node := &node{
		ctx:  ctx,
		door: d,
		entity: &redrawNode{
			taskNode: task,
		},
	}
	d.takeover(node)
	return ch
}

func (d *Door) replace(ctx context.Context, content any) <-chan error {
	task, ch := newTaskNode(ctx)
	node := &node{
		ctx:  ctx,
		door: d,
		entity: &replaceNode{
			taskNode: task,
			content:  content,
		},
	}
	d.takeover(node)
	return ch
}

func (d *Door) defaultNode() *node {
	contents := &contents{
		initializeFrame: &shredder.ValveFrame{},
		container:       &Container{},
	}
	contents.initializeFrame.Activate()
	n := &node{
		door: d,
		entity: &unmountNode{
			contents: contents,
		},
	}
	n.initFrame.Activate()
	return n
}

func (d *Door) takeoverSelf(prev *node, next *node) {
	swapped := d.node.CompareAndSwap(prev, next)
	if !swapped {
		return
	}
	prev.initFrame.Run(nil, nil, func(b bool) {
		defer next.initFrame.Activate()
		next.init(prev)
	})
}

func (d *Door) takeover(next *node) {
	prev := d.node.Swap(next)
	if prev == nil {
		prev = d.defaultNode()
	}
	prev.initFrame.Run(nil, nil, func(b bool) {
		defer next.initFrame.Activate()
		next.init(prev)
	})
}
