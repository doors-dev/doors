package inner

import "errors"

func NewRestoredCard(seq uint64, c *Call) *Card {
	return &Card{
		Beg:      seq,
		End:      seq,
		Call:     c,
		restored: true,
	}
}

func NewCard(seq uint64, c *Call) *Card {
	return &Card{
		Beg:  seq,
		End:  seq,
		Call: c,
	}
}

func newRangeFillerCard(beg uint64, end uint64, d deck) *Card {
	return &Card{
		Beg:  beg,
		End:  end,
		Call: nil,
		deck: d,
	}
}

type deck interface {
	swapBottom(*Card, *Card)
	storeBottom(*Card)
}

type Card struct {
	Beg      uint64
	End      uint64
	Call     *Call
	tail     *Card
	restored bool
	deck     deck
}

func (c *Card) setTail(n *Card) {
	c.tail = n
	c.deck.swapBottom(c, c.tail)
}

func (c *Card) setBottom() {
	c.deck.storeBottom(c)
	c.tail = nil
}

func (c *Card) insert(n *Card, h head) {
	if n.IsFiller() {
		panic("Cannot insert filler")
	}
	if c.Beg <= n.Seq() && n.Seq() <= c.End {
		panic("overlapping range")
	}
	if n.Seq() > c.Seq() {
		if c.tail == nil {
			c.setTail(n)
			return
		}
		c.tail.insert(n, c)
		return
	}
	n.setTail(c)
	h.setTail(n)
}

func (c *Card) cancel() {
	if !c.IsFiller() {
		c.Call.Cancel()
	}
	if c.tail == nil {
		return
	}
	c.tail.cancel()
}

func (c *Card) extractRestored(seq uint64, h head) (*Card, error) {
	if c.Seq() == seq {
		if c.IsFiller() {
			return nil, nil
		}
		if !c.restored {
			return nil, errors.New("Attempt to respond to non issued card")
		}
		if c.tail != nil {
			h.setTail(c.tail)
		} else {
			h.setBottom()
		}
		return c, nil
	}
	if seq < c.Seq() || c.tail == nil {
		return nil, nil
	}
	return c.tail.extractRestored(seq, c)
}

func (c *Card) Seq() uint64 {
	return c.End
}

func (c *Card) IsFiller() bool {
	return c.Call == nil
}

func (c *Card) insertTail(n *Card) {
	if n.IsFiller() {
		panic("Cannot insert filler tail")
	}
	if n.Seq() <= c.Seq() {
		panic("Cannot insert older tail")
	}
	c.setTail(n)
}

func (c *Card) fill(beg uint64, end uint64, h head) {
	if end+1 < c.Beg {
		filler := newRangeFillerCard(beg, end, c.deck)
		filler.setTail(c)
		h.setTail(filler)
		return
	}
	if beg <= c.End {
		c.Beg = min(beg, c.Beg)
		if end <= c.End {
			return
		}
		beg = c.End + 1
	}
	if c.IsFiller() && (beg == c.End+1) {
		c.End = end
		if c.tail != nil && c.End >= c.tail.Beg {
			c.tail.Beg = c.Beg
			h.setTail(c.tail)
			if c.End > c.tail.End {
				c.tail.fill(c.tail.End+1, c.End, h)
			}
		}
		return
	}
	if c.tail != nil {
		c.tail.fill(beg, end, c)
		return
	}
	c.setTail(newRangeFillerCard(beg, end, c.deck))
}
