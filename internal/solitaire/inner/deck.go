// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package inner

type head interface {
	setBottom()
	setTail(*Card)
}

type Deck struct {
	top    *Card
	bottom *Card
	count  int
}

func (q *Deck) Len() int {
	return q.count
}

func (q *Deck) Cut() *Card {
	if q.top == nil {
		return nil
	}
	top := q.top
	if !top.IsFiller() {
		q.count -= 1
	}
	if top.tail == nil {
		q.setBottom()
	} else {
		q.top = top.tail
	}
	return top
}

func (q *Deck) firstInsert(c *Card) bool {
	if q.top != nil && q.bottom != nil {
		return false
	}
	if q.bottom != nil || q.top != nil {
		panic("has top but no bottom or otherwise")
	}
	q.bottom = c
	q.top = c
	return true
}

func (q *Deck) Probe(seq uint64) {
	probe := newProbeCard(seq, q)
	if q.firstInsert(probe) {
		return
	}
	q.bottom.insertTail(probe)
}

func (q *Deck) Fill(beg uint64, end uint64) {
	if q.firstInsert(newRangeFillerCard(beg, end, q)) {
		return
	}
	q.top.fill(beg, end, q)
}

func (q *Deck) Insert(c *Card) {
	if c.IsFiller() {
		panic("cannot insert filler")
	}
	c.deck = q
	q.count += 1
	if q.firstInsert(c) {
		return
	}
	q.top.insert(c, q)
}

func (q *Deck) Cancel() {
	if q.top == nil {
		return
	}
	q.top.cancel()
}

func (q *Deck) IsCold(seq uint64) bool {
	if q.bottom == nil {
		return true
	}
	return q.bottom.Seq() != seq
}

func (q *Deck) ExtractRestored(seq uint64) (*Card, error) {
	if q.top == nil {
		return nil, nil
	}
	card, err := q.top.extractRestored(seq, q)
	if card != nil {
		q.count -= 1
	}
	return card, err
}

func (q *Deck) Append(c *Card) {
	c.deck = q
	q.count += 1
	if q.firstInsert(c) {
		return
	}
	q.bottom.insertTail(c)
}

func (q *Deck) setBottom() {
	if q.top != q.bottom {
		panic("expected only for the last card")
	}
	q.top = nil
	q.bottom = nil
}

func (q *Deck) storeBottom(c *Card) {
	q.bottom = c
}

func (q *Deck) swapBottom(prev *Card, new *Card) {
	if q.bottom != prev {
		return
	}
	q.bottom = new
}

func (q *Deck) setTail(c *Card) {
	q.top = c
}
