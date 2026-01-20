package door

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/gox"
)

type Door struct {
	node atomic.Pointer[node]
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

func (d *Door) Proxy(ctx context.Context, cur gox.Cursor, elem gox.Elem) error {
	node := &node{
		ctx:  ctx,
		door: d,
		kind: proxyNode,
		view: &view{
			elem: elem,
		},
	}
	d.takeover(node)
	return cur.Job(node)
}

func (d *Door) Job(ctx context.Context) gox.Job {
	node := &node{
		ctx:  ctx,
		door: d,
		kind: jobNode,
	}
	d.takeover(node)
	return node
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

