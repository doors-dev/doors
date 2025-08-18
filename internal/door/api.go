package door

import (
	"context"
	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
)

// Reload re-renders the door with its current content.
// This is useful when you want to refresh a door without changing its content,
// for example to reflect external state changes. If the door is not currently
// active, the operation completes immediately without visual effect.
func (n *Door) Reload(ctx context.Context) {
	n.reload(ctx)
}

// Update changes the content of the door and re-renders it in place.
// The door's children are replaced with the new content while preserving
// the door's DOM element. If the door is not currently active, the content
// change is stored and will be applied when the door is rendered.
func (n *Door) Update(ctx context.Context, content templ.Component) {
	n.update(ctx, content)
}

// Replace replaces the entire door element with new content.
// Unlike Update, this removes the door's DOM element entirely and replaces
// it with the rendered content. If the door is not currently active, the
// content change is stored and will be applied when the door is rendered.
func (n *Door) Replace(ctx context.Context, content templ.Component) {
	ctx = common.ClearBlockingCtx(ctx)
	n.replace(ctx, content)
}

// Remove removes the door and its DOM element from the page.
// If the door is not currently active, it is marked as removed and will
// not render if attempted to be rendered.
func (n *Door) Remove(ctx context.Context) {
	ctx = common.ClearBlockingCtx(ctx)
	n.remove(ctx)
}

// Clear removes all content from the door, equivalent to Update(ctx, nil).
// The door's DOM element remains but its children are removed.
// This is useful for emptying a container while keeping it available for future content.
func (n *Door) Clear(ctx context.Context) {
	n.clear(ctx)
}

// XReload returns a channel that can be used to track when the reload operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Door) XReload(ctx context.Context) <-chan error {
	common.LogBlockingWarning(ctx, "Door", "XUpdate")
	return n.reload(ctx)
}

// XUpdate returns a channel that can be used to track when the update operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Door) XUpdate(ctx context.Context, content templ.Component) <-chan error {
	common.LogBlockingWarning(ctx, "Door", "XUpdate")
	return n.update(ctx, content)
}

// XReplace returns a channel that can be used to track when the replace operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Door) XReplace(ctx context.Context, content templ.Component) <-chan error {
	common.LogBlockingWarning(ctx, "Door", "XReplace")
	return n.replace(ctx, content)
}

// XRemove returns a channel that can be used to track when the remove operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not active,
// the channel is closed immediately without sending any value.
// A blocking context warning is logged if called from a blocking context.
func (n *Door) XRemove(ctx context.Context) <-chan error {
	common.LogBlockingWarning(ctx, "Door", "XRemove")
	return n.remove(ctx)
}

// XClear returns a channel that can be used to track when the clear operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not active,
// the channel is closed immediately without sending any value.
// This is equivalent to XUpdate(ctx, nil) and empties the door's content.
func (n *Door) XClear(ctx context.Context) <-chan error {
	return n.clear(ctx)
}
