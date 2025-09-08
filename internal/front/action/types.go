// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package action

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/doors-dev/doors/internal/common"
)

type Actions []Action

func (a Actions) invocations() []*Invocation {
	inv := make([]*Invocation, len(a))
	for i, a := range a {
		inv[i] = a.Invocation()
	}
	return inv
}

func (a Actions) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.invocations())
}

func (a Actions) Set(h http.Header) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	h.Set("D00r-After", string(b))
	return nil
}

type Action interface {
	Invocation() *Invocation
	Log() string
}

type CallParams struct {
	Optimistic bool
	Timeout    time.Duration
}

type Call interface {
	Clean()
	Params() CallParams
	Action() (Action, bool)
	Payload() common.Writable
	Cancel()
	Result(json.RawMessage, error)
}

type Invocation struct {
	name string
	arg  []any
}

func (a *Invocation) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{a.name, a.arg})
}
