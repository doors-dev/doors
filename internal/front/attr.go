package front

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

func NewAttrs() *Attrs {
	return &Attrs{
		Attrs: *common.NewAttrs(),
	}
}

type Attrs struct {
	common.Attrs
}

func (a *Attrs) SetHook(name string, hook *Hook) {
	a.Attrs.SetObject(fmt.Sprintf("data-d00r-hook:%s", name), hook)
}

func (a *Attrs) SetData(name string, data any) {
	a.Attrs.SetObject(fmt.Sprintf("data-d00r-data:%s", name), data)
}

func (a *Attrs) AppendCapture(capture Capture, hook *Hook) {
	a.Attrs.AppendArray("data-d00r-capture", []any{capture.Listen(), capture.Name(), capture, hook})
}

func (a *Attrs) Init(_ context.Context, _ node.Core, _ instance.Core, attrs *Attrs) {
	a.Join(attrs)
}

func (a *Attrs) Render(ctx context.Context, w io.Writer) error {
	return AttrRender(ctx, w, a)
}

func (a *Attrs) Join(attrs *Attrs) {
	attrs.Attrs.Join(&a.Attrs)
}

type Attr interface {
	Init(context.Context, node.Core, instance.Core, *Attrs)
	templ.Component
}

func A(ctx context.Context, attr ...Attr) *Attrs{
	node := ctx.Value(common.NodeCtxKey).(node.Core)
	instance := ctx.Value(common.InstanceCtxKey).(instance.Core)
	attrs := NewAttrs()
	for _, attr := range attr {
		attr.Init(ctx, node, instance, attrs)
	}
	return attrs
} 


func AttrRender(ctx context.Context, w io.Writer, a Attr) error {
	node := ctx.Value(common.NodeCtxKey).(node.Core)
	instance := ctx.Value(common.InstanceCtxKey).(instance.Core)
	attrs, ok := ctx.Value(common.AttrsCtxKey).(*Attrs)
	if ok {
		a.Init(ctx, node, instance, attrs)
		return nil
	}
	attrs = NewAttrs()
	a.Init(ctx, node, instance, attrs)
	rm := ctx.Value(common.RenderMapCtxKey).(*common.RenderMap)
	return rm.WriteAttrs(w, &attrs.Attrs)
}
