// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package door

import (
	"context"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

// Edit renders door through the gox editor pipeline.
//
// It is a system method used by gox to support direct Door rendering, for
// example:
//
//	~(&doors.Door{})
func (d *Door) Edit(cur gox.Cursor) error {
	return cur.Printer().Send(renderJob{door: d})
}

// Proxy renders door through the gox proxy pipeline.
//
// It is a system method used by gox to support Door proxy syntax, for example:
//
//	~>(&doors.Door{}) <div>content</div>
func (d *Door) Proxy(cur gox.Cursor, el gox.Elem) error {
	return cur.Printer().Send(proxyJob{door: d, el: el})
}

// Inner replaces the door's current children while keeping the same door
// container mounted. If the door is not currently mounted, the content change
// is stored and will be applied when the door is rendered.
func (d *Door) Inner(ctx context.Context, content any) {
	d.inner(ctx, content)
}

// XInner tracks completion of [Door.Inner].
// The channel receives nil on success or an error on failure, then closes.
// It receives context.Canceled if the operation is overwritten by a newer
// update, unmount, or other door operation.
// If the door is not mounted, it closes immediately without sending a value.
//
// Do not wait on it during rendering. If you need to wait, use doors.Go(...),
// or your own goroutine with doors.Free(ctx).
func (d *Door) XInner(ctx context.Context, content any) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XInner")
	return d.inner(ctx, content)
}

// Outer replaces the rendered door with outer while keeping the same Go [Door]
// handle alive for later updates. Unlike [Door.Static], the result remains a
// live door that can be updated further. If the door is not currently mounted,
// the change is stored and will be applied when the door is rendered.
func (d *Door) Outer(ctx context.Context, outer gox.Elem) {
	d.outer(ctx, outer)
}

// XOuter tracks completion of [Door.Outer].
// The channel receives nil on success or an error on failure, then closes.
// It receives context.Canceled if the operation is overwritten by a newer
// update, unmount, or other door operation.
// If the door is not mounted, it closes immediately without sending a value.
//
// Do not wait on it during rendering. If you need to wait, use doors.Go(...),
// or your own goroutine with doors.Free(ctx).
func (d *Door) XOuter(ctx context.Context, outer gox.Elem) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XOuter")
	return d.outer(ctx, outer)
}

// Static removes the current door container and replaces it with static content.
// Unlike [Door.Outer], this removes the door's DOM element entirely. If the
// door is not currently mounted, the change is stored and will be applied when
// the door is rendered.
func (d *Door) Static(ctx context.Context, content any) {
	d.static(ctx, content)
}

// XStatic tracks completion of [Door.Static].
// The channel receives nil on success or an error on failure, then closes.
// It receives context.Canceled if the operation is overwritten by a newer
// update, unmount, or other door operation.
// If the door is not mounted, it closes immediately without sending a value.
//
// Do not wait on it during rendering. If you need to wait, use doors.Go(...),
// or your own goroutine with doors.Free(ctx).
func (d *Door) XStatic(ctx context.Context, content any) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XStatic")
	return d.static(ctx, content)
}

// Reload re-renders the door with its current content.
// If the door is not currently mounted, the operation completes immediately
// without a visual effect.
func (d *Door) Reload(ctx context.Context) {
	d.reload(ctx)
}

// XReload tracks completion of [Door.Reload].
// The channel receives nil on success or an error on failure, then closes.
// It receives context.Canceled if the operation is overwritten by a newer
// update, unmount, or other door operation.
// If the door is not mounted, it closes immediately without sending a value.
//
// Do not wait on it during rendering. If you need to wait, use doors.Go(...),
// or your own goroutine with doors.Free(ctx).
func (d *Door) XReload(ctx context.Context) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XReload")
	return d.reload(ctx)
}

// Unmount removes the door from the page but keeps its current content for a
// future mount.
func (d *Door) Unmount(ctx context.Context) {
	d.unmount(ctx)
}

// XUnmount tracks completion of [Door.Unmount].
// The channel receives nil on success or an error on failure, then closes.
// It receives context.Canceled if the operation is overwritten by a newer
// update, unmount, or other door operation.
// If the door is not mounted, it closes immediately without sending a value.
//
// Do not wait on it during rendering. If you need to wait, use doors.Go(...),
// or your own goroutine with doors.Free(ctx).
func (d *Door) XUnmount(ctx context.Context) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XUnmount")
	return d.unmount(ctx)
}
