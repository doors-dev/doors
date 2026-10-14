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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
	"github.com/doors-dev/doors/internal/solitaire/inner"
)

type stubRW struct {
	header   http.Header
	body     bytes.Buffer
	status   int
	flushes  int
	writeErr bool
}

func (s *stubRW) Header() http.Header {
	if s.header == nil {
		s.header = make(http.Header)
	}
	return s.header
}

func (s *stubRW) WriteHeader(status int) {
	s.status = status
}

func (s *stubRW) Write(data []byte) (int, error) {
	if s.writeErr {
		return 0, errors.New("write failed")
	}
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.body.Write(data)
}

func (s *stubRW) Flush() {
	s.flushes += 1
}

type stubSyncCall struct {
	act         actions.Action
	params      actions.CallParams
	cancelCount int
	resultCount int
	lastOutput  json.RawMessage
	lastErr     error
}

func (s *stubSyncCall) Params() actions.CallParams {
	return s.params
}

func (s *stubSyncCall) Action() (actions.Action, func(), bool) {
	if s.act == nil {
		return nil, nil, false
	}
	return s.act, func() {}, true
}

func (s *stubSyncCall) Cancel() {
	s.cancelCount += 1
}

func (s *stubSyncCall) Result(output json.RawMessage, err error) {
	s.resultCount += 1
	s.lastOutput = output
	s.lastErr = err
}

type stubExpireHandler struct {
	expired int
}

func (s *stubExpireHandler) Expire() {
	s.expired += 1
}

type stubInstance struct {
	syncErrors []error
	touches    int
}

func (s *stubInstance) SyncError(err error) {
	s.syncErrors = append(s.syncErrors, err)
}

func (s *stubInstance) Touch() {
	s.touches += 1
}

func testSolitaireConf() *common.SolitaireConf {
	conf := &common.Conf{
		SolitaireQueue:     8,
		SolitairePending:   4,
		SolitaireFrameSize: 4 * 1024,
		SolitaireFrameTime: 20 * time.Millisecond,
		SolitaireMaxRTT:    200 * time.Millisecond,
	}
	common.InitDefaults(conf)
	return common.GetSolitaireConf(conf)
}

func newTestWriteController(conf *common.SolitaireConf, w *stubRW) *writeController {
	return &writeController{
		Sync:    &frameSyncer{Conf: conf},
		Writer:  w,
		Flusher: w,
		Conf:    conf,
	}
}

func newInnerCard(seq uint64, call *stubSyncCall) *inner.Card {
	return inner.NewCard(seq, &inner.Call{
		Call:   call,
		Params: call.params,
	})
}

func TestWriteControllerStashAndSubmit(t *testing.T) {
	conf := testSolitaireConf()
	recorder := &stubRW{}
	fw := newTestWriteController(conf, recorder)

	issuedCall := &stubSyncCall{
		act: actions.Emit{
			Name:    "sync",
			DoorID:  7,
			Payload: actions.NewText("payload"),
		},
	}
	if got := fw.Stash(newInnerCard(5, issuedCall)); got != stashOk {
		t.Fatalf("expected issue stash result, got %v", got)
	}
	if fw.Len() != 1 {
		t.Fatalf("expected one stashed header, got %d", fw.Len())
	}
	if fw.frameStart.IsZero() {
		t.Fatal("expected issued card to start a frame")
	}
	if _, err := fw.Submit(signalSync); err != nil {
		t.Fatal(err)
	}
	out := recorder.body.Bytes()
	if !bytes.Contains(out, []byte{signalAction}) {
		t.Fatalf("expected action signal in output, got %v", out)
	}
	if !bytes.Contains(out, []byte("payload")) {
		t.Fatalf("expected action payload in output, got %q", out)
	}
	if !bytes.Contains(out, []byte{byte(signalSync)}) {
		t.Fatalf("expected sync frame in output, got %v", out)
	}
	if recorder.flushes != 1 {
		t.Fatalf("expected sync submit to flush once, got %d", recorder.flushes)
	}
	if fw.Len() != 0 {
		t.Fatalf("expected stashed headers to clear after submit, got %d", fw.Len())
	}

	noAction := &stubSyncCall{}
	cancelFrame := newTestWriteController(conf, &stubRW{})
	if got := cancelFrame.Stash(newInnerCard(6, noAction)); got != stashCancel {
		t.Fatalf("expected canceled stash result, got %v", got)
	}
	if cancelFrame.Len() != 0 {
		t.Fatalf("canceled card must not be stashed, got %d headers", cancelFrame.Len())
	}
	if !cancelFrame.frameStart.IsZero() {
		t.Fatal("canceled card must not start a frame")
	}
	if cancelFrame.buffer.Len() != 0 {
		t.Fatalf("canceled card must not write to frame buffer, got %d bytes", cancelFrame.buffer.Len())
	}
}

func TestWriteControllerSubmitRestoresOnWriteError(t *testing.T) {
	conf := testSolitaireConf()
	recorder := &stubRW{writeErr: true}
	fw := newTestWriteController(conf, recorder)
	call := &stubSyncCall{act: actions.Test{Arg: "retry"}}
	if got := fw.Stash(newInnerCard(1, call)); got != stashOk {
		t.Fatalf("expected issued card, got %v", got)
	}

	headers, err := fw.Submit(signalSync)
	if err == nil {
		t.Fatal("expected write error")
	}
	if len(headers) != 1 || headers[0].beg != 1 || headers[0].end != 1 {
		t.Fatalf("unexpected headers returned for restore: %#v", headers)
	}
	if fw.Len() != 0 {
		t.Fatalf("expected stashed headers to clear after failed submit, got %d", fw.Len())
	}
	if fw.buffer.Len() != 0 {
		t.Fatalf("expected frame buffer reset after failed submit, got %d bytes", fw.buffer.Len())
	}
}

// The link slot carries the door, then the dependency seq and door only when set; the batch slot is 0, the group, or group and batch.
func TestHeaderLink(t *testing.T) {
	invocation := actions.Test{Arg: 1}.Invocation()
	cases := []struct {
		h    header
		want string
	}{
		{header{beg: 4, end: 4, door: 7}, `[4,[7],0,["test",[1]],null]`},
		{header{beg: 4, end: 4, group: 5, door: 7, depSeq: 3}, `[4,[7,3],5,["test",[1]],null]`},
		{header{beg: 4, end: 4, group: 5, batch: 2, door: 7, depDoor: 9}, `[4,[7,0,9],[5,2],["test",[1]],null]`},
		{header{beg: 2, end: 4, batch: 2, door: 7, depSeq: 3, depDoor: 9}, `[[2,4],[7,3,9],[0,2],["test",[1]],null]`},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := c.h.writeCard(&buf, &invocation); err != nil {
			t.Fatal(err)
		}
		out := buf.Bytes()
		got := strings.TrimSpace(string(out[1 : len(out)-1]))
		if got != c.want {
			t.Fatalf("expected %s, got %s", c.want, got)
		}
	}
	var buf bytes.Buffer
	if err := (header{beg: 2, end: 4, group: 5, batch: 2, door: 7, depSeq: 3}).writeCard(&buf, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if got := strings.TrimSpace(string(out[1 : len(out)-1])); got != `[[2,4],null]` {
		t.Fatalf("expected a bare filler, got %s", got)
	}
}

func TestDeckInsertReturnsSeq(t *testing.T) {
	deck := newDeck(expirator.NewExpirator(&stubExpireHandler{}), testSolitaireConf())
	for want := uint64(1); want <= 3; want++ {
		seq, err := deck.Insert(&stubSyncCall{}, 0, 0, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if seq != want {
			t.Fatalf("expected seq %d, got %d", want, seq)
		}
	}
	deck.End()
	killed := &stubSyncCall{}
	seq, err := deck.Insert(killed, 0, 0, 0, 0, 0)
	if !errors.Is(err, errorKilled) {
		t.Fatalf("expected killed error, got %v", err)
	}
	if seq != 0 {
		t.Fatalf("expected seq 0 after end, got %d", seq)
	}
	if killed.cancelCount != 1 {
		t.Fatalf("expected killed call canceled once, got %d", killed.cancelCount)
	}
}

func TestDeckDumpIssuesCancelsAndWritesFillers(t *testing.T) {
	conf := testSolitaireConf()
	deck := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	optimistic := &stubSyncCall{
		act:    actions.Test{Arg: "optimistic"},
		params: actions.CallParams{Optimistic: true},
	}
	if _, err := deck.Insert(optimistic, 0, 0, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := deck.Dump(newTestWriteController(conf, &stubRW{})); err != nil {
		t.Fatal(err)
	}
	if deck.PendingCount() != 1 {
		t.Fatalf("expected issued action to become pending, got %d", deck.PendingCount())
	}
	if optimistic.resultCount != 1 {
		t.Fatalf("expected optimistic Written result once, got %d", optimistic.resultCount)
	}
	if string(optimistic.lastOutput) != "null" {
		t.Fatalf("unexpected optimistic output: %s", optimistic.lastOutput)
	}

	cancelDeck := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	canceled := &stubSyncCall{}
	if _, err := cancelDeck.Insert(canceled, 0, 0, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	fw := newTestWriteController(conf, &stubRW{})
	if err := cancelDeck.Dump(fw); err != nil {
		t.Fatal(err)
	}
	if canceled.cancelCount != 1 {
		t.Fatalf("expected canceled action once, got %d", canceled.cancelCount)
	}
	if cancelDeck.PendingCount() != 0 {
		t.Fatalf("canceled action must not become pending, got %d", cancelDeck.PendingCount())
	}
	if fw.Len() != 1 {
		t.Fatalf("expected deck to write one filler for canceled sequence, got %d", fw.Len())
	}
	if fw.stashed[0].beg != 1 || fw.stashed[0].end != 1 {
		t.Fatalf("unexpected filler header: %#v", fw.stashed[0])
	}
}

func TestReportParsingAndFrameSyncer(t *testing.T) {
	var okResult result
	if err := okResult.UnmarshalJSON([]byte(`[{"ok":true},null]`)); err != nil {
		t.Fatal(err)
	}
	if string(okResult.output) != `{"ok":true}` {
		t.Fatalf("unexpected ok result payload: %s", okResult.output)
	}

	var errResult result
	if err := errResult.UnmarshalJSON([]byte(`[null,"boom"]`)); err != nil {
		t.Fatal(err)
	}
	if errResult.err == nil || errResult.err.Error() != "boom" {
		t.Fatalf("unexpected result error: %v", errResult.err)
	}

	var one gap
	if err := one.UnmarshalJSON([]byte(`[3]`)); err != nil {
		t.Fatal(err)
	}
	if one.beg != 3 || one.end != 3 {
		t.Fatalf("unexpected single gap: %#v", one)
	}

	var many gap
	if err := many.UnmarshalJSON([]byte(`[3,5]`)); err != nil {
		t.Fatal(err)
	}
	if many.beg != 3 || many.end != 5 {
		t.Fatalf("unexpected ranged gap: %#v", many)
	}
	if err := (&gap{}).UnmarshalJSON([]byte(`[]`)); err == nil {
		t.Fatal("expected empty gap payload to fail")
	}

	var rep report
	if err := json.Unmarshal([]byte(`[7,123,{"2":[{"done":true},null],"3":[null,"bad"]},[[4,6],[8]]]`), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.ID != 7 || rep.TS != 123 {
		t.Fatalf("unexpected report header: %#v", rep)
	}
	if string(rep.Results[2].output) != `{"done":true}` {
		t.Fatalf("unexpected report result: %#v", rep.Results[2])
	}
	if rep.Results[3].err == nil || rep.Results[3].err.Error() != "bad" {
		t.Fatalf("unexpected report error: %#v", rep.Results[3])
	}
	if len(rep.Gaps) != 2 || rep.Gaps[0].beg != 4 || rep.Gaps[0].end != 6 || rep.Gaps[1].beg != 8 || rep.Gaps[1].end != 8 {
		t.Fatalf("unexpected report gaps: %#v", rep.Gaps)
	}

	conf := testSolitaireConf()
	syncer := &frameSyncer{Conf: conf}
	first := syncer.Collect()
	var firstTuple []json.RawMessage
	if err := json.Unmarshal(first, &firstTuple); err != nil {
		t.Fatal(err)
	}
	if len(firstTuple) != 1 {
		t.Fatalf("expected initial sync tuple without acks, got %s", first)
	}
	var ts int64
	if err := json.Unmarshal(firstTuple[0], &ts); err != nil {
		t.Fatal(err)
	}
	syncer.Report(11, ts)
	if rtt := syncer.RTT(); rtt < conf.FlushTime*2 || rtt > conf.MaxRTT {
		t.Fatalf("rtt outside configured bounds: %v", rtt)
	}
	second := syncer.Collect()
	var secondTuple []json.RawMessage
	if err := json.Unmarshal(second, &secondTuple); err != nil {
		t.Fatal(err)
	}
	if len(secondTuple) != 2 {
		t.Fatalf("expected sync tuple with acks, got %s", second)
	}
	var acks []uint64
	if err := json.Unmarshal(secondTuple[1], &acks); err != nil {
		t.Fatal(err)
	}
	if len(acks) != 1 || acks[0] != 11 {
		t.Fatalf("unexpected acks: %#v", acks)
	}
}

func TestSolitaireAndSenderHelpers(t *testing.T) {
	conf := testSolitaireConf()
	inst := &stubInstance{}
	s := NewSolitaire(inst, conf)

	queuedCall := &stubSyncCall{act: actions.Test{Arg: "queued"}}
	s.Call(queuedCall, 0, 0, 0, 0, 0)
	if s.deck.QueueLength() != 1 {
		t.Fatalf("expected solitaire call to queue work, got %d", s.deck.QueueLength())
	}

	s.Expire()
	if len(inst.syncErrors) == 0 || inst.syncErrors[0].Error() != "sync timeout" {
		t.Fatalf("unexpected expire errors: %#v", inst.syncErrors)
	}

	recorder := &stubRW{}
	ctx, cancel := context.WithCancelCause(context.Background())
	sender := newSender(inst, s.deck, newTestWriteController(conf, recorder), ctx, cancel, conf)
	if sender.Context() != ctx {
		t.Fatal("expected sender context to be returned unchanged")
	}
	cancel(context.Canceled)
	sender.handleCause()
	if !bytes.Contains(recorder.body.Bytes(), []byte{byte(signalRoll), terminator}) {
		t.Fatalf("expected roll signal to be written, got %v", recorder.body.Bytes())
	}

	endCtx, endCancel := context.WithCancelCause(context.Background())
	endSender := newSender(inst, s.deck, newTestWriteController(conf, &stubRW{}), endCtx, endCancel, conf)
	s.sender.Store(endSender)
	s.End(common.EndCauseSuspend)
	if !errors.Is(context.Cause(endCtx), common.EndCauseSuspend) {
		t.Fatalf("unexpected end cause: %v", context.Cause(endCtx))
	}
}

func TestSenderRunCleanupSetsCauseAfterSubmitError(t *testing.T) {
	conf := testSolitaireConf()
	inst := &stubInstance{}
	deck := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	fw := newTestWriteController(conf, &stubRW{writeErr: true})
	call := &stubSyncCall{act: actions.Test{Arg: "write-error"}}
	if got := fw.Stash(newInnerCard(1, call)); got != stashOk {
		t.Fatalf("expected issued card, got %v", got)
	}
	fw.frameStart = time.Now().Add(-conf.FlushTime)

	ctx, cancel := context.WithCancelCause(context.Background())
	sender := newSender(inst, deck, fw, ctx, cancel, conf)
	sender.Run()
	if !errors.Is(context.Cause(ctx), context.Canceled) {
		t.Fatalf("expected cleanup to cancel sender context after submit error, got %v", context.Cause(ctx))
	}
}

type blockingStasher struct {
	entered chan struct{}
	release chan struct{}
}

func (s *blockingStasher) Stash(*inner.Card) stashResult {
	s.entered <- struct{}{}
	<-s.release
	return stashOk
}

func (s *blockingStasher) Full() bool {
	return false
}

// Ending the deck while a dump is stashing a card cancels every call exactly
// once: the queued ones at once and the one being stashed when it returns.
func TestDeckEndDuringDump(t *testing.T) {
	conf := testSolitaireConf()
	deck := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	first := &stubSyncCall{act: actions.Test{Arg: "first"}}
	second := &stubSyncCall{act: actions.Test{Arg: "second"}}
	for _, c := range []*stubSyncCall{first, second} {
		if _, err := deck.Insert(c, 0, 0, 0, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	s := &blockingStasher{entered: make(chan struct{}, 1), release: make(chan struct{})}
	done := make(chan error, 1)
	go func() {
		done <- deck.Dump(s)
	}()
	<-s.entered
	deck.End()
	close(s.release)
	if err := <-done; !errors.Is(err, errorKilled) {
		t.Fatalf("expected the dump to stop with errorKilled, got %v", err)
	}
	if first.cancelCount != 1 || second.cancelCount != 1 {
		t.Fatalf("expected each call canceled once, got first=%d second=%d", first.cancelCount, second.cancelCount)
	}
	if first.resultCount != 0 {
		t.Fatalf("the stashed call must not be written after the deck ended, got %d results", first.resultCount)
	}
	if n := deck.PendingCount(); n != 0 {
		t.Fatalf("expected nothing pending after End, got %d", n)
	}
	deck.HeatUp()
	if n := deck.QueueLength(); n != 0 {
		t.Fatalf("expected an empty queue after End, got %d", n)
	}
}

// The delay before a probe grows while probes repeat without progress, stays within the maximum RTT, and resets on any progress.
func TestProbeBackoff(t *testing.T) {
	conf := testSolitaireConf()
	d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	fw := newTestWriteController(conf, &stubRW{})
	probe := func() {
		d.HeatUp()
		if err := d.Dump(fw); err != nil {
			t.Fatal(err)
		}
	}
	insert := func() uint64 {
		seq, err := d.Insert(&stubSyncCall{act: actions.Test{}}, 0, 0, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := d.Dump(fw); err != nil {
			t.Fatal(err)
		}
		return seq
	}
	expectReset := func(cause string) {
		t.Helper()
		if _, probes := d.Pending(); probes != 0 {
			t.Fatalf("expected %s to reset the back-off, got %d probes", cause, probes)
		}
	}
	first := insert()
	var prev time.Duration
	for i := range 6 {
		pending, probes := d.Pending()
		if !pending || probes != i {
			t.Fatalf("expected %d probes without progress, got %d (pending %v)", i, probes, pending)
		}
		delay := fw.probeDelay(probes)
		if i == 0 && delay != fw.Sync.RTT() {
			t.Fatalf("expected the first probe after one round trip, got %v", delay)
		}
		if delay > conf.MaxRTT {
			t.Fatalf("expected the delay within %v, got %v", conf.MaxRTT, delay)
		}
		if i > 0 && delay <= prev && prev != conf.MaxRTT {
			t.Fatalf("expected the delay to grow past %v, got %v", prev, delay)
		}
		prev = delay
		probe()
	}
	if prev != conf.MaxRTT {
		t.Fatalf("expected the delay to reach %v, got %v", conf.MaxRTT, prev)
	}
	second := insert()
	expectReset("a new call")
	probe()
	if err := d.CollectResults(map[uint64]result{first: {}}); err != nil {
		t.Fatal(err)
	}
	expectReset("a result")
	probe()
	if err := d.FillGaps([]gap{{beg: second, end: second}}); err != nil {
		t.Fatal(err)
	}
	expectReset("a reported gap")
	if delay := fw.probeDelay(0); delay != fw.Sync.RTT() {
		t.Fatalf("expected the delay back at one round trip, got %v", delay)
	}
}

// A sender with an unacknowledged call slows its probes down instead of repeating them every round trip.
func TestSenderProbesBackOff(t *testing.T) {
	conf := testSolitaireConf()
	inst := &stubInstance{}
	d := newDeck(expirator.NewExpirator(&stubExpireHandler{}), conf)
	rw := &stubRW{}
	ctx, cancel := context.WithCancelCause(context.Background())
	s := newSender(inst, d, newTestWriteController(conf, rw), ctx, cancel, conf)
	if _, err := d.Insert(&stubSyncCall{act: actions.Test{}, params: actions.CallParams{Timeout: time.Minute}}, 0, 0, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	go s.Run()
	time.Sleep(1200 * time.Millisecond)
	<-s.Roll()
	if probes := rw.flushes - 1; probes < 2 || probes > 10 {
		t.Fatalf("expected a few backed-off probes, got %d", probes)
	}
}
