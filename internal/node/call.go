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
	Arg() common.JsonWritable
	Payload() (common.Writable, bool)
	OnWriteErr() bool
	OnResult(error)
}

type nodeCaller Node

func (n *nodeCaller) Call() (Call, bool) {
	call, ok := <-(*Node)(n).call()
	return call, ok
}

type commitCall struct {
	name    string
	arg     common.JsonWritable
	payload common.Writable
	id      uint
	node    *Node
}

func (c *commitCall) Name() string {
	return c.name
}

func (c *commitCall) Arg() common.JsonWritable {
	return c.arg
}

func (c *commitCall) Payload() (common.Writable, bool) {
	if c.payload == nil {
		return nil, false
	}
	return c.payload, true
}

func (c *commitCall) OnResult(err error) {
	c.node.commitResult(c.id, err)
	if c.payload == nil {
		return
	}
	c.payload.Destroy()
}

func (c *commitCall) OnWriteErr() bool {
	return c.node.commitWriteErr(c)
}

type jsCall struct {
	name      string
	arg       common.JsonWritabeRaw
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

func (j *jsCall) Arg() common.JsonWritable {
	return common.JsonWritables([]common.JsonWritable{common.JsonWritableAny{j.name}, j.arg, common.JsonWritableAny{j.hookEntry.NodeId}, common.JsonWritableAny{j.hookEntry.HookId}})
}

func (k *jsCall) Payload() (common.Writable, bool) {
	return nil, false
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

func (j *jsCall) OnWriteErr() bool {
	return !j.done.Load()
}
