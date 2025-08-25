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
	"io"
	"log/slog"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/door"
)

type ADyn interface {
	Attr
	Value(ctx context.Context, value string)
	Enable(ctx context.Context, enable bool)
}

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
	call := &dynaCall{
		seq:        a.seq,
		attr:       a,
		prevValue:  prevValue,
		prevEnable: prevEnable,
	}
	if a.enable {
		call.data = &common.CallData{
			Name:    "dyna_set",
			Arg:     []any{a.id, a.value},
			Payload: common.WritableNone{},
		}
	} else {
		call.data = &common.CallData{
			Name:    "dyna_remove",
			Arg:     a.id,
			Payload: common.WritableNone{},
		}
	}
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
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
	call := &dynaCall{
		seq:  a.seq,
		attr: a,
		data: &common.CallData{
			Name:    "dyna_set",
			Arg:     []any{a.id, a.value},
			Payload: common.WritableNone{},
		},
		prevValue:  prevValue,
		prevEnable: prevEnable,
	}
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
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
	seq        int
	attr       *aDyn
	data       *common.CallData
	prevValue  string
	prevEnable bool
}

func (c *dynaCall) Data() *common.CallData {
	if c.seq != c.attr.getSeq() {
		return nil
	}
	return c.data
}

func (c *dynaCall) Result(err error) {
	if err == nil {
		return
	}
	slog.Error("Dynamic attribute call err " + err.Error())
	c.attr.restore(c.seq, c.prevValue, c.prevEnable)
}
