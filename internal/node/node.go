package node

import (
	"context"
	"errors"
	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"io"
	"log/slog"
	"sync"
)

type nodeMode int

const (
	dynamic nodeMode = iota
	static
	removed
)

type Node struct {
	mu             sync.Mutex
	suspended      bool
	mode           nodeMode
	core           *core
	content        templ.Component
	commitCounter  uint
	commitBuffer   *commit
	pushingCommits map[uint]*commit
	bufferedCall   *commitCall
}

func (n *Node) addHook(ctx context.Context, h Hook) (*HookEntry, bool) {
	n.lock()
	defer n.unlock()
	if !n.isActive() {
		return nil, false
	}
	hookId := n.core.instance.NewId()
	hook := newHook(common.ClearBlockingCtx(ctx), h)
	n.core.instance.RegisterHook(n.core.id, hookId, hook)
	return &HookEntry{
		NodeId: n.core.id,
		HookId: hookId,
		inst:   n.core.instance,
	}, true
}

func (n *Node) suspend() {
	n.lock()
	defer n.unlock()
	n.suspended = true
	n.resetCommits()
	if n.commitBuffer != nil {
		n.commitBuffer.suspend()
		n.commitBuffer = nil
	}
	if n.mode > dynamic {
		return
	}
	n.core.kill(false)
}

func (n *Node) isActive() bool {
	return n.core != nil && !n.suspended
}

func (n *Node) commitWriteErr(c *commitCall) bool {
	n.lock()
	defer n.unlock()
	if n.suspended {
		return false
	}
	if c.id != n.commitCounter {
		return false
	}
	n.bufferedCall = c
	return true
}

func (n *Node) commitResult(id uint, err error) {
	n.lock()
	defer n.unlock()
	if id == n.commitCounter && n.commitBuffer != nil {
		n.commitBuffer.result(err)
		return
	}
	for executingId := range n.pushingCommits {
		if executingId < id {
			n.pushingCommits[executingId].owerwrite()
			delete(n.pushingCommits, executingId)
			continue
		}
		if executingId == id {
			n.pushingCommits[executingId].result(err)
			delete(n.pushingCommits, executingId)
		}
	}
}

func (n *Node) call() <-chan Call {
	n.lock()
	defer n.unlock()
	ch := make(chan Call, 1)
	if (n.commitBuffer == nil && n.bufferedCall == nil) || !n.isActive() {
		close(ch)
		return ch
	}
	if n.bufferedCall != nil {
		ch <- n.bufferedCall
		n.bufferedCall = nil
		return ch
	}
	commitId := n.commitCounter
	n.pushingCommits[n.commitCounter] = n.commitBuffer
	n.commitBuffer = nil
	if n.mode == removed {
		n.core.renderRemoveCall(ch, commitId)
		return ch
	}
	if n.mode == static {
		n.core.renderReplaceCall(n.content, ch, commitId)
		return ch
	}
	n.core.renderUpdateCall(n.content, ch, commitId)
	return ch
}

func (n *Node) commit(ctx context.Context, content templ.Component) <-chan error {
	n.content = content
	if !n.isActive() {
		ch := make(chan error, 0)
		close(ch)
		return ch
	}
	waiting := (n.commitBuffer != nil || n.bufferedCall != nil)
	n.bufferedCall = nil
	if n.commitBuffer != nil {
		n.commitBuffer.owerwrite()
	}
	n.commitCounter += 1
	n.commitBuffer = newCommit(ctx)
	if !waiting {
		n.core.instance.Call((*nodeCaller)(n))
	}
	return n.commitBuffer.ch
}

func (n *Node) update(ctx context.Context, content templ.Component) (<-chan error, bool) {
	n.lock()
	defer n.unlock()
	if ctx.Err() != nil {
		return nil, false
	}
	if n.mode > dynamic {
		return nil, false
	}
	if n.isActive() {
		n.core.kill(true)
		n.core = newCore(n, n.core.parentCtx, n.core.instance, n.core.id)
	}
	ch := n.commit(ctx, content)
	return ch, true
}

func (n *Node) replace(ctx context.Context, content templ.Component) (<-chan error, bool) {
	n.lock()
	defer n.unlock()
	if ctx.Err() != nil {
		return nil, false
	}
	if n.mode > dynamic {
		return nil, false
	}
	n.mode = static
	if n.isActive() {
		n.core.kill(true)
		n.core.parent.removeChild(n)
	}
	ch := n.commit(ctx, content)
	return ch, true

}

func (n *Node) remove(ctx context.Context) (<-chan error, bool) {
	n.lock()
	defer n.unlock()
	if ctx.Err() != nil {
		return nil, false
	}
	if n.mode > dynamic {
		return nil, false
	}
	n.mode = removed
	if n.isActive() {
		n.core.kill(true)
		n.core.parent.removeChild(n)
	}
	ch := n.commit(ctx, nil)
	return ch, true
}

func (n *Node) resetCommits() {
	n.bufferedCall = nil
	if len(n.pushingCommits) == 0 {
		return
	}
	if n.commitBuffer == nil {
		n.commitBuffer = n.pushingCommits[n.commitCounter]
		delete(n.pushingCommits, n.commitCounter)
	}
	for id := range n.pushingCommits {
		n.pushingCommits[id].owerwrite()
		delete(n.pushingCommits, id)
	}
}

func (n *Node) Render(ctx context.Context, w io.Writer) error {
	n.lock()
	defer n.unlock()
	if n.mode == removed {
		return nil
	}
	if n.mode == static {
		return n.staticRender(ctx, w)
	}
	inst, instOk := ctx.Value(common.InstanceCtxKey).(instance)
	rm, rmOk := ctx.Value(common.RenderMapCtxKey).(*common.RenderMap)
	if !instOk || !rmOk {
		panic(errors.New("Node rendered outside doors context"))
	}
	var id uint64
	if n.core != nil {
		id = n.core.id
	} else {
		id = inst.NewId()
	}
	rw, ok := rm.Writer(id)
	if !ok {
		return common.RenderErrorLog(ctx, w, "Dynamic node rendered twice", slog.Uint64("node_id", id))
	}
	n.resetCommits()
	children := templ.GetChildren(ctx)
	ctx = templ.ClearChildren(ctx)
	n.suspended = false
	if n.core != nil {
		n.core.kill(true)
		if !n.core.isRoot() {
			n.core.parent.removeChild(n)
		}
	}
	n.core = newCore(n, ctx, inst, id)
	commit := n.commitBuffer
	n.commitBuffer = nil
	if !n.core.isRoot() {
		n.core.parent.addChild(n)
	}
	if w != nil {
		err := rw.Holdplace(w)
		if err != nil {
			return common.RenderErrorLog(ctx, w, err.Error(), slog.Uint64("node_id", id))
		}
	}
	n.core.render(ctx, rm, rw, children, n.content, commit)
	return nil
}

func (n *Node) staticRender(ctx context.Context, w io.Writer) error {
	if n.content == nil {
		return nil
	}
	err := n.content.Render(ctx, w)
	if err != nil {
		return common.RenderError(err.Error()).Render(ctx, w)
	}
	return nil

}

func (n *Node) lock() {
	n.mu.Lock()
	if n.pushingCommits == nil {
		n.pushingCommits = make(map[uint]*commit)
	}
}

func (n *Node) unlock() {
	n.mu.Unlock()
}
