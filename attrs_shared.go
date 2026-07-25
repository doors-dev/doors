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
	"sync"

	"github.com/doors-dev/doors/internal/common"
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

// sharedClosed returns an already-closed empty channel, used when no client
// call is issued (nothing to update, superseded, or canceled).
func sharedClosed() <-chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

// Enable adds the attribute to attached elements.
func (a AShared) Enable(ctx context.Context) {
	a.updateEnable(ctx, true)
}

// Disable removes the attribute from attached elements.
func (a AShared) Disable(ctx context.Context) {
	a.updateEnable(ctx, false)
}

// XEnable is like [aShared.Enable] but returns a channel reporting the client
// outcome. The channel receives nil on success or an error on failure, then
// closes; it closes without a value when no call is issued or the call is
// superseded. Do not wait on it during rendering.
func (a AShared) XEnable(ctx context.Context) <-chan error {
	return a.updateEnable(ctx, true)
}

// XDisable is like [aShared.Disable] but returns a channel reporting the client
// outcome. See [aShared.XEnable] for the channel contract.
func (a AShared) XDisable(ctx context.Context) <-chan error {
	return a.updateEnable(ctx, false)
}

func (a *aShared) updateEnable(ctx context.Context, enable bool) <-chan error {
	ctex.LogCanceled(ctx, "shared attribute enable")
	a.mu.Lock()
	prevValue := a.value
	prevEnable := a.enable
	if a.enable == enable {
		a.mu.Unlock()
		return sharedClosed()
	}
	a.enable = enable
	a.seq += 1
	seq := a.seq
	enabled := a.enable
	initialized := a.initialized
	id := a.id
	value := a.value
	a.mu.Unlock()
	if !initialized {
		return sharedClosed()
	}
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
	return a.dispatch(ctx, seq, act, prevValue, prevEnable)
}

// Update sets the attribute's value.
func (a AShared) Update(ctx context.Context, value string) {
	a.update(ctx, value)
}

// XUpdate is like [aShared.Update] but returns a channel reporting the client
// outcome. See [aShared.XEnable] for the channel contract.
func (a AShared) XUpdate(ctx context.Context, value string) <-chan error {
	return a.update(ctx, value)
}

func (a *aShared) update(ctx context.Context, value string) <-chan error {
	a.mu.Lock()
	ctex.LogCanceled(ctx, "shared attribute value")
	prevValue := a.value
	prevEnable := a.enable
	if ctx.Err() != nil {
		a.mu.Unlock()
		return sharedClosed()
	}
	if a.value == value {
		a.mu.Unlock()
		return sharedClosed()
	}
	a.value = value
	a.seq += 1
	seq := a.seq
	if !a.enable {
		a.mu.Unlock()
		return sharedClosed()
	}
	initialized := a.initialized
	id := a.id
	nextValue := a.value
	a.mu.Unlock()
	if !initialized {
		return sharedClosed()
	}
	return a.dispatch(ctx, seq, &action.DynaSet{
		ID:    id,
		Value: nextValue,
	}, prevValue, prevEnable)
}

// dispatch issues the client call and returns a channel that receives its
// outcome (nil on success, error on failure) then closes; a superseded or
// canceled call closes the channel without a value.
func (a *aShared) dispatch(ctx context.Context, seq int, act action.Action, prevValue string, prevEnable bool) <-chan error {
	core := ctx.Value(common.KeyCore).(core.Core)
	logger := core.App().Logger()
	ch := make(chan error, 1)
	core.Door().UserCall(
		ctx,
		func() bool {
			return a.check(seq)
		},
		act,
		func(rm json.RawMessage, err error) {
			if err != nil {
				logger.Error("shared attribute call error", "error", err)
				a.restore(seq, prevValue, prevEnable)
			}
			ch <- err
			close(ch)
		},
		func() {
			close(ch)
		},
		action.CallParams{},
	)
	return ch
}

func (a AShared) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(a, cur, elem)
}

func (a AShared) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.initialized {
		a.initialized = true
		a.id = core.Instance().NewID()
	}
	front.AttrsAppendDyn(attrs, a.id, a.name)
	front.AttrsSetParent(attrs, core.Door().ID())
	if a.enable {
		attrs.Get(a.name).Set(a.value)
	}
	return nil
}
