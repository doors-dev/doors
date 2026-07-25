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

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
	"github.com/doors-dev/doors/internal/solitaire/inner"
)

func newDeck(expirator *expirator.Expirator, conf *common.SolitaireConf) *deck {
	d := &deck{
		expirator: expirator,
		issued:    make(map[uint64]*inner.Card),
		conf:      conf,
	}
	return d
}

type deck struct {
	conf         *common.SolitaireConf
	seq          uint64
	issued       map[uint64]*inner.Card
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
		issued.Call.Cancel()
	}
	d.inner.Cancel()
}

func (d *deck) PendingCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.issued)
}
func (d *deck) QueueLength() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.inner.Len()
}

func (d *deck) Pending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.issued) != 0
}

func (d *deck) Restore(beg uint64, end uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return nil
	}
	issued, ok := d.issued[end]
	if !ok {
		d.inner.Fill(beg, end)
		return nil
	}
	delete(d.issued, end)
	if err := d.restore(end, issued.Call); err != nil {
		return err
	}
	if beg != end {
		d.inner.Fill(beg, end-1)
	}
	return nil
}

var errorKilled = errors.New("killed")
var errorLimit = errors.New("limit")

func (d *deck) Dump(s Stasher) (err error) {
	for err == nil {
		d.mu.Lock()
		if d.killed {
			d.mu.Unlock()
			return errorKilled
		}
		if len(d.issued) >= d.conf.Pending {
			err = errorLimit
		}
		card := d.inner.Cut()
		d.mu.Unlock()
		if card == nil {
			return
		}
		switch s.Stash(card) {
		case stashFiller:
		case stashCancel:
			d.expirator.Report(card.End)
			card.Call.Cancel()
			d.mu.Lock()
			d.inner.Fill(card.Beg, card.End)
			d.mu.Unlock()
		case stashIssue:
			d.mu.Lock()
			d.issued[card.End] = card
			d.mu.Unlock()
			card.Call.Written()
		}
		if s.Full() {
			return
		}
	}
	return
}

type bufferedResult struct {
	call   *inner.Call
	result result
}

func (b bufferedResult) process() {
	b.call.Result(b.result.output, b.result.err)
}

func (d *deck) CollectResults(r map[uint64]result) error {
	buffer := make([]bufferedResult, 0, len(r))
	defer func() {
		for _, r := range buffer {
			r.process()
		}
	}()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.killed {
		return context.Canceled
	}
	for seq, result := range r {
		if seq > d.seq {
			return errors.New("ready overflows last seq")
		}
		d.latestReport = max(d.latestReport, seq)
		issued, ok := d.issued[seq]
		if !ok {
			restored, err := d.inner.ExtractRestored(seq)
			if err != nil {
				return err
			}
			if restored == nil {
				continue
			}
			d.expirator.Report(seq)
			buffer = append(buffer, bufferedResult{restored.Call, result})
			continue
		}
		delete(d.issued, seq)
		d.expirator.Report(seq)
		buffer = append(buffer, bufferedResult{issued.Call, result})
	}
	return nil
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
	tolarance := min(uint64(d.conf.Pending), d.seq)
	prevEnd := max(d.latestReport, tolarance) - tolarance
	for _, gap := range g {
		if gap.end < gap.beg {
			return errors.New("gap range issue")
		}
		if gap.end > d.seq {
			return errors.New("gap overflows last seq")
		}
		if prevEnd >= gap.beg {
			return errors.New("gap overlap")
		}
		prevEnd = gap.end
		beg := gap.beg
		for seq := max(gap.beg, d.latestReport); seq <= gap.end; seq++ {
			card, ok := d.issued[seq]
			if !ok {
				continue
			}
			if beg != seq {
				d.inner.Fill(beg, seq-1)
			}
			beg = seq + 1
			delete(d.issued, seq)
			if err := d.restore(seq, card.Call); err != nil {
				return err
			}
		}
		if beg <= gap.end {
			d.inner.Fill(beg, gap.end)
		}
	}
	return nil
}

func (d *deck) Insert(c actions.Call) (err error) {
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
		dc.Params.Timeout = d.conf.SyncTimeout
	}
	card := inner.NewCard(d.seq, dc)
	d.inner.Append(card)
	deadline := time.Now().Add(dc.Params.Timeout)
	d.expirator.Track(d.seq, deadline)
	return d.checkQueueLength()
}

func (d *deck) checkQueueLength() error {
	if d.inner.Len()+len(d.issued) < d.conf.Queue {
		return nil
	}
	return errors.New("call queue limit reached")
}

func (d *deck) restore(seq uint64, c *inner.Call) error {
	n := inner.NewRestoredCard(seq, c)
	d.inner.Insert(n)
	return d.checkQueueLength()

}
