package instance

import (
	"github.com/doors-dev/doors/internal/common"
	"sync/atomic"
)

type setPathCaller struct {
	path     string
	replace  bool
	canceled atomic.Bool
}

func (t *setPathCaller) cancel() {
	t.canceled.Store(true)
}

func (t *setPathCaller) Data() *common.CallData {
	if t.canceled.Load() {
		return nil
	}
	return &common.CallData{
		Name:    "set_path",
		Arg:     t.arg(),
		Payload: common.WritableNone{},
	}
}

func (t *setPathCaller) arg() []any {
	return []any{t.path, t.replace}
}

func (t *setPathCaller) Result(error) {}

type LocatinReload struct {
}

func (l *LocatinReload) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_reload",
		Arg:     []any{},
		Payload: common.WritableNone{},
	}
}
func (t *LocatinReload) Result(error) {}

type LocationReplace struct {
	Href   string
	Origin bool
}

func (l *LocationReplace) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_replace",
		Arg:     []any{l.Href, l.Origin},
		Payload: common.WritableNone{},
	}
}
func (t *LocationReplace) Result(error) {}

type LocationAssign struct {
	Href   string
	Origin bool
}

func (l *LocationAssign) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_assign",
		Arg:     []any{l.Href, l.Origin},
		Payload: common.WritableNone{},
	}
}

func (t *LocationAssign) Result(error) {}
