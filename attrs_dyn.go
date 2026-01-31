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
	"sync"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/gox"
)

// ADyn is a dynamic attribute that can be updated at runtime.
type ADyn = *aDyn

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

func (a *aDyn) check(seq int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.seq == seq
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


// Enable adds or removes the attribute.
func (a ADyn) Enable(ctx context.Context, enable bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ctex.LogCanceled(ctx, "dynamic attribute enable")
	prevValue := a.value
	prevEnable := a.enable
	if a.enable == enable {
		return
	}
	a.enable = enable
	a.seq += 1
	seq := a.seq
	if !a.initialized {
		return
	}
    core := ctx.Value(ctex.KeyCore).(core.Core)
	var act action.Action
	if a.enable {
		act = &action.DynaSet{
			ID:    a.id,
			Value: a.value,
		}
	} else {
		act = &action.DynaRemove{
			ID: a.id,
		}
	}
	core.CallCheck(
		func() bool {
			return a.check(seq)
		},
		act,
		func(rm json.RawMessage, err error) {
			if err == nil {
				return
			}
			slog.Error("Dynamic attribute call err " + err.Error())
			a.restore(seq, prevValue, prevEnable)
		},
		nil,
		action.CallParams{},
	)
}

// Value sets the attribute's value.
func (a ADyn) Value(ctx context.Context, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ctex.LogCanceled(ctx, "dynamic attribute value")
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
	seq := a.seq
	if !a.enable {
		return
	}
	if !a.initialized {
		return
	}
    core := ctx.Value(ctex.KeyCore).(core.Core)
	core.CallCheck(
		func() bool {
			return a.check(seq)
		},
		&action.DynaSet{
			ID:    a.id,
			Value: a.value,
		},
		func(rm json.RawMessage, err error) {
			if err == nil {
				return
			}
			slog.Error("Dynamic attribute call err " + err.Error())
			a.restore(seq, prevValue, prevEnable)
		},
		nil,
		action.CallParams{},
	)
}

func (a ADyn) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	core := ctx.Value(ctex.KeyCore).(core.Core)
	if !a.initialized {
		a.initialized = true
		a.id = core.NewID()
	}
	front.AttrsAppendDyna(attrs, a.id, a.name)
	if a.enable {
		attrs.Get(a.name).Set(a.value)
	}
	return nil
}


