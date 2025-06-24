package instance

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
)
func newDeck(conf *common.SolitaireConf) *deck {
	return &deck{
		conf:   conf,
		issued: make(map[uint64]*issuedCall),
	}
}

type deck struct {
	conf         *common.SolitaireConf
	seq          uint64
	top          *card
	bottom       *card
	issued       map[uint64]*issuedCall
	mu           sync.Mutex
	deckSize     uint
	latestReport uint64
}

func (d *deck) Pending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.issued) != 0
}

type writeResult int

const (
	nothingToWrite writeResult = iota
	pendingLimit 
	writeOk
	writeErr
	writeSyncErr
)

func (d *deck) WriteNext(w io.Writer) (writeResult, error) {
	for {
		d.mu.Lock()
		if len(d.issued) == d.conf.Pending {
			d.mu.Unlock()
			return pendingLimit, nil
		}
		card := d.cutTop()
		d.mu.Unlock()
		if card == nil {
			return nothingToWrite, nil
		}
		header := newHeader(card.startSeq, card.endSeq)
		if card.isFiller() {
			if err := header.writeOnly(w); err != nil {
				return writeErr, err
			}
			return writeOk, nil
		}
		call, isRelevant := card.caller.Call()
		if !isRelevant {
			d.mu.Lock()
			d.skipRange(card.startSeq, card.endSeq)
			d.mu.Unlock()
			continue
		}
		issuedCall := &issuedCall{
			call:   call,
			caller: card.caller,
		}
		err := issuedCall.write(header, w)
		if err != nil {
			isRelevant = call.OnWriteErr()
			d.mu.Lock()
			if !isRelevant {
				d.skipRange(card.startSeq, card.endSeq)
			} else {
				if err := d.cancelCut(card); err != nil {
					return writeSyncErr, err
				}
			}
			d.mu.Unlock()
			return writeErr, err
		}
		d.mu.Lock()
		d.issued[card.seq()] = issuedCall
		d.mu.Unlock()
		return writeOk, nil
	}
}

func (d *deck) extractRestored(seq uint64) *card {
	if d.top == nil {
		return nil
	}
	card := d.top.extractRestored(seq, d)
	if card != nil {
		d.dec()
	}
	return card
}

func (d *deck) OnReport(s *report) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for seq := range s.Results {
		if seq > d.seq {
			return errors.New("ready overflows last seq")
		}
		if d.latestReport < seq {
			d.latestReport = seq
		}
		var resultErr error = nil
		if s.Results[seq] != nil {
			resultErr = errors.New(*s.Results[seq])
		}
		call, ok := d.issued[seq]
		if !ok {
			card := d.extractRestored(seq)
			if card == nil {
				continue
			}
			card.call.OnResult(resultErr)
			continue
		}
		delete(d.issued, seq)
		call.call.OnResult(resultErr)
	}
	if len(s.Gaps) == 0 {
		return nil
	}
	first := s.Gaps[0]
	last := s.Gaps[len(s.Gaps)-1]
	if last.end >= d.seq {
		return errors.New("gap overflows last seq")
	}
	if first.start <= d.latestReport {
		return errors.New("gap after report")
	}
	prevEnd := uint64(0)
	for _, gap := range s.Gaps {
		if gap.end < gap.start {
			return errors.New("gap range issue")
		}
		if prevEnd >= gap.start {
			return errors.New("gap overlap")
		}
		for seq := gap.start; seq <= gap.end; seq++ {
			call, ok := d.issued[seq]
			if !ok {
				d.skipSeq(seq)
				continue
			}
			isRelevant := call.call.OnWriteErr()
			if isRelevant {
				if err := d.restore(seq, call.caller, call.call); err != nil {
					return err
				}
			} else {
				d.skipSeq(seq)
			}
		}
	}
	return nil
}

func (d *deck) Insert(caller node.Caller) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seq += 1
	node := newCard(d.seq, caller)
	d.insertTail(node)
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
	if d.deckSize > d.conf.Queue {
		return errors.New("call queue limit reached")
	}
	return nil
}
func (d *deck) dec() {
	d.deckSize -= 1
}

func (d *deck) restore(seq uint64, caller node.Caller, call node.Call) error {
	n := newRestoredCard(seq, caller, call)
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
	d.dec()
	return card
}

func (d *deck) cancelCut(n *card) error {
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

func newRestoredCard(seq uint64, caller node.Caller, call node.Call) *card {
	return &card{
		startSeq: seq,
		endSeq:   seq,
		caller:   caller,
		call:     call,
	}
}

func newCard(seq uint64, caller node.Caller) *card {
	return &card{
		startSeq: seq,
		endSeq:   seq,
		caller:   caller,
		call:     nil,
	}
}
func newFillerCard(seq uint64) *card {
	return &card{
		startSeq: seq,
		endSeq:   seq,
		caller:   nil,
		call:     nil,
	}
}

type card struct {
	startSeq uint64
	endSeq   uint64
	caller   node.Caller
	tail     *card
	call     node.Call
}

func (s *card) extractRestored(seq uint64, h head) *card {
	if s.seq() == seq {
		if s.call != nil {
			slog.Error("Attempt to respond to non issued card")
			return nil
		}
		h.setTail(s.tail)
		return s
	}
	if seq < s.seq() || s.tail == nil {
		return nil
	}
	return s.tail.extractRestored(seq, s)
}

func (s *card) seq() uint64 {
	return s.endSeq
}

func (sn *card) isFiller() bool {
	return sn.caller == nil
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
		h.setTail(newFillerCard(seq))
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
		return []common.JsonWritable{common.JsonWritableAny{endSeq}}
	}
	return []common.JsonWritable{common.JsonWritableAny{endSeq}, common.JsonWritableAny{startSeq}}

}

type header []common.JsonWritable

var terminator = []byte{0xFF}

func (h header) writeOnly(w io.Writer) error {
	err := h.write(w)
	if err != nil {
		return err
	}
	_, err = w.Write(terminator)
	return err
}

func (h header) write(w io.Writer) error {
	buffer := &bytes.Buffer{}
	err := common.JsonWritables(common.JsonWritables(h)).WriteJson(buffer)
	if err != nil {
		panic("Json writable is not writable")
	}
	length := uint32(buffer.Len())
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return err
	}
	_, err = w.Write(buffer.Bytes())
	return err
}

type issuedCall struct {
	caller node.Caller
	call   node.Call
}

func (i *issuedCall) write(h header, w io.Writer) error {
	header := append(h, common.JsonWritableAny{i.call.Name()}, i.call.Arg())
	err := header.write(w)
	if err != nil {
		return err
	}
	payload, has := i.call.Payload()
	if has {
		err = payload.Write(w)
		if err != nil {
			return err
		}
	}
	_, err = w.Write(terminator)
	return err
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

type ready struct {
	seq uint64
	err error
}

func (m *ready) UnmarshalJSON(data []byte) error {
	var parts []json.RawMessage
	err := json.Unmarshal(data, &parts)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return errors.New("empty result array")
	}
	err = json.Unmarshal(parts[0], &m.seq)
	if err != nil {
		return err
	}
	if len(parts) > 1 {
		var message string
		err = json.Unmarshal(parts[1], &message)
		if err != nil {
			return err
		}
		m.err = errors.New(message)
		return nil
	}
	return nil
}

type report struct {
	Gaps    []gap              `json:"gaps"`
	Results map[uint64]*string `json:"results"`
}
