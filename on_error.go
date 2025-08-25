// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

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
