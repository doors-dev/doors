// Managed by GoX v0.1.28

//line fragments.gox:1
package attr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type pointerFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:19
func (f *pointerFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:21
		f.r.Update(ctx, 0, "")

//line fragments.gox:23
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:24
		__e = (doors.APointerDown{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "DOWN")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:32
				__e = __c.Set("id", "down"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerDown"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:35
		__e = (doors.APointerUp{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "UP")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:43
				__e = __c.Set("id", "up"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerUp"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:46
		__e = (doors.APointerEnter{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "ENTER")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:54
				__e = __c.Set("id", "enter"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerEnter"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:57
			__e = __c.Set("id", "beforeLeave"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("beforeLeave"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:58
		__e = (doors.APointerLeave{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "LEAVE")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:66
				__e = __c.Set("id", "leave"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerLeave"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:69
		__e = (doors.APointerMove{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "MOVE")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:77
				__e = __c.Set("id", "move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerMove"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:80
		__e = (doors.APointerOver{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "OVER")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:88
				__e = __c.Set("id", "over"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Over"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:91
			__e = __c.Set("id", "beforeOut"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("beforeOut"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:92
		__e = (doors.APointerOut{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "OUT")
			f.r.Update(ctx, 1, test.Float(r.Event().PageX))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:100
				__e = __c.Set("id", "out"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Out"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:103
}

type callFragment struct {
	data string
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:111
func (f *callFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:112
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:113
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:114
		__e = (doors.AHook[string]{
		Name: "myHook",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			f.r.Update(ctx, 0, r.Data())
			ch := doors.XCall[string](ctx, doors.ActionEmit{Name: "myCall", Arg: len(r.Data())})
			res := <-ch
			f.r.Update(ctx, 1, res.Ok)
			return len(r.Data()), true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line fragments.gox:123
			__e = (doors.AData{
		Name: "myData",
		Value: f.data,
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("script"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Raw("$on(\"myCall\", (data) => {\n\t\t\tdocument.getElementById(\"target\").innerHTML = `${data}`\n\t\t\treturn \"response\"\n\t\t})\n\t\tawait $hook(\"myHook\", await $data(\"myData\"))"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:133
}

type hookFragment struct {
	data string
	test.NoBeam
	r *test.Reporter
}

func (d *hookFragment) attr() []gox.Modify {
	return []gox.Modify{
		doors.AHook[string]{
			Name: "myHook",
			On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
				d.r.Update(ctx, 0, r.Data())
				return len(r.Data()), true
			},
		},
		doors.ARawHook{
			Name: "rawHook",
			On: func(ctx context.Context, r doors.RequestRawHook) bool {
				body, err := io.ReadAll(r.Body())
				if err != nil {
					return true
				}
				var str string
				json.Unmarshal(body, &str)
				d.r.Update(ctx, 1, str)
				fmt.Fprint(r.W(), len(str))
				return true
			},
		},
		doors.AData{
			Name: "myData",
			Value: d.data,
		},
	}
}

//line fragments.gox:171
func (f *hookFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:172
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:173
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:174
			__e = __c.Set("id", "target2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:175
			__e = __c.Modify(f.attr()...); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const a = await $hook(\"myHook\", await $data(\"myData\"))\n\t\tdocument.getElementById(\"target\").innerHTML = `${a}`\n\t\tconst b = await $hook(\"rawHook\", await $data(\"myData\"))\n\t\tdocument.getElementById(\"target2\").innerHTML = `${b}`"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:181
}

type dataFragment struct {
	data string
	test.NoBeam
}

//line fragments.gox:188
func (f *dataFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:189
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:190
			__e = __c.Set("data:myData", f.data); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("document.getElementById(\"target\").innerHTML = await $data(\"myData\")"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:193
}

type captureFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:200
func (f *captureFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:202
		f.r.Update(ctx, 0, "")
		f.r.Update(ctx, 1, "")
		f.r.Update(ctx, 2, "")
		f.r.Update(ctx, 3, "")
		f.r.Update(ctx, 4, "")

//line fragments.gox:208
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:209
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "parent")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:214
				__e = __c.Set("id", "bubble-parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line fragments.gox:215
				__e = (doors.AClick{
			StopPropagation: true,
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 1, "child")
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line fragments.gox:221
						__e = __c.Set("id", "bubble-child"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("bubble-child"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:223
		__e = (doors.AClick{
		ExactTarget: true,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 2, "exact")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:229
				__e = __c.Set("id", "exact-parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line fragments.gox:230
					__e = __c.Set("id", "exact-child"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("exact-child"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:232
			__e = __c.Set("id", "jump"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("jump"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:233
		__e = (doors.AClick{
		PreventDefault: true,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 3, "prevent")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line fragments.gox:239
				__e = __c.Set("id", "prevent-link"); if __e != nil { return }
//line fragments.gox:239
				__e = __c.Set("href", "#jump"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("prevent-link"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:240
		__e = (doors.AKeyDown{
		Filter: []string{"Enter"},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.r.Update(ctx, 4, r.Event().Key)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:246
				__e = __c.Set("id", "filter-input"); if __e != nil { return }
//line fragments.gox:246
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:247
}
