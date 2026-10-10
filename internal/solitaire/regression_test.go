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

package solitaire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
	"github.com/doors-dev/doors/internal/solitaire/inner"
)

type countCall struct {
	act     actions.Action
	params  actions.CallParams
	panics  bool
	done    atomic.Bool
	cancels atomic.Int32
	results atomic.Int32
}

func (c *countCall) Params() actions.CallParams {
	return c.params
}

func (c *countCall) Action() (actions.Action, func(), bool) {
	if c.act == nil || c.done.Load() {
		return nil, nil, false
	}
	return c.act, func() {}, true
}

func (c *countCall) Cancel() {
	c.done.Store(true)
	c.cancels.Add(1)
}

func (c *countCall) Result(json.RawMessage, error) {
	c.done.Store(true)
	c.results.Add(1)
	if c.panics {
		panic("result callback")
	}
}

func (c *countCall) outcomes() int32 {
	return c.cancels.Load() + c.results.Load()
}

type recordStasher struct {
	actions []actions.Action
	fillers int
	full    bool
}

func (s *recordStasher) Stash(card *inner.Card) stashResult {
	if card.IsFiller() {
		s.fillers += 1
		return stashFiller
	}
	action, free, ok := card.Call.Action()
	if !ok {
		return stashCancel
	}
	free()
	s.actions = append(s.actions, action)
	return stashOk
}

func (s *recordStasher) Full() bool {
	return s.full && len(s.actions)+s.fillers > 0
}

type gateStasher struct {
	recordStasher
	gated   bool
	entered chan struct{}
	release chan struct{}
}

func (s *gateStasher) Stash(card *inner.Card) stashResult {
	if !s.gated {
		s.gated = true
		s.entered <- struct{}{}
		<-s.release
	}
	return s.recordStasher.Stash(card)
}

func stashRestored(t *testing.T, full bool) (*deck, *countCall, *gateStasher, chan error) {
	t.Helper()
	d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), testSolitaireConf())
	c := &countCall{act: actions.Test{Arg: "a"}, params: actions.CallParams{Timeout: time.Minute}}
	if err := d.Insert(c); err != nil {
		t.Fatal(err)
	}
	if err := d.Dump(&recordStasher{}); err != nil {
		t.Fatal(err)
	}
	if err := d.FillGaps([]gap{{1, 1}}); err != nil {
		t.Fatal(err)
	}
	g := &gateStasher{recordStasher: recordStasher{full: full}, entered: make(chan struct{}), release: make(chan struct{})}
	done := make(chan error, 1)
	go func() {
		done <- d.Dump(g)
	}()
	<-g.entered
	return d, c, g, done
}

func noPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", name, r)
		}
	}()
	f()
}

// A result for a re-sent call that arrives while the sender is writing that
// call is delivered, and the call does not stay pending.
func TestDeckResultWhileStashing(t *testing.T) {
	d, c, g, done := stashRestored(t, true)
	if err := d.CollectResults(map[uint64]result{1: {output: json.RawMessage("1")}}); err != nil {
		t.Fatal(err)
	}
	close(g.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if c.results.Load() != 1 {
		t.Fatalf("expected the result delivered, got %d results", c.results.Load())
	}
	if n := d.PendingCount(); n != 0 {
		t.Fatalf("expected nothing pending, got %d", n)
	}
	d.End()
	if c.outcomes() != 1 {
		t.Fatalf("expected one outcome, got cancels=%d results=%d", c.cancels.Load(), c.results.Load())
	}
}

// A gap report for a call the sender is writing restores the call instead of
// queuing a filler over it, so later gap reports and restores keep working.
func TestDeckGapWhileStashing(t *testing.T) {
	for _, next := range []struct {
		name string
		f    func(d *deck) error
	}{
		{"gap report", func(d *deck) error { return d.FillGaps([]gap{{1, 1}}) }},
		{"restore", func(d *deck) error { return d.Restore(1, 1) }},
	} {
		t.Run(next.name, func(t *testing.T) {
			d, c, g, done := stashRestored(t, true)
			if err := d.FillGaps([]gap{{1, 1}}); err != nil {
				t.Fatal(err)
			}
			close(g.release)
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			noPanic(t, next.name, func() {
				if err := next.f(d); err != nil {
					t.Fatal(err)
				}
			})
			s := &recordStasher{}
			if err := d.Dump(s); err != nil {
				t.Fatal(err)
			}
			if s.fillers != 0 {
				t.Fatalf("expected no filler over the issued call, got %d", s.fillers)
			}
			d.End()
			if c.outcomes() != 1 {
				t.Fatalf("expected one outcome, got cancels=%d results=%d", c.cancels.Load(), c.results.Load())
			}
		})
	}
}

// An optimistic call is re-sent with its action after the browser reports it
// lost, although it was reported delivered when first written.
func TestDeckOptimisticResent(t *testing.T) {
	d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), testSolitaireConf())
	c := &countCall{act: actions.Test{Arg: "navigate"}, params: actions.CallParams{Optimistic: true, Timeout: time.Minute}}
	if err := d.Insert(c); err != nil {
		t.Fatal(err)
	}
	if err := d.Dump(&recordStasher{}); err != nil {
		t.Fatal(err)
	}
	if c.results.Load() != 1 {
		t.Fatalf("expected the optimistic call reported once written, got %d results", c.results.Load())
	}
	if err := d.FillGaps([]gap{{1, 1}}); err != nil {
		t.Fatal(err)
	}
	s := &recordStasher{}
	if err := d.Dump(s); err != nil {
		t.Fatal(err)
	}
	if len(s.actions) != 1 || s.actions[0] != c.act {
		t.Fatalf("expected the action re-sent, got actions=%v fillers=%d", s.actions, s.fillers)
	}
	if c.outcomes() != 1 {
		t.Fatalf("expected one outcome, got cancels=%d results=%d", c.cancels.Load(), c.results.Load())
	}
}

// A gap the browser reports for a seq it already has results past is stale and
// ignored instead of ending the instance.
func TestDeckStaleGapIgnored(t *testing.T) {
	d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), testSolitaireConf())
	calls := make([]*countCall, 7)
	for i := range calls {
		calls[i] = &countCall{params: actions.CallParams{Timeout: time.Minute}}
		if i == 0 || i == 1 || i == 6 {
			calls[i].act = actions.Test{Arg: i}
		}
		if err := d.Insert(calls[i]); err != nil {
			t.Fatal(err)
		}
	}
	for d.QueueLength() > 0 {
		if err := d.Dump(&recordStasher{}); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.FillGaps([]gap{{2, 2}}); err != nil {
		t.Fatal(err)
	}
	if err := d.Dump(&recordStasher{}); err != nil {
		t.Fatal(err)
	}
	ok := json.RawMessage("1")
	if err := d.CollectResults(map[uint64]result{1: {output: ok}, 2: {output: ok}, 7: {output: ok}}); err != nil {
		t.Fatal(err)
	}
	if err := d.FillGaps([]gap{{2, 2}}); err != nil {
		t.Fatalf("expected the stale gap ignored, got %v", err)
	}
	if n := d.PendingCount(); n != 0 {
		t.Fatalf("expected nothing pending, got %d", n)
	}
}

// A panic in one call's result callback does not keep the other calls of the
// same report from getting their results.
func TestDeckResultPanicKeepsBatch(t *testing.T) {
	for range 20 {
		d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), testSolitaireConf())
		calls := []*countCall{{act: actions.Test{Arg: "panic"}, panics: true}}
		for range 3 {
			calls = append(calls, &countCall{act: actions.Test{Arg: "ok"}})
		}
		for _, c := range calls {
			c.params.Timeout = time.Minute
			if err := d.Insert(c); err != nil {
				t.Fatal(err)
			}
		}
		if err := d.Dump(&recordStasher{}); err != nil && !errors.Is(err, errorLimit) {
			t.Fatal(err)
		}
		results := map[uint64]result{}
		for i := range calls {
			results[uint64(i+1)] = result{output: json.RawMessage("true")}
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected the callback panic to be re-raised")
				}
			}()
			d.CollectResults(results)
		}()
		for i, c := range calls {
			if c.results.Load() != 1 {
				t.Fatalf("call %d: expected its result, got %d", i+1, c.results.Load())
			}
		}
	}
}

// The queue limit ends the instance only when exceeded.
func TestDeckQueueLimit(t *testing.T) {
	conf := testSolitaireConf()
	d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	for i := range conf.Queue {
		if err := d.Insert(&countCall{act: actions.Test{Arg: i}, params: actions.CallParams{Timeout: time.Minute}}); err != nil {
			t.Fatalf("insert %d of %d: %v", i+1, conf.Queue, err)
		}
	}
	if err := d.Insert(&countCall{act: actions.Test{Arg: "over"}, params: actions.CallParams{Timeout: time.Minute}}); err == nil {
		t.Fatal("expected an error past the queue limit")
	}
}

// A sync stream whose request context was canceled with a cause foreign to
// Doors ends as a roll instead of panicking.
func TestSenderForeignCancelCause(t *testing.T) {
	s := NewSolitaire(&stubInstance{}, testSolitaireConf())
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errors.New("server shutdown"))
	noPanic(t, "connect", func() {
		s.Connect(&stubRW{}, httptest.NewRequest(http.MethodGet, "/s/x", nil).WithContext(ctx))
	})
}

type startedWriter struct {
	http.ResponseWriter
	started chan struct{}
	once    atomic.Bool
}

func (w *startedWriter) Write(p []byte) (int, error) {
	if !w.once.Swap(true) {
		close(w.started)
	}
	return w.ResponseWriter.Write(p)
}

func (w *startedWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *startedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// A new sync stream takes over even when the previous one is stuck writing to a
// client that stopped reading.
func TestRollPastBlockedWrite(t *testing.T) {
	conf := testSolitaireConf()
	conf.Roll = 600 * time.Millisecond
	s := NewSolitaire(&stubInstance{}, conf)
	started := make(chan struct{})
	var returned atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("t") == "stuck" {
			w = &startedWriter{ResponseWriter: w, started: started}
		}
		s.Connect(w, r)
		if r.URL.Query().Get("t") == "next" {
			returned.Store(true)
		}
	}))
	defer srv.Close()
	s.Call(&countCall{act: actions.Test{Arg: strings.Repeat("x", 16<<20)}, params: actions.CallParams{Timeout: time.Minute}})
	stuck, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer stuck.Close()
	fmt.Fprintf(stuck, "GET /s/x?t=stuck HTTP/1.1\r\nHost: x\r\n\r\n")
	<-started
	client := &http.Client{Timeout: time.Second}
	if resp, err := client.Get(srv.URL + "/s/x?t=next"); err == nil {
		resp.Body.Close()
	}
	deadline := time.Now().Add(3 * time.Second)
	for !returned.Load() {
		if time.Now().After(deadline) {
			t.Fatal("the new stream stayed blocked behind the stuck one")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
