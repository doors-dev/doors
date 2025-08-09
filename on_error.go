package doors

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front"
)

type OnError = front.OnError
type errorAction = front.ErrorAction

type IndicateOnError struct {
	Duration  time.Duration
	Indicator []Indicator
}

func (o IndicateOnError) ErrorAction() *errorAction {
	return front.OnErrorIndicate(o.Duration, front.IntoIndicate(o.Indicator))
}

type CallOnError struct {
	Name string
	Meta any
}

func (o CallOnError) ErrorAction() *errorAction {
	b, err := common.MarshalJSON(o.Meta)
	if err != nil {
		slog.Error("Error call arg marshaling error", slog.String("call_name", o.Name), slog.String("json_error", err.Error()))
	}
	return front.OnErrorCall(o.Name, json.RawMessage(b))
}


func OnErrorIndicate(duration time.Duration, indicator []Indicator) []OnError {
	return []OnError{IndicateOnError{
		Duration: duration,
		Indicator: indicator,
	}}
}
func OnErrorCall(name string, meta any) []OnError {
	return []OnError{CallOnError{
		Name: name,
		Meta: meta,
	}}
}
