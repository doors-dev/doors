package doors

import (
	"context"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

type PointerEvent = front.PointerEvent

type pointerEventHook struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (p *pointerEventHook) init(event string, ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {

	(&eventAttr[PointerEvent]{
		capture: &front.PointerCapture{
			Event:           event,
			StopPropagation: p.StopPropagation,
			PreventDefault:  p.PreventDefault,
		},
		inst:      inst,
		node:      n,
		scope:     p.Scope,
		ctx:       ctx,
		mark:      p.Mark,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
}

// AClick is an attribute struct used with A(ctx, ...) to handle click or pointer events via backend hooks.
type AClick struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c AClick) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("click", ctx, n, inst, attrs)
}

// APointerDown is an attribute struct used with A(ctx, ...) to handle 'pointerdown' events via backend hooks.
type APointerDown struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerDown) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerdown", ctx, n, inst, attrs)
}

// APointerUp is an attribute struct used with A(ctx, ...) to handle 'pointerup' events via backend hooks.
type APointerUp struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerUp) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerup", ctx, n, inst, attrs)
}

// APointerMove is an attribute struct used with A(ctx, ...) to handle 'pointermove' events via backend hooks.
type APointerMove struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerMove) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointermove", ctx, n, inst, attrs)
}

// APointerOver is an attribute struct used with A(ctx, ...) to handle 'pointerover' events via backend hooks.
type APointerOver pointerEventHook

func (c APointerOver) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerover", ctx, n, inst, attrs)
}

// APointerOut is an attribute struct used with A(ctx, ...) to handle 'pointerout' events via backend hooks.
type APointerOut struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerOut) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerout", ctx, n, inst, attrs)
}

// APointerEnter is an attribute struct used with A(ctx, ...) to handle 'pointerenter' events via backend hooks.
type APointerEnter struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerEnter) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerenter", ctx, n, inst, attrs)
}

// APointerLeave is an attribute struct used with A(ctx, ...) to handle 'pointerleave' events via backend hooks.
type APointerLeave struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerLeave) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointerleave", ctx, n, inst, attrs)
}

// APointerCancel is an attribute struct used with A(ctx, ...) to handle 'pointercancel' events via backend hooks.
type APointerCancel struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Mark is used to identify hook error on the frontend
	Mark string

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool
}

func (c APointerCancel) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("pointercancel", ctx, n, inst, attrs)
}

// AGotPointerCapture is an attribute struct used with A(ctx, ...) to handle 'gotpointercapture' events via backend hooks.
type AGotPointerCapture pointerEventHook

func (c AGotPointerCapture) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("gotpointercapture", ctx, n, inst, attrs)
}

// ALostPointerCapture is an attribute struct used with A(ctx, ...) to handle 'lostpointercapture' events via backend hooks.
type ALostPointerCapture pointerEventHook

func (c ALostPointerCapture) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*pointerEventHook)(&c)
	p.init("lostpointercapture", ctx, n, inst, attrs)
}
