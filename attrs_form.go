package doors

import (
	"context"
	"net/http"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
	"github.com/go-playground/form/v4"
)

type FocusEvent = front.FocusEvent

type focusEventHook struct {
	// Mark is an optional identifier that appears in frontend hook lifecycle events.
	// Use it to filter events like `hook:start` or `hook:end` in JavaScript.
	Mark string

	// Mode determines how this hook is scheduled (e.g., blocking, debounce).
	// See ModeDefault, ModeBlock, etc.
	Mode HookMode

	// Indicate specifies how to visually indicate the hook is running (e.g., spinner, class, content). Optional.
	Indicate []Indicate

	// On is the required backend handler that runs when the event is triggered.
	//
	// The function receives a typed EventRequest[FocusEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[FocusEvent]) bool
}

func (p *focusEventHook) init(event string, ctx context.Context, n node.Core, _ instance.Core, attrs *front.Attrs) {
	(&eventAttr[FocusEvent]{
		capture: &front.FocusCapture{
			Event: event,
		},
		node:     n,
		ctx:      ctx,
		mark:     p.Mark,
		indicate: p.Indicate,
		on:       p.On,
	}).init(attrs)
}

// AFocus is an attribute struct used with A(ctx, ...) to handle 'focus' events via backend hooks.
type AFocus focusEventHook

func (f AFocus) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(&f)
	p.init("focus", ctx, n, inst, attrs)
}

// ABlur is an attribute struct used with A(ctx, ...) to handle 'blur' events via backend hooks.
type ABlur focusEventHook

func (b ABlur) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(&b)
	p.init("blur", ctx, n, inst, attrs)
}

// AFocusIn is an attribute struct used with A(ctx, ...) to handle 'focusin' events via backend hooks.
type AFocusIn focusEventHook

func (f AFocusIn) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(&f)
	p.init("focusin", ctx, n, inst, attrs)
}

// AFocusOut is an attribute struct used with A(ctx, ...) to handle 'focusout' events via backend hooks.
type AFocusOut focusEventHook

func (f AFocusOut) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	p := (*focusEventHook)(&f)
	p.init("focusout", ctx, n, inst, attrs)
}

// ARawSubmit is an attribute struct used with A(ctx, ...) to handle form submissions via backend hooks,
// providing low-level access to the raw multipart form data.
//
// Unlike ASubmit, this variant does not decode the form into a typed struct.
// Instead, it gives full control over file uploads, streaming, and multipart parsing via RawFormRequest.
//
// This is useful when handling large forms, file uploads, or custom parsing logic.
//
// Example:
//
//	<form { A(ctx, ARawSubmit{
//	    On: func(ctx context.Context, req RawFormRequest) bool {
//	        form, _ := req.ParseForm(32 << 20) // 32 MB
//	        file, _, _ := form.FormFile("upload")
//	        // handle file...
//	        return true
//	    },
//	})... }>
type ARawSubmit struct {
	// Mark is an optional identifier that appears in frontend hook lifecycle events.
	// Use it to filter events like `hook:start` or `hook:end` in JavaScript.
	Mark string

	// Mode determines how this hook is scheduled (e.g., blocking, debounce).
	// See ModeDefault, ModeBlock, ModeFrame, etc.
	Mode HookMode

	// Indicate specifies how to visually indicate the hook is running
	// (e.g., by applying a class, attribute, or replacing content). Optional.
	Indicate []Indicate

	// On is the required backend handler for the form submission.
	//
	// It receives a RawFormRequest and should return true (is done)
	// when processing is complete and the hook can be removed.
	On func(context.Context, RRawForm) bool
}

func (s ARawSubmit) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(&front.FormCapture{}, &front.Hook{
		Mark:      s.Mark,
		Mode:      s.Mode,
		Indicate:  s.Indicate,
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

// ASubmit is an attribute struct used with A(ctx, ...) to handle form submissions via backend hooks.
//
// It binds a <form> element to a backend handler that receives decoded form data of type D.
// The hook runs when the form is submitted and can support file uploads or large payloads.
//
// This is typically used as:
//
//	<form { A(ctx, ASubmit[MyFormData]{
//	    On: func(ctx context.Context, req FormRequest[MyFormData]) bool {
//	        // handle form submission
//	        return true
//	    },
//	})... }>
type ASubmit[D any] struct {
	// MaxMemory sets the maximum number of bytes to parse into memory
	// before falling back to temporary files when handling multipart forms.
	//
	// This affects file upload behavior. It is passed to ParseMultipartForm.
	// Defaults to 8 MB if zero.
	MaxMemory int

	// Mark is an optional identifier that appears in frontend hook lifecycle events.
	// Use it to filter events like `hook:start` or `hook:end` in JavaScript.
	Mark string

	// Mode determines how this hook is scheduled (e.g., blocking, debounce).
	// See ModeDefault, ModeBlock, ModeFrame, etc.
	Mode HookMode

	// Indicate specifies how to visually indicate the hook is running
	// (e.g., by applying a class, attribute, or replacing content). Optional.
	Indicate []Indicate

	// On is the required backend handler for the form submission.
	//
	// It receives a typed FormRequest[D] and should return true (is done)
	// when processing is complete and the hook can be removed.
	On func(context.Context, RForm[D]) bool
}

func (s ASubmit[V]) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(&front.FormCapture{}, &front.Hook{
		Mark:      s.Mark,
		Mode:      s.Mode,
		Indicate:  s.Indicate,
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
type AChange struct {
	// Mark is an optional identifier that appears in frontend hook lifecycle events.
	// Use it to filter events like `hook:start` or `hook:end` in JavaScript.
	Mark string

	// Mode determines how this hook is scheduled (e.g., blocking, debounce).
	// See ModeDefault, ModeBlock, ModeFrame, etc.
	Mode HookMode

	// Indicate specifies how to visually indicate the hook is running
	// (e.g., by applying a class, attribute, or replacing content). Optional.
	Indicate []Indicate

	// On is the required backend handler for the change event.
	//
	// It receives a typed EventRequest[ChangeEvent] and should return true
	// when the hook is complete and can be removed.
	On func(context.Context, REvent[ChangeEvent]) bool
}

func (p AChange) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[ChangeEvent]{
		capture:  &front.ChangeCapture{},
		node:     n,
		ctx:      ctx,
		mark:     p.Mark,
		indicate: p.Indicate,
		on:       p.On,
	}).init(attrs)
}

type InputEvent = front.InputEvent
type AInput struct {
	Mark         string
	Mode         HookMode
	Indicate     []Indicate
	On           func(context.Context, REvent[InputEvent]) bool
	ExcludeValue bool
}

func (p AInput) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	(&eventAttr[InputEvent]{
		capture: &front.InputCapture{
			ExcludeValue: p.ExcludeValue,
		},
		node:     n,
		ctx:      ctx,
		mark:     p.Mark,
		indicate: p.Indicate,
		on:       p.On,
	}).init(attrs)
}
