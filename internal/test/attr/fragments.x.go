// Managed by GoX v0.3.2

//line fragments.gox:1
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
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type pointerFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:35
func (f *pointerFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:37
		f.r.Update(ctx, 0, "")

//line fragments.gox:39
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:40
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
//line fragments.gox:48
				__e = __c.Set("id", "down"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerDown"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:51
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
//line fragments.gox:59
				__e = __c.Set("id", "up"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerUp"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:62
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
//line fragments.gox:70
				__e = __c.Set("id", "enter"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerEnter"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:73
			__e = __c.Set("id", "beforeLeave"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("beforeLeave"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:74
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
//line fragments.gox:82
				__e = __c.Set("id", "leave"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerLeave"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:85
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
//line fragments.gox:93
				__e = __c.Set("id", "move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("PointerMove"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:96
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
//line fragments.gox:104
				__e = __c.Set("id", "over"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Over"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:107
			__e = __c.Set("id", "beforeOut"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("beforeOut"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:108
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
//line fragments.gox:116
				__e = __c.Set("id", "out"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Out"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:119
}

type callFragment struct {
	data string
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:127
func (f *callFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:128
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:129
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:130
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
//line fragments.gox:142
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
//line fragments.gox:157
}

type timeFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:164
func (f *timeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:165
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:166
		__e = __c.Any(doors.Go(func(ctx context.Context) {
		var now time.Time
		err := <-doors.Call(ctx, doors.ActionTime{}.Into(&now))
		if err != nil {
			f.r.Update(ctx, 0, err.Error())
			return
		}
		_, clientOffset := now.Zone()
		_, serverOffset := time.Now().Zone()
		f.r.Update(ctx, 0, fmt.Sprint(time.Since(now).Abs() < 5*time.Second, clientOffset == serverOffset))
	})); if __e != nil { return }
	return })
//line fragments.gox:177
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

//line fragments.gox:215
func (f *hookFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:216
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:217
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:218
			__e = __c.Set("id", "target2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:219
			__e = __c.Modify(f.attr()...); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const a = await $hook(\"myHook\", await $data(\"myData\"))\n\t\tdocument.getElementById(\"target\").innerHTML = `${a}`\n\t\tconst b = await $hook(\"rawHook\", await $data(\"myData\"))\n\t\tdocument.getElementById(\"target2\").innerHTML = `${b}`"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:225
}

type multiHookFragment struct {
	test.NoBeam
	r    *test.Reporter
	mu   sync.Mutex
	gate chan struct{}
	n    int
}

func (d *multiHookFragment) attr() []gox.Modify {
	return []gox.Modify{
		doors.ARawHook{
			Name: "multiHook",
			Race: doors.RaceParallel,
			On: func(ctx context.Context, r doors.RequestRawHook) bool {
				body, _ := io.ReadAll(r.Body())
				var str string
				json.Unmarshal(body, &str)
				d.mu.Lock()
				d.n++
				n := d.n
				if n == 3 {
					close(d.gate)
				}
				d.mu.Unlock()
				if n <= 3 {
					select {
					case <-d.gate:
					case <-time.After(3 * time.Second):
						json.NewEncoder(r.ResponseWriter()).Encode("timeout")
						return false
					}
				}
				d.r.Update(ctx, n-1, str)
				json.NewEncoder(r.ResponseWriter()).Encode(str)
				return str == "d"
			},
		},
	}
}

//line fragments.gox:267
func (f *multiHookFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:268
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:269
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:270
			__e = __c.Set("id", "target2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:271
			__e = __c.Set("id", "target3"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:272
			__e = __c.Modify(f.attr()...); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const res = await Promise.all([$hook(\"multiHook\", \"a\"), $hook(\"multiHook\", \"b\"), $hook(\"multiHook\", \"c\")])\n\t\tdocument.getElementById(\"target\").innerHTML = res.sort().join(\"\")\n\t\tdocument.getElementById(\"target2\").innerHTML = await $hook(\"multiHook\", \"d\")\n\t\ttry {\n\t\t\tawait $hook(\"multiHook\", \"e\")\n\t\t\tdocument.getElementById(\"target3\").innerHTML = \"alive\"\n\t\t} catch (e) {\n\t\t\tdocument.getElementById(\"target3\").innerHTML = \"gone\"\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:283
}

type timeoutFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:290
func (f *timeoutFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:291
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:292
			__e = __c.Set("id", "slow-default"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:293
			__e = __c.Set("id", "slow-long"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:294
		__e = (doors.AHook[string]{
		Name: "slowDefault",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			time.Sleep(2 * time.Second)
			return "late", true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line fragments.gox:300
			__e = (doors.AHook[string]{
		Name: "slowLong",
		RequestTimeout: 3 * time.Second,
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			time.Sleep(2 * time.Second)
			return "done", true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("script"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Raw("$hook(\"slowLong\", \"x\").then(\n\t\t\t(res) => {\n\t\t\t\tdocument.getElementById(\"slow-long\").innerHTML = `ok:${res}`\n\t\t\t},\n\t\t\t() => {\n\t\t\t\tdocument.getElementById(\"slow-long\").innerHTML = \"err\"\n\t\t\t},\n\t\t)\n\t\t$hook(\"slowDefault\", \"x\").then(\n\t\t\t() => {\n\t\t\t\tdocument.getElementById(\"slow-default\").innerHTML = \"ok\"\n\t\t\t},\n\t\t\t() => {\n\t\t\t\tdocument.getElementById(\"slow-default\").innerHTML = \"err\"\n\t\t\t},\n\t\t)"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:325
		__e = (doors.ASubmit[timeoutForm]{
		RequestTimeout: 3 * time.Second,
		On: func(ctx context.Context, r doors.RequestForm[timeoutForm]) bool {
			time.Sleep(2 * time.Second)
			f.r.Update(ctx, 0, r.Data().Value)
			return true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("form"); if __e != nil { return }
			{
//line fragments.gox:332
				__e = __c.Set("id", "slow-form"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line fragments.gox:333
					__e = __c.Set("type", "text"); if __e != nil { return }
//line fragments.gox:333
					__e = __c.Set("name", "Value"); if __e != nil { return }
//line fragments.gox:333
					__e = __c.Set("value", "submitted"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line fragments.gox:334
					__e = __c.Set("id", "slow-form-submit"); if __e != nil { return }
//line fragments.gox:334
					__e = __c.Set("type", "submit"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("go"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:336
}

type timeoutForm struct {
	Value string
}

type dataFragment struct {
	data string
	test.NoBeam
}

//line fragments.gox:347
func (f *dataFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line fragments.gox:348
			__e = __c.Set("id", "target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line fragments.gox:349
			__e = __c.Set("data:myData", f.data); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("document.getElementById(\"target\").innerHTML = await $data(\"myData\")"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line fragments.gox:352
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

//line fragments.gox:365
func (f *captureFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:367
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

//line fragments.gox:378
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:379
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "parent")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line fragments.gox:384
				__e = __c.Set("id", "bubble-parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line fragments.gox:385
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
//line fragments.gox:391
						__e = __c.Set("id", "bubble-child"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("bubble-child"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:393
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
//line fragments.gox:399
				__e = __c.Set("id", "exact-parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line fragments.gox:400
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
//line fragments.gox:402
			__e = __c.Set("id", "jump"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("jump"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line fragments.gox:403
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
//line fragments.gox:409
				__e = __c.Set("id", "prevent-link"); if __e != nil { return }
//line fragments.gox:409
				__e = __c.Set("href", "#jump"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("prevent-link"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:410
		__e = (doors.AKeyDown{
		Key: doors.Key{Key: "Enter"},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.filter++
			f.r.Update(ctx, 4, fmt.Sprint(f.filter))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:417
				__e = __c.Set("id", "filter-input"); if __e != nil { return }
//line fragments.gox:417
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:418
		__e = (doors.AKeyDown{
		Key: doors.Key{Key: "s", CtrlMod: doors.ModOn},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.ctrlOn++
			f.r.Update(ctx, 5, fmt.Sprint(f.ctrlOn))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:425
				__e = __c.Set("id", "keys-ctrl"); if __e != nil { return }
//line fragments.gox:425
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:426
		__e = (doors.AKeyDown{
		Key: doors.Key{Key: "d", CtrlMod: doors.ModOff},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.ctrlOff++
			f.r.Update(ctx, 6, fmt.Sprint(f.ctrlOff))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:433
				__e = __c.Set("id", "keys-ctrl-off"); if __e != nil { return }
//line fragments.gox:433
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:434
		__e = (doors.AKeyDown{
		Key: doors.Key{Key: "e", MetaMod: doors.ModOn},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.metaOn++
			f.r.Update(ctx, 7, fmt.Sprint(f.metaOn))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:441
				__e = __c.Set("id", "keys-meta"); if __e != nil { return }
//line fragments.gox:441
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:442
		__e = (doors.AKeyDown{
		Key: doors.Key{Key: "a"}.And(doors.Key{Key: "b", ShiftMod: doors.ModOn}),
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.multi++
			f.r.Update(ctx, 8, fmt.Sprint(f.multi))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:449
				__e = __c.Set("id", "keys-multi"); if __e != nil { return }
//line fragments.gox:449
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line fragments.gox:450
		__e = (doors.AKeyDown{
		Key: doors.Key{Key: "", AltMod: doors.ModOn},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.anyKey++
			f.r.Update(ctx, 9, fmt.Sprint(f.anyKey))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line fragments.gox:457
				__e = __c.Set("id", "keys-any"); if __e != nil { return }
//line fragments.gox:457
				__e = __c.Set("type", "text"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:458
}

type pointerCoordsFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line fragments.gox:465
func (f *pointerCoordsFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line fragments.gox:466
		__e = __c.Any(f.r); if __e != nil { return }
//line fragments.gox:467
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
//line fragments.gox:494
				__e = __c.Set("id", "coord-target"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("click-me"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line fragments.gox:495
}
