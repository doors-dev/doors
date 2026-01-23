package door2

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/gox"
)

type Door struct {
	node atomic.Pointer[node]
}

var _ gox.Proxy = &Door{}
var _ gox.Editor = &Door{}

func (d *Door) Remove(ctx context.Context) {
	node := &node{
		ctx:  ctx,
		door: d,
		kind: unmountedNode,
		view: &view{},
	}
	d.takeover(node)
}

func (d *Door) Clear(ctx context.Context) {
	node := &node{
		ctx:  ctx,
		door: d,
		kind: updatedNode,
		view: &view{},
	}
	d.takeover(node)
}

func (d *Door) Update(ctx context.Context, content any) {
	node := &node{
		ctx:  ctx,
		door: d,
		kind: updatedNode,
		view: &view{
			content: content,
		},
	}
	d.takeover(node)
}

func (d *Door) Replace(ctx context.Context, content any) {
	node := &node{
		ctx:  ctx,
		door: d,
		kind: replacedNode,
		view: &view{
			content: content,
		},
	}
	d.takeover(node)
}

func (d *Door) Proxy(cur gox.Cursor, elem gox.Elem) error {
	node := &node{
		ctx:  cur.Context(),
		door: d,
		kind: proxyNode,
		view: &view{
			elem: elem,
		},
	}
	d.takeover(node)
	return cur.Job(node)
}

func (d *Door) Use(cur gox.Cursor) error {
	node := &node{
		ctx:  cur.Context(),
		door: d,
		kind: editorNode,
	}
	d.takeover(node)
	return cur.Job(node)
}

func (d *Door) takeover(next *node) {
	prev := d.node.Swap(next)
	if prev == nil {
		prev = &node{
			door: d,
			kind: unmountedNode,
		}
		prev.takoverFrame.Activate()
	}
	next.takover(prev)
}
