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
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/instance"
)

// CallResult holds the outcome of an XCall.
// Either Ok is set with the result, or Err is non-nil.
type CallResult[T any] struct {
	Ok  T     // Result value
	Err error // Error if the call failed
}

// Call dispatches an action to the client side.
// Returns a cancel function to abort the call (best-effort).
func Call(ctx context.Context, action Action) context.CancelFunc {
	_, cancel := call[json.RawMessage](ctx, action)
	return cancel
}

// XCall dispatches an action to the client side and returns a result channel.
// The channel is closed without a value if the call is canceled.
// Cancellation is best-effort. Wait on the channel only in contexts
// where blocking is allowed (hooks, goroutines).
// The output value is unmarshaled into type T.
// For all actions, except ActionEmit, use json.RawMessage as T.
func XCall[T any](ctx context.Context, action Action) (<-chan CallResult[T], context.CancelFunc) {
	common.LogBlockingWarning(ctx, "action", "XCall")
	return call[T](ctx, action)
}

func call[T any](ctx context.Context, action Action) (<-chan CallResult[T], context.CancelFunc) {
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	door := ctx.Value(common.CtxKeyDoor).(door.Core)
	ch := make(chan CallResult[T], 1)
	a, params, err := action.action(ctx, inst, door)
	res := CallResult[T]{}
	if err != nil {
		slog.Error("Action preparation errror", slog.String("error", err.Error()))
		res.Err = err
		ch <- res
		close(ch)
		return ch, func() {}
	}
	if ctx.Err() != nil {
		close(ch)
		slog.Error("Call attempt from the canceled context", slog.String("action", a.Log()))
		return ch, func() {}
	}
	cancel := inst.CallCtx(
		ctx,
		a,
		func(rm json.RawMessage, err error) {
			if err != nil {
				res.Err = err
			} else {
				res.Err = json.Unmarshal(rm, &res.Ok)
			}
			ch <- res
			close(ch)
		},
		func() {
			close(ch)
		},
		params,
	)
	return ch, cancel
}

/*
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
} */
