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

package utils

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/actions"
)

type tabStateInstance struct {
	core.Instance
	calls chan map[string]json.RawMessage
}

func (i *tabStateInstance) UserCallCheck(check func() bool, action actions.Action, _ func(json.RawMessage, error), _ func(), _ actions.CallParams) {
	if check != nil && !check() {
		return
	}
	i.calls <- action.(actions.UpdateState).State
}

func newTabStateTest() (TabStateManager, *tabStateInstance) {
	inst := &tabStateInstance{calls: make(chan map[string]json.RawMessage, 16)}
	return NewTabStateManager(inst, context.Background()), inst
}

func intEqual(a, b int) bool { return a == b }

func ptr(v int) *int { return &v }

func raw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func expectValue(t *testing.T, got *int, want *int) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("expected nil, got %d", *got)
		}
		return
	}
	if got == nil {
		t.Fatalf("expected %d, got nil", *want)
	}
	if *got != *want {
		t.Fatalf("expected %d, got %d", *want, *got)
	}
}

func expectPush(t *testing.T, inst *tabStateInstance, want map[string]any) {
	t.Helper()
	select {
	case got := <-inst.calls:
		if len(got) != len(want) {
			t.Fatalf("expected push %v, got %v", want, got)
		}
		for key, value := range want {
			gotValue, ok := got[key]
			if !ok {
				t.Fatalf("expected push %v, got %v", want, got)
			}
			var decoded any
			if err := json.Unmarshal(gotValue, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != value {
				t.Fatalf("expected push %v, got %v", want, got)
			}
		}
	case <-time.After(time.Second):
		t.Fatalf("expected push %v, got none", want)
	}
}

func expectNoPush(t *testing.T, inst *tabStateInstance) {
	t.Helper()
	select {
	case got := <-inst.calls:
		t.Fatalf("expected no push, got %v", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestTabStateNilBeforeInit(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	expectValue(t, n.Get(), nil)
	m.Initialize(nil)
	expectValue(t, n.Get(), ptr(0))
	expectNoPush(t, inst)
}

func TestTabStateInitFromClient(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	m.Initialize(map[string]json.RawMessage{"n": raw(t, 7)})
	expectValue(t, n.Get(), ptr(7))
	expectNoPush(t, inst)
	m.Initialize(map[string]json.RawMessage{"n": raw(t, 9)})
	expectValue(t, n.Get(), ptr(7))
	expectNoPush(t, inst)
}

func TestTabStateUpdateBeforeInit(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	<-n.Update(context.Background(), ptr(5))
	expectValue(t, n.Get(), ptr(5))
	<-n.Update(context.Background(), nil)
	expectValue(t, n.Get(), nil)
	m.Initialize(nil)
	expectValue(t, n.Get(), ptr(0))
	expectNoPush(t, inst)
}

func TestTabStateServerWinsOnInit(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	<-n.Update(context.Background(), ptr(5))
	m.Initialize(map[string]json.RawMessage{"n": raw(t, 7), "other": raw(t, true)})
	expectValue(t, n.Get(), ptr(5))
	expectPush(t, inst, map[string]any{"n": float64(5), "other": true})
	<-n.Update(context.Background(), ptr(6))
	expectPush(t, inst, map[string]any{"n": float64(6), "other": true})
	<-n.Update(context.Background(), nil)
	expectValue(t, n.Get(), ptr(0))
	expectPush(t, inst, map[string]any{"other": true})
}

func TestTabStateServerValueWithEmptyClientPushes(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	<-n.Update(context.Background(), ptr(5))
	m.Initialize(nil)
	expectValue(t, n.Get(), ptr(5))
	expectPush(t, inst, map[string]any{"n": float64(5)})
}

func TestTabStateUpdateAfterInitPushes(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	m.Initialize(nil)
	expectNoPush(t, inst)
	<-n.Update(context.Background(), ptr(1))
	expectPush(t, inst, map[string]any{"n": float64(1)})
	<-n.Update(context.Background(), ptr(1))
	expectNoPush(t, inst)
	<-n.Update(context.Background(), nil)
	expectValue(t, n.Get(), ptr(0))
	expectPush(t, inst, map[string]any{})
}

func TestTabStateBadStoredValue(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	m.Initialize(map[string]json.RawMessage{"n": raw(t, "x")})
	expectValue(t, n.Get(), ptr(0))
	expectNoPush(t, inst)
}

func TestTabStateNilBeforeInitYieldsToClient(t *testing.T) {
	m, inst := newTabStateTest()
	n := m.Derive("n", intEqual)
	<-n.Update(context.Background(), ptr(5))
	<-n.Update(context.Background(), nil)
	expectValue(t, n.Get(), nil)
	m.Initialize(map[string]json.RawMessage{"n": raw(t, 7)})
	expectValue(t, n.Get(), ptr(7))
	expectNoPush(t, inst)
}
