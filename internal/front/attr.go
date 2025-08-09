package front

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

func newAs() *Attrs {
	return &Attrs{
		regular: make(templ.Attributes),
		objects: make(map[string]any),
		arrays:  make(map[string][]any),
	}
}

type Attrs struct {
	regular templ.Attributes
	objects map[string]any
	arrays  map[string][]any
}

func (a *Attrs) marshal(value any) (string, error) {
	b, err := common.MarshalJSON(value)
	if err != nil {
		return "", err
	}
	return common.AsString(&b), nil
}

func (a *Attrs) render() templ.Attributes {
	output := make(templ.Attributes)
	for name := range a.regular {
		output[name] = a.regular[name]
	}
	for name := range a.objects {
		s, err := a.marshal(a.objects[name])
		if err != nil {
			slog.Error("object attribute marshaling err", slog.String("json_error", err.Error()), slog.String("attr_name", name))
			continue
		}
		output[name] = s
	}
	for name := range a.arrays {
		s, err := a.marshal(a.arrays[name])
		if err != nil {
			slog.Error("array attribute marshaling err", slog.String("json_error", err.Error()), slog.String("attr_name", name))
			continue
		}
		output[name] = s
	}
	return output
}

func (a *Attrs) SetRaw(attrs templ.Attributes) {
	for name := range attrs {
		a.regular[name] = attrs[name]
	}
}

func (a *Attrs) Set(name string, value any) {
	a.regular[name] = value
}

func (a *Attrs) SetObject(name string, value any) {
	a.objects[name] = value
}

func (a *Attrs) AppendArray(name string, value any) {
	arr, ok := a.arrays[name]
	if !ok {
		arr = []any{value}
	} else {
		arr = append(arr, value)
	}
	a.arrays[name] = arr
}

func (a *Attrs) SetHook(name string, hook *Hook) {
	a.SetObject(fmt.Sprintf("data-d00r-hook:%s", name), hook)
}

func (a *Attrs) SetData(name string, data any) {
	a.SetObject(fmt.Sprintf("data-d00r-data:%s", name), data)
}

func (a *Attrs) AppendCapture(capture Capture, hook *Hook) {
	a.AppendArray("data-d00r-capture", []any{capture.Listen(), capture.Name(), capture, hook})
}

type Attr interface {
	Init(context.Context, node.Core, instance.Core, *Attrs)
}

func A(ctx context.Context, attr ...Attr) templ.Attributes {
	node, nodeOk := ctx.Value(common.NodeCtxKey).(node.Core)
	instance, instOk := ctx.Value(common.InstanceCtxKey).(instance.Core)
	if !nodeOk || !instOk {
		log.Fatalf("Attributes are used outside context")
	}
	attrs := newAs()
	for _, attr := range attr {
		attr.Init(ctx, node, instance, attrs)
	}
	return attrs.render()
}
