// Managed by GoX v0.1.28

//line caputre_error.gox:1
package attr

import (
	"context"
	"net/http"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type errorFragment struct {
	test.NoBeam
	n1 doors.Door
	n2 doors.Door
}

//line caputre_error.gox:19
func (f *errorFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line caputre_error.gox:20
			__e = __c.Set("id", "report"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("initial"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const r = document.getElementById(\"report\")\n\t\t$on(\"root\", (arg, e) => {\n\t\t\tconsole.log(e)\n\t\t\tr.innerHTML = \"root/\" + arg\n\t\t})\n\t\t$on(\"error\", (arg, e) => {\n\t\t\tconsole.log(e)\n\t\t\tr.innerHTML = \"root_error/\" + arg\n\t\t})"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line caputre_error.gox:32
		__e = __c.Any(f.button("err_1", doors.ActionEmit{Name: "error", Arg: "err_1"})); if __e != nil { return }
//line caputre_error.gox:33
		__e = __c.Any(f.button("err_2", doors.ActionEmit{Name: "root", Arg: "err_2"})); if __e != nil { return }
//line caputre_error.gox:34
		__e = (f.n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Init("script"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Raw("const r = document.getElementById(\"report\")\n\t\t\t$on(\"n1\", (arg, e) => {\n\t\t\t\tconsole.log(e)\n\t\t\t\tr.innerHTML = \"n1/\" + arg\n\t\t\t})\n\t\t\t$on(\"error\", (arg, e) => {\n\t\t\t\tconsole.log(e)\n\t\t\t\tr.innerHTML = \"n1_error/\" + arg\n\t\t\t})"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line caputre_error.gox:46
				__e = (f.n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
						__e = __c.Init("script"); if __e != nil { return }
						{
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Raw("const r = document.getElementById(\"report\")\n\t\t\t\t$on(\"n2\", (arg, e) => {\n\t\t\t\t\tconsole.log(e)\n\t\t\t\t\tr.innerHTML = \"n2/\" + arg\n\t\t\t\t})"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line caputre_error.gox:54
							__e = __c.Set("id", "indicator"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("init"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
//line caputre_error.gox:55
						__e = __c.Any(f.button("err_5", doors.Join[doors.Actions](
				doors.ActionEmit{
					Name: "n2",
					Arg: "err_5",
				},
				doors.ActionIndicate{
					Duration: 500 * time.Millisecond,
					Indicator: doors.Join[doors.Indicators](
						doors.IndicatorAttr{
							Selector: doors.SelectorQuery("#indicator"),
							Name: "data-indicator",
							Value: "true",
						},
						doors.IndicatorContent{
							Selector: doors.SelectorQuery("#indicator"),
							Content: "indicator",
						},
					),
				},
			))); if __e != nil { return }
//line caputre_error.gox:75
						__e = __c.Any(f.button("err_6", doors.ActionEmit{Name: "error", Arg: "err_6"})); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line caputre_error.gox:77
				__e = __c.Any(f.button("err_3", doors.ActionEmit{Name: "error", Arg: "err_3"})); if __e != nil { return }
//line caputre_error.gox:78
				__e = __c.Any(f.button("err_4", doors.ActionEmit{Name: "n1", Arg: "err_4"})); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line caputre_error.gox:80
}

//line caputre_error.gox:82
func (f *errorFragment) button(id string, on doors.Actions) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line caputre_error.gox:83
			__e = __c.Set("id", id); if __e != nil { return }
//line caputre_error.gox:83
			__e = __c.Modify(doors.A(ctx, f.handler(on))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line caputre_error.gox:84
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line caputre_error.gox:86
}

func (f *errorFragment) handler(on doors.Actions) doors.Attr {
	return doors.AClick{
		OnError: on,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			w := r.(R)
			w.ResponseWriter().WriteHeader(http.StatusBadGateway)
			return false
		},
	}
}

type R interface {
	ResponseWriter() http.ResponseWriter
}
