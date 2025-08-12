package node

import (
	"context"
	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
)

// Reload re-renders the node with its current content.
// This is useful when you want to refresh a node without changing its content,
// for example to reflect external state changes. If the node is not currently
// active, the operation completes immediately without visual effect.
func (n *Node) Reload(ctx context.Context) {
	n.reload(ctx)
}

// Update changes the content of the node and re-renders it in place.
// The node's children are replaced with the new content while preserving
// the node's DOM element. If the node is not currently active, the content
// change is stored and will be applied when the node is rendered.
func (n *Node) Update(ctx context.Context, content templ.Component) {
	n.update(ctx, content)
}

// Replace replaces the entire node element with new content.
// Unlike Update, this removes the node's DOM element entirely and replaces
// it with the rendered content. If the node is not currently active, the
// content change is stored and will be applied when the node is rendered.
func (n *Node) Replace(ctx context.Context, content templ.Component) {
	ctx = common.ClearBlockingCtx(ctx)
	n.replace(ctx, content)
}

// Remove removes the node and its DOM element from the page.
// If the node is not currently active, it is marked as removed and will
// not render if attempted to be rendered.
func (n *Node) Remove(ctx context.Context) {
	ctx = common.ClearBlockingCtx(ctx)
	n.remove(ctx)
}

// Clear removes all content from the node, equivalent to Update(ctx, nil).
// The node's DOM element remains but its children are removed.
// This is useful for emptying a container while keeping it available for future content.
func (n *Node) Clear(ctx context.Context) {
	n.clear(ctx)
}

// XReload returns a channel that can be used to track when the reload operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the node is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Node) XReload(ctx context.Context) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XUpdate")
	return n.reload(ctx)
}

// XUpdate returns a channel that can be used to track when the update operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the node is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Node) XUpdate(ctx context.Context, content templ.Component) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XUpdate")
	return n.update(ctx, content)
}

// XReplace returns a channel that can be used to track when the replace operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the node is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Node) XReplace(ctx context.Context, content templ.Component) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XReplace")
	return n.replace(ctx, content)
}

// XRemove returns a channel that can be used to track when the remove operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the node is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Node) XRemove(ctx context.Context) <-chan error {
	common.LogBlockingWarning(ctx, "Node", "XRemove")
	return n.remove(ctx)
}

// XClear returns a channel that can be used to track when the clear operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the node is not active,
// the channel is closed immediately without sending any value.
// This is equivalent to XUpdate(ctx, nil) and empties the node's content.
func (n *Node) XClear(ctx context.Context) <-chan error {
	return n.clear(ctx)
}
