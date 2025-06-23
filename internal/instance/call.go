package instance

import (
	"encoding/json"
	"errors"
	"io"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
	"github.com/doors-dev/doors/internal/path"
)

type CallResponse struct {
	Ack   uint           `json:"ack"`
	Ready []*ReadyResult `json:"ready"`
}

type ReadyResult struct {
	seq uint
	err error
}

func (m *ReadyResult) UnmarshalJSON(data []byte) error {
	var parts []json.RawMessage
	err := json.Unmarshal(data, &parts)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return errors.New("empty result array")
	}
	err = json.Unmarshal(parts[0], &m.seq)
	if err != nil {
		return err
	}
	if len(parts) > 1 {
		var message string
		err = json.Unmarshal(parts[1], &message)
		if err != nil {
			return err
		}
		m.err = errors.New(message)
		return nil
	}
	return nil
}

type call struct {
	link   *connector
	caller node.Caller
	seq    uint
	call   node.Call
}

func (c *call) Name() string {
	return "call"
}

func (c *call) result(err error) {
	c.call.OnResult(err)
}

func (c *call) writeErr() {
	c.call.OnWriteErr()
}

func (c *call) WriteData(w io.Writer) error {
	call, ok := c.caller.Call()
	if !ok {
		c.link.cancelCall(c.seq)
		return nil
	}
	c.call = call
	data := []common.Writable{common.WritableAny{c.seq}, common.WritableAny{call.Name()}, common.Writables(call.Args())}
	err := common.Writables(data).WriteJson(w)
	if err != nil {
		c.link.onWrite(c.seq, false)
		return err
	}
	c.link.onWrite(c.seq, true)
	return nil
}

type touchCaller struct {
	fired bool
}

func (t *touchCaller) Call() (node.Call, bool) {
	if t.fired {
		return nil, false
	}
	t.fired = true
	return t, true
}
func (t *touchCaller) Name() string {
	return "touch"
}

func (t *touchCaller) Args() []common.Writable {
	return make([]common.Writable, 0)
}
func (t *touchCaller) OnWriteErr()    {}
func (t *touchCaller) OnResult(error) {}

type setPathCaller struct {
	path     string
	replace  bool
	canceled atomic.Bool
}

func (t *setPathCaller) cancel() {
	t.canceled.Store(true)
}

func (t *setPathCaller) Call() (node.Call, bool) {
	if t.canceled.Load() {
		return nil, false
	}
	return t, true
}
func (t *setPathCaller) Name() string {
	return "set_path"
}
func (t *setPathCaller) OnResult(error) {}

func (t *setPathCaller) Args() []common.Writable {
	return []common.Writable{common.WritableAny{t.path}, common.WritableAny{t.replace}}
}
func (t *setPathCaller) OnWriteErr() {}

type relocateCaller struct {
	location *path.Location
}

func (t *relocateCaller) Call() (node.Call, bool) {
	return t, true
}
func (t *relocateCaller) Name() string {
	return "relocate"
}

func (t *relocateCaller) Args() []common.Writable {
	return []common.Writable{common.WritableAny{t.location.String()}}
}
func (t *relocateCaller) OnWriteErr()    {}
func (t *relocateCaller) OnResult(error) {}
