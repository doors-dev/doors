// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
)

// CallResult holds the outcome of an XCall.
// Either Ok is set with the result, or Err is non-nil.
type CallResult[T any] struct {
	Ok  T     // Result value
	Err error // Error if the call failed
}

// Call dispatches action to the client and returns a best-effort cancel
// function.
func Call(ctx context.Context, action Action) context.CancelFunc {
	_, cancel := call[json.RawMessage](ctx, action)
	return cancel
}

// XCall dispatches action to the client and returns a result channel.
//
// The channel closes without a value if the call is canceled. Do not wait on
// it during rendering. If you need to wait, do it in a hook, inside [Go], or
// in your own goroutine with [Free]. T is the expected decoded payload type.
// For actions other than [ActionEmit], [json.RawMessage] is usually the right
// choice.
func XCall[T any](ctx context.Context, action Action) (<-chan CallResult[T], context.CancelFunc) {
	ctex.LogFreeWarning(ctx, "action", "XCall")
	return call[T](ctx, action)
}

func call[T any](ctx context.Context, action Action) (<-chan CallResult[T], context.CancelFunc) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	ch := make(chan CallResult[T], 1)
	a, params, err := action.action(ctx, core, !core.Conf().SolitaireDisableGzip)
	res := CallResult[T]{}
	if err != nil {
		slog.Error("Action preparation error", "error", err)
		res.Err = err
		ch <- res
		close(ch)
		return ch, func() {}
	}
	if ctx.Err() != nil {
		ctex.LogCanceled(ctx, "call "+a.Log())
	}
	cancel := core.CallCtx(
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
