// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Params() CallParams
	Action() (Action, bool)
	Cancel()
	Result(json.RawMessage, error)
}

type Invocation struct {
	name    string
	arg     []any
	payload Payload
}

func (a Invocation) Payload() Payload {
	return a.payload
}

func (a Invocation) Func() []any {
	return []any{a.name, a.arg}
}

func (a Invocation) MarshalJSON() ([]byte, error) {
	if a.payload.IsNone() {
		return json.Marshal([]any{a.name, a.arg})
	}
	return json.Marshal([]any{a.name, a.arg, a.payload})
}
