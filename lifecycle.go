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
// content completes: the HTML is fully rendered and the resulting page or
// update is enqueued for delivery to the client. If the content of ctx is
// already on the page (for example in an event handler), f fires promptly.
//
// OnReady is best-effort and fires at most once: if the render cycle fails or
// is superseded by a newer door operation, f never runs. Tie resource cleanup
// to [OnClean], never to OnReady. When content is replaced, the previous
// content's [OnClean] callbacks run before the replacing content's OnReady
// callbacks.
//
// f runs on the instance goroutine pool with a context equivalent to
// [DetachedContext] of ctx. It must not block; if you need to wait, start a
// goroutine with the provided ctx.
//
// ctx must belong to a Doors render or handler; otherwise OnReady panics.
//
// Example:
//
//	doors.OnReady(ctx, func(ctx context.Context) {
//	    // content is rendered and its update is on the way to the client
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
// enclosing door is updated, removed, re-rendered by an ancestor, fails to
// render, or the instance ends. Exactly one of these eventually happens to
// every rendered piece of content, so f runs exactly once per registration.
//
// f runs inline on framework goroutines and receives no context: by the time
// it runs, the registering context is canceled. It must not block;
// fire-and-forget Doors calls with a context captured beforehand (for
// example from [InstanceContext]) are safe.
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
