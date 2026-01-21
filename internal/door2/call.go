package door2

import (
	"encoding/json"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
)

type reportHook uint64

func (c reportHook) Params() action.CallParams {
	return action.CallParams{}
}

func (c reportHook) Action() (action.Action, bool) {
	return &action.ReportHook{HookId: uint64(c)}, true
}

func (C reportHook) Payload() common.Writable {
	return common.WritableNone{}
}

func (c reportHook) Cancel()                             {}
func (c reportHook) Result(r json.RawMessage, err error) {}
func (c reportHook) Clean()                              {}

