package common

import (
	"context"
	"fmt"
	"io"
	"sync"
)

type Event interface {
	Name() string
	WriteData(w io.Writer) error
}

var eventTerminator []byte = []byte("\n\n")

func WriteEvent(e Event, w io.Writer) error {
	_, err := fmt.Fprintf(w, "event: %s\ndata: ", e.Name())
	if err != nil {
		return err
	}
	err = e.WriteData(w)
	if err != nil {
		return err
	}
	_, err = w.Write(eventTerminator)
	return err
}

type SuspendEvent struct{}

func (e SuspendEvent) Name() string {
	return "suspend"
}
func (e SuspendEvent) WriteData(io.Writer) error {
	return nil
}

type UnauthorizedEvent struct{}

func (e UnauthorizedEvent) Name() string {
	return "unauthorized"
}
func (e UnauthorizedEvent) WriteData(io.Writer) error {
	return nil
}

type GoneEvent struct{}

func (e GoneEvent) Name() string {
	return "gone"
}
func (e GoneEvent) WriteData(io.Writer) error {
	return nil
}

type EventSender interface {
	Close()
	Tx(Event) bool
}

type EventReceiver interface {
	Rx() (Event, bool)
}

func NewEventChannel(ctx context.Context) (EventSender, EventReceiver) {
	ec := &eventChannel{
		ctx:   ctx,
		ch:    make(chan Event, 1),
		close: sync.Once{},
	}
	return ec, ec
}

type eventChannel struct {
	ctx   context.Context
	ch    chan Event
	close sync.Once
}

func (ec *eventChannel) Close() {
	ec.close.Do(func() {
		close(ec.ch)
	})
}

func (ec *eventChannel) Tx(e Event) bool {
	if ec.ctx.Err() != nil {
		return false
	}
	select {
	case <-ec.ctx.Done():
		return false
	case ec.ch <- e:
		return true
	}
}

func (ec *eventChannel) Rx() (Event, bool) {
	if ec.ctx.Err() != nil {
		return nil, false
	}
	select {
	case <-ec.ctx.Done():
		return nil, false
	case event, ok := <-ec.ch:
		return event, ok
	}
}
