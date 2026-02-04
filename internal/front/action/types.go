// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package action

import (
	"encoding/base64"
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

type PayloadType int

const (
	PayloadNone     PayloadType = 0x00
	PayloadBinary   PayloadType = 0x01
	PayloadJSON     PayloadType = 0x02
	PayloadText     PayloadType = 0x03
	PayloadBinaryGZ PayloadType = 0x11
	PayloadJSONGZ   PayloadType = 0x12
	PayloadTextGZ   PayloadType = 0x13
)

type Call interface {
	// Clean()
	Params() CallParams
	Action() (Action, bool)
	Cancel()
	Result(json.RawMessage, error)
}

type Invocation struct {
	name        string
	arg         []any
	payload     []byte
	payloadType PayloadType
}

func (a Invocation) Payload() ([]byte, PayloadType) {
	return a.payload, a.payloadType
}

func (a Invocation) Func() []any {
	return []any{a.name, a.arg}
}

func (a Invocation) MarshalJSON() ([]byte, error) {
	if a.payloadType == PayloadNone {
		return json.Marshal([]any{a.name, a.arg})
	}
	encoded := base64.StdEncoding.EncodeToString(a.payload)
	return json.Marshal([]any{a.name, a.arg, []any{a.payloadType, encoded}})
}
