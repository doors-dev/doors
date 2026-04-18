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
	"errors"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
	"github.com/doors-dev/doors/internal/solitaire/inner"
)

func newDeck(expirator *expirator.Expirator, queueLimit int, issueLimit int, syncTimeout time.Duration) *deck {
	d := &deck{
		expirator:   expirator,
		issueLimit:  issueLimit,
		queueLimit:  queueLimit,
		issued:      make(map[uint64]*issuedCall),
		syncTimeout: syncTimeout,
	}
	return d
}

type deck struct {
	issueLimit   int
	queueLimit   int
	syncTimeout  time.Duration
	seq          uint64
	issued       map[uint64]*issuedCall
	mu           sync.Mutex
	killed       bool
	latestReport uint64
	expirator    *expirator.Expirator
	inner        inner.Deck
}

func (d *deck) End() {
	d.mu.Lock()
	if d.killed {
		d.mu.Unlock()
		return
	}
	d.killed = true
	d.mu.Unlock()
	d.expirator.Shutdown()
	for _, issued := range d.issued {
		if issued.call == nil {
			continue
		}
		issued.call.Cancel()
	}
	d.inner.Cancel()
}

func (d *deck) PendingCount() int {
	return len(d.issued)
}
func (d *deck) QueueLength() int {
	return d.inner.Len()
}

func (d *deck) Pending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.issued) != 0
}

type writeResult int

const (
	writeOk writeResult = iota
	writeNothing
	writeErr
	writeLimit
	writeSyncErr
	writeContinue
	writeKilled
)

func (d *deck) issue() (*inner.Card, *issuedCall, writeResult) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return nil, nil, writeKilled
	}
	if len(d.issued) == d.issueLimit {
		return nil, nil, writeLimit
	}
	card := d.inner.Cut()
	if card == nil {
		return nil, nil, writeNothing
	}
	var issued *issuedCall
	if !card.IsFiller() {
		issued = &issuedCall{}
		d.issued[card.Seq()] = issued
	}
	return card, issued, writeContinue
}

func (d *deck) unIssue(card *inner.Card) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return nil
	}
	if card.IsFiller() {
		d.inner.Fill(card.Beg, card.End)
		return nil
	}
	delete(d.issued, card.Seq())
	d.inner.Insert(card)
	return d.checkQueueLength()
}

func (d *deck) invocation(card *inner.Card, issued *issuedCall) writeResult {
	action, ok := card.Call.Action()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return writeKilled
	}
	if !ok {
		delete(d.issued, card.Seq())
		d.inner.Fill(card.Beg, card.End)
		return writeNothing
	}
	issued.call = card.Call
	issued.invocation = action.Invocation()
	return writeContinue
}

func (d *deck) error(card *inner.Card) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return
	}
	d.inner.Fill(card.Beg, card.End)
	delete(d.issued, card.Seq())
}

func (d *deck) WriteNext(w *writer) (res writeResult, syncErr error) {
	card, issued, res := d.issue()
	if res != writeContinue {
		return res, nil
	}
	header := newHeader(card.Beg, card.End)
	if card.IsFiller() {
		if err := header.writeFiller(w); err != nil {
			d.unIssue(card)
			return writeErr, nil
		}
		return writeOk, nil
	}
	res = d.invocation(card, issued)
	if res == writeKilled {
		card.Call.Cancel()
		return writeKilled, nil
	}
	if res == writeNothing {
		d.expirator.Report(card.Seq())
		card.Call.Cancel()
		return writeContinue, nil
	}
	if res != writeContinue {
		return res, nil
	}
	err := issued.write(header, w)
	issued.invocation = action.Invocation{}
	if err != nil {
		if errors.Is(err, writerError) {
			if err := d.unIssue(card); err != nil {
				return writeSyncErr, err
			}
			return writeErr, nil
		}
		d.error(card)
		d.expirator.Report(card.Seq())
		card.Call.Result(nil, errors.Join(errors.New("call serialization error"), err))
		_, err := w.Write(errorTerminator)
		if err != nil {
			return writeErr, nil
		}
		return writeOk, nil
	}
	card.Call.Written()
	return writeOk, nil
}

type bufferedResult struct {
	call   *inner.Call
	result result
}

func (b bufferedResult) process() {
	b.call.Result(b.result.output, b.result.err)
}

func (d *deck) CollectResults(r map[uint64]result) (int, error) {
	buffer := make([]bufferedResult, 0, len(r))
	defer func() {
		for _, r := range buffer {
			r.process()
		}
	}()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return 0, context.Canceled
	}
	counter := 0
	for seq, result := range r {
		if seq > d.seq {
			return counter, errors.New("ready overflows last seq")
		}
		d.latestReport = max(d.latestReport, seq)
		issued, ok := d.issued[seq]
		if !ok {
			restored, err := d.inner.ExtractRestored(seq)
			if err != nil {
				return counter, err
			}
			if restored == nil {
				continue
			}
			counter += 1
			d.expirator.Report(seq)
			buffer = append(buffer, bufferedResult{restored.Call, result})
			continue
		}
		if issued.call == nil {
			return counter, errors.New("reported to unwritten card")
		}
		delete(d.issued, seq)
		counter += 1
		d.expirator.Report(seq)
		buffer = append(buffer, bufferedResult{issued.call, result})
	}
	return counter, nil
}

func (d *deck) HeatUp() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inner.IsCold(d.seq) && len(d.issued) > 0 {
		d.seq += 1
		d.inner.Probe(d.seq)
	}
}

func (d *deck) FillGaps(g []gap) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return context.Canceled
	}
	tolarance := min(uint64(d.issueLimit), d.seq)
	prevEnd := max(d.latestReport, tolarance) - tolarance
	for _, gap := range g {
		if gap.end < gap.start {
			return errors.New("gap range issue")
		}
		if gap.end > d.seq {
			return errors.New("gap overflows last seq")
		}
		if prevEnd >= gap.start {
			return errors.New("gap overlap")
		}
		prevEnd = gap.end
		beg := gap.start
		for seq := max(gap.start, d.latestReport); seq <= gap.end; seq++ {
			call, ok := d.issued[seq]
			if !ok {
				continue
			}
			if call.call == nil {
				return errors.New("can't report gap to non-issued card")
			}
			if beg != seq {
				d.inner.Fill(beg, seq-1)
			}
			beg = seq + 1
			delete(d.issued, seq)
			if err := d.restore(seq, call.call); err != nil {
				return err
			}
		}
		if beg <= gap.end {
			d.inner.Fill(beg, gap.end)
		}
	}
	return nil
}

func (d *deck) Insert(c action.Call) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		c.Cancel()
		return errors.New("killed")
	}
	d.seq += 1
	dc := &inner.Call{
		Call:   c,
		Params: c.Params(),
	}
	if dc.Params.Timeout == 0 {
		dc.Params.Timeout = d.syncTimeout
	}
	card := inner.NewCard(d.seq, dc)
	d.inner.Append(card)
	deadline := time.Now().Add(dc.Params.Timeout)
	d.expirator.Track(d.seq, deadline)
	return d.checkQueueLength()
}

func (d *deck) checkQueueLength() error {
	if d.inner.Len() < d.queueLimit {
		return nil
	}
	return errors.New("call queue limit reached")
}

func (d *deck) restore(seq uint64, c *inner.Call) error {
	n := inner.NewRestoredCard(seq, c)
	d.inner.Insert(n)
	return d.checkQueueLength()

}

func (d *deck) cancelCut(n *inner.Card) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.inner.Insert(n)
	return d.checkQueueLength()
}
