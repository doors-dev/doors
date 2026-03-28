// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package door

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

// Door is a dynamic fragment placeholder that can be rendered and then updated
// over time.
type Door struct {
	node atomic.Pointer[node]
}

// Proxy renders d as a proxy around elem.
func (d *Door) Proxy(cur gox.Cursor, elem gox.Elem) error {
	node := &node{
		ctx:  cur.Context(),
		door: d,
		entity: &proxyNode{
			elem: elem,
		},
	}
	d.takeover(node, shredder.FreeFrame{})
	return cur.Send(node)
}

// Edit renders d as an editable dynamic container.
func (d *Door) Edit(cur gox.Cursor) error {
	node := &node{
		ctx:    cur.Context(),
		door:   d,
		entity: &editorNode{},
	}
	d.takeover(node, shredder.FreeFrame{})
	return cur.Send(node)
}

// Main returns d as a [gox.Elem].
func (d *Door) Main() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		return d.Edit(cur)
	})
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
	d.takeover(node, task.TaskFrame())
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
	d.takeover(node, task.TaskFrame())
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
	d.takeover(node, task.TaskFrame())
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
	d.takeover(node, task.TaskFrame())
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
	d.takeover(node, task.TaskFrame())
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

func (d *Door) takeover(next *node, taskFrame shredder.SimpleFrame) {
	prev := d.node.Swap(next)
	if prev == nil {
		prev = d.defaultNode()
	}
	initFrame := shredder.Join(true, &prev.initFrame, taskFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.initFrame.Activate()
		next.init(prev)
	})
}
