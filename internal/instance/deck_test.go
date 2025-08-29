// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

func testCounters(t *testing.T, d *deck, queue int, pending int) {
	if d.QueueLength() != queue {
		t.Fatal("Wrong queue length:", queue, d.QueueLength())
	}
	if d.PendingCount() != pending {
		t.Fatal("Wrong pending count:", pending, d.PendingCount())
	}
}

func testNothing(t *testing.T, deck *deck, w io.Writer) {
	r, _ := deck.WriteNext(w)
	if r != nothingToWrite {
		t.Fatal("queue must be empty")
	}
}

func testWriteAndCheck(t *testing.T, deck *deck, w *testWriter, start uint64, end uint64, arg int) {
	if deck != nil {
		r, _ := deck.WriteNext(w)
		if r != writeOk {
			t.Fatal("nothingToWrite")
		}
	}
	p, err := w.readPackage()
	if err != nil {
		t.Fatal(err)
	}
	if p.startSeq != start || p.endSeq != end || p.arg != arg {
		t.Fatal("wrong expectations: start ", start, p.startSeq, " end:", end, p.endSeq, " arg:", arg, p.arg)
	}
}

type testCalls [10]*testCall

func (t testCalls) get(i int) *testCall {
	return t[i-1]
}

func (t testCalls) cancel(i int) {
	t[i-1].cancel = true
}

func (t testCalls) insertWrite(deck *deck, w io.Writer, insertCount int, writeCount int) error {
	for i := range insertCount {
		err := deck.Insert(t[i])
		if err != nil {
			return err
		}
	}
	for i := range writeCount {
		r, err := deck.WriteNext(w)
		if r != writeOk {
			return errors.Join(err, errors.New("write err "+fmt.Sprint(i, r)))
		}
	}
	return nil
}

func newTestCalls() testCalls {
	var ps [10]*testCall
	for i := range 10 {
		ps[i] = &testCall{
			arg:       i + 1,
			resultErr: nil,
		}
	}
	return ps
}

func TestRange(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 10, 11, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	ps.insertWrite(deck, w, 10, 10)
	testCounters(t, deck, 0, 10)
	testNothing(t, deck, w)

	for i := range 10 {
		p, err := w.readPackage()
		if err != nil {
			t.Fatal(err)
		}
		if p.arg != i+1 {
			t.Fatal("Wrong arg")
		}
	}
	testNothing(t, deck, w)
}

func TestRetry(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 1, 1, time.Minute)
	w := &testWriter{}
	deck.Insert(&testCall{
		arg:     123,
		payload: "123",
	})
	testCounters(t, deck, 1, 0)
	w.errNextWrite = true
	r, _ := deck.WriteNext(w)
	testCounters(t, deck, 1, 0)
	if r != writeErr {
		t.Fatal("Expect error")
	}
	r, _ = deck.WriteNext(w)
	if r != writeOk {
		t.Fatal("Write errored")
	}
	p, err := w.readPackage()
	if err != nil {
		t.Fatal(err.Error())
	}
	if p.arg != 123 {
		t.Fatal("wrong arg")
	}
}

func TestSkip(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 5, 6, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	ps.cancel(1)
	ps.cancel(3)
	ps.cancel(4)
	ps.cancel(5)
	err := ps.insertWrite(deck, w, 5, 2)
	testCounters(t, deck, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	testWriteAndCheck(t, nil, w, 1, 2, 2)
	testCounters(t, deck, 0, 1)
	testWriteAndCheck(t, nil, w, 3, 5, 0)
	testCounters(t, deck, 0, 1)
	testNothing(t, deck, w)
}

func TestReportResult(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 5, 6, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 5, 5)
	testCounters(t, deck, 0, 5)
	if err != nil {
		t.Fatal(err)
	}
	re := errors.New("test")
	rep := &report{
		Gaps: []gap{},
		Results: map[uint64]result{
			1: result{output: nil, err: nil},
			2: result{output: nil, err: re},
			3: result{output: nil, err: re},
			4: result{output: nil, err: nil},
			5: result{output: nil, err: re},
		},
	}
	err = deck.OnReport(rep)
	if err != nil {
		t.Fatal(err)
	}
	testCounters(t, deck, 0, 0)
	if ps[0].resultErr != nil {
		t.Fatal("wrong result")
	}
	if ps[3].resultErr != nil {
		t.Fatal("wrong result")
	}
	if ps[1].resultErr.Error() != re.Error() {
		t.Fatal("wrong result")
	}
	if ps[2].resultErr.Error() != re.Error() {
		t.Fatal("wrong result")
	}
	if ps[4].resultErr.Error() != re.Error() {
		t.Fatal("wrong result")
	}
}

func TestReportGap(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 6, 7, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 6, 5)
	testCounters(t, deck, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = w.skip(5)
	if err != nil {
		t.Fatal(err)
	}
	s := `{
	"gaps":[[2,4]],
	"results":{"1":[null,null]}
	}`
	rep := &report{}
	err = json.Unmarshal([]byte(s), rep)
	if err != nil {
		t.Fatal(err)
	}
	ps.cancel(2)
	err = deck.OnReport(rep)
	testCounters(t, deck, 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	ps.cancel(4)
	testWriteAndCheck(t, deck, w, 2, 3, 3)
	testCounters(t, deck, 2, 2)
	testWriteAndCheck(t, deck, w, 4, 4, 0)
	testCounters(t, deck, 1, 2)
	w.errNextWrite = true
	r, err := deck.WriteNext(w)
	if r != writeErr {
		t.Fatal("write must fail")
	}
	testCounters(t, deck, 1, 2)
	w.errNextWrite = true
	r, err = deck.WriteNext(w)
	if r != writeErr {
		t.Fatal("write must fail")
	}
	testCounters(t, deck, 1, 2)
	testWriteAndCheck(t, deck, w, 6, 6, 6)
	testCounters(t, deck, 0, 3)
}

func TestExtraction(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 6, 7, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 5, 4)
	testCounters(t, deck, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	err = w.skip(4)
	if err != nil {
		t.Fatal(err)
	}
	rep := &report{
		Gaps: []gap{{
			start: 3,
			end:   4,
		}},
		Results: map[uint64]result{
			1: result{output: nil, err: nil},
		},
	}
	deck.OnReport(rep)
	testCounters(t, deck, 3, 1)
	rep = &report{
		Gaps: []gap{{
			start: 4,
			end:   4,
		}},
		Results: map[uint64]result{
			2: result{output: nil, err: nil},
			3: result{output: nil, err: nil},
		},
	}
	deck.OnReport(rep)
	testCounters(t, deck, 2, 0)
	testWriteAndCheck(t, deck, w, 4, 4, 4)
	testCounters(t, deck, 1, 1)
	testWriteAndCheck(t, deck, w, 5, 5, 5)
	testCounters(t, deck, 0, 2)
}

func TestSkipTail(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 6, 7, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 4, 3)
	testCounters(t, deck, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = w.skip(3)
	if err != nil {
		t.Fatal(err)
	}
	rep := &report{
		Gaps: []gap{{
			start: 1,
			end:   1,
		}},
		Results: map[uint64]result{},
	}
	deck.OnReport(rep)
	testCounters(t, deck, 2, 2)
	ps.cancel(3)
	rep = &report{
		Gaps: []gap{{
			start: 3,
			end:   3,
		}},
		Results: map[uint64]result{},
	}
	err = deck.OnReport(rep)
	testCounters(t, deck, 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	testWriteAndCheck(t, deck, w, 1, 1, 1)
	testCounters(t, deck, 2, 2)
	testWriteAndCheck(t, deck, w, 3, 4, 4)
	testCounters(t, deck, 0, 3)
}

func TestLimits(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 6, 6, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 6, 0)
	testCounters(t, deck, 6, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = deck.Insert(ps.get(7))
	testCounters(t, deck, 7, 0)
	if err == nil {
		t.Fatal("must hit queue limit")
	}
	testWriteAndCheck(t, deck, w, 1, 1, 1)
	testWriteAndCheck(t, deck, w, 2, 2, 2)
	testCounters(t, deck, 5, 2)
	err = deck.Insert(ps.get(8))
	if err != nil {
		t.Fatal("there must be space")
	}
	testCounters(t, deck, 6, 2)
	testWriteAndCheck(t, deck, w, 3, 3, 3)
	testCounters(t, deck, 5, 3)
	testWriteAndCheck(t, deck, w, 4, 4, 4)
	testCounters(t, deck, 4, 4)
	testWriteAndCheck(t, deck, w, 5, 5, 5)
	testCounters(t, deck, 3, 5)
	testWriteAndCheck(t, deck, w, 6, 6, 6)
	testCounters(t, deck, 2, 6)
	r, _ := deck.WriteNext(w)
	if r != pendingLimit {
		t.Fatal("must reach pending limit")
	}
	deck.OnReport(&report{
		Gaps: []gap{},
		Results: map[uint64]result{
			1: result{output: nil, err: nil},
		},
	})
	r, _ = deck.WriteNext(w)
	if r != writeOk {
		t.Fatal("must be available")
	}
	testCounters(t, deck, 1, 6)
}

func TestReportNonIssued(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 6, 6, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 6, 4)
	testCounters(t, deck, 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	rep := &report{
		Gaps: []gap{},
		Results: map[uint64]result{
			5: result{output: nil, err: nil},
		},
	}
	err = deck.OnReport(rep)
	if err == nil {
		t.Fatal("must error - reported non issued")
	}
	rep = &report{
		Gaps: []gap{},
		Results: map[uint64]result{
			7: result{output: nil, err: nil},
		},
	}
	err = deck.OnReport(rep)
	if err == nil {
		t.Fatal("must error - reported non overflow")
	}
	testCounters(t, deck, 2, 4)
}
func TestReport(t *testing.T) {
	ti := &testInstance{}
	deck := newDeck(ti, 7, 7, time.Minute)
	w := &testWriter{}
	ps := newTestCalls()
	err := ps.insertWrite(deck, w, 7, 7)
	testCounters(t, deck, 0, 7)
	if err != nil {
		t.Fatal(err)
	}
	rep := &report{
		Gaps: []gap{{
			start: 2,
			end:   7,
		}},
		Results: map[uint64]result{
			3: result{output: nil, err: nil},
		},
	}
	err = deck.OnReport(rep)
	if err == nil {
		t.Fatal("must error - reported  overflow")
	}
	testCounters(t, deck, 0, 6)
	rep = &report{
		Gaps: []gap{{
			start: 2,
			end:   3,
		}},
		Results: map[uint64]result{},
	}
	err = deck.OnReport(rep)
	if err == nil {
		t.Fatal("must error - reported gap after result")
	}
	rep = &report{
		Gaps: []gap{{
			start: 4,
			end:   5,
		}, {
			start: 5,
			end:   6,
		}},
		Results: map[uint64]result{},
	}
	err = deck.OnReport(rep)
	if err == nil {
		t.Fatal("must error - reported gap overlap")
	}
	testCounters(t, deck, 2, 4)
	rep = &report{
		Gaps: []gap{{
			start: 6,
			end:   5,
		}},
		Results: map[uint64]result{},
	}
	err = deck.OnReport(rep)
	if err == nil {
		t.Fatal("must error - reported gap order")
	}
	testCounters(t, deck, 2, 4)
}
