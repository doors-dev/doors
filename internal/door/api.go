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

	"github.com/doors-dev/doors/internal/ctex"
)

// Reload re-renders the door with its current content.
// This is useful when you want to refresh a door without changing its content,
// for example to reflect external state changes. If the door is not currently
// mounted, the operation completes immediately without visual effect.
func (d *Door) Reload(ctx context.Context) {
	d.reload(ctx)
}

// Update changes the content of the door and re-renders it in place.
// The door's children are replaced with the new content while preserving
// the door's DOM element. If the door is not currently mounted, the content
// change is stored and will be applied when the door is rendered.
func (d *Door) Update(ctx context.Context, content any) {
	d.update(ctx, content)
}

// Replace replaces the entire door element with new content.
// Unlike Update, this removes the door's DOM element entirely and replaces
// it with the rendered content. If the door is not currently mounted, the
// content change is stored and will be applied when the door is rendered.
func (d *Door) Replace(ctx context.Context, content any) {
	ctx = ctex.ClearBlockingCtx(ctx)
	d.replace(ctx, content)
}

// Remove removes the door and its DOM element from the page.
// If the door is not currently mounted, it is marked as removed and will
// not render if attempted to be rendered.
func (d *Door) Remove(ctx context.Context) {
	ctx = ctex.ClearBlockingCtx(ctx)
	d.replace(ctx, nil)
}

// Unmount removes the door and its DOM element from the page, while
// preserving the door's content. 
func (d *Door) Unmount(ctx context.Context) {
	ctx = ctex.ClearBlockingCtx(ctx)
	d.unmount(ctx)
}

// Clear removes all content from the door, equivalent to Update(ctx, nil).
// The door's DOM element remains but its children are removed.
// This is useful for emptying a container while keeping it available for future content.
func (d *Door) Clear(ctx context.Context) {
	d.update(ctx, nil)
}

// XReload returns a channel that can be used to track when the reload operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not mounted,
// the channel is closed immediately without sending any value.
// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
func (d *Door) XReload(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XUpdate")
	return d.reload(ctx)
}

// XUpdate returns a channel that can be used to track when the update operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not mounted,
// the channel is closed immediately without sending any value.
// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
func (d *Door) XUpdate(ctx context.Context, content any) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XUpdate")
	return d.update(ctx, content)
}

// XReplace returns a channel that can be used to track when the replace operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not mounted,
// the channel is closed immediately without sending any value.
// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
func (d *Door) XReplace(ctx context.Context, content any) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XReplace")
	return d.replace(ctx, content)
}

// XRemove returns a channel that can be used to track when the remove operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not mounted,
// the channel is closed immediately without sending any value.
// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
func (d *Door) XRemove(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XRemove")
	return d.replace(ctx, nil)
}

// XUnmount returns a channel that can be used to track when the unmount operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not mounted,
// the channel is closed immediately without sending any value.
// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
func (d *Door) XUnmount(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XUnmount")
	return d.unmount(ctx)
}

// XClear returns a channel that can be used to track when the clear operation completes.
// The channel will receive nil on success or an error if the operation fails.
// The channel is closed after sending the result. If the door is not mounted,
// the channel is closed immediately without sending any value.
// Wait on the channel only in contexts where blocking is allowed (hooks, goroutines).
func (d *Door) XClear(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XClear")
	return d.update(ctx, nil)
}
