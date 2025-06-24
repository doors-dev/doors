package instance

import (
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
	"github.com/doors-dev/doors/internal/path"
)

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
func (t *setPathCaller) Payload() (common.Writable, bool) {
	return nil, false
}
func (t *setPathCaller) Name() string {
	return "set_path"
}
func (t *setPathCaller) OnResult(error) {}

func (t *setPathCaller) Arg() common.JsonWritable {
	return common.JsonWritables([]common.JsonWritable{common.JsonWritableAny{t.path}, common.JsonWritableAny{t.replace}})
}
func (t *setPathCaller) OnWriteErr() bool {
	return !t.canceled.Load()
}

type relocateCaller struct {
	location *path.Location
}

func (t *relocateCaller) Call() (node.Call, bool) {
	return t, true
}
func (t *relocateCaller) Name() string {
	return "relocate"
}
func (t *relocateCaller) Payload() (common.Writable, bool) {
	return nil, false
}

func (t *relocateCaller) Arg() common.JsonWritable {
	return common.JsonWritables([]common.JsonWritable{common.JsonWritableAny{t.location.String()}})
}
func (t *relocateCaller) OnWriteErr() bool    {
	return true
}
func (t *relocateCaller) OnResult(error) {}
