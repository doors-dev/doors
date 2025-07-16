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

	// Mark is an optional identifier that appears in frontend hook lifecycle events.
	// Use it to filter events like `hook:start` or `hook:end` in JavaScript.
	Mark string

	// Mode determines how this hook is scheduled (e.g., blocking, debounce).
	// See ModeDefault, ModeBlock, etc.
	Mode HookMode

	// Indicate specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicate []Indicate

	// On is the required backend handler for the click event.
	// It receives a typed EventRequest[KeyboardEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[KeyboardEvent]) bool
}

func (k *keyEventHook) init(event string, ctx context.Context, n node.Core, _ instance.Core, attrs *front.Attrs) {
	(&eventAttr[KeyboardEvent]{
		node: n,
		ctx:  ctx,
		capture: &front.KeyboardEventCapture{
			Event:           event,
			PreventDefault:  k.PreventDefault,
			StopPropagation: k.StopPropagation,
		},
		mode:     k.Mode,
		mark:      k.Mark,
		indicate: k.Indicate,
		on:        k.On,
	}).init(attrs)
}

// AKeyDown is an attribute struct used with A(ctx, ...) to handle 'keydown' events via backend hooks.
type AKeyDown keyEventHook

func (k AKeyDown) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(&k)
	p.init("keydown", ctx, n, inst, attrs)
}

// AKeyUp is an attribute struct used with A(ctx, ...) to handle 'keyup' events via backend hooks.
type AKeyUp keyEventHook

func (k AKeyUp) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(&k)
	p.init("keyup", ctx, n, inst, attrs)
}

// AKeyPress is an attribute struct used with A(ctx, ...) to handle 'keypress' events via backend hooks.
type AKeyPress keyEventHook

func (k AKeyPress) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*keyEventHook)(&k)
	p.init("keyup", ctx, n, inst, attrs)
}
