package node

import (
	"log/slog"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
)

type Caller interface {
	Call() (Call, bool)
}

type Call interface {
	Name() string
	Args() []common.Writable
	OnWriteErr()
	OnResult(error)
}

type nodeCaller Node

func (n *nodeCaller) Call() (Call, bool) {
	call, ok := <-(*Node)(n).call()
	return call, ok
}

type commitCall struct {
	name string
	args []common.Writable
	id   uint
	node *Node
}

func (c *commitCall) Name() string {
	return c.name
}

func (c *commitCall) Args() []common.Writable {
	return c.args
}

func (c *commitCall) OnResult(err error) {
	c.node.commitResult(c.id, err)
	for _, writable := range c.args {
		writable.Destroy()
	}
}

func (c *commitCall) OnWriteErr() {
	c.node.commitWriteErr(c)
}

type jsCall struct {
	name      string
	arg       common.WritableRaw
	core      *core
	hookEntry *HookEntry
	done      atomic.Bool
}

func (j *jsCall) kill() {
	if j.done.Swap(true) {
		return
	}
}

func (j *jsCall) cancel() bool {
	if j.done.Swap(true) {
		return false
	}
	j.core.removeJsCall(j)
	j.hookEntry.Cancel()
    return true
}

func (j *jsCall) Call() (Call, bool) {
	if j.done.Load() {
		return nil, false
	}
	return j, true
}

func (j *jsCall) Name() string {
	return "call"
}

func (j *jsCall) Args() []common.Writable {
	return []common.Writable{common.WritableAny{j.name}, j.arg, common.WritableAny{j.hookEntry.NodeId}, common.WritableAny{j.hookEntry.HookId}}
}

func (j *jsCall) OnResult(err error) {
	if j.done.Swap(true) {
		return
	}
	j.core.removeJsCall(j)
	if err == nil {
		return
	}
	slog.Error("Call failed", slog.String("call_name", j.name), slog.String("js_error", err.Error()))
    j.hookEntry.cancel(err)

}

func (j *jsCall) OnWriteErr() {
}
