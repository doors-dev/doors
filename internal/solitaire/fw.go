package solitaire

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/solitaire/inner"
)

type signal byte

const (
	signalSync    signal = 0x00
	signalRoll    signal = 0x02
	signalSuspend signal = 0x03
	signalKill    signal = 0x04
)

const signalAction byte = 0x01

const terminator byte = 0xFF

type writeController struct {
	Sync       *frameSyncer
	Flusher    http.Flusher
	Writer     io.Writer
	Conf       *common.SolitaireConf
	frameStart time.Time
	buffer     bytes.Buffer
	stashed    []header
}

type stashResult int

const (
	stashIssue stashResult = iota
	stashFiller
	stashCancel
)

type Stasher interface {
	Stash(card *inner.Card) stashResult
	Full() bool
}

var _ Stasher = (*writeController)(nil)

func (f *writeController) Len() int {
	return len(f.stashed)
}

func (f *writeController) Penalty() <-chan time.Time {
	return time.After(f.Sync.RTT())
}

func (f *writeController) Wait(pending bool) <-chan time.Time {
	if !pending && f.Len() == 0 {
		var ch chan time.Time
		return ch
	}
	if f.Len() == 0 {
		return time.After(f.Sync.RTT())
	}
	flushTime := time.Until(f.frameStart.Add(f.Conf.FlushTime))
	return time.After(flushTime)
}

func (f *writeController) Submit(signal signal) ([]header, error) {
	if signal == signalKill || signal == signalSuspend {
		f.Writer.Write([]byte{byte(signal)})
		return nil, nil
	}
	if signal == signalSync && f.buffer.Len() == 0 {
		return nil, nil
	}
	if signal == signalRoll {
		if err := f.buffer.WriteByte(byte(signalRoll)); err != nil {
			panic(err)
		}
	}
	if len(f.stashed) != 0 {
		f.frameStart = time.Time{}
		if err := f.buffer.WriteByte(byte(signalSync)); err != nil {
			panic(err)
		}
		if _, err := f.buffer.Write(f.Sync.Collect()); err != nil {
			panic(err)
		}
	}
	if err := f.buffer.WriteByte(terminator); err != nil {
		panic(err)
	}
	if _, err := f.buffer.WriteTo(f.Writer); err != nil {
		h := f.stashed
		f.stashed = nil
		f.buffer.Reset()
		return h, err
	}
	f.stashed = nil
	if signal == signalSync {
		f.Flusher.Flush()
	}
	return nil, nil
}

func (f *writeController) Stash(card *inner.Card) stashResult {
	h := header{
		beg: card.Beg,
		end: card.End,
	}
	if card.IsFiller() {
		f.stash(h)
		if err := h.writeCard(&f.buffer, nil); err != nil {
			panic(err)
		}
		return stashFiller
	}
	action, ok := card.Call.Action()
	if !ok {
		return stashCancel
	}
	f.stash(h)
	invocation := action.Invocation()
	if err := h.writeCard(&f.buffer, &invocation); err != nil {
		panic(err)
	}
	return stashIssue
}

func (f *writeController) stash(h header) {
	if len(f.stashed) == 0 {
		f.frameStart = time.Now()
	}
	f.stashed = append(f.stashed, h)
}

func (f *writeController) Hot() bool {
	if !f.frameStart.IsZero() && time.Since(f.frameStart) >= f.Conf.FlushTime {
		return true
	}
	if f.buffer.Len() >= f.Conf.FrameSize {
		return true
	}
	return false
}

func (f *writeController) Full() bool {
	return f.buffer.Len() >= f.Conf.FrameSize
}

type header struct {
	beg uint64
	end uint64
}

func (h header) writeCard(w *bytes.Buffer, invocation *actions.Invocation) error {
	if err := w.WriteByte(signalAction); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	tuple := make([]any, 0, 4)
	if h.beg == h.end {
		tuple = append(tuple, h.end)
	} else {
		tuple = append(tuple, []any{h.beg, h.end})
	}
	var write *actions.Payload
	if invocation != nil {
		tuple = append(tuple, invocation.Func())
		payload := invocation.Payload()
		if payload.Type() != actions.PayloadNone {
			write = &payload
			tuple = append(tuple, []any{payload.Type(), payload.Len()})
		}
	}
	if write == nil {
		tuple = append(tuple, nil)
	}
	if err := enc.Encode(tuple); err != nil {
		return err
	}
	if err := w.WriteByte(terminator); err != nil {
		return err
	}
	if write == nil {
		return nil
	}
	return write.Output(w)
}
