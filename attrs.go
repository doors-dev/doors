package doors

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

// A constructs a set of HTML attributes.
//
// These attributes enable backend-connected interactivity — such as pointer events,
// data binding, and hook-based logic — by wiring frontend behavior to Go code via context.
//
// `A` is typically used inside HTML tags to attach event handlers ,
// and is required for features like AClick, AChange, AHook, etc.
//
// It should be passed within an attribute block and spread into the element using `...`.
//
// Example:
//
//	<button { A(ctx, AClick{
//	    On: func(ctx context.Context, _ d.EventRequest[d.PointerEvent]) bool {
//	        log.Println("Clicked")
//	        return false
//	    },
//	})... }>
//	    Click Me
//	</button>
//
// Parameters:
//   - ctx: the current rendering context. It is used to bind interactive behavior
//     to the component’s lifecycle and scope.
//   - attrs: a list of special Attribute values (e.g., AClick, AHook, ABind).
//
// Returns:
//   - A templ.Attributes object that can be spread into a templ element.
func A(ctx context.Context, attrs ...front.Attr) templ.Attributes {
	return front.A(ctx, attrs...)
}

type ARaw templ.Attributes

func (s ARaw) Init(ctx context.Context, _ node.Core, _ instance.Core, attrs *front.Attrs) {
	if s == nil {
		return
	}
	attrs.SetRaw(templ.Attributes(s))
}

// HookMode defines how templ hooks behave when multiple invocations are triggered.
//
// Each mode determines whether hooks can run concurrently, block, debounce, or coalesce
// depending on the timing and number of interactions.
//
// Set via the Mode field of A... structs (attributes).
type HookMode = front.HookMode

// ModeDefault is the standard hook behavior.
//
// Hook triggers are sent to the backend immediately, even if a previous one is still in progress.
// However, hook functions on the backend are always executed one after another — never concurrently.
// This ensures consistent behavior without requiring user-defined concurrency handling.
//
// Use this mode when it's safe for hooks to queue up freely, and you don’t need throttling or blocking.
func ModeDefault() HookMode {
	return front.Default()
}

// ModeBlock ensures that only one instance of a given hook runs at a time.
//
// If a hook is triggered again while a previous one is still in flight (including
// any node or beam updates it performs), the new invocation is ignored.
//
// This is useful for guarding against duplicate form submissions.
func ModeBlock() HookMode {
	return front.Block()
}

// ModeFrame queues hooks until all previous hook executions (and any resulting updates)
// have fully completed.
//
// This provides the strongest consistency guarantee: one complete render-update cycle at a time.
// All new hooks are blocked until the current frame is fully processed.
//
// Use this for cases where timing-sensitive or multi-step updates must be serialized cleanly.
func ModeFrame() HookMode {
	return front.Frame()
}

// ModeButter disables ModeFrame behavior even if inherited from parent scope.
//
// It behaves like ModeDefault: hooks are triggered independently without blocking.
// Use this to override ModeFrame in nested contexts where blocking is undesirable.
func ModeButter() HookMode {
	return front.Butter()
}

// ModeDebounce throttles hook invocations by debouncing them.
//
// Hooks are delayed by the given duration, and at most one is processed
// within the given limit. This is useful for search fields, scroll events,
// or other noisy input sources.
//
// Parameters:
//   - duration: how long to wait after the last trigger before firing.
//   - limit: the maximum time allowed between the first and final call.
//
// Example: Debounce for 300ms, with a hard limit of 1s:
//
//	ModeDebounce(300 * time.Millisecond, 1 * time.Second)
func ModeDebounce(duration time.Duration, limit time.Duration) HookMode {
	return front.Debounce(int(duration.Milliseconds()), int(limit.Milliseconds()))
}

type Indicate = front.Indicate

func IndicateContent(content string) Indicate {
	return front.IndicateContent(front.SelectTarget(), content)
}
func IndicateAttr(attr string, value string) Indicate {
	return front.IndicateAttr(front.SelectTarget(), attr, value)
}
func IndicateClassRemove(query string, class string) Indicate {
	return front.IndicateClassRemove(front.SelectTarget(), class)
}
func IndicateClass(class string) Indicate {
	return front.IndicateClass(front.SelectTarget(), class)
}

func IndicateContentQuery(query string, content string) Indicate {
	return front.IndicateContent(front.SelectQuery(query), content)
}
func IndicateAttrQuery(query string, attr string, value string) Indicate {
	return front.IndicateAttr(front.SelectQuery(query), attr, value)
}
func IndicateClassQuery(query string, class string) Indicate {
	return front.IndicateClass(front.SelectQuery(query), class)
}
func IndicateClassRemoveQuery(query string, class string) Indicate {
	return front.IndicateClassRemove(front.SelectQuery(query), class)
}
func IndicateContentQueryParent(query string, content string) Indicate {
	return front.IndicateContent(front.SelectParentQuery(query), content)
}
func IndicateAttrQueryParent(query string, attr string, value string) Indicate {
	return front.IndicateAttr(front.SelectParentQuery(query), attr, value)
}
func IndicateClassQueryParent(query string, class string) Indicate {
	return front.IndicateClass(front.SelectParentQuery(query), class)
}
func IndicateClassRemoveQueryParent(query string, class string) Indicate {
	return front.IndicateClassRemove(front.SelectParentQuery(query), class)
}



var noAttrs []*front.Attr = make([]*front.Attr, 0)

type eventAttr[E any] struct {
	node      node.Core
	ctx       context.Context
	capture   front.Capture
	mark      string
	mode      HookMode
	indicate []Indicate
	on        func(context.Context, EventRequest[E]) bool
}

func (p *eventAttr[E]) init(attrs *front.Attrs) {
	entry, ok := p.node.RegisterAttrHook(p.ctx, &node.AttrHook{
		Trigger: p.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(p.capture, &front.Hook{
		Mark:      p.mark,
		Mode:      p.mode,
		Indicate: p.indicate,
		HookEntry: entry,
	})
}

func (p *eventAttr[E]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	var e E
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&e)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(400)
		return false
	}
	w.WriteHeader(200)
	return p.on(ctx, &eventRequest[E]{
		request: request{
			r: r,
			w: w,
		},
		e: &e,
	})
}
