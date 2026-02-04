package door

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/doors-dev/doors/internal/front/action"
)

type reportHook uint64

func (c reportHook) Params() action.CallParams {
	return action.CallParams{}
}

func (c reportHook) Action() (action.Action, bool) {
	return action.ReportHook{HookId: uint64(c)}, true
}

func (c reportHook) Cancel() {}

func (c reportHook) Result(r json.RawMessage, err error) {}

type callKind int

const (
	callReplace callKind = iota
	callUpdate
)

type call struct {
	ctx     context.Context
	ch      chan error
	kind    callKind
	id      uint64
	payload *printer
}

func (n *call) Cancel() {
	n.payload.release()
	n.send(context.Canceled)
}

func (n *call) Result(_ json.RawMessage, err error) {
	n.payload.release()
	if err != nil {
		slog.Error("door rendering error", slog.String("error", err.Error()))
	}
	n.send(err)
}

func (n *call) send(err error) {
	if n.ch == nil {
		return
	}
	n.ch <- err
	close(n.ch)
}

func (c *call) Action() (action.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	payload, payloadType := c.payload.payload()
	switch c.kind {
	case callReplace:
		return action.DoorReplace{
			ID:          c.id,
			Payload:     payload,
			PayloadType: payloadType,
		}, true
	case callUpdate:
		return action.DoorUpdate{
			ID:          c.id,
			Payload:     payload,
			PayloadType: payloadType,
		}, true
	default:
		panic("unsupporte door call type")
	}
}

func (c *call) Params() action.CallParams {
	return action.CallParams{}
}
