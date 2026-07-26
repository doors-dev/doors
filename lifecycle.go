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
// content completes and the resulting page or update is enqueued for delivery
// to the client. If that content is already on the page, for example in an
// event handler, on fires promptly.
//
// OnReady is best-effort and fires at most once: if the render cycle fails or
// is superseded by a newer operation, on never runs. Tie cleanup to [OnClean].
//
// on runs on the instance goroutine pool with a context equivalent to
// [DetachedContext] of ctx. It must not block; if you need to wait, start a
// goroutine with the provided ctx.
//
// ctx must belong to a Doors render or handler; otherwise OnReady panics.
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

// OnSettle runs ops within the current dispatch batch and registers on to run
// once that batch settles: every Doors operation started from the batch is
// processed and its updates are enqueued for delivery to the client. In a
// handler the batch spans the whole handler; otherwise OnSettle opens one
// spanning the ops calls. on never runs before the content that produced ctx
// is ready itself.
//
// Where [OnReady] tracks one render cycle, OnSettle tracks everything the batch
// started, including the Door updates and beam propagation it triggered.
// Settling is server-side: the updates are queued, not yet written to the
// connection or acknowledged by the browser.
//
// on runs exactly once, even if the owner is canceled, the producing render
// fails, or the instance is shutting down; the provided context reports such
// states. It runs on the instance goroutine pool with a context equivalent to
// [DetachedContext] of ctx and must not block. When the instance is shutting
// down, or the owner is already canceled, it runs inline on the calling
// goroutine.
//
// ctx must belong to a Doors render or handler; otherwise OnSettle panics.
func OnSettle(ctx context.Context, on func(ctx context.Context), ops ...func(ctx context.Context)) {
	ctex.LogCanceled(ctx, "OnSettle")
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
// enclosing Door is updated, removed, rerendered by an ancestor, fails to
// render, or the instance ends. Exactly one of these eventually happens to
// every rendered piece of content, so f runs exactly once per registration.
//
// f runs on the instance goroutine pool and receives no context; the
// registering context is already canceled by then. On a shutting down instance
// it runs inline. It must not block; fire-and-forget Doors calls with a context
// captured beforehand, for example from [InstanceContext], are safe.
//
// ctx must belong to a Doors render or handler; otherwise OnClean panics.
func OnClean(ctx context.Context, f func()) {
	ctex.LogCanceled(ctx, "OnClean")
	core := ctx.Value(common.KeyCore).(core.Core)
	core.Door().CleanFrame().Submit(context.Background(), core.Instance().Runtime(), func(bool) {
		f()
	})
}
