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
	"log/slog"
	"time"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
)

// Action performs a client-side operation.
type Action interface {
	action(ctx context.Context, core core.Core, gz bool) (action.Action, action.CallParams, error)
}

// Actions is a composable list of client-side operations.
type Actions interface {
	// Actions returns the flattened action list.
	Actions() []Action
	Joiner[Actions]
}

func actionsOrNil(actions Actions) []Action {
	if actions == nil {
		return nil
	}
	return actions.Actions()
}

func intoActions(ctx context.Context, actions []Action) action.Actions {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	arr := make(action.Actions, 0)
	for _, action := range actions {
		a, _, err := action.action(ctx, core, false)
		if err != nil {
			slog.Error("Action preparation error", "error", err)
			continue
		}
		arr = append(arr, a)
	}
	return arr
}

type actions []Actions

func (s actions) Actions() []Action {
	output := make([]Action, 0)
	for _, a := range s {
		if a == nil {
			continue
		}
		output = append(output, a.Actions()...)
	}
	return output
}

func (s actions) And(a Actions) Actions {
	c := make(actions, len(s), len(s)+1)
	copy(c, s)
	c = append(c, a)
	return c
}

var _ Actions = actions(nil)

// ActionEmit invokes a client-side handler registered with
// `$on(name, handler)`.
type ActionEmit struct {
	Name string
	Arg  any
}

func (ae ActionEmit) Actions() []Action {
	return []Action{ae}
}

func (ae ActionEmit) And(a Actions) Actions {
	return actions([]Actions{ae, a})
}

func (ae ActionEmit) action(ctx context.Context, core core.Core, gz bool) (action.Action, action.CallParams, error) {
	payload, err := action.IntoPayload(ae.Arg, gz)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	act := action.Emit{
		Name:    ae.Name,
		DoorID:  core.DoorID(),
		Payload: payload,
	}
	return act, action.CallParams{}, nil
}

var _ Actions = ActionEmit{}

// ActionLocationReload reloads the current page.
type ActionLocationReload struct{}

func (ar ActionLocationReload) Actions() []Action {
	return []Action{ar}
}

func (ar ActionLocationReload) And(a Actions) Actions {
	return actions([]Actions{ar, a})
}

func (ar ActionLocationReload) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return &action.LocationReload{}, action.CallParams{Timeout: core.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionLocationReplace replaces the current history entry with a model-derived
// URL.
type ActionLocationReplace struct {
	Model any
}

func (ar ActionLocationReplace) Actions() []Action {
	return []Action{ar}
}

func (ar ActionLocationReplace) And(a Actions) Actions {
	return actions([]Actions{ar, a})
}

func (a ActionLocationReplace) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	l, err := NewLocation(a.Model)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	return &action.LocationReplace{
		URL:    l.String(),
		Origin: true,
	}, action.CallParams{Timeout: core.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionLocationAssign navigates to a model-derived URL.
type ActionLocationAssign struct {
	Model any
}

func (aa ActionLocationAssign) Actions() []Action {
	return []Action{aa}
}

func (aa ActionLocationAssign) And(a Actions) Actions {
	return actions([]Actions{aa, a})
}

func (aa ActionLocationAssign) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	l, err := NewLocation(aa.Model)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	return &action.LocationAssign{
		URL:    l.String(),
		Origin: true,
	}, action.CallParams{Timeout: core.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionLocationRawAssign navigates to url without first encoding a path model.
type ActionLocationRawAssign struct {
	URL string
}

func (aa ActionLocationRawAssign) Actions() []Action {
	return []Action{aa}
}

func (aa ActionLocationRawAssign) And(a Actions) Actions {
	return actions([]Actions{aa, a})
}

func (a ActionLocationRawAssign) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return &action.LocationAssign{
		URL:    a.URL,
		Origin: false,
	}, action.CallParams{Timeout: core.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionScroll scrolls the first element matching Selector into view.
//
// Options is passed to the browser scroll call as-is and should match the
// shape accepted by `Element.scrollIntoView(...)`, for example:
// `map[string]any{"behavior": "smooth", "block": "center"}`.
type ActionScroll struct {
	// CSS selector for the element to scroll into view.
	Selector string
	// Browser scroll options forwarded to `scrollIntoView(...)`.
	Options any
}

func (aa ActionScroll) Actions() []Action {
	return []Action{aa}
}

func (aa ActionScroll) And(a Actions) Actions {
	return actions([]Actions{aa, a})
}

func (a ActionScroll) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return action.Scroll{
		Selector: a.Selector,
		Options:  a.Options,
	}, action.CallParams{}, nil
}

// ActionIndicate applies indicators for Duration.
type ActionIndicate struct {
	Indicator Indicators
	Duration  time.Duration
}

func (ai ActionIndicate) Actions() []Action {
	return []Action{ai}
}

func (ai ActionIndicate) And(a Actions) Actions {
	return actions([]Actions{ai, a})
}

func (ai ActionIndicate) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return action.Indicate{
		Indicate: indicatorsOrNil(ai.Indicator),
		Duration: ai.Duration,
	}, action.CallParams{}, nil
}
