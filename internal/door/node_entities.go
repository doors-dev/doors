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
	"github.com/doors-dev/doors/internal/door/pipe"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Container = pipe.ProxyContainer

type contentsNode interface {
	Contents() *contents
}

type nodeMount interface {
	contentsNode
	Kill() bool
	ReplaceId() uint64
	Tracker() *tracker
	WriteFrame() *shredder.ValveFrame
	MountFrame() *atomic.Pointer[shredder.ValveFrame]
}

type mountNode struct {
	tracker    *tracker
	contents   *contents
	mountFrame *atomic.Pointer[shredder.ValveFrame]
	killed     atomic.Bool
}

func (c *mountNode) initMountFrame() {
	if c.mountFrame != nil {
		panic("already initialized")
	}
	c.mountFrame = &atomic.Pointer[shredder.ValveFrame]{}
	c.mountFrame.Store(&shredder.ValveFrame{})
}

func (c *mountNode) Kill() bool {
	killed := c.killed.CompareAndSwap(false, true)
	if !killed {
		return false
	}
	c.tracker.kill()
	return true
}

func (c *mountNode) ReplaceId() uint64 {
	return c.tracker.id
}

func (c *mountNode) SetMountFrame(f *shredder.ValveFrame) {
	prev := c.mountFrame.Swap(f)
	f.Run(nil, nil, func(b bool) {
		prev.Activate()
	})
}

func (c *mountNode) MountFrame() *atomic.Pointer[shredder.ValveFrame] {
	return c.mountFrame
}

func (c *mountNode) WriteFrame() *shredder.ValveFrame {
	return c.mountFrame.Load()
}

func newTaskNode(ctx context.Context) (*taskNode, <-chan error) {
	ch := make(chan error, 1)
	return &taskNode{&ch, ctex.GetFrames(ctx)}, ch
}

type taskNode struct {
	ch     *chan error
	frames ctex.Frames
}

func (t *taskNode) ContextJoinedFrame() shredder.Frame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.JoinedFrame()
}

func (t *taskNode) ContextSendFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Send()
}

func (t *taskNode) ContextRenderFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frames.Render()
}

func (t *taskNode) Report(err error) {
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

func (t *taskNode) Cancel() {
	if t == nil {
		return
	}
	t.Report(context.Canceled)
}

func (t *taskNode) Accept() {
	if t == nil {
		return
	}
	if t.ch != nil {
		close(*t.ch)
		t.ch = nil
	}
}

func (n *mountNode) Tracker() *tracker {
	return n.tracker
}

func (n *mountNode) Contents() *contents {
	return n.contents
}

type replaceNode struct {
	*taskNode
	replaceId uint64
	content   any
}

type updateNode struct {
	mountNode
	*taskNode
	content any
}

var _ nodeMount = &updateNode{}

type redrawNode struct {
	*taskNode
}

type rebaseNode struct {
	*taskNode
	elem gox.Elem
}

type unmountNode struct {
	*taskNode
	remove    bool
	contents  *contents
	proxyElem gox.Elem
}

var _ contentsNode = &unmountNode{}

func (n *unmountNode) wasProxy() bool {
	return n.proxyElem != nil
}

func (n *unmountNode) Contents() *contents {
	return n.contents
}

type contents struct {
	initializeFrame *shredder.ValveFrame
	container       *Container
	content         any
}

type proxyNode struct {
	mountNode
	*taskNode
	elem         gox.Elem
	prevContents *contents
	replaceId    atomic.Uint64
}

func (c *proxyNode) ReplaceId() uint64 {
	return c.replaceId.Swap(0)
}

func (c *proxyNode) setReplaceId() uint64 {
	return c.replaceId.Swap(c.tracker.id)
}

func (c *proxyNode) inheritReplaceId(prev nodeMount) {
	c.replaceId.Store(prev.ReplaceId())
}

var _ nodeMount = &proxyNode{}

type editorNode struct{}
