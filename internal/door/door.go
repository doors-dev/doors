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

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Door struct {
	node atomic.Pointer[node]
}

func (d *Door) proxy(p *pipe, el gox.Elem) {
	task := nodeProxy{
		pipe:   p,
		buffer: p.branch(),
		el:     el,
	}
	d.schedule(task, p.renderFrame)
}

func (d *Door) render(p *pipe) {
	task := nodeRender{
		pipe:   p,
		buffer: p.branch(),
	}
	d.schedule(task, p.renderFrame)
}

func (d *Door) outer(ctx context.Context, outer gox.Elem) <-chan error {
	ctex.LogCanceled(ctx, "Door outer")
	userTask, ch := newUserTask(ctx)
	task := nodeOuter{
		userTask: userTask,
		outer:    outer,
	}
	d.schedule(task, userTask.JoinedFrame())
	return ch
}

func (d *Door) inner(ctx context.Context, content any) <-chan error {
	ctex.LogCanceled(ctx, "Door inner")
	userTask, ch := newUserTask(ctx)
	task := nodeInner{
		userTask: userTask,
		content:  content,
	}
	d.schedule(task, userTask.JoinedFrame())
	return ch
}

func (d *Door) static(ctx context.Context, content any) <-chan error {
	ctex.LogCanceled(ctx, "Door static")
	userTask, ch := newUserTask(ctx)
	task := nodeStatic{
		userTask: userTask,
		content:  content,
	}
	d.schedule(task, userTask.JoinedFrame())
	return ch
}

func (d *Door) unmount(ctx context.Context) <-chan error {
	ctex.LogCanceled(ctx, "Door unmount")
	userTask, ch := newUserTask(ctx)
	task := nodeUnmount{
		userTask: userTask,
	}
	d.schedule(task, userTask.JoinedFrame())
	return ch
}

func (d *Door) reload(ctx context.Context) <-chan error {
	ctex.LogCanceled(ctx, "Door reload")
	userTask, ch := newUserTask(ctx)
	task := nodeReload{
		userTask: userTask,
	}
	d.schedule(task, userTask.JoinedFrame())
	return ch
}

func (d *Door) reloadSelf(ctx context.Context, prev *node) <-chan error {
	ctex.LogCanceled(ctx, "Door reload")
	userTask, ch := newUserTask(ctx)
	task := nodeReload{
		userTask: userTask,
	}
	if !d.atomicSchedule(prev, task, userTask.JoinedFrame()) {
		userTask.Cancel()
	}
	return ch
}

func (d *Door) unmountedSelf(prev *node) {
	node := &node{
		door:    d,
		mode:    prev.mode,
		outer:   prev.outer,
		content: prev.content,
	}
	node.guard.Activate()
	d.node.CompareAndSwap(prev, node)
}

func (d *Door) schedule(task nodeTask, externalFrame shredder.Frame) {
	next := &node{
		door: d,
	}
	prev := d.node.Swap(next)
	if prev == nil {
		prev = &node{
			door: d,
			mode: modeOuter,
		}
		prev.guard.Activate()
	}
	initFrame := shredder.Join(true, &prev.guard, externalFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.guard.Activate()
		task.apply(next, prev)
	})
}

func (d *Door) atomicSchedule(prev *node, task nodeTask, externalFrame shredder.Frame) bool {
	next := &node{
		door: d,
	}
	ok := d.node.CompareAndSwap(prev, next)
	if !ok {
		return false
	}
	initFrame := shredder.Join(true, &prev.guard, externalFrame)
	defer initFrame.Release()
	initFrame.Run(nil, nil, func(b bool) {
		defer next.guard.Activate()
		task.apply(next, prev)
	})
	return true
}

var _ gox.Proxy = &Door{}
var _ gox.Editor = &Door{}
