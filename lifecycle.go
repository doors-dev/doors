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
	"encoding/json"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/shredder"
)

// OnReady registers on to run when the render cycle producing the current
// content completes and the resulting page or update is enqueued for delivery
// to the client. If that content is already on the page, for example in an
// event handler, on fires promptly.
//
// OnReady is best-effort and fires at most once: if the render cycle fails or
// is superseded by a newer operation, on never runs. Content rendered by
// [Door.Static], [Door.DeferredStatic] or [DeferredStatic] is the exception: it
// belongs to the parent Door, so on still runs if that render fails. Tie
// cleanup to [OnClean].
//
// on runs inline on the framework goroutine completing the cycle — on the
// calling goroutine when the content is already on the page — with a context
// equivalent to [DetachedContext] of ctx. It must not block; if you need to
// wait, start a goroutine with the provided ctx.
//
// ctx must belong to a Doors render or handler; otherwise OnReady panics.
func OnReady(ctx context.Context, on func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnReady")
	core := ctx.Value(common.KeyCore).(core.Core)
	detached := DetachedContext(ctx)
	ready := core.Door().ReadyFrame()
	defer ready.Release()
	ready.Run(ctx, core.Instance().Runtime(), func(ok bool) {
		if !ok {
			return
		}
		on(detached)
	})
}

// OnSettle runs ops as part of the work ctx belongs to and registers on to run
// once that work settles: every Doors operation started from it is processed
// and its updates are enqueued for delivery to the client. The content of a
// deferred operation is not awaited. In a handler or in the ops of OnSettle
// and [OnBatch], OnSettle joins that work; otherwise it starts new work
// spanning the ops calls. on never runs before the content that produced ctx
// is ready itself.
//
// Where [OnReady] tracks one render cycle, OnSettle tracks everything the work
// started, including the Door updates and beam propagation it triggered.
// Settling is server-side: the updates are queued, not yet written to the
// connection or acknowledged by the browser.
//
// on runs exactly once, even if the owner is canceled, the producing render
// fails, or the instance is shutting down; the provided context reports such
// states. It runs inline on the goroutine that settles the work — on the
// calling goroutine when everything already settled — with a context
// equivalent to [DetachedContext] of ctx, except that calls issued while on
// runs keep their place with the handler's updates. It must not block.
//
// Using a handler ctx after the handler returned is misuse; use
// [DetachedContext] or [HoldSettle].
//
// ctx must belong to a Doors render or handler; otherwise OnSettle panics.
func OnSettle(ctx context.Context, on func(ctx context.Context), ops ...func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnSettle")
	core := ctx.Value(common.KeyCore).(core.Core)
	frame, ok := ctex.GetFrames(ctx).AfterFrame()
	if !ok && len(ops) == 0 {
		detached := ctex.NewDetachedContext(ctx, ctx, true)
		ready := core.Door().ReadyFrame()
		ready.Run(ctx, core.Instance().Runtime(), func(bool) {
			on(detached)
		})
		ready.Release()
		return
	}
	if !ok {
		ctx, frame = ctex.NewAfterFrame(ctx, core.Instance().ID(), 0)
		defer frame.Activate()
		holdUntilReady(ctx, core, frame)
	}
	detached := ctex.NewDetachedContext(ctx, ctx, true)
	frame.After().Run(ctx, core.Instance().Runtime(), func(bool) {
		on(detached)
	})
	for _, op := range ops {
		op(ctx)
	}
}

// OnBatch batches updates yourself, anywhere: the updates made in ops are
// applied at once, and on runs when the browser has applied them. In a handler
// with Batch set, or in ops of another OnBatch, OnBatch joins that batch;
// otherwise it opens one.
//
// In a handler the batch closes when the handler settles, not when ops return;
// use [DetachedContext] of ctx to batch on its own. In an [OnSettle] callback
// of a handler, the handler's completion waits for the batch.
//
// Where [OnSettle] reports that the updates are queued for delivery, OnBatch
// reports that the browser applied them.
//
// on runs exactly once, also if the instance ends first; the provided context
// reports such states. It runs with a context equivalent to [DetachedContext]
// of ctx and must not block.
//
// ctx must belong to a Doors render or handler; otherwise OnBatch panics.
func OnBatch(ctx context.Context, on func(ctx context.Context), ops ...func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnBatch")
	detached := DetachedContext(ctx)
	core := ctx.Value(common.KeyCore).(core.Core)
	inst := core.Instance()
	fs := ctex.GetFrames(ctx)
	batchFrame, joined := fs.BatchFrame()
	if !joined {
		after, ok := fs.AfterFrame()
		if !ok {
			ctx, after = ctex.NewAfterFrame(ctx, inst.ID(), fs.Group())
			defer after.Activate()
			holdUntilReady(ctx, core, after)
		}
		ctx, batchFrame = ctex.NewBatch(ctx, inst.NewID())
		markerCtx := ctex.NewDetachedContext(ctx, inst.Runtime().Context(), true)
		door := core.Door().ID()
		inst.UserCall(markerCtx, actions.BatchInit{}, nil, nil, actions.CallParams{}, door, 0)
		after.After().Run(nil, nil, func(bool) {
			inst.UserCall(markerCtx, actions.BatchFlush{}, func(json.RawMessage, error) {
				batchFrame.Activate()
			}, batchFrame.Activate, actions.CallParams{}, door, 0)
		})
	}
	batchFrame.Submit(detached, inst.Runtime(), func(bool) {
		on(detached)
	})
	for _, op := range ops {
		op(ctx)
	}
}

func holdUntilReady(ctx context.Context, c core.Core, after *shredder.AfterFrame) {
	release := shredder.Join(ctx, after).Release
	ready := c.Door().ReadyFrame()
	ready.Run(ctx, c.Instance().Runtime(), func(bool) {
		release()
	})
	ready.Release()
}

// OnClean registers f to run when the current content is cleared: the
// enclosing Door is updated, removed, rerendered by an ancestor, fails to
// render, or the instance ends. Exactly one of these eventually happens to
// every rendered piece of content, so f runs exactly once per registration.
// Content rendered by [Door.Static], [Door.DeferredStatic] or [DeferredStatic]
// belongs to the parent Door, so f runs when the parent's content is cleared,
// even if that render fails.
//
// f runs inline on the framework goroutine performing the clean — on the
// calling goroutine when the content is already cleaned — and receives no
// context; the registering context is already canceled by then. It must not
// block; fire-and-forget Doors calls with a context captured beforehand, for
// example from [InstanceContext], are safe.
//
// ctx must belong to a Doors render or handler; otherwise OnClean panics.
func OnClean(ctx context.Context, f func()) {
	ctex.LogCanceled(ctx, "OnClean")
	core := ctx.Value(common.KeyCore).(core.Core)
	core.Door().CleanFrame().Run(context.Background(), core.Instance().Runtime(), func(bool) {
		f()
	})
}

// HoldSettle keeps the handler's work open after the handler returns, until
// release is called. While held, it does not settle: [OnSettle] callbacks
// wait, and in a hook the client keeps the indicator, the scope, and the $hook
// promise pending; in a handler with Batch set, the updates stay unapplied
// until release. Use it to hand the work over to a goroutine.
//
// Call it synchronously in the handler; release is idempotent and may run
// settle callbacks inline. Operations started with the handler ctx join the
// work, operations started with [DetachedContext] do not. Always call release,
// typically deferred in the goroutine; held work settles otherwise only when
// the instance ends.
//
// Outside a handler or the ops of [OnSettle] and [OnBatch], for example in an
// OnSettle callback, release is a no-op.
func HoldSettle(ctx context.Context) (release func()) {
	ctex.LogCanceled(ctx, "HoldSettle")
	after, ok := ctex.GetFrames(ctx).AfterFrame()
	if !ok {
		return func() {}
	}
	frameRelease := shredder.Join(ctx, after).Release
	stop := context.AfterFunc(InstanceContext(ctx), frameRelease)
	return func() {
		stop()
		frameRelease()
	}
}
