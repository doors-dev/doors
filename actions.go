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
	"log/slog"
	"time"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/action"
)

// Action performs a client-side operation.
type Action interface {
	action(ctx context.Context, core core.Core, gz bool) (action.Action, action.CallParams, error)
}

func intoActions(ctx context.Context, actions []Action) action.Actions {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	arr := make(action.Actions, 0)
	for _, action := range actions {
		a, _, err := action.action(ctx, core, false)
		if err != nil {
			slog.Error("Action preparation error", slog.String("error", err.Error()))
			continue
		}
		arr = append(arr, a)
	}
	return arr
}

// ActionEmit invokes a client-side handler registered with
// `$on(name, handler)`.
type ActionEmit struct {
	Name string
	Arg  any
}

// ActionOnlyEmit returns a single ActionEmit.
func ActionOnlyEmit(name string, arg any) []Action {
	return []Action{ActionEmit{Name: name, Arg: arg}}
}

func (a ActionEmit) action(ctx context.Context, core core.Core, gz bool) (action.Action, action.CallParams, error) {
	payload, err := action.IntoPayload(a.Arg, gz)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	act := action.Emit{
		Name:    a.Name,
		DoorID:  core.DoorID(),
		Payload: payload,
	}
	return act, action.CallParams{}, nil
}

// ActionLocationReload reloads the current page.
type ActionLocationReload struct{}

// ActionOnlyLocationReload returns a single ActionLocationReload.
func ActionOnlyLocationReload() []Action {
	return []Action{ActionLocationReload{}}
}

func (a ActionLocationReload) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return &action.LocationReload{}, action.CallParams{Timeout: core.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionLocationReplace replaces the current history entry with a model-derived
// URL.
type ActionLocationReplace struct {
	Model any
}

// ActionOnlyLocationReplace returns a single ActionLocationReplace.
func ActionOnlyLocationReplace(model any) []Action {
	return []Action{ActionLocationReplace{Model: model}}
}

func (a ActionLocationReplace) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	l, err := NewLocation(ctx, a.Model)
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

// ActionOnlyLocationAssign returns a single ActionLocationAssign.
func ActionOnlyLocationAssign(model any) []Action {
	return []Action{ActionLocationAssign{Model: model}}
}

func (a ActionLocationAssign) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	l, err := NewLocation(ctx, a.Model)
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

// ActionOnlyLocationRawAssign returns a single [ActionLocationRawAssign].
func ActionOnlyLocationRawAssign(url string) []Action {
	return []Action{ActionLocationRawAssign{URL: url}}
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

// ActionOnlyScroll returns a single ActionScroll with default scroll
// behavior.
func ActionOnlyScroll(selector string) []Action {
	return []Action{ActionScroll{Selector: selector}}
}

func (a ActionScroll) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return action.Scroll{
		Selector: a.Selector,
		Options:  a.Options,
	}, action.CallParams{}, nil
}

// ActionIndicate applies indicators for Duration.
type ActionIndicate struct {
	Indicator []Indicator
	Duration  time.Duration
}

// ActionOnlyIndicate returns a single ActionIndicate.
func ActionOnlyIndicate(indicator []Indicator, duration time.Duration) []Action {
	return []Action{ActionIndicate{Indicator: indicator, Duration: duration}}
}

func (a ActionIndicate) action(ctx context.Context, core core.Core, _ bool) (action.Action, action.CallParams, error) {
	return action.Indicate{
		Indicate: front.IntoIndicate(a.Indicator),
		Duration: a.Duration,
	}, action.CallParams{}, nil
}
