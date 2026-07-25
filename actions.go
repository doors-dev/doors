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
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/actions"
)

type action struct {
	action actions.Action
	params actions.CallParams
	decode func(json.RawMessage) error
}

// Action performs a client-side operation.
type Action interface {
	action(ctx context.Context, core core.Core, gz bool) (action, error)
}

// ActionInto is an [Action] with a typed client result.
//
// Into returns an action that unmarshals the result into dst. Use it with
// [Call]: dst is valid after the returned channel delivers nil. Inside action
// lists such as Before, After, or OnError there is no result transport, so
// Into has no effect there.
type ActionInto[T any] interface {
	Action
	Into(dst *T) Action
}

type actionFunc func(ctx context.Context, core core.Core, gz bool) (action, error)
type actionIntoFunc[T any] func(ctx context.Context, core core.Core, gz bool) (action, error)

func (a actionFunc) action(ctx context.Context, core core.Core, gz bool) (action, error) {
	return a(ctx, core, gz)
}

func (a actionIntoFunc[T]) Into(dst *T) Action {
	return into((actionFunc)(a), dst)
}

func (a actionIntoFunc[T]) action(ctx context.Context, core core.Core, gz bool) (action, error) {
	return a(ctx, core, gz)
}

var _ Action = actionFunc(nil)
var _ ActionInto[any] = actionIntoFunc[any](nil)

func into[T any](a Action, dst *T) Action {
	return actionFunc(func(ctx context.Context, core core.Core, gz bool) (action, error) {
		prep, err := a.action(ctx, core, gz)
		if err != nil {
			return prep, err
		}
		prep.decode = func(rm json.RawMessage) error {
			return json.Unmarshal(rm, dst)
		}
		return prep, nil
	})
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

func intoActions(ctx context.Context, as []Action) actions.Actions {
	core := ctx.Value(common.KeyCore).(core.Core)
	arr := make(actions.Actions, 0)
	for _, a := range as {
		prep, err := a.action(ctx, core, false)
		if err != nil {
			core.App().Logger().Error("Action preparation error", "error", err)
			continue
		}
		arr = append(arr, prep.action)
	}
	return arr
}

type joinedActions []Actions

func (s joinedActions) Actions() []Action {
	output := make([]Action, 0)
	for _, a := range s {
		if a == nil {
			continue
		}
		output = append(output, a.Actions()...)
	}
	return output
}

func (s joinedActions) And(a Actions) Actions {
	c := make(joinedActions, len(s), len(s)+1)
	copy(c, s)
	c = append(c, a)
	return c
}

var _ Actions = joinedActions(nil)

// ActionEmit invokes a client-side handler registered with
// `$on(name, handler)`.
//
// T is the handler's result type; capture it with Into (see [ActionInto]).
// For fire-and-forget, use ActionEmit[any] and dispatch without Into.
type ActionEmit[T any] struct {
	Name string
	Arg  any
}

func (ae ActionEmit[T]) Actions() []Action {
	return []Action{ae}
}

func (ae ActionEmit[T]) And(a Actions) Actions {
	return joinedActions([]Actions{ae, a})
}

// Into returns an action that unmarshals the handler result into dst; see
// [ActionInto].
func (ae ActionEmit[T]) Into(dst *T) Action {
	return into(ae, dst)
}

func (ae ActionEmit[T]) action(ctx context.Context, core core.Core, gz bool) (action, error) {
	payload, err := actions.IntoPayload(ae.Arg, gz)
	if err != nil {
		return action{}, err
	}
	return action{
		action: actions.Emit{
			Name:    ae.Name,
			DoorID:  core.Door().ID(),
			Payload: payload,
		},
	}, nil
}

var _ Actions = ActionEmit[any]{}

// ActionLocationReload reloads the current page.
//
// It performs a full page load and starts a fresh instance: use it for error
// recovery, session resets, or picking up a new deployment — not for in-app
// navigation.
type ActionLocationReload struct{}

func (ar ActionLocationReload) Actions() []Action {
	return []Action{ar}
}

func (ar ActionLocationReload) And(a Actions) Actions {
	return joinedActions([]Actions{ar, a})
}

func (ar ActionLocationReload) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	return action{
		action: &actions.LocationReload{},
		params: actions.CallParams{Timeout: core.App().Conf().InstanceTTL, Optimistic: true},
	}, nil
}

// ActionLocationReplace replaces the current history entry with a model-derived
// URL.
//
// It performs a full page load. For in-app navigation, use [ALink] or update
// the route source instead.
type ActionLocationReplace struct {
	Model any
}

func (ar ActionLocationReplace) Actions() []Action {
	return []Action{ar}
}

func (ar ActionLocationReplace) And(a Actions) Actions {
	return joinedActions([]Actions{ar, a})
}

func (a ActionLocationReplace) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	l, err := NewLocation(a.Model)
	if err != nil {
		return action{}, err
	}
	return action{
		action: &actions.LocationReplace{
			URL:    l.String(),
			Origin: true,
		},
		params: actions.CallParams{Timeout: core.App().Conf().InstanceTTL, Optimistic: true},
	}, nil
}

// ActionLocationAssign navigates to a model-derived URL.
//
// It performs a full page load. For in-app navigation, use [ALink] or update
// the route source instead.
type ActionLocationAssign struct {
	Model any
}

func (aa ActionLocationAssign) Actions() []Action {
	return []Action{aa}
}

func (aa ActionLocationAssign) And(a Actions) Actions {
	return joinedActions([]Actions{aa, a})
}

func (aa ActionLocationAssign) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	l, err := NewLocation(aa.Model)
	if err != nil {
		return action{}, err
	}
	return action{
		action: &actions.LocationAssign{
			URL:    l.String(),
			Origin: true,
		},
		params: actions.CallParams{Timeout: core.App().Conf().InstanceTTL, Optimistic: true},
	}, nil
}

// ActionLocationRawAssign navigates to url without first encoding a path model.
//
// URL must be absolute. It is the escape hatch to other origins (OAuth,
// external pages); it performs a full page load. For in-app navigation, use
// [ALink] or update the route source instead.
type ActionLocationRawAssign struct {
	URL string
}

func (aa ActionLocationRawAssign) Actions() []Action {
	return []Action{aa}
}

func (aa ActionLocationRawAssign) And(a Actions) Actions {
	return joinedActions([]Actions{aa, a})
}

func (a ActionLocationRawAssign) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	return action{
		action: &actions.LocationAssign{
			URL:    a.URL,
			Origin: false,
		},
		params: actions.CallParams{Timeout: core.App().Conf().InstanceTTL, Optimistic: true},
	}, nil
}

// ActionLocationRawReplace replaces the current history entry with url without
// first encoding a path model.
//
// URL must be absolute. Like [ActionLocationRawAssign] it can leave the
// current origin, but the current page does not stay in history — use it for
// redirects the user should not navigate back to, such as OAuth flows.
type ActionLocationRawReplace struct {
	URL string
}

func (aa ActionLocationRawReplace) Actions() []Action {
	return []Action{aa}
}

func (aa ActionLocationRawReplace) And(a Actions) Actions {
	return joinedActions([]Actions{aa, a})
}

func (a ActionLocationRawReplace) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	return action{
		action: &actions.LocationReplace{
			URL:    a.URL,
			Origin: false,
		},
		params: actions.CallParams{Timeout: core.App().Conf().InstanceTTL, Optimistic: true},
	}, nil
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
	return joinedActions([]Actions{aa, a})
}

func (a ActionScroll) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	return action{
		action: actions.Scroll{
			Selector: a.Selector,
			Options:  a.Options,
		},
	}, nil
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
	return joinedActions([]Actions{ai, a})
}

func (ai ActionIndicate) action(ctx context.Context, core core.Core, _ bool) (action, error) {
	return action{
		action: actions.Indicate{
			Indicate: indicatorsOrNil(ai.Indicator),
			Duration: ai.Duration,
		},
	}, nil
}
