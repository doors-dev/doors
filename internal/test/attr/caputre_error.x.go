// Managed by GoX v0.3.2

//line caputre_error.gox:1
// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

//line caputre_error.gox:33
func (f *errorFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line caputre_error.gox:34
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
//line caputre_error.gox:46
		__e = __c.Any(f.button("err_1", doors.ActionEmit[any]{Name: "error", Arg: "err_1"})); if __e != nil { return }
//line caputre_error.gox:47
		__e = __c.Any(f.button("err_2", doors.ActionEmit[any]{Name: "root", Arg: "err_2"})); if __e != nil { return }
//line caputre_error.gox:48
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
//line caputre_error.gox:60
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
//line caputre_error.gox:68
							__e = __c.Set("id", "indicator"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("init"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
//line caputre_error.gox:69
						__e = __c.Any(f.button("err_5", doors.Join[doors.Actions](
				doors.ActionEmit[any]{
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
//line caputre_error.gox:89
						__e = __c.Any(f.button("err_6", doors.ActionEmit[any]{Name: "error", Arg: "err_6"})); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line caputre_error.gox:91
				__e = __c.Any(f.button("err_3", doors.ActionEmit[any]{Name: "error", Arg: "err_3"})); if __e != nil { return }
//line caputre_error.gox:92
				__e = __c.Any(f.button("err_4", doors.ActionEmit[any]{Name: "n1", Arg: "err_4"})); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line caputre_error.gox:94
}

//line caputre_error.gox:96
func (f *errorFragment) button(id string, on doors.Actions) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line caputre_error.gox:97
			__e = __c.Set("id", id); if __e != nil { return }
//line caputre_error.gox:97
			__e = __c.Modify(doors.A(ctx, f.handler(on))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line caputre_error.gox:98
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line caputre_error.gox:100
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
