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

package doors

import (
	"context"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
)

// OnReady registers f to run when the render cycle producing the current
// content is finished: the HTML was fully rendered on the server and the
// resulting page or update was enqueued for delivery to the client — the same
// moment the first nil is sent on X-prefixed method channels (see
// [Door.XInner]). It does not mean the browser applied the update; for
// client-applied confirmation use the second value of an X-prefixed method
// channel.
//
// The cycle is determined by ctx. During rendering it is the render cycle the
// content belongs to (the update of the enclosing door, or the initial page
// render), so f effectively fires once per render of the surrounding content.
// In an event handler — or any other ctx whose content is already on the
// page — the cycle has already finished and f fires promptly. With a ctx from
// [InstanceContext], f fires once the initial page render is complete.
//
// f fires at most once per registration. If the render cycle fails, is
// overwritten by a newer door operation before completing, or the dynamic
// owner is already unmounted, f is never called — OnReady is best-effort.
// Tie resource cleanup to [OnClean], which always fires, never to OnReady.
// When content is replaced, [OnClean] callbacks of the previous content
// complete before OnReady callbacks of the replacing content run.
//
// f runs on the instance goroutine pool, never on the calling goroutine. A
// panic in f is recovered, logged, and ends the instance (the same policy as
// event handlers and render code). The context passed to f is detached from
// the render frame — equivalent to [DetachedContext] of the registration
// ctx: it keeps the current Doors ownership, is canceled when the dynamic
// owner is unmounted, and is safe to pass to any Doors API. Do not block
// inside f; if you need to wait (for example on X-prefixed channels), start
// your own goroutine with the provided ctx.
//
// ctx must belong to a Doors render or handler; otherwise OnReady panics.
//
// Example:
//
//	doors.OnReady(ctx, func(ctx context.Context) {
//	    // this content is rendered and its update is on the way to the client
//	})
func OnReady(ctx context.Context, f func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnReady")
	c := ctx.Value(common.KeyCore).(core.Core)
	fctx := DetachedContext(ctx)
	c.Door().Ready(func() {
		f(fctx)
	})
}

// OnClean registers f to run when the current content is cleared: the
// enclosing door is updated by [Door.Inner], [Door.Outer], or [Door.Reload],
// removed by [Door.Unmount] or [Door.Static], removed by a re-render of an
// ancestor door, fails to render, or the instance ends. Exactly one of these
// eventually happens to every rendered piece of content, so f runs exactly
// once per registration. With a ctx from [InstanceContext], f runs when the
// instance ends.
//
// When content is replaced by a new render, f is deferred until the replacing
// render cycle completes and its update is enqueued; the replaced content's
// OnClean callbacks run before the replacing content's [OnReady] callbacks.
// On unmount, removal, and shutdown, f runs synchronously inside the cleanup
// cascade. Nested doors' callbacks run before their parents'; order across
// sibling doors is not defined. If the content is already cleaned when
// OnClean is called, f runs immediately on the calling goroutine.
//
// f runs inline on framework goroutines and receives no context: by the time
// it runs, the registering context is canceled. It must be fast, must not
// block, and must never wait on X-prefixed channels. Fire-and-forget Doors
// calls using a context captured beforehand (for example from
// [InstanceContext] or [SessionContext]) are safe. A panic in f is recovered,
// logged, and ends the instance.
//
// Registered from the attribute context of the door's container element, f is
// bound to the container element's lifetime and survives [Door.Inner]
// updates, matching how <title> and <meta> overrides revert.
//
// ctx must belong to a Doors render or handler; otherwise OnClean panics.
//
// Example:
//
//	sub := pubsub.Subscribe(topic)
//	doors.OnClean(ctx, func() {
//	    sub.Close()
//	})
func OnClean(ctx context.Context, f func()) {
	c := ctx.Value(common.KeyCore).(core.Core)
	rt := c.Instance().Runtime()
	c.Door().Clean(func() {
		rt.SafeCtxFun(context.Background(), func(context.Context) { f() })
	})
}
