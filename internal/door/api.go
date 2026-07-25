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

	"github.com/doors-dev/gox"
)

// Edit renders door through the gox editor pipeline.
//
// It is a system method used by gox to support direct Door rendering, for
// example:
//
//	~(&doors.Door{})
func (d *Door) Edit(cur gox.Cursor) error {
	return cur.Printer().Send(renderJob{door: d, fakeJob: fakeJob{cur.Context()}})
}

// Proxy renders door through the gox proxy pipeline.
//
// It is a system method used by gox to support Door proxy syntax, for example:
//
//	~>(&doors.Door{}) <div>content</div>
func (d *Door) Proxy(cur gox.Cursor, el gox.Elem) error {
	return cur.Printer().Send(proxyJob{door: d, el: el, fakeJob: fakeJob{cur.Context()}})
}

// Inner replaces the door's current children while keeping the same door
// container mounted. If the door is not currently mounted, the content change
// is stored and will be applied when the door is rendered.
//
// The returned channel is optional to use and tracks completion.
// On success it sends two nil values then closes: the first means the call was
// scheduled (render was completed), the second means it was applied to the
// page.
// On failure it sends an error then closes.
// It receives context.Canceled if the operation is overwritten by a newer
// update, unmount, or other door operation.
// If the door is not mounted, it closes immediately without sending a value.
//
// Do not wait on it during rendering. If you need to wait, use doors.Go(...),
// or your own goroutine with doors.DetachedContext(ctx).
func (d *Door) Inner(ctx context.Context, content any) <-chan error {
	return d.inner(ctx, content)
}

// Outer replaces the rendered door with outer while keeping the same Go [Door]
// handle alive for later updates. Unlike [Door.Static], the result remains a
// live door that can be updated further. If the door is not currently mounted,
// the change is stored and will be applied when the door is rendered.
//
// The returned channel is optional to use; see [Door.Inner] for the contract.
func (d *Door) Outer(ctx context.Context, outer gox.Elem) <-chan error {
	return d.outer(ctx, outer)
}

// Static removes the current door container and replaces it with static content.
// Unlike [Door.Outer], this removes the door's DOM element entirely. If the
// door is not currently mounted, the change is stored and will be applied when
// the door is rendered.
//
// The returned channel is optional to use; see [Door.Inner] for the contract.
func (d *Door) Static(ctx context.Context, content any) <-chan error {
	return d.static(ctx, content)
}

// Reload re-renders the door with its current content.
// If the door is not currently mounted, the operation completes immediately
// without a visual effect.
//
// The returned channel is optional to use; see [Door.Inner] for the contract.
func (d *Door) Reload(ctx context.Context) <-chan error {
	return d.reload(ctx)
}

// Unmount removes the door from the page but keeps its current content for a
// future mount.
//
// The returned channel is optional to use; see [Door.Inner] for the contract.
func (d *Door) Unmount(ctx context.Context) <-chan error {
	return d.unmount(ctx)
}
