package door

import (
	"sync/atomic"

	"github.com/doors-dev/doors/internal/sh"
)

type Door struct {
	node atomic.Pointer[node]
}

func (d *Door) takeover(next *node) {
	prev := d.node.Swap(next)
	next.takover(prev)
}

type nodeKind int

const (
	unmountedNode nodeKind = iota
	replacedNode
	updatedNode
	jobNode
	proxyNode
)

type node struct {
	kind          nodeKind
	initFrame     sh.ValveFrame
	updateFrame   sh.SimpleFrame
	renderShread  sh.Shread
}

func (n *node) takover(prev *node) {
	frame := sh.Join(n.initFrame, prev.initShread.Frame())
	defer frame.Release()
	frame.Run(nil, func() {
		switch n.kind {
		case unmountedNode:
			panic("door: unmounted node can't takeover")
		case jobNode:
			n.jobTakeover(prev)
		case proxyNode:
			n.proxyTakeover(prev)
		case updatedNode:
			n.updatedTakeover(prev)
		case replacedNode:
			n.replacedTakeover(prev)
		}
	})
}

func (n *node) updatedTakeover(prev *node) {
	defer n.initFrame.Release()
	switch prev.kind {
	case updatedNode, proxyNode, jobNode:
		n.updateFrame = prev.updateFrame
	}
}

func (n *node) proxyTakeover(prev *node) {
}

func (n *node) render(parent parent, parentPipe *pipe) error {

}

type tracker struct {
	frame sh.Frame
}
