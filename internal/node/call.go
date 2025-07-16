package node

import (

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
