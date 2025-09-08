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
	"io"
	"net/http"

	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/go-playground/form/v4"
)

// ARawSubmit handles form submissions with raw multipart data,
// giving full control over uploads, streaming, and parsing.
type ARawSubmit struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend form handler.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RRawForm) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (s ARawSubmit) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s ARawSubmit) Attr() AttrInit {
	return s
}

func (s ARawSubmit) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(&front.FormCapture{}, &front.Hook{
		OnError:   intoActions(ctx, s.OnError),
		Before:    intoActions(ctx, s.Before),
		Scope:     front.IntoScopeSet(inst, s.Scope),
		Indicate:  front.IntoIndicate(s.Indicator),
		HookEntry: entry,
	})
}

func (s *ARawSubmit) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	done := s.On(ctx, &request{
		w:   w,
		r:   r,
		ctx: ctx,
	})
	return done
}

var formDecoder *form.Decoder

func init() {
	formDecoder = form.NewDecoder()
}

// ASubmit handles form submissions with decoded data of type T,
// which must be a struct annotated for go-playground/form.
type ASubmit[T any] struct {
	// MaxMemory sets the maximum number of bytes to parse into memory.
	// It is passed to ParseMultipartForm.
	// Defaults to 8 MB if zero.
	MaxMemory int
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend form handler.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, RForm[T]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (s ASubmit[V]) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s ASubmit[V]) Attr() AttrInit {
	return s
}

func (s ASubmit[V]) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(&front.FormCapture{}, &front.Hook{
		OnError:   intoActions(ctx, s.OnError),
		Before:    intoActions(ctx, s.Before),
		Scope:     front.IntoScopeSet(inst, s.Scope),
		Indicate:  front.IntoIndicate(s.Indicator),
		HookEntry: entry,
	})
}

const defaultMaxMemory = 8 << 20

func (s *ASubmit[V]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	maxMemory := defaultMaxMemory
	if s.MaxMemory > 0 {
		maxMemory = s.MaxMemory
	}
	err := r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		w.Write([]byte("Multipart form parsing error"))
		w.WriteHeader(400)
		return false
	}
	var v V
	err = formDecoder.Decode(&v, r.Form)
	if err != nil {
		w.Write([]byte("Form decoding error"))
		w.WriteHeader(400)
		return false
	}
	return s.On(ctx, &formHookRequest[V]{
		data: &v,
		request: request{
			w:   w,
			r:   r,
			ctx: ctx,
		},
	})
}

type ChangeEvent = front.ChangeEvent

// AChange is an attribute struct used with A(ctx, ...) to handle 'change' events via backend hooks.
//
// It binds to inputs, selects, or other form elements and triggers the On handler
// when the value is committed (typically when focus leaves or enter is pressed).
//
// This is useful for handling committed input changes (unlike 'input', which fires continuously).
//
// Example:
//
//	<input type="text" { A(ctx, AChange{
//	    On: func(ctx context.Context, ev EventRequest[ChangeEvent]) bool {
//	        // handle changed input value
//	        return true
//	    },
//	})... }>

// AChange prepares a change event hook for DOM elements,
// with configurable propagation, scheduling, indicators, and handlers.
type AChange struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend event handler.
	// Receives a typed REvent[ChangeEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[ChangeEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p AChange) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, p)
}

func (p AChange) Attr() AttrInit {
	return p
}

func (p AChange) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[ChangeEvent]{
		capture:   &front.ChangeCapture{},
		door:      n,
		ctx:       ctx,
		onError:   p.OnError,
		before:    p.Before,
		indicator: p.Indicator,
		inst:      inst,
		scope:     p.Scope,
		on:        p.On,
	}).init(attrs)
}

type InputEvent = front.InputEvent

type AInput struct {
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend event handler.
	// Receives a typed REvent[InputEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On           func(context.Context, REvent[InputEvent]) bool
	// If true, does not include value in event 
	// Optional.
	ExcludeValue bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}

func (p AInput) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, p)
}

func (p AInput) Attr() AttrInit {
	return p
}

func (p AInput) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[InputEvent]{
		capture: &front.InputCapture{
			ExcludeValue: p.ExcludeValue,
		},
		door:      n,
		inst:      inst,
		ctx:       ctx,
		onError:   p.OnError,
		before:    p.Before,
		scope:     p.Scope,
		indicator: p.Indicator,
		on:        p.On,
	}).init(attrs)
}
