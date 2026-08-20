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

// Edit renders the Door directly in GoX:
//
//	~(&doors.Door{})
func (d *Door) Edit(cur gox.Cursor) error {
	return cur.Printer().Send(renderJob{door: d, fakeJob: fakeJob{cur.Context()}})
}

// Proxy renders the Door in GoX with the following element as its container:
//
//	~>(&doors.Door{}) <div>content</div>
func (d *Door) Proxy(cur gox.Cursor, el gox.Elem) error {
	return cur.Printer().Send(proxyJob{door: d, el: el, fakeJob: fakeJob{cur.Context()}})
}

// Inner replaces the Door's children while keeping the same container mounted.
// A nil content empties it.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Inner(ctx context.Context, content any) <-chan error {
	return d.inner(ctx, content)
}

// Outer replaces the Door container and its children with outer. A nil outer
// leaves an empty container. Unlike [Door.Static], the result remains a live
// Door that can be updated further.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Outer(ctx context.Context, outer any) <-chan error {
	return d.outer(ctx, outer)
}

// Static removes the Door container and renders content in its place. A nil
// content leaves nothing in place. Unlike [Door.Outer], the result is no longer
// a live Door; later operations change the stored state without putting the
// Door back on the page.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Static(ctx context.Context, content any) <-chan error {
	return d.static(ctx, content)
}

// Reload rerenders the Door with its current content.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Reload(ctx context.Context) <-chan error {
	return d.reload(ctx)
}

// Unmount removes the Door from the page and keeps its current content for a
// future mount. Unlike [Door.Static], the Door stays live and can be mounted
// again.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Unmount(ctx context.Context) <-chan error {
	return d.unmount(ctx)
}

// Freeze keeps the Door's current markup on the page but releases everything
// behind it: hooks, subscriptions, and nested Doors. Interactive elements
// inside stay visible but silently stop working. Unlike [Door.Static], no
// content is sent. The Door keeps its stored state and can be mounted again.
//
// Made for content that is dynamic only for a while and then final: blocks in
// a growing feed, log or chat entries, streamed output. Freeze the finished
// block, drop the reference, and server memory stays flat as the page grows.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Freeze(ctx context.Context) <-chan error {
	return d.freeze(ctx)
}
