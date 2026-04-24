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

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type nodeInner struct {
	*userTask
	content any
}

func (t nodeInner) apply(next *node, prev *node) {
	next.mode = modeInner
	next.outer = prev.outer
	next.content = t.content
	if !prev.isMounted() {
		t.userTask.Accept()
		return
	}
	next.tracker = trackerInherit(next, prev.tracker, true)
	next.sync(t.userTask)
}

var _ nodeTask = nodeInner{}

type nodeOuter struct {
	*userTask
	outer gox.Elem
}

func (t nodeOuter) apply(next *node, prev *node) {
	next.mode = modeOuter
	next.outer = t.outer
	if !prev.isMounted() {
		t.userTask.Accept()
		return
	}
	next.tracker = trackerInherit(next, prev.tracker, false)
	next.sync(t.userTask)
}

var _ nodeTask = nodeOuter{}

type nodeReload struct {
	*userTask
}

func (t nodeReload) apply(next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.content = prev.content
	if !prev.isMounted() {
		t.userTask.Accept()
		return
	}
	next.tracker = trackerInherit(next, prev.tracker, next.mode == modeInner)
	next.sync(t.userTask)
}

var _ nodeTask = nodeReload{}

type nodeUnmount struct {
	*userTask
}

func (t nodeUnmount) apply(next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.content = prev.content
	if !prev.isMounted() {
		t.userTask.Accept()
		return
	}
	trackerRemove(prev.tracker, t.userTask)
}

var _ nodeTask = nodeUnmount{}

type nodeStatic struct {
	*userTask
	content any
}

func (t nodeStatic) apply(next *node, prev *node) {
	next.mode = modeStatic
	next.content = t.content
	if !prev.isMounted() {
		t.userTask.Accept()
		return
	}
	trackerShutdown(prev.tracker)
	next.tracker = prev.tracker
	next.sync(t.userTask)
}

var _ nodeTask = nodeStatic{}

type nodeProxy struct {
	el     gox.Elem
	pipe   *pipe
	buffer *deque.Deque[any]
}

func (t nodeProxy) apply(next *node, prev *node) {
	next.mode = modeBlend
	next.outer = t.el
	next.content = prev.content
	if prev.isMounted() {
		trackerRemove(prev.tracker, nil)
	}
	next.tracker = trackerCreate(next, t.pipe)
	next.render(t.pipe, t.buffer)
}

var _ nodeTask = nodeProxy{}

type nodeRender struct {
	pipe   *pipe
	buffer *deque.Deque[any]
}

func (t nodeRender) apply(next *node, prev *node) {
	next.mode = prev.mode
	next.outer = prev.outer
	next.content = prev.content
	if prev.isMounted() {
		trackerRemove(prev.tracker, nil)
	}
	if next.mode != modeStatic {
		next.tracker = trackerCreate(next, t.pipe)
	}
	next.render(t.pipe, t.buffer)
}

var _ nodeTask = nodeRender{}

type nodeTask interface {
	apply(next *node, prev *node)
}

func newUserTask(ctx context.Context) (*userTask, <-chan error) {
	ch := make(chan error, 1)
	return &userTask{&ch, ctex.GetFrames(ctx)}, ch
}

type userTask struct {
	ch     *chan error
	frames ctex.Frames
}

func (t *userTask) JoinedFrame() shredder.Frame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.JoinedFrame()
}

func (t *userTask) CallFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Call()
}

func (t *userTask) RenderFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Render()
}

func (t *userTask) Report(err error) {
	if t == nil {
		return
	}
	if t.ch == nil {
		return
	}
	*t.ch <- err
	close(*t.ch)
	t.ch = nil
}

func (t *userTask) Cancel() {
	if t == nil {
		return
	}
	t.Report(context.Canceled)
}

func (t *userTask) Accept() {
	if t == nil {
		return
	}
	if t.ch != nil {
		close(*t.ch)
		t.ch = nil
	}
}
