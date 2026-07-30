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

package beam

import (
	"context"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type freeScreen struct {
	ctx    context.Context
	source anySource
	w      *watcher
	thread shredder.Thread
}

func newFreeScreen(ctx context.Context, source anySource, w *watcher) *freeScreen {
	f := &freeScreen{ctx: ctx, source: source, w: w}
	w.register(f)
	return f
}

func (f *freeScreen) init(seq uint) {
	defer func() {
		if r := recover(); r != nil {
			f.w.abort()
			panic(r)
		}
	}()
	f.w.init(f.ctx, seq)
}

func (f *freeScreen) removeWatcher(*watcher) {
	f.source.removeFreeSub(f)
}

func (f *freeScreen) sync(ctx context.Context, cleanFrame shredder.SimpleFrame, sourceFrame shredder.SimpleFrame, seq uint, isStopped func() bool) {
	frame := shredder.Join(ctx, true, sourceFrame, f.thread.Frame(), f.w.syncFrame())
	defer frame.Release()
	frame.Submit(f.ctx, nil, func(ok bool) {
		if f.ctx.Err() != nil {
			f.w.Cancel()
			return
		}
		if !ok || isStopped() {
			return
		}
		defer func() {
			if r := recover(); r != nil {
				common.Logger(f.ctx).Error("beam subscriber panic", "recover", r)
				f.w.abort()
			}
		}()
		f.w.sync(f.ctx, seq, cleanFrame)
	})
}
