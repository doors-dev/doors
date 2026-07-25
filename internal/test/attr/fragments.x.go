// Managed by GoX v0.2.3

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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			f.r.Update(ctx, 1, test.Float(r.Event().PageX()))
			f.r.Update(ctx, 2, test.Float(r.Event().PageY()))
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
			var res string
			<-doors.Call(ctx, doors.ActionEmit[string]{Name: "myCall", Arg: len(r.Data())}.Into(&res))
			f.r.Update(ctx, 1, res)
			var asyncRes string
			<-doors.Call(ctx, doors.ActionEmit[string]{Name: "myAsyncCall", Arg: r.Data()}.Into(&asyncRes))
			f.r.Update(ctx, 2, asyncRes)
			return len(r.Data()), true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line fragments.gox:126
			__e = (doors.AData{
		Name: "myData",
		Value: f.data,
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("script"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Raw("$on(\"myCall\", (data) => {\n\t\t\tdocument.getElementById(\"target\").innerHTML = `${data}`\n\t\t\treturn \"response\"\n\t\t})\n\t\t$on(\n\t\t\t\"myAsyncCall\",\n\t\t\t(data) =>\n\t\t\t\tnew Promise((resolve) => setTimeout(() => resolve(`async:${data}`), 100)),\n\t\t)\n\t\tawait $hook(\"myHook\", await $data(\"myData\"))"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:141
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
				fmt.Fprint(r.ResponseWriter(), len(str))
				return true
			},
		},
		doors.AData{
			Name: "myData",
			Value: d.data,
		},
	}
}

//line fragments.gox:179
func (f *hookFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:180
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:181
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:182
			__e = __c.Set("id", "target2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:183
			__e = __c.Modify(f.attr()...); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const a = await $hook(\"myHook\", await $data(\"myData\"))\n\t\tdocument.getElementById(\"target\").innerHTML = `${a}`\n\t\tconst b = await $hook(\"rawHook\", await $data(\"myData\"))\n\t\tdocument.getElementById(\"target2\").innerHTML = `${b}`"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:189
}

type dataFragment struct {
	data string
	test.NoBeam
}

//line fragments.gox:196
func (f *dataFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:197
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:198
			__e = __c.Set("data:myData", f.data); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("document.getElementById(\"target\").innerHTML = await $data(\"myData\")"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:201
}

type captureFragment struct {
	test.NoBeam
	r *test.Reporter
	filter int
	ctrlOn int
	ctrlOff int
	metaOn int
	multi int
	anyKey int
}

//line fragments.gox:214
func (f *captureFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:216
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")
	f.r.Update(ctx, 2, "")
	f.r.Update(ctx, 3, "")
	f.r.Update(ctx, 4, "")
	f.r.Update(ctx, 5, "")
	f.r.Update(ctx, 6, "")
	f.r.Update(ctx, 7, "")
	f.r.Update(ctx, 8, "")
	f.r.Update(ctx, 9, "")

//line fragments.gox:227
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:228
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "parent")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:233
				__e = __c.Set("id", "bubble-parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line fragments.gox:234
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
//line fragments.gox:240
						__e = __c.Set("id", "bubble-child"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("bubble-child"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:242
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
//line fragments.gox:248
				__e = __c.Set("id", "exact-parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line fragments.gox:249
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
//line fragments.gox:251
			__e = __c.Set("id", "jump"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("jump"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:252
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
//line fragments.gox:258
				__e = __c.Set("id", "prevent-link"); if __e != nil { return }
//line fragments.gox:258
				__e = __c.Set("href", "#jump"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("prevent-link"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:259
		__e = (doors.AKeyDown{
		Filter: []string{"Enter"},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.filter++
			f.r.Update(ctx, 4, fmt.Sprint(f.filter))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:266
				__e = __c.Set("id", "filter-input"); if __e != nil { return }
//line fragments.gox:266
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:267
		__e = (doors.AKeyDown{
		Keys: []doors.Key{{Key: "s", CtrlMod: doors.ModOn}},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.ctrlOn++
			f.r.Update(ctx, 5, fmt.Sprint(f.ctrlOn))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:274
				__e = __c.Set("id", "keys-ctrl"); if __e != nil { return }
//line fragments.gox:274
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:275
		__e = (doors.AKeyDown{
		Keys: []doors.Key{{Key: "d", CtrlMod: doors.ModOff}},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.ctrlOff++
			f.r.Update(ctx, 6, fmt.Sprint(f.ctrlOff))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:282
				__e = __c.Set("id", "keys-ctrl-off"); if __e != nil { return }
//line fragments.gox:282
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:283
		__e = (doors.AKeyDown{
		Keys: []doors.Key{{Key: "e", MetaMod: doors.ModOn}},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.metaOn++
			f.r.Update(ctx, 7, fmt.Sprint(f.metaOn))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:290
				__e = __c.Set("id", "keys-meta"); if __e != nil { return }
//line fragments.gox:290
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:291
		__e = (doors.AKeyDown{
		Keys: []doors.Key{{Key: "a"}, {Key: "b", ShiftMod: doors.ModOn}},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.multi++
			f.r.Update(ctx, 8, fmt.Sprint(f.multi))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:298
				__e = __c.Set("id", "keys-multi"); if __e != nil { return }
//line fragments.gox:298
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:299
		__e = (doors.AKeyDown{
		Keys: []doors.Key{{Key: "", AltMod: doors.ModOn}},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.anyKey++
			f.r.Update(ctx, 9, fmt.Sprint(f.anyKey))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:306
				__e = __c.Set("id", "keys-any"); if __e != nil { return }
//line fragments.gox:306
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:307
}

type pointerCoordsFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:314
func (f *pointerCoordsFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:315
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:316
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			e := r.Event()
			f.r.Update(ctx, 0, test.Float(e.OffsetX()))
			f.r.Update(ctx, 1, test.Float(e.OffsetY()))
			f.r.Update(ctx, 2, test.Float(e.ClientX()))
			f.r.Update(ctx, 3, test.Float(e.ClientY()))
			f.r.Update(ctx, 4, test.Float(e.PageX()))
			f.r.Update(ctx, 5, test.Float(e.PageY()))
			f.r.Update(ctx, 6, test.Float(e.ScreenX()))
			f.r.Update(ctx, 7, test.Float(e.ScreenY()))
			f.r.Update(ctx, 8, test.Float(e.Pointer.Width))
			f.r.Update(ctx, 9, test.Float(e.Pointer.Height))
			f.r.Update(ctx, 10, test.Float(e.Target.X))
			f.r.Update(ctx, 11, test.Float(e.Target.Y))
			f.r.Update(ctx, 12, test.Float(e.Target.Width))
			f.r.Update(ctx, 13, test.Float(e.Target.Height))
			f.r.Update(ctx, 14, test.Float(e.Page.X))
			f.r.Update(ctx, 15, test.Float(e.Page.Y))
			f.r.Update(ctx, 16, test.Float(e.Page.Width))
			f.r.Update(ctx, 17, test.Float(e.Page.Height))
			f.r.Update(ctx, 18, test.Float(e.Screen.X))
			f.r.Update(ctx, 19, test.Float(e.Screen.Y))
			f.r.Update(ctx, 20, test.Float(e.Screen.Width))
			f.r.Update(ctx, 21, test.Float(e.Screen.Height))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:343
				__e = __c.Set("id", "coord-target"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("click-me"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:344
}
