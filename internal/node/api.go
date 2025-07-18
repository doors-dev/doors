package node

import (
	"context"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
)

func (n *Node) Update(ctx context.Context, content templ.Component) {
	n.update(ctx, content)
}

func (n *Node) Replace(ctx context.Context, content templ.Component) {
	ctx = common.ClearBlockingCtx(ctx)
	n.replace(ctx, content)
}

func (n *Node) Remove(ctx context.Context) {
	ctx = common.ClearBlockingCtx(ctx)
	n.remove(ctx)
}

func (n *Node) Clear(ctx context.Context) {
	n.Update(ctx, nil)
}

func (n *Node) XUpdate(ctx context.Context, content templ.Component) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XUpdate")
	ctx = common.ClearBlockingCtx(ctx)
	return n.update(ctx, content)
}

func (n *Node) XReplace(ctx context.Context, content templ.Component) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XReplace")
	ctx = common.ClearBlockingCtx(ctx)
	return n.replace(ctx, content)
}

func (n *Node) XRemove(ctx context.Context) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XRemove")
	ctx = common.ClearBlockingCtx(ctx)
	return n.remove(ctx)
}

func (n *Node) XClear(ctx context.Context) <-chan error {
	return n.XUpdate(ctx, nil)
}
