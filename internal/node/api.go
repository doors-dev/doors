package node

import (
	"context"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
)

// Update sets the Node's content to the provided templ.Component.
//
// The method returns true if the Node was successfully updated. This occurs only if the Node
// has not been previously removed via Remove or replaced via Replace.
//
// You can safely call Update even if the Node has not yet been rendered or has been
// removed due to its parent being removed. In such cases, the Node will render
// with the latest content when it is next used.
//
// Parameters:
//   - ctx: the latest context in scope, 
//   - content: the new templ.Component content to assign to the Node.
// Returns:
//   - A boolean indicating whether the clear operation was accepted.
func (n *Node) Update(ctx context.Context, content templ.Component) bool {
	_, ok := n.update(ctx, content)
	return ok
}

// Replace swaps the Node with the provided templ.Component, replacing its entire outer HTML.
//
// This makes the Node behave like a static component, as its identity is lost in the process.
// Once replaced, the Node can no longer be updated or removed.
//
// The method returns true if the Node was successfully replaced. If the Node was already removed
// or previously replaced, it returns false.
//
// Replace is safe to call even if the Node hasn't yet been rendered or was removed due to its
// parent's removal. 
//
// Parameters:
//   - ctx: the latest context in scope, 
//   - content: the new templ.Component that replaces the Node entirely.
// Returns:
//   - A boolean indicating whether the clear operation was accepted.
func (n *Node) Replace(ctx context.Context, content templ.Component) bool {
	ctx = common.ClearBlockingCtx(ctx)
	_, ok := n.replace(ctx, content)
	return ok
}

// Remove removes the Node from the page if it has already been rendered.
// If the Node is rendered in the future, it will produce no output.
//
// Once removed, the Node cannot be updated or replaced. The removal is final and
// effectively makes the Node inert in the rendering tree.
//
// The method returns true if the Node was successfully removed. If the Node had already
// been removed or was previously replaced, it returns false.
//
// Remove is safe to call even before the Node is rendered or if it was implicitly removed
// due to its parent's removal.
//
// Parameters:
//   - ctx: the latest context in scope, 
// Returns:
//   - A boolean indicating whether the clear operation was accepted.
func (n *Node) Remove(ctx context.Context) bool {
	ctx = common.ClearBlockingCtx(ctx)
	_, ok := n.remove(ctx)
	return ok
}

// Clear removes any content from the Node by updating it with nil.
//
// This effectively makes the Node render nothing, but unlike Remove, the Node remains active
// and can still be updated or replaced in the future.
//
// The method returns true if the Node was successfully cleared. If the Node was already
// removed or replaced, it returns false.
//
// Parameters:
//   - ctx: the latest context in scope
// Returns:
//   - A boolean indicating whether the clear operation was accepted.
func (n *Node) Clear(ctx context.Context) bool {
	return n.Update(ctx, nil)
}

// XUpdate updates of the Node's content with the provided templ.Component.
//
// This method returns a channel and a boolean:
//   - The returned channel will receive a single `error` when the frontend confirms the update,
//     or will be closed silently if the Node is not currently rendered (e.g., due to parent removal).
//     In that case, the update is still retained and will be applied when the Node is next rendered.
//     Possible error values:
//       - `nil` if the update was applied successfully,
//       - non-nil if the update was overridden by a newer operation before it was applied,
//         or if a rare frontend error occurred during processing.
//   - The boolean return value indicates whether the update was accepted for processing.
//     It returns true if the Node was not removed or replaced; false otherwise.
//
// **Important:** Blocking on the returned channel can cause a deadlock under extreme conditions if `XUpdate` is used
// outside a blocking-safe context, such as a hook handler or inside a `doors.Go` component.
//
// Parameters:
//   - ctx: the latest context in scope, 
//   - content: the new templ.Component to set as the Node's content.
//
// Returns:
//   - A receive-only channel that will emit a single error (or close silently if not applicable).
//   - A boolean indicating whether the update was accepted.
func (n *Node) XUpdate(ctx context.Context, content templ.Component) (<-chan error, bool) {
	common.LogBlockingWarning(ctx, "Node", "XUpdate")
	ctx = common.ClearBlockingCtx(ctx)
	return n.update(ctx, content)
}

// XReplace replaces the Node with the provided templ.Component,
// replacing its entire outer HTML and making it behave like a static component.
//
// This method returns a channel and a boolean:
//   - The returned channel will receive a single `error` when the frontend confirms the replacement,
//     or will be closed silently if the Node is not currently rendered (e.g., due to parent removal).
//     In that case, the replacement is still retained and will be applied when the Node is next rendered.
//     Possible error values:
//       - `nil` if the replacement was applied successfully ,
//       - non-nil if a rare frontend error occurred during processing
//   - The boolean return value indicates whether the replacement was accepted for processing.
//     It returns true if the Node was not previously removed or replaced; false otherwise.
//
// Once replaced, the Node becomes static: it cannot be updated or removed afterward.
//
// **Important:** Blocking on the returned channel can cause a deadlock under extreme conditions
// **Important:** Blocking on the returned channel can cause a deadlock under extreme conditions if `XUpdate` is used
// outside a blocking-safe context, such as a hook handler or inside a `doors.Go` component.
//
// Parameters:
//   - ctx: the latest context in scope, 
//   - content: the new templ.Component that will replace the entire Node.
//
// Returns:
//   - A receive-only channel that will emit a single error (or close silently if deferred).
//   - A boolean indicating whether the replacement was accepted.
func (n *Node) XReplace(ctx context.Context, content templ.Component) (<-chan error, bool) {
	common.LogBlockingWarning(ctx, "Node", "XReplace")
	ctx = common.ClearBlockingCtx(ctx)
	return n.replace(ctx, content)
}


// XRemove removes the Node from the page if it has already been rendered.
// If the Node is rendered again in the future, it will produce no output.
//
// This method returns a channel and a boolean:
//   - The returned channel will receive a single `error` when the frontend confirms the removal,
//     or will be closed silently if the Node is not currently rendered (e.g., due to parent removal).
//     In that case, the removal is still retained and will be applied when the Node is next rendered.
//     Possible error values:
//       - `nil` if the removal was applied successfully,
//       - non-nil if the operation was overridden by a newer one before it was applied,
//         or if a rare frontend error occurred during processing.
//   - The boolean return value indicates whether the removal was accepted for processing.
//     It returns true if the Node was not already removed or replaced; false otherwise.
//
// Once removed, the Node is considered inert — it will render nothing and can no longer be updated or replaced.
//
// **Important:** Blocking on the returned channel can cause a deadlock under extreme conditions if `XUpdate` is used
// outside a blocking-safe context, such as a hook handler or inside a `doors.Go` component.
//
// Parameters:
//   - ctx: the latest context in scope
//
// Returns:
//   - A receive-only channel that will emit a single error (or close silently if deferred).
//   - A boolean indicating whether the removal was accepted.

func (n *Node) XRemove(ctx context.Context) (<-chan error, bool) {
	common.LogBlockingWarning(ctx, "Node", "XRemove")
	ctx = common.ClearBlockingCtx(ctx)
	return n.remove(ctx)
}

// XClear clears the content of the Node by updating it with nil.
//
// If the Node has already been rendered, its content will be removed from the page.
// If the Node is rendered again in the future, it will render nothing until new content is set.
//
// This method returns a channel and a boolean:
//   - The returned channel will receive a single `error` when the frontend confirms the clear operation,
//     or will be closed silently if the Node is not currently rendered (e.g., due to parent removal).
//     In that case, the clear operation is still retained and will be applied when the Node is next rendered.
//     Possible error values:
//       - `nil` if the clear was applied successfully ,
//       - non-nil rare frontend error occurred during processing,
//   - The boolean return value indicates whether the clear operation was accepted for processing.
//     It returns true if the Node was not already removed or replaced; false otherwise.
//
// Unlike Remove, XClear does not make the Node inert — it can still be updated or replaced after being cleared.
//
// **Important:** Blocking on the returned channel can cause a deadlock under extreme conditions if `XUpdate` is used
// outside a blocking-safe context, such as a hook handler or inside a `doors.Go` component.
//
// Parameters:
//   - ctx: the latest context in scope
//
// Returns:
//   - A receive-only channel that will emit a single error (or close silently if deferred).
//   - A boolean indicating whether the clear operation was accepted.

func (n *Node) XClear(ctx context.Context) (<-chan error, bool) {
	return n.XUpdate(ctx, nil)
}
