// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors/internal/common"
)

type solitaireInstance interface {
	syncError(error)
}

func newSolitaire(inst solitaireInstance, conf *common.SolitaireConf) *solitaire {
	return &solitaire{
		conf: conf,
		inst: inst,
		deck: newDeck(conf.Queue, conf.Pending),
	}
}

type solitaire struct {
	conf       *common.SolitaireConf
	inst       solitaireInstance
	deck       *deck
	connection atomic.Pointer[conn]
}

func (s *solitaire) Call(call common.Call) {
	err := s.deck.Insert(call)
	if err != nil {
		s.inst.syncError(err)
		return
	}
	c := s.connection.Load()
	c.Trigger()
}

func (s *solitaire) End(cause endCause) {
	conn := s.connection.Swap(nil)
	if conn == nil {
		return
	}
	conn.End(cause)
}

func (s *solitaire) Connect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	decoder := json.NewDecoder(r.Body)
	rep := &report{}
	err := decoder.Decode(rep)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = s.deck.OnReport(rep)
	if err != nil {
		s.inst.syncError(err)
		return
	}
	c := newConn(s.inst, s.conf, w, r, s.deck, start)
	prev := s.connection.Swap(c)
	prev.Cancel()
	prev.Wait()
	c.Run()
}

func newConn(inst solitaireInstance, conf *common.SolitaireConf, w http.ResponseWriter, r *http.Request, desk *deck, start time.Time) *conn {
	ctx, cancel := context.WithCancelCause(r.Context())
	return &conn{
		conf:    conf,
		inst:    inst,
		ctx:     ctx,
		cancel:  cancel,
		writer:  &writer{w: w},
		flusher: w.(http.Flusher),
		start:   start,
		desk:    desk,
	}
}

const rollSignal uint8 = 0x00
const suspendSignal uint8 = 0x01
const killSignal uint8 = 0x02

type conn struct {
	conf     *common.SolitaireConf
	inst     solitaireInstance
	ctx      context.Context
	cancel   context.CancelCauseFunc
	writer   *writer
	flusher  http.Flusher
	desk     *deck
	start    time.Time
	deadline time.Time
	zombie   bool
	trigger  atomic.Pointer[chan struct{}]
}

type waitResult int

const (
	waitContinue waitResult = iota
	waitReturn
	waitZombie
)

func (c *conn) wait() waitResult {
	ch := c.trigger.Load()
	if ch == nil {
		return waitContinue
	}
	if c.zombie {
		select {
		case <-*ch:
			return waitContinue
		case <-c.ctx.Done():
			return waitReturn
		}
	}
	dur := time.Until(c.deadline)
	if dur <= 0 {
		return waitZombie
	}
	select {
	case <-*ch:
		return waitContinue
	case <-c.ctx.Done():
		return waitReturn
	case <-time.After(dur):
		return waitZombie
	}

}

func (c *conn) handleCause() {
	cause := context.Cause(c.ctx)
	endCause, ok := cause.(endCause)
	if !ok {
		return
	}
	switch endCause {
	case causeKilled:
		c.signal(killSignal)
	case causeSuspend:
		c.signal(suspendSignal)
	case causeSyncError:
		return
	}
	c.flusher.Flush()
}

func (c *conn) Run() {
	ttl := max(min(c.conf.Request-time.Since(c.start), c.conf.RollTime), 0)
	if c.desk.Pending() {
		ttl = min(ttl, c.conf.RollPendingTime)
	}
	c.deadline = time.Now().Add(ttl)
	defer c.handleCause()
	defer c.cleanup()
	var waitChannel *chan struct{} = nil
	sentenced := false
	for c.ctx.Err() == nil {
		if waitChannel != nil {
			c.trigger.Store(waitChannel)
		}
		writeResult, err := c.desk.WriteNext(c.writer)
		if writeResult == writeErr {
			return
		}
		if writeResult == writeSyncErr {
			c.inst.syncError(err)
			return
		}
		if writeResult == nothingToWrite {
			if waitChannel != nil {
				waitChannel = nil
				r := c.wait()
				if r == waitReturn {
					return
				}
				if r == waitContinue || c.zombie {
					continue
				}
				if !c.writer.somethingWritten() && c.desk.Pending() {
					c.desk.Insert(&touch{})
					sentenced = true
					continue
				}
			} else {
				if c.zombie || !c.writer.somethingWritten() {
					ch := make(chan struct{})
					waitChannel = &ch
					continue
				}
			}
		}
		if writeResult == writeOk {
			waitChannel = nil
			if c.zombie || (!sentenced && c.writer.total <= c.conf.RollSize) {
				c.flusher.Flush()
				continue
			}
		}
		if writeResult == pendingLimit && !c.writer.somethingWritten() {
			select {
			case <-time.After(c.conf.RollPendingTime):
			case <-c.ctx.Done():
				return
			}
		}
		c.zombie = true
		err = c.signal(rollSignal)
		if err != nil {
			return
		}
		c.flusher.Flush()
		if writeResult == pendingLimit {
			return
		}
	}
}

func (c *conn) signal(s uint8) error {
	err := signal(s).write(c.writer.w)
	if err != nil {
		return err
	}
	c.flusher.Flush()
	return nil
}

func (c *conn) cleanup() {
	c.cancel(nil)
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

func (c *conn) Cancel() {
	if c == nil {
		return
	}
	c.cancel(nil)
}
func (c *conn) End(cause endCause) {
	if c == nil {
		return
	}
	c.cancel(cause)
}
func (c *conn) Wait() {
	if c == nil {
		return
	}
	<-c.ctx.Done()
}

type writer struct {
	total int
	w     http.ResponseWriter
}

func (w *writer) somethingWritten() bool {
	return w.total != 0
}

func (w *writer) Write(data []byte) (int, error) {
	size, err := w.w.Write(data)
	w.total += size
	return size, err
}

type signal uint8

var zeroLength = []byte{0x00, 0x00, 0x00, 0x00}

func (s signal) write(w io.Writer) error {
	_, err := w.Write(zeroLength)
	if err != nil {
		return nil
	}
	err = binary.Write(w, binary.BigEndian, s)
	return err
}

type touch struct {
	fired bool
}

func (t *touch) Data() *common.CallData {
	if t.fired {
		return nil
	}
	return &common.CallData{
		Name: "touch",
		Arg: []any{},
		Payload: common.WritableNone{},
	}
}
func (t *touch) Result(error) {}
