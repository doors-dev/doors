// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
)

func newDeck(inst solitaireInstance, queueLimit int, issueLimit int, syncTimeout time.Duration) *deck {
	return &deck{
		inst:        inst,
		issueLimit:  issueLimit,
		queueLimit:  queueLimit,
		issued:      make(map[uint64]*issuedCall),
		syncTimeout: syncTimeout,
		expirator: common.NewExpirator(func() {
			inst.syncError(errors.New("sync timeout"))
		}),
	}
}

type deck struct {
	inst         solitaireInstance
	issueLimit   int
	queueLimit   int
	syncTimeout  time.Duration
	seq          uint64
	top          *card
	bottom       *card
	issued       map[uint64]*issuedCall
	mu           sync.Mutex
	killed       bool
	deckSize     int
	latestReport uint64
	expirator    *common.Expirator
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
		d.issued[seq].call.cancel()
		d.issued[seq].call.clean()
	}
	if d.top == nil {
		return
	}
	d.top.cancel()
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
			d.inst.syncError(err)
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
		header := newHeader(card.startSeq, card.endSeq)
		if card.isFiller() {
			if err := header.writeFiller(w); err != nil {
				d.skipRange(card.startSeq, card.endSeq)
				return writeErr, err
			}
			return writeOk, nil
		}
		action, ok := card.call.action()
		if !ok {
			d.expirator.Report(card.seq())
			card.call.cancel()
			card.call.clean()
			d.skipRange(card.startSeq, card.endSeq)
			continue
		}
		issuedCall := &issuedCall{
			invocation: action.Invocation(),
			call:       card.call,
		}
		err := issuedCall.write(header, w)
		if err != nil {
			if errors.Is(err, writerError) {
				if err := d.cancelCut(card); err != nil {
					return writeSyncErr, err
				}
				return writeErr, err
			}
			d.expirator.Report(card.seq())
			issuedCall.call.result(nil, errors.Join(errors.New("call serialization error"), err))
			issuedCall.call.clean()
			d.skipRange(card.startSeq, card.endSeq)
			_, err := w.Write(errorTerminator)
			if err != nil {
				return writeErr, err
			}
			return writeOk, nil
		}
		d.issued[card.seq()] = issuedCall
		w.afterFlush(issuedCall.call.written)
		return writeOk, nil
	}
}

func (d *deck) extractRestored(seq uint64) (*card, error) {
	if d.top == nil {
		return nil, nil
	}
	card, err := d.top.extractRestored(seq, d)
	if card != nil {
		d.dec()
	}
	return card, err
}

func (d *deck) OnReport(s *report) (counter int, err error) {
	defer func() {
		if err != nil {
			d.inst.syncError(err)
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
			restored.call.result(result.output, result.err)
			restored.call.clean()
			continue
		}
		delete(d.issued, seq)
		counter += 1
		d.expirator.Report(seq)
		issued.call.result(result.output, result.err)
		issued.call.clean()
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
	if (d.bottom == nil || d.bottom.seq() != d.seq) && len(d.issued) > 0 {
		d.seq += 1
		d.skipSeq(d.seq)
	}
	return counter, nil
}
func (d *deck) Insert(c action.Call) (err error) {
	defer func() {
		if err != nil {
			d.inst.syncError(err)
		}
	}()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		c.Cancel()
		return errors.New("killed")
	}
	d.seq += 1
	dc := &deckCall{
		call:   c,
		params: c.Params(),
	}
	if dc.params.Timeout == 0 {
		dc.params.Timeout = d.syncTimeout
	}
	door := newCard(d.seq, dc)
	d.insertTail(door)
	deadline := time.Now().Add(dc.params.Timeout)
	d.expirator.Track(d.seq, deadline)
	return d.inc()
}

func (d *deck) insertTail(n *card) {
	if d.bottom == nil {
		d.bottom = n
		d.top = n
		return
	}
	d.bottom.insertTail(n)
	d.bottom = n
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

func (d *deck) restore(seq uint64, c *deckCall) error {
	n := newRestoredCard(seq, c)
	if d.top == nil {
		d.top = n
		d.bottom = n
		return d.inc()
	}
	err := d.top.insert(n, d)
	if err != nil {
		return err
	}
	return d.inc()

}

func (d *deck) skipSeq(seq uint64) {
	if d.top == nil {
		d.bottom = newFillerCard(seq)
		d.top = d.bottom
		return
	}
	d.top.skipSeq(seq, d)
}

func (d *deck) setTail(n *card) {
	d.top = n
}

func (d *deck) cutTop() *card {
	card := d.top
	if card == nil {
		return nil
	}
	d.top = card.tail
	if d.top == nil {
		d.bottom = nil
	}
	if !card.isFiller() {
		d.dec()
	}
	return card
}

func (d *deck) cancelCut(n *card) error {
	if n.isFiller() {
		panic("Cannot cancel filler cut")
	}
	if d.top == nil {
		d.top = n
		d.bottom = n
		return d.inc()
	}
	if err := d.top.insert(n, d); err != nil {
		return err
	}
	return d.inc()
}

func (d *deck) skipRange(start uint64, end uint64) {
	for seq := start; seq <= end; seq++ {
		d.skipSeq(seq)
	}
}

type head interface {
	setTail(*card)
}

func newRestoredCard(seq uint64, c *deckCall) *card {
	return &card{
		startSeq: seq,
		endSeq:   seq,
		call:     c,
		restored: true,
	}
}

func newCard(seq uint64, c *deckCall) *card {
	return &card{
		startSeq: seq,
		endSeq:   seq,
		call:     c,
	}
}
func newFillerCard(seq uint64) *card {
	return &card{
		startSeq: seq,
		endSeq:   seq,
		call:     nil,
	}
}

type card struct {
	startSeq uint64
	endSeq   uint64
	call     *deckCall
	tail     *card
	restored bool
}

func (s *card) cancel() {
	if !s.isFiller() {
		s.call.cancel()
		s.call.clean()
	}
	if s.tail == nil {
		return
	}
	s.tail.cancel()
}

func (s *card) extractRestored(seq uint64, h head) (*card, error) {
	if s.seq() == seq {
		if s.isFiller() {
			return nil, nil
		}
		if !s.restored {
			return nil, errors.New("Attempt to respond to non issued card")
		}
		h.setTail(s.tail)
		return s, nil
	}
	if seq < s.seq() || s.tail == nil {
		return nil, nil
	}
	return s.tail.extractRestored(seq, s)
}

func (s *card) seq() uint64 {
	return s.endSeq
}

func (sn *card) isFiller() bool {
	return sn.call == nil
}

func (sn *card) insert(n *card, h head) error {
	if n.isFiller() {
		panic("Cannot insert filler")
	}
	if sn.startSeq >= n.seq() && sn.endSeq <= n.seq() {
		return errors.New("overlapping range")
	}
	if n.seq() > sn.seq() {
		if sn.tail == nil {
			sn.setTail(n)
			return nil
		}
		return sn.tail.insert(n, sn)
	}
	n.setTail(sn)
	h.setTail(n)
	return nil
}

func (sn *card) insertTail(n *card) {
	if n.isFiller() {
		panic("Cannot insert filler tail")
	}
	if n.seq() <= sn.seq() {
		panic("Cannot insert older tail")
	}
	sn.setTail(n)
}

func (c *card) setTail(n *card) {
	c.tail = n
}
func (c *card) skipSeq(seq uint64, h head) {
	if seq == c.startSeq-1 {
		c.startSeq = seq
		return
	}
	if seq < c.startSeq-1 {
		filler := newFillerCard(seq)
		filler.setTail(c)
		h.setTail(filler)
		return
	}
	if seq > c.endSeq {
		if c.isFiller() && c.endSeq+1 == seq {
			c.endSeq = seq
			if c.tail != nil && c.tail.startSeq-1 == c.endSeq {
				c.tail.startSeq = c.startSeq
				h.setTail(c.tail)
			}
			return
		}
		if c.tail != nil {
			c.tail.skipSeq(seq, c)
			return
		}
		c.setTail(newFillerCard(seq))
	}
}

func newHeader(startSeq uint64, endSeq uint64) header {
	if startSeq == endSeq {
		return []any{[]uint64{endSeq}}
	}
	return []any{[]uint64{endSeq, startSeq}}

}

type header []any

var ackSignal = []byte{0x00}
var actionSignal = []byte{0x01}
var rollSignal = []byte{0x02}
var suspendSignal = []byte{0x03}
var killSignal = []byte{0x04}

var continueWithPayload = []byte{0xFC}
var terminator = []byte{0xFF}
var errorTerminator = []byte{0xFD}

func (h header) writeFiller(w io.Writer) error {
	err := h.write(w)
	if err != nil {
		return err
	}
	_, err = w.Write(terminator)
	return err
}

func (h header) write(w io.Writer) error {
	_, err := w.Write(actionSignal)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err = enc.Encode(h)
	return err
}

type issuedCall struct {
	call       *deckCall
	invocation *action.Invocation
}

func (i *issuedCall) write(h header, w io.Writer) error {
	header := append(h, i.invocation)
	err := header.write(w)
	if err != nil {
		return err
	}
	payload := i.call.payload()
	if _, ok := payload.(common.WritableNone); ok {
		_, err = w.Write(terminator)
		return err
	}
	_, err = w.Write(continueWithPayload)
	if err != nil {
		return err
	}
	err = payload.Write(w)
	if err != nil {
		return err
	}
	_, err = w.Write(terminator)
	return err
}

type result struct {
	output json.RawMessage
	err    error
}

func (r *result) UnmarshalJSON(data []byte) error {
	var a [2]json.RawMessage
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	var e *string
	err = json.Unmarshal(a[1], &e)
	if err != nil {
		return err
	}
	if e != nil {
		r.err = errors.New(*e)
		return nil
	}
	r.output = a[0]
	return nil
}

type gap struct {
	start uint64
	end   uint64
}

func (m *gap) UnmarshalJSON(data []byte) error {
	var parts []json.RawMessage
	err := json.Unmarshal(data, &parts)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return errors.New("empty result array")
	}
	err = json.Unmarshal(parts[0], &m.start)
	if err != nil {
		return err
	}
	if len(parts) > 1 {
		err = json.Unmarshal(parts[1], &m.end)
		if err != nil {
			return err
		}
		return nil
	} else {
		m.end = m.start
	}
	return nil
}

type report struct {
	Gaps    []gap             `json:"Gaps"`
	Results map[uint64]result `json:"results"`
}

type deckCall struct {
	call     action.Call
	reported atomic.Bool
	params   action.CallParams
}

func (p *deckCall) written() {
	if !p.params.Optimistic {
		return
	}
	p.result([]byte("null"), nil)
}

func (c *deckCall) payload() common.Writable {
	return c.call.Payload()
}

func (c *deckCall) action() (action.Action, bool) {
	return c.call.Action()
}

func (c *deckCall) clean() {
	c.call.Clean()
}

func (c *deckCall) cancel() {
	if c.reported.Swap(true) {
		return
	}
	c.call.Cancel()
}

func (c *deckCall) result(ok json.RawMessage, err error) {
	if c.reported.Swap(true) {
		return
	}
	c.call.Result(ok, err)
}
