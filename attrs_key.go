package doors

import (
	"context"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

type KeyboardEvent = front.KeyboardEvent

type keyEventHook struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed EventRequest[KeyboardEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[KeyboardEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (k *keyEventHook) init(event string, ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[KeyboardEvent]{
		node: n,
		ctx:  ctx,
		capture: &front.KeyboardEventCapture{
			Event:           event,
			PreventDefault:  k.PreventDefault,
			StopPropagation: k.StopPropagation,
		},
		inst:      inst,
		scope:     k.Scope,
		onError:   k.OnError,
		indicator: k.Indicator,
		on:        k.On,
	}).init(attrs)
}

// AKeyDown is an attribute struct used with A(ctx, ...) to handle 'keydown' events via backend hooks.
type AKeyDown struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed EventRequest[KeyboardEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[KeyboardEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (k AKeyDown) Attr() AttrInit {
	return &k
}

func (k *AKeyDown) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(k)
	p.init("keydown", ctx, n, inst, attrs)
}

// AKeyUp is an attribute struct used with A(ctx, ...) to handle 'keyup' events via backend hooks.
type AKeyUp struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed EventRequest[KeyboardEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[KeyboardEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (k AKeyUp) Attr() AttrInit {
	return &k
}

func (k *AKeyUp) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(k)
	p.init("keyup", ctx, n, inst, attrs)
}

