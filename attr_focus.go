package doors

import (
	"context"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

type FocusEvent = front.FocusEvent

type focusIOEventHook struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicator []Indicator

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (p *focusIOEventHook) init(event string, ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[FocusEvent]{
		capture: &front.FocusIOCapture{
			Event:           event,
			StopPropagation: p.StopPropagation,
		},
		node:      n,
		ctx:       ctx,
		inst:      inst,
		onError:   p.OnError,
		scope:     p.Scope,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
}

type focusEventHook struct {
	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicator []Indicator

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (p *focusEventHook) init(event string, ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[FocusEvent]{
		capture: &front.FocusCapture{
			Event: event,
		},
		node:      n,
		ctx:       ctx,
		inst:      inst,
		onError:   p.OnError,
		scope:     p.Scope,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
}

// AFocus is an attribute struct used with A(ctx, ...) to handle 'focus' events via backend hooks.
type AFocus struct {
	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicator []Indicator

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (f AFocus) Attr() AttrInit {
	return &f
}

func (f *AFocus) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(f)
	p.init("focus", ctx, n, inst, attrs)
}

// ABlur is an attribute struct used with A(ctx, ...) to handle 'blur' events via backend hooks.
type ABlur struct {
	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicator []Indicator

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (b ABlur) Attr() AttrInit {
	return &b
}

func (b *ABlur) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(b)
	p.init("blur", ctx, n, inst, attrs)
}

// AFocusIn is an attribute struct used with A(ctx, ...) to handle 'focusin' events via backend hooks.
type AFocusIn struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicator []Indicator

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (f AFocusIn) Attr() AttrInit {
	return &f
}

func (f *AFocusIn) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusIOEventHook)(f)
	p.init("focusin", ctx, n, inst, attrs)
}

// AFocusOut is an attribute struct used with A(ctx, ...) to handle 'focusout' events via backend hooks.
type AFocusOut struct {
	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicator []Indicator

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool

	// OnError determines what to do if error occured during hook requrest
	OnError []OnError
}

func (f AFocusOut) Attr() AttrInit {
	return &f
}

func (f *AFocusOut) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusIOEventHook)(f)
	p.init("focusout", ctx, n, inst, attrs)
}
