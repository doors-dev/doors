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
	"io"
	"log/slog"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/instance"
)

// ADyn is a dynamic attribute that can be updated at runtime.
type ADyn interface {
	Attr
	// Value sets the attribute's value.
	Value(ctx context.Context, value string)
	// Enable adds or removes the attribute.
	Enable(ctx context.Context, enable bool)
}

// NewADyn returns a new dynamic attribute with the given name, value, and state.
func NewADyn(name string, value string, enable bool) ADyn {
	return &aDyn{
		name:   name,
		value:  value,
		enable: enable,
	}
}

type aDyn struct {
	mu          sync.Mutex
	name        string
	value       string
	enable      bool
	id          uint64
	initialized bool
	seq         int
}

func (a *aDyn) getSeq() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.seq
}

func (a *aDyn) restore(seq int, value string, enable bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if seq != a.seq {
		return
	}
	a.seq -= 1
	a.enable = enable
	a.value = value
}

func (a *aDyn) Enable(ctx context.Context, enable bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	prevValue := a.value
	prevEnable := a.enable
	if ctx.Err() != nil {
		return
	}
	if a.enable == enable {
		return
	}
	a.enable = enable
	a.seq += 1
	if !a.initialized {
		return
	}
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	call := &dynaCall{
		done:       ctxwg.Add(ctx),
		seq:        a.seq,
		attr:       a,
		prevValue:  prevValue,
		prevEnable: prevEnable,
		optimistic: inst.Conf().OptimisicSync,
	}
	if a.enable {
		call.action = &action.DynaSet{
			Id:    a.id,
			Value: a.value,
		}
	} else {
		call.action = &action.DynaRemove{
			Id: a.id,
		}
	}
	inst.Call(call)
}

func (a *aDyn) Value(ctx context.Context, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	prevValue := a.value
	prevEnable := a.enable
	if ctx.Err() != nil {
		return
	}
	if a.value == value {
		return
	}
	a.value = value
	a.seq += 1
	if !a.enable {
		return
	}
	if !a.initialized {
		return
	}
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	call := &dynaCall{
		done: ctxwg.Add(ctx),
		seq:  a.seq,
		attr: a,
		action: &action.DynaSet{
			Id:    a.id,
			Value: a.value,
		},
		prevValue:  prevValue,
		prevEnable: prevEnable,
		optimistic: inst.Conf().OptimisicSync,
	}
	inst.Call(call)
}

func (a *aDyn) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.initialized {
		a.initialized = true
		a.id = inst.NewId()
	}
	attrs.AppendDyna(a.id, a.name)
	if a.enable {
		attrs.Set(a.name, a.value)
	}
}

func (a *aDyn) Attr() AttrInit {
	return a
}

func (a *aDyn) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, a)
}

type dynaCall struct {
	done       func()
	seq        int
	attr       *aDyn
	action     action.Action
	prevValue  string
	prevEnable bool
	optimistic bool
}

func (c *dynaCall) Params() action.CallParams {
	return action.CallParams{
		Optimistic: c.optimistic,
	}
}
func (c *dynaCall) Clean() {
}

func (c *dynaCall) Cancel() {
	c.done()
}

func (c *dynaCall) Action() (action.Action, bool) {
	if c.seq != c.attr.getSeq() {
		return nil, false
	}
	return c.action, true
}

func (c *dynaCall) Payload() common.Writable {
	return common.WritableNone{}
}

func (c *dynaCall) Result(_ json.RawMessage, err error) {
	defer c.done()
	if err == nil {
		return
	}
	slog.Error("Dynamic attribute call err " + err.Error())
	c.attr.restore(c.seq, c.prevValue, c.prevEnable)
}
