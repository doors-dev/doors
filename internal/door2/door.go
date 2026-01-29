package door2

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

type Door struct {
	node atomic.Pointer[node]
}

var _ gox.Proxy = &Door{}
var _ gox.Editor = &Door{}

func (d *Door) remove(ctx context.Context) <-chan error {
	node := &node{
		ctx:    ctx,
		reportCh: make(chan error, 1),
		done:   ctex.WgAdd(ctx),
		door:   d,
		kind:   unmountedNode,
		view:   &view{},
	}
	d.takeover(node)
	return node.reportCh
}

func (d *Door) update(ctx context.Context, content any) <-chan error {
	node := &node{
		ctx:    ctx,
		reportCh: make(chan error, 1),
		done:   ctex.WgAdd(ctx),
		door:   d,
		kind:   updatedNode,
		view: &view{
			content: content,
		},
	}
	d.takeover(node)
	return node.reportCh
}

func (d *Door) reload(ctx context.Context) <-chan error {
	node := &node{
		ctx:    ctx,
		reportCh: make(chan error, 1),
		done:   ctex.WgAdd(ctx),
		door:   d,
		kind:   updatedNode,
		view:   nil,
	}
	d.takeover(node)
	return node.reportCh
}

func (d *Door) replace(ctx context.Context, content any) <-chan error {
	node := &node{
		ctx:    ctx,
		reportCh: make(chan error, 1),
		done:   ctex.WgAdd(ctx),
		door:   d,
		kind:   replacedNode,
		view: &view{
			content: content,
		},
	}
	d.takeover(node)
	return node.reportCh
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
	return cur.Send(node)
}

func (d *Door) Use(cur gox.Cursor) error {
	node := &node{
		ctx:  cur.Context(),
		door: d,
		kind: editorNode,
	}
	d.takeover(node)
	return cur.Send(node)
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
