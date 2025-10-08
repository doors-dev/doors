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
	"log/slog"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/action"

	"github.com/doors-dev/doors/internal/instance"
)

// Action performs a client-side operation
type Action interface {
	action(context.Context, instance.Core, door.Core) (action.Action, action.CallParams, error)
}

func intoActions(ctx context.Context, actions []Action) action.Actions {
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	door := ctx.Value(common.CtxKeyDoor).(door.Core)
	arr := make(action.Actions, 0)
	for _, action := range actions {
		a, _, err := action.action(ctx, inst, door)
		if err != nil {
			slog.Error("Action preparation error", slog.String("error", err.Error()))
			continue
		}
		arr = append(arr, a)
	}
	return arr
}

// ActionEmit invokes a client-side handler registered with
// $d.on(name: string, func: (arg: any, err?: Error) => any).
type ActionEmit struct {
	Name string
	Arg  any
}

// ActionOnlyEmit returns a single ActionEmit.
func ActionOnlyEmit(name string, arg any) []Action {
	return []Action{ActionEmit{Name: name, Arg: arg}}
}

func (a ActionEmit) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	return &action.Emit{
		Name:   a.Name,
		Arg:    a.Arg,
		DoorId: door.Id(),
	}, action.CallParams{}, nil
}

// ActionLocationReload reloads the current page.
type ActionLocationReload struct{}

// ActionOnlyLocationReload returns a single ActionLocationReload.
func ActionOnlyLocationReload() []Action {
	return []Action{ActionLocationReload{}}
}

func (a ActionLocationReload) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	return &action.LocationReload{}, action.CallParams{Timeout: inst.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionLocationReplace replaces the current location with a model-derived URL.
type ActionLocationReplace struct {
	Model any
}

// ActionOnlyLocationReplace returns a single ActionLocationReplace.
func ActionOnlyLocationReplace(model any) []Action {
	return []Action{ActionLocationReplace{Model: model}}
}

func (a ActionLocationReplace) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	l, err := NewLocation(ctx, a.Model)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	return &action.LocationReplace{
		URL:    l.String(),
		Origin: true,
	}, action.CallParams{Timeout: inst.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionLocationAssign navigates to a model-derived URL.
type ActionLocationAssign struct {
	Model any
}

// ActionOnlyLocationAssign returns a single ActionLocationAssign.
func ActionOnlyLocationAssign(model any) []Action {
	return []Action{ActionLocationAssign{Model: model}}
}

func (a ActionLocationAssign) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	l, err := NewLocation(ctx, a.Model)
	if err != nil {
		return nil, action.CallParams{}, err
	}
	return &action.LocationAssign{
		URL:    l.String(),
		Origin: true,
	}, action.CallParams{Timeout: inst.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionRawLocationAssign navigates to a specified URL
type ActionRawLocationAssign struct {
	URL string
}

// ActionOnlyRawLocationAssign returns a single ActionLocationAssign.
func ActionOnlyRawLocationAssign(url string) []Action {
	return []Action{ActionRawLocationAssign{URL: url}}
}

func (a ActionRawLocationAssign) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	return &action.LocationAssign{
		URL:    a.URL,
		Origin: false,
	}, action.CallParams{Timeout: inst.Conf().InstanceTTL, Optimistic: true}, nil
}

// ActionScroll scrolls to the first element matching Selector.
type ActionScroll struct {
	Selector string
	Smooth   bool
}

// ActionOnlyScroll returns a single ActionScroll.
func ActionOnlyScroll(selector string, smooth bool) []Action {
	return []Action{ActionScroll{Selector: selector, Smooth: smooth}}
}

func (a ActionScroll) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	return &action.Scroll{
		Selector: a.Selector,
		Smooth:   a.Smooth,
	}, action.CallParams{}, nil
}

// ActionIndicate applies indicators for a fixed duration.
type ActionIndicate struct {
	Indicator []Indicator
	Duration  time.Duration
}

// ActionOnlyIndicate returns a single ActionIndicate.
func ActionOnlyIndicate(indicator []Indicator, duration time.Duration) []Action {
	return []Action{ActionIndicate{Indicator: indicator, Duration: duration}}
}

func (a ActionIndicate) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, action.CallParams, error) {
	return &action.Indicate{
		Indicate: front.IntoIndicate(a.Indicator),
		Duration: a.Duration,
	}, action.CallParams{}, nil
}
