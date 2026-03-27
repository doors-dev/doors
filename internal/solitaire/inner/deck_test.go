package inner

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/doors-dev/doors/internal/front/action"
)

type stubActionCall struct {
	act         action.Action
	cancelCount int
	resultCount int
	lastResult  json.RawMessage
	lastErr     error
}

func (s *stubActionCall) Params() action.CallParams {
	return action.CallParams{}
}

func (s *stubActionCall) Action() (action.Action, bool) {
	return s.act, true
}

func (s *stubActionCall) Cancel() {
	s.cancelCount += 1
}

func (s *stubActionCall) Result(result json.RawMessage, err error) {
	s.resultCount += 1
	s.lastResult = result
	s.lastErr = err
}

func newStubInnerCall(id int) (*Call, *stubActionCall) {
	call := &stubActionCall{act: action.Test{Arg: id}}
	return &Call{
		Call: call,
	}, call
}

func TestCallLifecycle(t *testing.T) {
	optimisticCall, optimisticStub := newStubInnerCall(1)
	optimisticCall.Params = action.CallParams{Optimistic: true}
	optimisticCall.Written()
	optimisticCall.Written()
	if optimisticStub.resultCount != 1 {
		t.Fatalf("expected optimistic written result once, got %d", optimisticStub.resultCount)
	}
	if string(optimisticStub.lastResult) != "null" {
		t.Fatalf("unexpected optimistic result payload: %s", optimisticStub.lastResult)
	}

	plainCall, plainStub := newStubInnerCall(2)
	gotAction, ok := plainCall.Action()
	if !ok {
		t.Fatal("expected action to be available")
	}
	if gotAction.Log() != "test" {
		t.Fatalf("unexpected action log: %s", gotAction.Log())
	}
	plainCall.Cancel()
	plainCall.Cancel()
	if plainStub.cancelCount != 1 {
		t.Fatalf("expected cancel once, got %d", plainStub.cancelCount)
	}

	resultCall, resultStub := newStubInnerCall(3)
	expectedErr := errors.New("boom")
	resultCall.Result(json.RawMessage(`{"ok":true}`), expectedErr)
	resultCall.Result(json.RawMessage(`{"ok":false}`), nil)
	if resultStub.resultCount != 1 {
		t.Fatalf("expected result once, got %d", resultStub.resultCount)
	}
	if string(resultStub.lastResult) != `{"ok":true}` {
		t.Fatalf("unexpected result payload: %s", resultStub.lastResult)
	}
	if !errors.Is(resultStub.lastErr, expectedErr) {
		t.Fatal("expected result error to be preserved")
	}
}

func TestDeckAppendCutAndCancel(t *testing.T) {
	var deck Deck
	card1, stub1 := newStubInnerCall(1)
	card2, stub2 := newStubInnerCall(2)
	deck.Append(NewCard(1, card1))
	deck.Append(NewCard(2, card2))

	if deck.Len() != 2 {
		t.Fatalf("unexpected deck length: %d", deck.Len())
	}
	if deck.IsCold(2) {
		t.Fatal("expected deck to be warm for latest sequence")
	}

	first := deck.Cut()
	if first == nil || first.Seq() != 1 {
		t.Fatalf("unexpected first cut card: %#v", first)
	}
	if deck.Len() != 1 {
		t.Fatalf("unexpected deck length after cut: %d", deck.Len())
	}

	deck.Cancel()
	if stub1.cancelCount != 0 {
		t.Fatalf("expected already cut card to stay untouched, got %d cancels", stub1.cancelCount)
	}
	if stub2.cancelCount != 1 {
		t.Fatalf("expected remaining card to be canceled once, got %d", stub2.cancelCount)
	}

	second := deck.Cut()
	if second == nil || second.Seq() != 2 {
		t.Fatalf("unexpected second cut card: %#v", second)
	}
	if deck.Cut() != nil {
		t.Fatal("expected empty deck after cutting both cards")
	}
}

func TestDeckFillInsertAndExtractRestored(t *testing.T) {
	var fillers Deck
	fillers.Fill(3, 3)
	fillers.Fill(1, 1)
	fillers.Fill(2, 3)
	filler := fillers.Cut()
	if filler == nil || !filler.IsFiller() {
		t.Fatal("expected merged filler card")
	}
	if filler.Beg != 1 || filler.End != 3 {
		t.Fatalf("unexpected filler bounds: %d-%d", filler.Beg, filler.End)
	}
	if fillers.Cut() != nil {
		t.Fatal("expected filler deck to be empty after cut")
	}

	var inserted Deck
	lateCall, _ := newStubInnerCall(2)
	earlyCall, _ := newStubInnerCall(1)
	inserted.Insert(NewCard(2, lateCall))
	inserted.Insert(NewCard(1, earlyCall))
	first := inserted.Cut()
	if first == nil || first.Seq() != 1 {
		t.Fatalf("expected lower sequence to move to top, got %#v", first)
	}

	var restored Deck
	call1, _ := newStubInnerCall(1)
	call2, _ := newStubInnerCall(2)
	call3, _ := newStubInnerCall(3)
	restored.Append(NewCard(1, call1))
	restored.Append(NewRestoredCard(2, call2))
	restored.Append(NewCard(3, call3))

	card, err := restored.ExtractRestored(2)
	if err != nil {
		t.Fatal(err)
	}
	if card == nil || card.Seq() != 2 {
		t.Fatalf("unexpected restored card: %#v", card)
	}
	if restored.Len() != 2 {
		t.Fatalf("unexpected deck length after extraction: %d", restored.Len())
	}

	if _, err := restored.ExtractRestored(3); err == nil {
		t.Fatal("expected extracting a non-restored card to fail")
	}
	card, err = restored.ExtractRestored(99)
	if err != nil {
		t.Fatal(err)
	}
	if card != nil {
		t.Fatalf("expected missing restored card lookup to return nil, got %#v", card)
	}
}
