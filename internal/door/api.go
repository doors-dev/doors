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

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

// Main renders the Door directly in GoX:
//
//	~(&doors.Door{})
func (d *Door) Main() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		return cur.Printer().Send(renderJob{door: d, caller: common.CaptureCaller(), fakeJob: fakeJob{cur.Context()}})
	})
}

// Proxy renders the Door in GoX with the following element as its container:
//
//	~>(new(doors.Door)) <div>content</div>
func (d *Door) Proxy(cur gox.Cursor, el gox.Elem) error {
	return cur.Printer().Send(proxyJob{door: d, el: el, caller: common.CaptureCaller(), fakeJob: fakeJob{cur.Context()}})
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
	ctex.LogCanceled(ctx, "Door inner")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, innerNode{userTask: task, inner: content}, task.InitFrame())
	return ch
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
	ctex.LogCanceled(ctx, "Door outer")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, outerNode{userTask: task, outer: outer}, task.InitFrame())
	return ch
}

// Static removes the Door container and renders content in its place. A nil
// content leaves nothing in place. Unlike [Door.Outer], the result is no longer
// a live Door; later operations change the stored state without putting the
// Door back on the page.
//
// The content belongs to the parent Door: hooks, subscriptions, calls, and
// lifecycle callbacks it registers attach to the parent, even if the render
// fails.
//
// The returned channel is optional to use. On success it sends two nil values
// then closes: the first means the call was scheduled, the second means it was
// applied to the page. On failure it sends an error then closes;
// context.Canceled means a newer operation superseded this one. If the Door is
// not mounted, it closes immediately without sending a value. Do not wait on
// the channel during rendering; to wait, use doors.Go or your own goroutine
// with doors.DetachedContext.
func (d *Door) Static(ctx context.Context, content any) <-chan error {
	ctex.LogCanceled(ctx, "Door static")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, staticNode{userTask: task, outer: content}, task.InitFrame())
	return ch
}

// DeferredInner is the same as [Door.Inner], but the operation waits until the
// current content is rendered and scheduled for delivery. Use it to show a
// placeholder first and replace it with content that is slow to render.
func (d *Door) DeferredInner(ctx context.Context, content any) <-chan error {
	ctex.LogCanceled(ctx, "Door deferred inner")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, innerNode{userTask: task, inner: content, defered: true}, task.InitFrame())
	return ch
}

// DeferredOuter is the same as [Door.Outer], but the operation waits until the
// current content is rendered and scheduled for delivery. Use it to show a
// placeholder first and replace it with content that is slow to render.
func (d *Door) DeferredOuter(ctx context.Context, outer any) <-chan error {
	ctex.LogCanceled(ctx, "Door deferred outer")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, outerNode{userTask: task, outer: outer, defered: true}, task.InitFrame())
	return ch
}

// DeferredStatic is the same as [Door.Static], but the operation waits until
// the current content is rendered and scheduled for delivery. Use it to show a
// placeholder first and replace it with content that is slow to render.
func (d *Door) DeferredStatic(ctx context.Context, content any) <-chan error {
	ctex.LogCanceled(ctx, "Door deferred static")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, staticNode{userTask: task, outer: content, defered: true}, task.InitFrame())
	return ch
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
	ctex.LogCanceled(ctx, "Door reload")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, nodeReload{userTask: task}, task.InitFrame())
	return ch
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
	ctex.LogCanceled(ctx, "Door unmount")
	task, ch := newUserTask(ctx)
	d.schedule(ctx, unmountNode{userTask: task}, task.InitFrame())
	return ch
}

var _ gox.Proxy = &Door{}
var _ gox.Comp = &Door{}
