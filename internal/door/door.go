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
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

// Door renders content that can be updated, replaced, or removed after render.
type Door struct {
	node atomic.Pointer[node]
}

func (d *Door) render(ctx context.Context, p *pipe, caller common.Caller) {
	d.schedule(ctx, renderNode{
		pipe:   p,
		buffer: p.branch(),
		ctx:    ctx,
		caller: caller,
	}, p.renderFrame)
}

func (d *Door) proxy(p *pipe, el gox.Elem, ctx context.Context, caller common.Caller) {
	d.schedule(ctx, nodeProxy{
		pipe:   p,
		buffer: p.branch(),
		el:     el,
		ctx:    ctx,
		caller: caller,
	}, p.renderFrame)
}

func (d *Door) schedule(ctx context.Context, task nodeTask, externalFrame shredder.ReleaseFrame) {
	next := &node{
		door: d,
	}
	prev := d.node.Swap(next)
	if prev == nil {
		prev = &node{
			door: d,
			mode: modeOuter,
		}
		prev.initGuard.Activate()
	}
	initFrame := shredder.JoinRelease(ctx, &prev.initGuard, externalFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.initGuard.Activate()
		task.apply(initFrame, next, prev)
	})
}

func (d *Door) reloadSelf(ctx context.Context, prev *node) <-chan error {
	ctex.LogCanceled(ctx, "Door reload")
	userTask, ch := newUserTask(ctx)
	frame := userTask.InitFrame()
	if !d.atomicSchedule(ctx, prev, nodeReload{userTask: userTask}, frame) {
		frame.Release()
		userTask.Cancel()
	}
	return ch
}

func (d *Door) atomicSchedule(ctx context.Context, prev *node, task nodeTask, externalFrame shredder.ReleaseFrame) bool {
	next := &node{
		door: d,
	}
	if !d.node.CompareAndSwap(prev, next) {
		return false
	}
	initFrame := shredder.JoinRelease(ctx, &prev.initGuard, externalFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.initGuard.Activate()
		task.apply(initFrame, next, prev)
	})
	return true
}
