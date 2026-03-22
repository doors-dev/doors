// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package solitaire

import (
	"errors"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
	"github.com/doors-dev/doors/internal/solitaire/inner"
	"sync"
	"time"
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
	defer d.mu.Unlock()
	if d.killed {
		return
	}
	d.expirator.Shutdown()
	d.killed = true
	for seq := range d.issued {
		d.issued[seq].call.Cancel()
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
)

func (d *deck) WriteNext(w *writer) (res writeResult, syncErr error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for {
		if d.killed {
			return writeErr, nil
		}
		if len(d.issued) == d.issueLimit {
			return writeLimit, nil
		}
		card := d.inner.Cut()
		if card == nil {
			return writeNothing, nil
		}
		header := newHeader(card.Beg, card.End)
		if card.IsFiller() {
			if err := header.writeFiller(w); err != nil {
				d.inner.Fill(card.Beg, card.End)
				return writeErr, nil
			}
			return writeOk, nil
		}
		action, ok := card.Call.Action()
		if !ok {
			d.expirator.Report(card.Seq())
			card.Call.Cancel()
			d.inner.Fill(card.Beg, card.End)
			continue
		}
		issuedCall := &issuedCall{
			invocation: action.Invocation(),
			call:       card.Call,
		}
		err := issuedCall.write(header, w)
		if err != nil {
			if errors.Is(err, writerError) {
				if err := d.cancelCut(card); err != nil {
					return writeSyncErr, err
				}
				return writeErr, nil
			}
			d.expirator.Report(card.Seq())
			issuedCall.call.Result(nil, errors.Join(errors.New("call serialization error"), err))
			d.inner.Fill(card.Beg, card.End)
			_, err := w.Write(errorTerminator)
			if err != nil {
				return writeErr, nil
			}
			return writeOk, nil
		}
		d.issued[card.Seq()] = issuedCall
		w.AfterFlush(issuedCall.call.Written)
		return writeOk, nil
	}
}

func (d *deck) CollectResults(r map[uint64]result) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
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
			restored.Call.Result(result.output, result.err)
			continue
		}
		delete(d.issued, seq)
		counter += 1
		d.expirator.Report(seq)
		issued.call.Result(result.output, result.err)
	}
	return counter, nil
}

func (d *deck) HeatUp() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.inner.IsCold(d.seq) && len(d.issued) > 0 {
		d.seq += 1
		d.inner.Fill(d.seq, d.seq)
	}
}

func (d *deck) FillGaps(g []gap) error {
	d.mu.Lock()
	defer d.mu.Unlock()
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
	d.inner.Insert(n)
	return d.checkQueueLength()
}
