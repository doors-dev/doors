package doors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

type AHook[I any, O any] struct {
	On        func(ctx context.Context, r RHook[I]) (O, bool)
	Name      string
	Mode      HookMode
	Indicator []Indicate
}

func (h AHook[I, O]) Init(ctx context.Context, n node.Core, _ instance.Core, attr *front.Attrs) {
	if h.On == nil {
		println("Hook withoud handler")
		return
	}
	entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
		Trigger: h.handle,
	})
	if !ok {
		return
	}
	attr.SetHook(h.Name, &front.Hook{
		Mode:      h.Mode,
		Indicate:  h.Indicator,
		HookEntry: entry,
	})
}

func (h *AHook[I, O]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	var input I
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&input)
	r.Body.Close()
	if err != nil {
		println(err.Error())
		w.WriteHeader(400)
		return false
	}
	output, done := h.On(ctx, &formHookRequest[I]{
		data: &input,
		request: request{
			w: w,
			r: r,
			ctx: ctx,
		},
	})
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&output)
	if err != nil {
		println(err.Error())
		w.WriteHeader(500)
	}
	return done

}

type ARawHook struct {
	Name      string
	On        func(ctx context.Context, r RRawHook) bool
	Mode      HookMode
	Indicator []Indicate
}

func (h ARawHook) Init(ctx context.Context, n node.Core, _ instance.Core, attr *front.Attrs) {
	if h.On == nil {
		println("Hook withoud handler")
		return
	}
	entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
		Trigger: h.handle,
	})
	if !ok {
		return
	}
	attr.SetHook(h.Name, &front.Hook{
		Mode:      h.Mode,
		Indicate:  h.Indicator,
		HookEntry: entry,
	})
}

func (h *ARawHook) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	return h.On(ctx, &request{
		r:   r,
		w:   w,
		ctx: ctx,
	})
}

type AData struct {
	Name  string
	Value any
}

func (a AData) Init(_ context.Context, n node.Core, _ instance.Core, attr *front.Attrs) {
	attr.SetData(a.Name, a.Value)
}

type ADataMap map[string]any

func (dm ADataMap) Init(_ context.Context, n node.Core, _ instance.Core, attr *front.Attrs) {
	for name := range dm {
		attr.SetData(name, dm[name])
	}
}
