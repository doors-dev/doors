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
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/front/action"
)

func (c *core[M]) SimpleCall(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func()) context.CancelFunc {
	if ctx.Err() != nil {
		if onCancel != nil {
			onCancel()
		}
		return func() {}
	}
	done := ctxwg.Add(ctx)
	ctx, cancel := context.WithCancel(context.Background())
	call := &SimpleCall{
		ctx:      ctx,
		done:     done,
		action:   action,
		onResult: onResult,
		onCancel: onCancel,
	}
	c.Call(call)
	return cancel
}

type SimpleCall struct {
	ctx      context.Context
	action   action.Action
	done     func()
	onResult func(json.RawMessage, error)
	onCancel func()
}

func (c *SimpleCall) Action() (action.Action, bool) {
	if c.ctx.Err() != nil {
		return nil, false
	}
	return c.action, true
}

func (C *SimpleCall) Payload() common.Writable {
	return common.WritableNone{}
}

func (c SimpleCall) Cancel() {
	defer c.done()
	if c.onCancel == nil {
		return
	}
	c.onCancel()
}
func (c *SimpleCall) Result(r json.RawMessage, err error) {
	if err != nil {
		slog.Error("Call failed", slog.String("action", c.action.Log()), slog.String("error", err.Error()))
	}
	defer c.done()
	if c.onResult == nil {
		return
	}
	if err != nil {
		c.onResult(r, errors.Join(errors.New("execution error"), err))
		return
	}
	c.onResult(r, err)
}

/*
type setPathCall struct {
	path     string
	replace  bool
	canceled atomic.Bool
}

func (t *setPathCall) cancel() {
	t.canceled.Store(true)
}

func (t *setPathCall) Data() *common.CallData {
	if t.canceled.Load() {
		return nil
	}
	return &common.CallData{
		Name:    "set_path",
		Arg:     t.arg(),
		Payload: common.WritableNone{},
	}
}

func (t *setPathCall) arg() []any {
	return []any{t.path, t.replace}
}

func (t *setPathCall) Result(json.RawMessage, error) {}
func (t *setPathCall) Cancel()                       {}

type LocatinReload struct {
}

func (l *LocatinReload) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_reload",
		Arg:     []any{},
		Payload: common.WritableNone{},
	}
}
func (t *LocatinReload) Result(json.RawMessage, error) {}
func (t *LocatinReload) Cancel()                       {}

type LocationReplace struct {
	Href   string
	Origin bool
}

func (l *LocationReplace) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_replace",
		Arg:     []any{l.Href, l.Origin},
		Payload: common.WritableNone{},
	}
}
func (t *LocationReplace) Result(json.RawMessage, error) {}
func (t *LocationReplace) Cancel()                       {}

type LocationAssign struct {
	Href   string
	Origin bool
}

func (l *LocationAssign) Data() *common.CallData {
	return &common.CallData{
		Name:    "location_assign",
		Arg:     []any{l.Href, l.Origin},
		Payload: common.WritableNone{},
	}
}

func (t *LocationAssign) Result(json.RawMessage, error) {}
func (t *LocationAssign) Cancel()                       {}

func NewSimpleCall(ctx context.Context, actor action.Actor, onResult func(json.RawMessage, error), onCancel func()) (action.Call, context.CancelFunc) {
	inst := ctx.Value(common.CtxKeyInstance).(Core)
	callCtx, cancel := context.WithCancel(context.Background())
	call := &SimpleCall{
		ctx:    callCtx,
		name:   name,
		arg:    arg,
		done:   ctxwg.Add(ctx),
		cancel: cancel,
	}
}

type SimpleCall struct {
	ctx      context.Context
	inst     Core
	actor
	done     func()
	cancel   context.CancelFunc
	onResult func(json.RawMessage, error)
	onCancel func()
}

func (c *SimpleCall[OUTPUT]) Result(r json.RawMessage, err error) {
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

func (c *SimpleCall[OUTPUT]) Data() *common.CallData {

	if c.ctx.Err() != nil {
		return nil
	}
	return &common.CallData{
		Name:    c.name,
		Arg:     c.arg,
		Payload: common.WritableNone{},
	}
}

func (c *SimpleCall[OUTPUT]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{c.Name, c.Arg})
} */
