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
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/solitaire/expirator"
)

type Solitaire = *solitaire

type Instance interface {
	SyncError(error)
	Touch()
}

func NewSolitaire(inst Instance, conf *common.SolitaireConf) Solitaire {
	s := &solitaire{
		inst: inst,
		conf: conf,
	}
	expirator := expirator.NewExpirator(s)
	s.deck = newDeck(expirator, conf.Queue, conf.Pending, conf.SyncTimeout)
	return s
}

type solitaire struct {
	inst Instance
	conf *common.SolitaireConf
	deck *deck
	conn atomic.Pointer[con]
}

func (s Solitaire) Call(call action.Call) {
	err := s.deck.Insert(call)
	if err != nil {
		s.inst.SyncError(err)
		return
	}
	c := s.conn.Load()
	c.Trigger()
}

func (d *solitaire) Expire() {
	d.inst.SyncError(errors.New("sync timeout"))
}

func (s Solitaire) End(cause common.EndCause) {
	defer s.deck.End()
	conn := s.conn.Swap(nil)
	if conn == nil {
		return
	}
	conn.End(cause)
}

func (s Solitaire) Connect(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	rep := report{}
	err := decoder.Decode(&rep)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	count, err := s.deck.CollectResults(rep.Results)
	if err != nil {
		if err == context.Canceled {
			return
		}
		s.inst.SyncError(err)
		return
	}
	if count > 0 {
		s.inst.Touch()
	}
	wr := &writer{w: w, f: w.(http.Flusher), sizeLimit: s.conf.FlushSize, timeLimit: s.conf.FlushTimeout}
	if err := wr.WriteAck(); err != nil {
		return
	}
	ctx, cancelTimer := context.WithTimeout(r.Context(), s.conf.Ping*4/3)
	ctx, cancelCause := context.WithCancelCause(ctx)
	cancel := func(cause error) {
		cancelCause(cause)
		cancelTimer()
	}
	con := newCon(s.inst, s.deck, wr, ctx, cancel, rep.Gaps)
	prev := s.conn.Swap(con)
	<-prev.Roll()
	con.Run()
}

func newCon(inst Instance, d *deck, w *writer, ctx context.Context, cancel func(error), gaps []gap) *con {
	c := &con{
		inst:     inst,
		writer:   w,
		ctx:      ctx,
		cancel:   cancel,
		endGuard: make(chan struct{}),
		deck:     d,
		gaps:     gaps,
	}
	c.zombie.Store(len(gaps) == 0)
	return c
}

type con struct {
	writer   *writer
	ctx      context.Context
	cancel   func(error)
	zombie   atomic.Bool
	endGuard chan struct{}
	deck     *deck
	inst     Instance
	trigger  atomic.Pointer[chan struct{}]
	gaps     []gap
}

func (c *con) Context() context.Context {
	return c.ctx
}

func (c *con) Roll() <-chan struct{} {
	if c == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	swapped := c.zombie.CompareAndSwap(false, true)
	if swapped {
		c.Trigger()
	} else {
		c.cancel(nil)
	}
	return c.endGuard
}

func (c *con) Trigger() {
	if c == nil {
		return
	}
	ch := c.trigger.Swap(nil)
	if ch != nil {
		close(*ch)
	}
}

func (c *con) End(cause common.EndCause) {
	if c == nil {
		return
	}
	c.cancel(cause)
}

func (c *con) Run() {
	defer c.writer.Flush()
	defer c.handleCause()
	defer c.Cleanup()

	if c.ctx.Err() != nil {
		return
	}

	reportGaps := len(c.gaps) != 0
	if reportGaps {
		if err := c.deck.FillGaps(c.gaps); err != nil {
			if err == context.Canceled {
				return
			}
			c.inst.SyncError(err)
			return
		}
		c.gaps = nil
	}

	c.deck.HeatUp()

	drained := false
	armed := false
	for c.ctx.Err() == nil {
		result, syncErr := c.deck.WriteNext(c.writer)
		switch result {
		case writeContinue:
			continue
		case writeSyncErr:
			c.inst.SyncError(syncErr)
			return
		case writeErr, writeKilled:
			return
		case writeLimit:
			c.writer.Flush()
			if reportGaps && c.zombie.Swap(true) {
				return
			}
			<-c.ctx.Done()
			return
		case writeOk:
			c.writer.TryFlush()
			armed = false
		case writeNothing:
			drained = true
			c.writer.Flush()
			if !armed {
				armed = true
				c.arm()
			} else {
				armed = false
				c.wait()
			}
		}
		if drained && reportGaps && c.zombie.Load() {
			return
		}
	}
}

func (c *con) arm() {
	ch := make(chan struct{})
	c.trigger.Store(&ch)
}

func (c *con) wait() {
	ch := c.trigger.Load()
	if ch == nil {
		return
	}
	select {
	case <-*ch:
	case <-c.ctx.Done():
	}
}

func (c *con) handleCause() {
	cause := context.Cause(c.ctx)
	switch cause {
	case context.DeadlineExceeded:
	case context.Canceled:
		c.writer.Write(rollSignal)
	case common.EndCauseKilled:
		c.writer.Write(killSignal)
	case common.EndCauseSuspend:
		c.writer.Write(suspendSignal)
	case common.EndCauseSyncError:
	default:
		panic(errors.New("unknown solitaire connection cancel cause"))
	}
}

func (c *con) Cleanup() {
	c.cancel(nil)
	close(c.endGuard)
}
