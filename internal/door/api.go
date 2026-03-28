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
	"github.com/doors-dev/gox"
)

// Reload re-renders the door with its current content.
// If the door is not currently mounted, the operation completes immediately
// without a visual effect.
func (d *Door) Reload(ctx context.Context) {
	d.reload(ctx)
}

// Update replaces the door's current children while keeping the same door
// container mounted. If the door is not currently mounted, the content change
// is stored and will be applied when the door is rendered.
func (d *Door) Update(ctx context.Context, content any) {
	d.update(ctx, content)
}

// Rebase replaces the rendered door with el while keeping the same Go [Door]
// handle alive for later updates. Unlike [Door.Replace], the result remains a
// live door that can be updated further. If the door is not currently mounted,
// the change is stored and will be applied when the door is rendered.
func (d *Door) Rebase(ctx context.Context, el gox.Elem) {
	d.rebase(ctx, el)
}

// Replace removes the current door container and replaces it with content.
// Unlike [Door.Update], this removes the door's DOM element entirely. If the
// door is not currently mounted, the change is stored and will be applied when
// the door is rendered.
func (d *Door) Replace(ctx context.Context, content any) {
	ctx = ctex.ClearBlockingCtx(ctx)
	d.replace(ctx, content)
}

// Delete removes the door and forgets its current content. If the door is not
// currently mounted, it is marked as removed and will not render later.
func (d *Door) Delete(ctx context.Context) {
	ctx = ctex.ClearBlockingCtx(ctx)
	d.replace(ctx, nil)
}

// Unmount removes the door from the page but keeps its current content for a
// future mount.
func (d *Door) Unmount(ctx context.Context) {
	ctx = ctex.ClearBlockingCtx(ctx)
	d.unmount(ctx)
}

// Clear removes all door content while keeping the container mounted.
// It is equivalent to Update(ctx, nil).
func (d *Door) Clear(ctx context.Context) {
	d.update(ctx, nil)
}

// XReload tracks completion of [Door.Reload].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XReload(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XUpdate")
	return d.reload(ctx)
}

// XUpdate tracks completion of [Door.Update].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XUpdate(ctx context.Context, content any) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XUpdate")
	return d.update(ctx, content)
}

// XRebase tracks completion of [Door.Rebase].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XRebase(ctx context.Context, el gox.Elem) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XRebase")
	return d.rebase(ctx, el)
}

// XReplace tracks completion of [Door.Replace].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XReplace(ctx context.Context, content any) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XReplace")
	return d.replace(ctx, content)
}

// XDelete tracks completion of [Door.Delete].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XDelete(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XDelete")
	return d.replace(ctx, nil)
}

// XUnmount tracks completion of [Door.Unmount].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XUnmount(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XUnmount")
	return d.unmount(ctx)
}

// XClear tracks completion of [Door.Clear].
// The channel receives nil on success or an error on failure, then closes.
// If the door is not mounted, it closes immediately without sending a value.
// Wait on it only in contexts where blocking is allowed.
func (d *Door) XClear(ctx context.Context) <-chan error {
	ctex.LogBlockingWarning(ctx, "Door", "XClear")
	return d.update(ctx, nil)
}
