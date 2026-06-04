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
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

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
		sync: frameSyncer{
			Conf: conf,
		},
	}
	expirator := expirator.NewExpirator(s)
	s.deck = newDeck(expirator, conf)
	return s
}

type solitaire struct {
	inst     Instance
	conf     *common.SolitaireConf
	deck     *deck
	sync     frameSyncer
	sender   atomic.Pointer[sender]
	receiver atomic.Pointer[receiver]
}

func (s Solitaire) Call(call action.Call) {
	err := s.deck.Insert(call)
	if err != nil {
		s.inst.SyncError(err)
		return
	}
	c := s.sender.Load()
	c.Trigger()
}

func (d *solitaire) Expire() {
	d.inst.SyncError(errors.New("sync timeout"))
}

func (s Solitaire) End(cause common.EndCause) {
	defer s.deck.End()
	conn := s.sender.Swap(nil)
	if conn == nil {
		return
	}
	conn.End(cause)
}

func (s Solitaire) Connect(w http.ResponseWriter, r *http.Request) {
	ctx, cancelTimer := context.WithTimeout(r.Context(), s.conf.Roll*4/3)
	ctx, cancelCause := context.WithCancelCause(ctx)
	cancel := func(cause error) {
		cancelCause(cause)
		cancelTimer()
	}
	w.Header().Set("Cache-Control", "no-store")
	if r.Header.Get("Accept") == "application/octet-stream" {
		s.inst.Touch()
		fw := &writeController{
			Sync:    &s.sync,
			Writer:  w,
			Flusher: w.(http.Flusher),
			Conf:    s.conf,
		}
		sender := newSender(s.inst, s.deck, fw, ctx, cancel, s.conf)
		prev := s.sender.Swap(sender)
		if prev != nil {
			<-prev.Roll()
		}
		sender.Run()
		return
	}
	receiver := newReceiver(s.inst, s.deck, &s.sync, r, w, ctx, cancel, s.conf)
	prev := s.receiver.Swap(receiver)
	if prev != nil {
		<-prev.Roll()
	}
	receiver.Run()
}

func newReceiver(inst Instance, d *deck, s *frameSyncer, r *http.Request, w http.ResponseWriter, ctx context.Context, cancel func(error), conf *common.SolitaireConf) *receiver {
	rv := &receiver{
		conn: conn{
			ctx:      ctx,
			cancel:   cancel,
			endGuard: make(chan struct{}),
			deck:     d,
			inst:     inst,
			conf:     conf,
		},
		sync:     s,
		r:        r,
		w:        w,
		readChan: make(chan task),
	}
	return rv
}

type receiver struct {
	mu sync.Mutex
	conn
	timeoutCtx context.Context
	sync       *frameSyncer
	r          *http.Request
	w          http.ResponseWriter
	readChan   chan task
}

func (rv *receiver) Run() {
	defer rv.Cleanup()
	go rv.reader()
	if rv.r.Header.Get("Content-Type") == "application/octet-stream" {
		rv.readStream()
		return
	}
	rep := report{}
	select {
	case err := <-rv.decode(&rep):
		if err != nil {
			rv.w.WriteHeader(http.StatusBadRequest)
			return
		}
		rv.onReport(rep)
	case <-time.After(rv.conf.ReportTimeout):
		rv.w.WriteHeader(http.StatusRequestTimeout)
		return
	}
}

func (rv *receiver) readStream() {
	var lengthBuffer [4]byte
	var reportBuffer []byte
	for rv.ctx.Err() == nil {
		select {
		case err := <-rv.read(lengthBuffer[:]):
			if err != nil {
				return
			}
		case <-rv.conn.ctx.Done():
			rv.r.Body.Close()
			return
		}
		length := binary.BigEndian.Uint32(lengthBuffer[:])
		if length > uint32(rv.conf.ReportSize) {
			rv.inst.SyncError(errors.New("solitaire report is too long"))
			rv.w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		if uint32(cap(reportBuffer)) < length {
			reportBuffer = make([]byte, length)
		} else {
			reportBuffer = reportBuffer[:length]
		}
		select {
		case err := <-rv.read(reportBuffer):
			if err != nil {
				return
			}
		case <-time.After(rv.conf.ReportTimeout):
			rv.r.Body.Close()
			rv.w.WriteHeader(http.StatusRequestTimeout)
			return
		}
		rep := report{}
		if err := json.Unmarshal(reportBuffer, &rep); err != nil {
			rv.w.WriteHeader(http.StatusBadRequest)
			return
		}
		rv.onReport(rep)
	}
}

func (rv *receiver) read(buffer []byte) <-chan error {
	ch := make(chan error, 1)
	rv.readChan <- task{
		resp: ch,
		read: buffer,
	}
	return ch
}

func (rv *receiver) decode(rep *report) <-chan error {
	ch := make(chan error, 1)
	rv.readChan <- task{
		resp:   ch,
		decode: rep,
	}
	return ch
}

type task struct {
	resp   chan<- error
	read   []byte
	decode *report
}

func (rv *receiver) reader() {
	for {
		var task task
		select {
		case <-rv.conn.endGuard:
			return
		case task = <-rv.readChan:
		}
		if task.read != nil {
			_, err := io.ReadFull(rv.r.Body, task.read)
			task.resp <- err
			continue
		}
		decoder := json.NewDecoder(rv.r.Body)
		task.resp <- decoder.Decode(task.decode)
	}
}

func (rv *receiver) onReport(rep report) {
	rv.sync.Report(rep.ID, rep.TS)
	err := rv.deck.CollectResults(rep.Results)
	if err != nil {
		if err == context.Canceled {
			return
		}
		rv.inst.SyncError(err)
		return
	}
	err = rv.deck.FillGaps(rep.Gaps)
	if err != nil {
		if err == context.Canceled {
			return
		}
		rv.inst.SyncError(err)
		return
	}
}

func newSender(inst Instance, d *deck, w *writeController, ctx context.Context, cancel func(error), conf *common.SolitaireConf) *sender {
	s := &sender{
		conn: conn{
			ctx:      ctx,
			cancel:   cancel,
			endGuard: make(chan struct{}),
			deck:     d,
			inst:     inst,
			conf:     conf,
		},
		writeController: w,
	}
	return s
}

type sender struct {
	conn
	writeController *writeController
	trigger         atomic.Pointer[chan struct{}]
}

func (c *sender) Trigger() {
	if c == nil {
		return
	}
	ch := c.trigger.Swap(nil)
	if ch != nil {
		close(*ch)
	}
}

func (c *sender) End(cause common.EndCause) {
	if c == nil {
		return
	}
	c.cancel(cause)
}

func (c *sender) submit(signal signal) error {
	h, err := c.writeController.Submit(signal)
	if err == nil {
		return nil
	}
	for _, h := range h {
		c.deck.Restore(h.beg, h.end)
	}
	return err
}

func (c *sender) penalty() {
	select {
	case <-time.After(c.conf.MaxRTT):
	case <-c.ctx.Done():
	}
}

func (c *sender) Run() {
	defer c.handleCause()
	defer c.Cleanup()
	armed := false
	for c.ctx.Err() == nil {
		if c.writeController.Hot() {
			if err := c.submit(signalSync); err != nil {
				return
			}
		}
		before := c.writeController.Len()
		err := c.deck.Dump(c.writeController)
		if err != nil {
			if errors.Is(err, errorKilled) {
				return
			}
			if errors.Is(err, errorLimit) {
				c.penalty()
				continue
			}
			panic(err)
		}
		if before != c.writeController.Len() {
			continue
		}
		if !armed {
			armed = true
			c.arm()
		} else {
			armed = c.wait()
		}
	}
}

func (c *sender) arm() {
	ch := make(chan struct{})
	c.trigger.Store(&ch)
}

func (c *sender) wait() bool {
	ch := c.trigger.Load()
	if ch == nil {
		return false
	}
	select {
	case <-c.writeController.Wait(c.deck.Pending()):
		if c.writeController.Len() == 0 {
			c.deck.HeatUp()
		}
		return true
	case <-*ch:
		return false
	case <-c.ctx.Done():
		return false
	}
}

func (c *sender) handleCause() {
	cause := context.Cause(c.ctx)
	switch cause {
	case context.DeadlineExceeded, context.Canceled:
		c.submit(signalRoll)
	case common.EndCauseKilled:
		c.submit(signalKill)
	case common.EndCauseSuspend:
		c.submit(signalSuspend)
	case common.EndCauseSyncError:
	default:
		panic(errors.New("unknown solitaire connection cancel cause"))
	}
}

type conn struct {
	ctx      context.Context
	cancel   func(error)
	endGuard chan struct{}
	deck     *deck
	inst     Instance
	conf     *common.SolitaireConf
}

func (c *conn) Context() context.Context {
	return c.ctx
}

func (c *conn) Roll() <-chan struct{} {
	if c == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	c.cancel(nil)
	return c.endGuard
}

func (c *conn) Cleanup() {
	c.cancel(nil)
	close(c.endGuard)
}
