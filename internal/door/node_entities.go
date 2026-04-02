// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
	frame := ctex.Frame(ctx)
	ch := make(chan error, 1)
	return &taskNode{&ch, frame}, ch
}

type taskNode struct {
	ch    *chan error
	frame shredder.SimpleFrame
}

func (t *taskNode) TaskFrame() shredder.SimpleFrame {
	if t == nil {
		return shredder.FreeFrame{}
	}
	return t.frame
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
