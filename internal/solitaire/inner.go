// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package solitaire
/*
import (
	"errors"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/solitaire/inner"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
	"io"
	"sync"
	"time"
)

func newDeck(inst Instance, queueLimit int, issueLimit int, syncTimeout time.Duration) *deck {
	d := &deck{
		inst:        inst,
		issueLimit:  issueLimit,
		queueLimit:  queueLimit,
		issued:      make(map[uint64]*issuedCall),
		syncTimeout: syncTimeout,
	}
	d.expirator = expirator.NewExpirator(d)
	return d
}

type deck struct {
	inst         Instance
	issueLimit   int
	queueLimit   int
	syncTimeout  time.Duration
	seq          uint64
	issued       map[uint64]*issuedCall
	mu           sync.Mutex
	killed       bool
	deckSize     int
	latestReport uint64
	expirator    *expirator.Expirator
	inner        inner.Deck
}

func (d *deck) Expire() {
	d.inst.SyncError(errors.New("sync timeout"))
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
	return d.deckSize
}

func (d *deck) Pending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.issued) != 0
}

type writeResult int

const (
	writeOk writeResult = iota
	nothingToWrite
	writeErr
	pendingLimit
	writeSyncErr
)

type flushWriter interface {
	io.Writer
	afterFlush(func())
}

func (d *deck) WriteNext(w flushWriter) (res writeResult, err error) {
	defer func() {
		if res == writeSyncErr {
			d.inst.SyncError(err)
		}
	}()
	d.mu.Lock()
	defer d.mu.Unlock()
	for {
		if d.killed {
			return writeSyncErr, errors.New("killed")
		}
		if len(d.issued) == d.issueLimit {
			return pendingLimit, nil
		}
		card := d.cutTop()
		if card == nil {
			return nothingToWrite, nil
		}
		header := newHeader(card.StartSeq, card.EndSeq)
		if card.IsFiller() {
			if err := header.writeFiller(w); err != nil {
				d.skipRange(card.StartSeq, card.EndSeq)
				return writeErr, err
			}
			return writeOk, nil
		}
		action, ok := card.Call.Action()
		if !ok {
			d.expirator.Report(card.Seq())
			card.Call.Cancel()
			d.skipRange(card.StartSeq, card.EndSeq)
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
				return writeErr, err
			}
			d.expirator.Report(card.Seq())
			issuedCall.call.Result(nil, errors.Join(errors.New("call serialization error"), err))
			d.skipRange(card.StartSeq, card.EndSeq)
			_, err := w.Write(errorTerminator)
			if err != nil {
				return writeErr, err
			}
			return writeOk, nil
		}
		d.issued[card.Seq()] = issuedCall
		w.afterFlush(issuedCall.call.Written)
		return writeOk, nil
	}
}

func (d *deck) extractRestored(seq uint64) (*inner.Card, error) {
	card, err := d.inner.ExtractRestored(seq)
	if card != nil {
		d.dec()
	}
	return card, err
}

func (d *deck) OnReport(s *report) (counter int, err error) {
	defer func() {
		if err != nil {
			d.inst.SyncError(err)
		}
	}()
	d.mu.Lock()
	defer d.mu.Unlock()
	latest := uint64(0)
	for seq := range s.Results {
		if seq > d.seq {
			return counter, errors.New("ready overflows last seq")
		}
		latest = max(latest, seq)
		result := s.Results[seq]
		issued, ok := d.issued[seq]
		if !ok {
			restored, err := d.extractRestored(seq)
			if err != nil {
				return counter, err
			}
			if restored == nil {
				continue
			}
			counter += 1
			d.expirator.Report(seq)
			restored.Call.Result(result.output, result.err)
			// restored.call.clean()
			continue
		}
		delete(d.issued, seq)
		counter += 1
		d.expirator.Report(seq)
		issued.call.Result(result.output, result.err)
		// issued.call.clean()
	}
	prevEnd := latest
	if latest < d.latestReport {
		tolarance := min(uint64(d.issueLimit), d.seq)
		prevEnd = max(d.latestReport, tolarance) - tolarance
	} else {
		d.latestReport = latest
	}
	for _, gap := range s.Gaps {
		if gap.end < gap.start {
			return counter, errors.New("gap range issue")
		}
		if gap.end > d.seq {
			return counter, errors.New("gap overflows last seq")
		}
		if prevEnd >= gap.start {
			return counter, errors.New("gap overlap")
		}
		prevEnd = gap.end
		for seq := gap.start; seq <= gap.end; seq++ {
			if seq <= d.latestReport {
				continue
			}
			call, ok := d.issued[seq]
			if !ok {
				d.skipSeq(seq)
				continue
			}
			delete(d.issued, seq)
			if err := d.restore(seq, call.call); err != nil {
				return counter, err
			}
		}
	}
	if d.inner.IsCold(d.seq) && len(d.issued) > 0 {
		d.seq += 1
		d.skipSeq(d.seq)
	}
	return counter, nil
}
func (d *deck) Insert(c action.Call) (err error) {
	defer func() {
		if err != nil {
			d.inst.SyncError(err)
		}
	}()
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
	return d.inc()
}

func (d *deck) inc() error {
	d.deckSize += 1
	if d.deckSize > d.queueLimit {
		return errors.New("call queue limit reached")
	}
	return nil
}
func (d *deck) dec() {
	d.deckSize -= 1
}

func (d *deck) restore(seq uint64, c *inner.Call) error {
	n := inner.NewRestoredCard(seq, c)
	err := d.inner.Insert(n)
	if err != nil {
		return err
	}
	return d.inc()

}

func (d *deck) skipSeq(seq uint64) {
	d.inner.Skip(seq)
}

func (d *deck) cutTop() *inner.Card {
	card := d.inner.Cut()
	if card == nil {
		return nil
	}
	if !card.IsFiller() {
		d.dec()
	}
	return card
}

func (d *deck) cancelCut(n *inner.Card) error {
	if n.IsFiller() {
		panic("Cannot cancel filler cut")
	}
	if err := d.inner.Insert(n); err != nil {
		return err
	}
	return d.inc()
}

func (d *deck) skipRange(start uint64, end uint64) {
	for seq := start; seq <= end; seq++ {
		d.skipSeq(seq)
	}
} */
