// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/instance"
)

type call[O any] struct {
	inst     instance.Core
	doorId   uint64
	ctx      context.Context
	arg      any
	name     string
	done     func()
	onResult func(O, error)
	onCancel func()
}

func (c *call[O]) Data() *common.CallData {
	if c.ctx.Err() != nil {
		return nil
	}
	return &common.CallData{
		Name:    "call",
		Arg:     []any{c.name, c.arg, c.doorId},
		Payload: common.WritableNone{},
	}
}

func (c call[O]) Cancel() {
	defer c.done()
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}

func (c *call[O]) Result(r json.RawMessage, err error) {
	if err != nil {
		slog.Error("Call failed", slog.String("call_name", c.name), slog.String("error", err.Error()))
	}
	if c.onResult == nil {
		c.done()
		return
	}
	ok := c.inst.Spawn(func() {
		defer c.done()
		var output O
		if err != nil {
			c.onResult(output, errors.Join(errors.New("execution error"), err))
			return
		}
		err = json.Unmarshal(r, &output)
		if err != nil {
			c.onResult(output, errors.Join(errors.New("result unmarshal error"), err))
			return
		}
		c.onResult(output, err)
	})
	if !ok {
		c.onCancel()
	}
}

// Call invokes a JavaScript handler previously registered on the client
// with $d.on(name, fn). The argument is marshaled to JSON and passed to
// the handler. The handler’s return value is unmarshaled into OUTPUT and
// delivered asynchronously to onResult.
//
// onResult and onCancel are optional and may be nil. If provided,
// onResult is invoked with the decoded result or an error, and onCancel
// is invoked if the call is canceled or cannot be scheduled.
//
// The returned CancelFunc attempts to cancel the call, but cancellation
// is best-effort and not guaranteed if the call is already in progress.
//
// If provied context is already canceled, onCancel invoked immediately.
//
// Example:
//
//	cancel := doors.Call[string](ctx, "my_js_call", "Hello from Go",
//		func(out string, err error) {
//			if err != nil { log.Println("error:", err); return }
//			log.Println("reply:", out)
//		},
//		func() { log.Println("canceled") },
//	)
//
func Call[OUTPUT any](ctx context.Context, name string, arg any, onResult func(OUTPUT, error), onCancel func()) context.CancelFunc {
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	doorId := ctx.Value(common.CtxKeyDoor).(door.Core).Id()
	callCtx, cancel := context.WithCancel(context.Background())
	c := &call[OUTPUT]{
		done:     ctxwg.Add(ctx),
		inst:     inst,
		name:     name,
		doorId:   doorId,
		ctx:      callCtx,
		arg:      arg,
		onCancel: onCancel,
		onResult: onResult,
	}
	if ctx.Err() != nil {
		c.Cancel()
		return cancel
	}
	inst.Call(c)
	return cancel

}

// Fire performs JavaScrtipt call via simplified API (without handlers).
// Please refer to doors.Call[OUTPUT] for details.
//
//
func Fire(ctx context.Context, name string, arg any) context.CancelFunc {
	return Call[json.RawMessage](ctx, name, arg, nil, nil)
}

