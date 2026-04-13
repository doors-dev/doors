// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// AShared is a reusable dynamic attribute handle shared across every element it
// is attached to.
type AShared = *aShared

var _ Attr = &aShared{}

// NewAShared returns an enabled shared attribute handle.
//
// Update, Enable, and Disable affect every attached element together.
func NewAShared(name string, value string) AShared {
	return &aShared{
		name:   name,
		value:  value,
		enable: true,
	}
}

type aShared struct {
	mu          sync.Mutex
	name        string
	value       string
	enable      bool
	id          uint64
	initialized bool
	seq         int
}

func (a *aShared) check(seq int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.seq == seq
}

func (a *aShared) restore(seq int, value string, enable bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if seq != a.seq {
		return
	}
	a.seq -= 1
	a.enable = enable
	a.value = value
}

// Enable adds the attribute to attached elements.
func (a AShared) Enable(ctx context.Context) {
	a.updateEnable(ctx, true)
}

// Disable removes the attribute from attached elements.
func (a AShared) Disable(ctx context.Context) {
	a.updateEnable(ctx, false)
}

func (a *aShared) updateEnable(ctx context.Context, enable bool) {
	ctex.LogCanceled(ctx, "shared attribute enable")
	a.mu.Lock()
	prevValue := a.value
	prevEnable := a.enable
	if a.enable == enable {
		a.mu.Unlock()
		return
	}
	a.enable = enable
	a.seq += 1
	seq := a.seq
	enabled := a.enable
	initialized := a.initialized
	id := a.id
	value := a.value
	a.mu.Unlock()
	core := ctx.Value(ctex.KeyCore).(core.Core)
	var act action.Action
	if enabled {
		act = &action.DynaSet{
			ID:    id,
			Value: value,
		}
	} else {
		act = &action.DynaRemove{
			ID: id,
		}
	}
	if !initialized {
		return
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
			slog.Error("shared attribute call error", "error", err)
			a.restore(seq, prevValue, prevEnable)
		},
		nil,
		action.CallParams{},
	)
}

// Update sets the attribute's value.
func (a AShared) Update(ctx context.Context, value string) {
	a.mu.Lock()
	ctex.LogCanceled(ctx, "shared attribute value")
	prevValue := a.value
	prevEnable := a.enable
	if ctx.Err() != nil {
		a.mu.Unlock()
		return
	}
	if a.value == value {
		a.mu.Unlock()
		return
	}
	a.value = value
	a.seq += 1
	seq := a.seq
	if !a.enable {
		a.mu.Unlock()
		return
	}
	initialized := a.initialized
	id := a.id
	nextValue := a.value
	a.mu.Unlock()
	if !initialized {
		return
	}
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.CallCheck(
		func() bool {
			return a.check(seq)
		},
		&action.DynaSet{
			ID:    id,
			Value: nextValue,
		},
		func(rm json.RawMessage, err error) {
			if err == nil {
				return
			}
			slog.Error("shared attribute call error", "error", err)
			a.restore(seq, prevValue, prevEnable)
		},
		nil,
		action.CallParams{},
	)
}

func (a AShared) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(a, cur, elem)
}

func (a AShared) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.initialized {
		a.initialized = true
		a.id = core.NewID()
	}
	front.AttrsAppendDyn(attrs, a.id, a.name)
	front.AttrsSetParent(attrs, core.DoorID())
	if a.enable {
		attrs.Get(a.name).Set(a.value)
	}
	return nil
}
