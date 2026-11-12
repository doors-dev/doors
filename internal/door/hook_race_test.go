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

package door

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/gox"
)

type raceHook struct {
	t     *testing.T
	inst  *testInstance
	hook  *hook
	id    uint64
	gate  chan struct{}
	seen  []*chan struct{}
	busy  atomic.Int32
	clash atomic.Bool
	mu    sync.Mutex
	ran   []uint64
	codes sync.Map
	wg    sync.WaitGroup
}

func newRaceHook(t *testing.T, race core.Race) *raceHook {
	t.Helper()
	inst := newTestInstance(t)
	d := &Door{}
	renderTestPage(t, inst, gox.Elem(func(cur gox.Cursor) error {
		return cur.Comp(d)
	}))
	rh := &raceHook{t: t, inst: inst, gate: make(chan struct{})}
	reg, ok := d.node.Load().tracker.RegisterHook(func(ctx context.Context, _ http.ResponseWriter, r *http.Request) bool {
		track, _ := strconv.ParseUint(r.Header.Get("T"), 10, 64)
		if rh.busy.Add(1) > 1 {
			rh.clash.Store(true)
		}
		defer rh.busy.Add(-1)
		if track == 2 {
			<-rh.gate
		}
		rh.mu.Lock()
		rh.ran = append(rh.ran, track)
		rh.mu.Unlock()
		return false
	}, race, false, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	rh.id = reg.HookID
	inst.root.mu.Lock()
	rh.hook = inst.root.hooks[rh.id]
	inst.root.mu.Unlock()
	return rh
}

func (rh *raceHook) call(track uint64) int {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("T", strconv.FormatUint(track, 10))
	w := httptest.NewRecorder()
	if !rh.inst.root.TriggerHook(rh.id, w, r, track) {
		rh.t.Fatalf("hook %d not found", track)
	}
	return w.Code
}

func (rh *raceHook) async(track uint64) {
	rh.wg.Add(1)
	go func() {
		defer rh.wg.Done()
		rh.codes.Store(track, rh.call(track))
	}()
}

func (rh *raceHook) waitFor(cond func() bool) {
	rh.t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			rh.t.Fatal("condition not reached")
		}
		time.Sleep(time.Millisecond)
	}
}

func (rh *raceHook) entered() {
	rh.t.Helper()
	rh.waitFor(func() bool {
		tail := rh.hook.ch.Load()
		if tail == nil || slices.Contains(rh.seen, tail) {
			return false
		}
		rh.seen = append(rh.seen, tail)
		return true
	})
}

func (rh *raceHook) held() {
	rh.t.Helper()
	rh.async(2)
	rh.entered()
}

func (rh *raceHook) finish() []uint64 {
	close(rh.gate)
	rh.wg.Wait()
	if rh.clash.Load() {
		rh.t.Fatal("calls ran concurrently")
	}
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return rh.ran
}

func (rh *raceHook) code(track uint64) int {
	v, _ := rh.codes.Load(track)
	code, _ := v.(int)
	return code
}

func TestHookRaceQueued(t *testing.T) {
	cases := []struct {
		race core.Race
		ran  []uint64
		code int
	}{
		{core.RaceStrict, []uint64{2, 3, 4}, http.StatusOK},
		{core.RaceLatest, []uint64{2, 4}, http.StatusPreconditionFailed},
	}
	for _, tc := range cases {
		rh := newRaceHook(t, tc.race)
		rh.held()
		rh.async(3)
		rh.entered()
		rh.async(4)
		rh.entered()
		if ran := rh.finish(); !slices.Equal(ran, tc.ran) {
			t.Fatalf("race %d: ran %v, want %v", tc.race, ran, tc.ran)
		}
		if got := rh.code(3); got != tc.code {
			t.Fatalf("race %d: call 3 status %d, want %d", tc.race, got, tc.code)
		}
		if got := rh.code(4); got != http.StatusOK {
			t.Fatalf("race %d: call 4 status %d", tc.race, got)
		}
		if got := rh.call(3); got != http.StatusPreconditionFailed {
			t.Fatalf("race %d: stale call status %d", tc.race, got)
		}
		if got := rh.call(5); got != http.StatusOK {
			t.Fatalf("race %d: newer call status %d", tc.race, got)
		}
	}
}

func TestHookRaceOutOfOrder(t *testing.T) {
	for race, want := range map[core.Race][]uint64{core.RaceStrict: {2, 4, 5}, core.RaceLatest: {2, 5}} {
		rh := newRaceHook(t, race)
		rh.held()
		rh.async(4)
		rh.entered()
		rh.async(3)
		rh.entered()
		rh.async(5)
		rh.entered()
		if ran := rh.finish(); !slices.Equal(ran, want) {
			t.Fatalf("race %d: ran %v, want %v", race, ran, want)
		}
		if got := rh.code(3); got != http.StatusPreconditionFailed {
			t.Fatalf("race %d: late older call status %d", race, got)
		}
	}
}

func TestHookRaceSerialRunsAll(t *testing.T) {
	rh := newRaceHook(t, core.RaceSerial)
	rh.held()
	rh.async(1)
	if ran := rh.finish(); !slices.Equal(ran, []uint64{2, 1}) {
		t.Fatalf("serial ran %v", ran)
	}
	if got := rh.code(1); got != http.StatusOK {
		t.Fatalf("serial call status %d", got)
	}
}
