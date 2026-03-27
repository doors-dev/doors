package solitaire

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
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
	act         action.Action
	params      action.CallParams
	cancelCount int
	resultCount int
	lastOutput  json.RawMessage
	lastErr     error
}

func (s *stubSyncCall) Params() action.CallParams {
	return s.params
}

func (s *stubSyncCall) Action() (action.Action, bool) {
	if s.act == nil {
		return nil, false
	}
	return s.act, true
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
	conf := &common.SystemConf{}
	common.InitDefaults(conf)
	return common.GetSolitaireConf(conf)
}

func TestWriterHelpersAndDecoders(t *testing.T) {
	recorder := &stubRW{}
	w := &writer{
		sizeLimit: 2,
		timeLimit: time.Hour,
		w:         recorder,
		f:         recorder,
	}

	flushed := 0
	w.AfterFlush(func() {
		flushed += 1
	})
	if err := w.WriteAck(); err != nil {
		t.Fatal(err)
	}
	if recorder.flushes != 1 {
		t.Fatalf("expected ack flush, got %d", recorder.flushes)
	}

	if _, err := w.Write([]byte("ab")); err != nil {
		t.Fatal(err)
	}
	if !w.toFlush {
		t.Fatal("expected size limit to request flush")
	}
	w.TryFlush()
	if flushed != 1 {
		t.Fatalf("expected after-flush hook once, got %d", flushed)
	}
	if recorder.flushes != 2 {
		t.Fatalf("unexpected flush count after TryFlush: %d", recorder.flushes)
	}

	timeRecorder := &stubRW{}
	timeWriter := &writer{
		sizeLimit:   100,
		timeLimit:   time.Millisecond,
		lastFlushed: time.Now().Add(-time.Second),
		w:           timeRecorder,
		f:           timeRecorder,
	}
	if _, err := timeWriter.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if !timeWriter.toFlush {
		t.Fatal("expected time limit to request flush")
	}

	errorRecorder := &stubRW{writeErr: true}
	errorWriter := &writer{w: errorRecorder, f: errorRecorder}
	if err := errorWriter.WriteAck(); !errors.Is(err, writerError) {
		t.Fatalf("expected writerError from WriteAck, got %v", err)
	}
	if _, err := errorWriter.Write([]byte("x")); !errors.Is(err, writerError) {
		t.Fatalf("expected writerError from Write, got %v", err)
	}

	single := newHeader(5, 5)
	if len(single) != 1 {
		t.Fatalf("unexpected single header: %#v", single)
	}
	ranged := newHeader(2, 4)
	if len(ranged) != 1 || len(ranged[0].([]uint64)) != 2 {
		t.Fatalf("unexpected ranged header: %#v", ranged)
	}

	var filler bytes.Buffer
	if err := single.writeFiller(&filler); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(filler.Bytes(), terminator) {
		t.Fatal("expected filler output to include terminator")
	}

	var issued bytes.Buffer
	invocation := action.Emit{Name: "sync", DoorID: 7, Payload: action.NewText("payload")}.Invocation()
	if err := (&issuedCall{invocation: invocation}).write(single, &issued); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(issued.Bytes(), []byte("payload")) {
		t.Fatalf("expected issued call payload in output, got %q", issued.Bytes())
	}

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
	if one.start != 3 || one.end != 3 {
		t.Fatalf("unexpected single gap: %#v", one)
	}

	var many gap
	if err := many.UnmarshalJSON([]byte(`[3,5]`)); err != nil {
		t.Fatal(err)
	}
	if many.start != 3 || many.end != 5 {
		t.Fatalf("unexpected ranged gap: %#v", many)
	}

	if err := (&gap{}).UnmarshalJSON([]byte(`[]`)); err == nil {
		t.Fatal("expected empty gap payload to fail")
	}
}

func TestDeckCallAndDeckState(t *testing.T) {
	optimisticCall := &stubSyncCall{params: action.CallParams{Optimistic: true}}
	deckCallOptimistic := &deckCall{call: optimisticCall, params: optimisticCall.params}
	deckCallOptimistic.written()
	deckCallOptimistic.written()
	if optimisticCall.resultCount != 1 {
		t.Fatalf("expected optimistic write result once, got %d", optimisticCall.resultCount)
	}
	if string(optimisticCall.lastOutput) != "null" {
		t.Fatalf("unexpected optimistic output: %s", optimisticCall.lastOutput)
	}

	cancelCall := &stubSyncCall{act: action.Test{Arg: "cancel"}}
	deckCallCancel := &deckCall{call: cancelCall}
	if act, ok := deckCallCancel.action(); !ok || act.Log() != "test" {
		t.Fatalf("unexpected deck call action: %v %v", act, ok)
	}
	deckCallCancel.cancel()
	deckCallCancel.cancel()
	if cancelCall.cancelCount != 1 {
		t.Fatalf("expected cancel once, got %d", cancelCall.cancelCount)
	}

	resultCall := &stubSyncCall{act: action.Test{Arg: "result"}}
	deckCallResult := &deckCall{call: resultCall}
	expectedErr := errors.New("boom")
	deckCallResult.result(json.RawMessage(`{"ok":true}`), expectedErr)
	deckCallResult.result(json.RawMessage(`{"ok":false}`), nil)
	if resultCall.resultCount != 1 {
		t.Fatalf("expected result once, got %d", resultCall.resultCount)
	}
	if !errors.Is(resultCall.lastErr, expectedErr) {
		t.Fatal("expected deck call result error to be preserved")
	}

	expireHandler := &stubExpireHandler{}
	deck := newDeck(expirator.NewExpirator(expireHandler), 4, 4, time.Second)
	if deck.PendingCount() != 0 || deck.QueueLength() != 0 || deck.Pending() {
		t.Fatal("expected new deck to start empty")
	}

	call := &stubSyncCall{act: action.Test{Arg: "queued"}}
	if err := deck.Insert(call); err != nil {
		t.Fatal(err)
	}
	if deck.QueueLength() != 1 {
		t.Fatalf("unexpected queue length after insert: %d", deck.QueueLength())
	}
	if deck.Pending() {
		t.Fatal("expected deck to have no pending issued calls before WriteNext")
	}

	recorder := &stubRW{}
	w := &writer{
		sizeLimit: 100,
		timeLimit: time.Hour,
		w:         recorder,
		f:         recorder,
	}
	res, syncErr := deck.WriteNext(w)
	if res != writeOk || syncErr != nil {
		t.Fatalf("unexpected write result: %v %v", res, syncErr)
	}
	if deck.PendingCount() != 1 || !deck.Pending() {
		t.Fatal("expected issued call to become pending after WriteNext")
	}

	extraCall := &stubSyncCall{act: action.Test{Arg: "extra"}}
	if err := deck.cancelCut(inner.NewCard(99, &inner.Call{Call: extraCall})); err != nil {
		t.Fatal(err)
	}
	if deck.QueueLength() != 1 {
		t.Fatalf("unexpected queue length after cancelCut: %d", deck.QueueLength())
	}

	deck.End()
	if call.cancelCount != 1 {
		t.Fatalf("expected issued call to be canceled on deck end, got %d", call.cancelCount)
	}
}

func TestSolitaireAndConnectionHelpers(t *testing.T) {
	inst := &stubInstance{}
	s := NewSolitaire(inst, testSolitaireConf())

	queuedCall := &stubSyncCall{act: action.Test{Arg: "queued"}}
	s.Call(queuedCall)
	if s.deck.QueueLength() != 1 {
		t.Fatalf("expected solitaire call to queue work, got %d", s.deck.QueueLength())
	}

	s.Expire()
	if len(inst.syncErrors) == 0 || inst.syncErrors[0].Error() != "sync timeout" {
		t.Fatalf("unexpected expire errors: %#v", inst.syncErrors)
	}

	recorder := &stubRW{}
	ctx, cancel := context.WithCancelCause(context.Background())
	connection := &con{
		writer: &writer{
			sizeLimit: 100,
			timeLimit: time.Hour,
			w:         recorder,
			f:         recorder,
		},
		ctx:      ctx,
		cancel:   cancel,
		endGuard: make(chan struct{}),
		deck:     s.deck,
		inst:     inst,
	}
	if connection.Context() != ctx {
		t.Fatal("expected connection context to be returned unchanged")
	}

	cancel(context.Canceled)
	connection.handleCause()
	if !bytes.Contains(recorder.body.Bytes(), rollSignal) {
		t.Fatalf("expected roll signal to be written, got %v", recorder.body.Bytes())
	}

	endCtx, endCancel := context.WithCancelCause(context.Background())
	s.conn.Store(&con{
		writer: &writer{w: &stubRW{}, f: &stubRW{}},
		ctx:    endCtx,
		cancel: endCancel,
		deck:   s.deck,
		inst:   inst,
	})
	s.End(common.EndCauseSuspend)
	if !errors.Is(context.Cause(endCtx), common.EndCauseSuspend) {
		t.Fatalf("unexpected end cause: %v", context.Cause(endCtx))
	}
}
