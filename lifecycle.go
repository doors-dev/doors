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
	"github.com/doors-dev/doors/internal/shredder"
)

// OnReady registers on to run when the render cycle producing the current
// content completes: the HTML is fully rendered and the resulting page or
// update is enqueued for delivery to the client. If the content of ctx is
// already on the page (for example in an event handler), on fires promptly.
//
// OnReady is best-effort and fires at most once: if the render cycle fails or
// is superseded by a newer door operation, on never runs. Tie resource cleanup
// to [OnClean], never to OnReady. When content is replaced, the previous
// content's [OnClean] callbacks are dispatched to the pool before the
// replacing content's OnReady callbacks.
//
// on runs on the instance goroutine pool with a context equivalent to
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
func OnReady(ctx context.Context, on func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnReady")
	core := ctx.Value(common.KeyCore).(core.Core)
	ctx = DetachedContext(ctx)
	core.Door().ReadyFrame().Submit(ctx, core.Instance().Runtime(), func(ok bool) {
		if !ok {
			return
		}
		on(ctx)
	})
}

// OnFlush runs ops within the current dispatch batch and registers on to run
// once that batch is flushed: every doors operation initiated from the batch
// context is processed and its updates are enqueued for delivery to the
// client. In a handler, the batch spans the whole handler; if ctx carries no
// batch (render contexts, [DetachedContext], goroutines), OnFlush opens one
// spanning the ops calls. on never runs before the content that produced ctx
// is flushed itself.
//
// on runs exactly once, even if the owner is canceled, the producing render
// fails, or the instance is shutting down — the provided context reports such
// states. on runs on the instance goroutine pool with a context equivalent to
// [DetachedContext] of ctx; it must not block. Callbacks sharing one batch are
// dispatched in reverse registration order. As an exception, on a shutting
// down instance (or when a canceled owner meets a saturated pool) on runs
// inline on the dispatching goroutine instead of the pool.
//
// ctx must belong to a Doors render or handler; otherwise OnFlush panics.
//
// Example:
//
//	doors.OnFlush(ctx, func(ctx context.Context) {
//	    // updates below are enqueued for delivery
//	}, func(ctx context.Context) {
//	    d.Update(ctx, content)
//	})
func OnFlush(ctx context.Context, on func(ctx context.Context), ops ...func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnFlush")
	detached := DetachedContext(ctx)
	core := ctx.Value(common.KeyCore).(core.Core)
	var afterFrame shredder.SimpleFrame = shredder.FreeFrame{}
	if frame, ok := ctex.AfterFrame(ctx); ok {
		afterFrame = frame.After()
	} else if len(ops) != 0 {
		ctx, frame = ctex.AfterFrameInsert(ctx)
		afterFrame = frame.After()
		defer frame.Activate()
	}
	joined := shredder.Join(ctx, false, core.Door().ReadyFrame(), afterFrame)
	joined.Submit(ctx, core.Instance().Runtime(), func(bool) {
		on(detached)
	})
	joined.Release()
	for _, op := range ops {
		op(ctx)
	}
}

// OnClean registers f to run when the current content is cleared: the
// enclosing door is updated, removed, re-rendered by an ancestor, fails to
// render, or the instance ends. Exactly one of these eventually happens to
// every rendered piece of content, so f runs exactly once per registration.
//
// f runs on the instance goroutine pool and receives no context: by the time
// it runs, the registering context is canceled. As an exception, on a
// shutting down instance f runs inline on the framework goroutine. It must
// not block; fire-and-forget Doors calls with a context captured beforehand
// (for example from [InstanceContext]) are safe.
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
	ctex.LogCanceled(ctx, "OnClean")
	core := ctx.Value(common.KeyCore).(core.Core)
	core.Door().CleanFrame().Submit(context.Background(), core.Instance().Runtime(), func(bool) {
		f()
	})
}
