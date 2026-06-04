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
/*

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/solitaire/inner"
)

type writer struct {
	sizeLimit   int
	timeLimit   time.Duration
	lastFlushed time.Time
	total       int
	f           http.Flusher
	w           http.ResponseWriter
	toFlush     bool
}

func (w *writer) Flush() {
	if w.total == 0 {
		return
	}
	w.toFlush = false
	w.f.Flush()
	w.total = 0
	w.lastFlushed = time.Now()
}

func (w *writer) TryFlush() {
	if !w.toFlush {
		return
	}
	w.Flush()
}

var writerError = errors.New("actual write error")

func (w *writer) WriteAck() error {
	_, err := w.w.Write(signalFlush)
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

func newHeader(startSeq uint64, endSeq uint64) header {
	if startSeq == endSeq {
		return []any{endSeq}
	}
	return []any{[]uint64{startSeq, endSeq}}

}

type header []any


func (h header) writeFiller(w io.Writer) error {
	err := h.write(w, action.PayloadNone, 0)
	if err != nil {
		return err
	}
	return err
}

func (h header) write(w io.Writer, payloadType action.PayloadType, payloadLength int) error {
	if _, err := w.Write(signalAction); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	var payloadInfo []any = nil
	if payloadType != action.PayloadNone {
		payloadInfo = []any{payloadType, payloadLength}
	}
	if err := enc.Encode(append(h, payloadInfo)); err != nil {
		return err
	}
	if _, err := w.Write(terminator); err != nil {
		return err
	}
	return nil
}

type issuedCall struct {
	call       *inner.Call
	invocation action.Invocation
}

func (i *issuedCall) write(h header, w io.Writer) error {
	h = append(h, i.invocation.Func())
	payload := i.invocation.Payload()
	if err := h.write(w, payload.Type(), payload.Len()); err != nil {
		return err
	}
	if err := payload.Output(w); err != nil {
		return err
	}
	return nil
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

func (c *deckCall) action() (action.Action, bool) {
	return c.call.Action()
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
} */
