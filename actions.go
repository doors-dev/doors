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

type Action interface {
	action(context.Context, instance.Core, door.Core) (action.Action, error)
}

func intoActions(ctx context.Context, actions []Action) action.Actions {
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	door := ctx.Value(common.CtxKeyDoor).(door.Core)
	arr := make(action.Actions, 0)
	for _, action := range actions {
		a, err := action.action(ctx, inst, door)
		if err != nil {
			slog.Error("Action preparation error", slog.String("error", err.Error()))
			continue
		}
		arr = append(arr, a)
	}
	return arr
}

type ActionEmit struct {
	Name string
	Arg  any
}

func ActionOnlyEmit(name string, arg any) []Action {
	return []Action{ActionEmit{Name: name, Arg: arg}}
}

func (a ActionEmit) isOptimisic() bool {
	return false
}

func (a ActionEmit) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, error) {
	return &action.Emit{
		Name:   a.Name,
		Arg:    a.Arg,
		DoorId: door.Id(),
	}, nil
}

type ActionLocationReload struct {
}

func (a ActionLocationReload) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, error) {
	return &action.LocationReload{}, nil
}

func ActionOnlyLocationReload() []Action {
	return []Action{ActionLocationReload{}}
}

type ActionLocationReplace struct {
	Model any
}

func (a ActionLocationReplace) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, error) {
	l, err := NewLocation(ctx, a.Model)
	if err != nil {
		return nil, err
	}
	return &action.LocationReplace{
		URL:    l.String(),
		Origin: true,
	}, nil
}

func ActionOnlyLocationReplace(model any) []Action {
	return []Action{ActionLocationReplace{Model: model}}
}

type ActionLocationAssign struct {
	Model any
}

func (a ActionLocationAssign) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, error) {
	l, err := NewLocation(ctx, a.Model)
	if err != nil {
		return nil, err
	}
	return &action.LocationAssign{
		URL:    l.String(),
		Origin: true,
	}, nil
}

func ActionOnlyLocationAssign(model any) []Action {
	return []Action{ActionLocationAssign{Model: model}}
}

type ActionScroll struct {
	Selector string
	Smooth   bool
}

func (a ActionScroll) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, error) {
	return &action.Scroll{
		Selector: a.Selector,
		Smooth:   a.Smooth,
	}, nil
}

func ActionOnlyScroll(selector string, smooth bool) []Action {
	return []Action{ActionScroll{Selector: selector, Smooth: smooth}}
}

type ActionIndicate struct {
	Indicator []Indicator
	Duration  time.Duration
}


func (a ActionIndicate) action(ctx context.Context, inst instance.Core, door door.Core) (action.Action, error) {
	return &action.Indicate{
		Indicate: front.IntoIndicate(a.Indicator),
		Duration: a.Duration,
	}, nil
}

func ActionOnlyIndicate(indicator []Indicator, duration time.Duration) []Action {
	return []Action{ActionIndicate{Indicator: indicator, Duration: duration}}
}
