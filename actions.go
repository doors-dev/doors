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

// Action is a client-side effect triggered from Go. Only Doors implements it.
type Action interface {
	action(ctx context.Context, core core.Core, gz bool) (action, error)
}

// ActionInto is an [Action] with a capturable client result.
type ActionInto[T any] interface {
	Action
	// Into returns an action for [Call] that decodes the client result into
	// dst. dst is valid once the completion channel delivers nil.
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

// Actions is a composable list of client-side actions.
type Actions interface {
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

// ActionEmit calls the JavaScript handler registered with $on(name, handler).
//
// Lookup starts at the Door the action is dispatched from and walks out to the
// root, so the nearest matching handler wins. T is the handler's return type;
// capture it with [ActionInto.Into], or use ActionEmit[any] to ignore it.
type ActionEmit[T any] struct {
	// Name is the handler name passed to $on. The action fails when no handler
	// with this name is reachable. Required.
	Name string
	// Arg is the value passed to the handler. A string arrives as a string,
	// []byte as an ArrayBuffer, anything else as decoded JSON. Optional; nil
	// arrives as null.
	Arg any
}

func (ae ActionEmit[T]) Actions() []Action {
	return []Action{ae}
}

func (ae ActionEmit[T]) And(a Actions) Actions {
	return joinedActions([]Actions{ae, a})
}

// Into returns an action that decodes the handler result into dst; see
// [ActionInto.Into].
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
// It is a hard navigation: the page loads again and a fresh instance starts.
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

// ActionLocationReplace replaces the current history entry with a URL built
// from Model.
//
// It is a hard navigation. Unlike [ALink], the page loads again instead of
// rerouting the current instance.
type ActionLocationReplace struct {
	// Model is the target path model value, a [Location], or a
	// [LocationEncoder]. Required.
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

// ActionLocationAssign navigates to a URL built from Model, pushing a new
// history entry.
//
// It is a hard navigation. Unlike [ALink], the page loads again instead of
// rerouting the current instance.
type ActionLocationAssign struct {
	// Model is the target path model value, a [Location], or a
	// [LocationEncoder]. Required.
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

// ActionLocationRawAssign navigates to URL, pushing a new history entry.
//
// Unlike [ActionLocationAssign], the URL is taken as given, so it can point to
// another origin.
type ActionLocationRawAssign struct {
	// URL is the target URL. Must be absolute. Required.
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

// ActionLocationRawReplace replaces the current history entry with URL.
//
// Unlike [ActionLocationRawAssign], the current page is dropped from history,
// so the user cannot navigate back to it.
type ActionLocationRawReplace struct {
	// URL is the target URL. Must be absolute. Required.
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
// If no element matches, the action fails; dispatched with [Call], the error
// arrives on the completion channel.
type ActionScroll struct {
	// Selector is a CSS selector matched against the whole document. Required.
	Selector string
	// Options is passed to Element.scrollIntoView. Optional; nil uses the
	// browser default.
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

// ActionIndicate applies Indicator for Duration, independent of any request.
//
// Only the Query and QueryAll indicator variants apply without an event
// element, which exists only in event attr and [ALink] action lists.
type ActionIndicate struct {
	// Indicator lists the temporary DOM changes to apply. Required; without it
	// the action does nothing.
	Indicator Indicators
	// Duration is how long the changes stay applied, truncated to
	// milliseconds. Required; zero removes them immediately.
	Duration time.Duration
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
