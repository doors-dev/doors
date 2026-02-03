// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package action

import (
	"encoding/json"
	"net/http"
	"time"

)

type Actions []Action

func (a Actions) invocations() []Invocation {
	inv := make([]Invocation, len(a))
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
	h.Set("D0-After", string(b))
	return nil
}

type Action interface {
	Invocation() Invocation
	Log() string
}

type CallParams struct {
	Optimistic bool
	Timeout    time.Duration
}

type Call interface {
	// Clean()
	Params() CallParams
	Action() (Action, bool)
	Payload() []byte
	Cancel()
	Result(json.RawMessage, error)
}

type Invocation struct {
	name string
	arg  []any
}

func (a Invocation) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{a.name, a.arg})
}
