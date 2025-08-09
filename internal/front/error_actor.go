package front

import (
	"encoding/json"
	"time"
)

type OnError interface {
	ErrorAction() *ErrorAction
}

type ErrorAction struct {
	kind string
	args []any
}

func (ea *ErrorAction) MarshalJSON() ([]byte, error) {
	a := []any{ea.kind, ea.args}
	return json.Marshal(a)
}

func IntoErrorAction(errorActor []OnError) []*ErrorAction {
	a := make([]*ErrorAction, len(errorActor))
	for i, s := range errorActor {
		a[i] = s.ErrorAction()
	}
	return a
}

func OnErrorCall(name string, meta json.RawMessage) *ErrorAction {
	return &ErrorAction{
		kind: "call",
		args: []any{name, meta},
	}
}

func OnErrorIndicate(d time.Duration, i []*Indicate) *ErrorAction {
	return &ErrorAction{
		kind: "indicator",
		args: []any{d.Milliseconds(), i},
	}
}

