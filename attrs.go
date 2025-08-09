package doors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

type Attr = front.Attr

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
//	<button { doors.A(ctx, doors.AClick{
//	    On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
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
func A(ctx context.Context, attrs ...Attr) templ.Attributes {
	return front.A(ctx, attrs...)
}

type ARaw templ.Attributes

func (s ARaw) Init(ctx context.Context, _ node.Core, _ instance.Core, attrs *front.Attrs) {
	if s == nil {
		return
	}
	attrs.SetRaw(templ.Attributes(s))
}

type eventAttr[E any] struct {
	node      node.Core
	ctx       context.Context
	capture   front.Capture
	onError   []OnError
	scope     []Scope
	indicator []Indicator
	inst      instance.Core
	on        func(context.Context, REvent[E]) bool
}

func (p *eventAttr[E]) init(attrs *front.Attrs) {
	entry, ok := p.node.RegisterAttrHook(p.ctx, &node.AttrHook{
		Trigger: p.handle,
	})
	if !ok {
		return
	}
	attrs.AppendCapture(p.capture, &front.Hook{
		Error:     front.IntoErrorAction(p.onError),
		Scope:     front.IntoScopeSet(p.inst, p.scope),
		Indicate:  front.IntoIndicate(p.indicator),
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
	return p.on(ctx, &eventRequest[E]{
		request: request{
			r:   r,
			w:   w,
			ctx: ctx,
		},
		e: &e,
	})
}
