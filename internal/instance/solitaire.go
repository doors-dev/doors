// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
)

type solitaireInstance interface {
	syncError(error)
	touch()
}

func newSolitaire(inst solitaireInstance, conf *common.SolitaireConf) *solitaire {
	return &solitaire{
		inst: inst,
		conf: conf,
		deck: newDeck(inst, conf.Queue, conf.Pending, conf.SyncTimeout),
	}
}

type solitaire struct {
	inst   solitaireInstance
	conf   *common.SolitaireConf
	deck   *deck
	buffer atomic.Pointer[conn]
	conn   atomic.Pointer[conn]
}

func (s *solitaire) Call(call action.Call) {
	err := s.deck.Insert(call)
	if err != nil {
		return
	}
	c := s.conn.Load()
	c.Trigger()
}

func (s *solitaire) End(cause endCause) {
	defer s.deck.End()
	conn := s.conn.Swap(nil)
	if conn == nil {
		return
	}
	conn.End(cause)
}

func (s *solitaire) Connect(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	rep := &report{}
	err := decoder.Decode(rep)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	counter, err := s.deck.OnReport(rep)
	if err != nil {
		return
	}
	if counter > 0 {
		s.inst.touch()
	}
	conn := newConn(s.conf, w, r, s.deck)
	if conn.ack() != nil {
		return
	}
	s.buffer.Swap(conn)
	active := s.conn.Load()
	ch := active.Roll()
	<-ch
	ok := s.buffer.CompareAndSwap(conn, nil)
	if !ok {
		conn.Roll()
		conn.Run()
		return
	}
	ok = s.conn.CompareAndSwap(active, conn)
	if !ok {
		conn.Roll()
		conn.Run()
		return
	}
	conn.Run()
}

func newConn(conf *common.SolitaireConf, w http.ResponseWriter, r *http.Request, desk *deck) *conn {
	ctx, cancelTimer := context.WithTimeout(r.Context(), conf.Ping*4/3)
	ctx, cancel := context.WithCancelCause(ctx)
	return &conn{
		conf:       conf,
		requestCtx: r.Context(),
		ctx:        ctx,
		cancel: func(cause error) {
			cancel(cause)
			cancelTimer()
		},
		writer: &writer{w: w, f: w.(http.Flusher), sizeLimit: conf.FlushSize, timeLimit: conf.FlushTimeout},
		desk:   desk,
		endCh:  make(chan struct{}),
	}
}

type conn struct {
	conf       *common.SolitaireConf
	requestCtx context.Context
	ctx        context.Context
	cancel     context.CancelCauseFunc
	writer     *writer
	desk       *deck
	trigger    atomic.Pointer[chan struct{}]
	endCh      chan struct{}
}

func (c *conn) wait() {
	ch := c.trigger.Load()
	if ch == nil {
		return
	}
	select {
	case <-*ch:
	case <-c.ctx.Done():
	}
}

func (c *conn) ack() (err error) {
	return c.writer.writeAck()
}
func (c *conn) handleCause() {
	cause := context.Cause(c.ctx)
	switch cause {
	case context.DeadlineExceeded:
	case context.Canceled:
		c.writer.Write(rollSignal)
	case causeKilled:
		c.writer.Write(killSignal)
	case causeSuspend:
		c.writer.Write(suspendSignal)
	case causeSyncError:
	default:
		panic(errors.New("Unknown solitaire connection cancel cause"))
	}
}

func (c *conn) Run() {
	defer c.writer.flush()
	defer c.handleCause()
	defer c.cleanup()
	wait := false
	zombie := false
	start := time.Now()
	for c.ctx.Err() == nil {
		for c.requestCtx.Err() == nil {
			writeResult, _ := c.desk.WriteNext(c.writer)
			if writeResult == writeErr {
				return
			}
			if writeResult == writeSyncErr {
				return
			}
			if writeResult == nothingToWrite {
				zombie = true
				c.writer.flush()
				if wait {
					wait = false
					c.wait()
				} else {
					ch := make(chan struct{})
					c.trigger.Store(&ch)
					wait = true
				}
			}
			if writeResult == pendingLimit {
				zombie = true
				c.writer.flush()
				<-c.ctx.Done()
			}
			if writeResult == writeOk {
				if c.writer.toFlush {
					c.writer.flush()
				}
			}
			if !zombie && time.Since(start) > c.conf.RollDuration {
				zombie = true
			}
			if zombie {
				break
			}
		}
	}
}

func (c *conn) cleanup() {
	c.cancel(nil)
	close(c.endCh)
}

func (c *conn) Trigger() {
	if c == nil {
		return
	}
	if c.ctx.Err() != nil {
		return
	}
	ch := c.trigger.Swap(nil)
	if ch != nil {
		close(*ch)
	}
}

func (c *conn) Roll() <-chan struct{} {
	if c == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	c.cancel(nil)
	return c.endCh
}
func (c *conn) End(cause endCause) {
	if c == nil {
		return
	}
	c.cancel(cause)
}

type writer struct {
	sizeLimit   int
	timeLimit   time.Duration
	lastFlushed time.Time
	total       int
	f           http.Flusher
	w           http.ResponseWriter
	after       []func()
	toFlush     bool
}

func (w *writer) afterFlush(f func()) {
	w.after = append(w.after, f)
}

func (w *writer) flush() {
	if w.total == 0 {
		return
	}
	w.toFlush = false
	w.f.Flush()
	w.total = 0
	w.lastFlushed = time.Now()
	for _, f := range w.after {
		f()
	}
	w.after = nil
}

var writerError = errors.New("actual write error")

func (w *writer) writeAck() error {
	_, err := w.w.Write(ackSignal)
	if err != nil {
		return writerError
	}
	w.f.Flush()
	return nil
}
func (w *writer) Write(data []byte) (int, error) {
	size, err := w.w.Write(data)
	if err != nil {
		return 0, writerError
	}
	w.total += size
	if w.lastFlushed.IsZero() {
		w.lastFlushed = time.Now()
	}
	if w.total >= w.sizeLimit || time.Since(w.lastFlushed) >= w.timeLimit {
		w.toFlush = true
	}
	return size, nil
}
