// Managed by GoX v0.4.0

//line sys_emit.gox:1
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
	"fmt"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type sysEmitFragment struct {
	test.NoBeam
	r *test.Reporter
	done atomic.Int32
}

func (f *sysEmitFragment) hit(slot int) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			time.Sleep(200 * time.Millisecond)
			f.done.Add(1)
			f.r.Update(ctx, slot, fmt.Sprint(r.Event().Button))
			return false
		},
	}
}

//line sys_emit.gox:45
func (f *sysEmitFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:47
		for i := 0; i < 4; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:51
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:52
			__e = __c.Set("id", "outer"); if __e != nil { return }
//line sys_emit.gox:52
			__e = __c.Modify(doors.A(ctx, f.hit(0))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line sys_emit.gox:53
				__e = __c.Set("id", "inner"); if __e != nil { return }
//line sys_emit.gox:53
				__e = __c.Modify(doors.A(ctx, f.hit(1))); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("target"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:55
		__e = (doors.AHook[int]{
		Name: "count",
		On: func(ctx context.Context, r doors.RequestHook[int]) (any, bool) {
			f.r.Update(ctx, 2, fmt.Sprint(r.Data()))
			f.r.Update(ctx, 3, fmt.Sprint(f.done.Load()))
			return nil, true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("const n = await $sys.emit(\n\t\t\tdocument.getElementById(\"inner\"),\n\t\t\tnew PointerEvent(\"click\", { bubbles: true, button: 2 }),\n\t\t)\n\t\tawait $hook(\"count\", n)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:69
}

type sysEmitForm struct {
	N int `form:"n"`
}

type sysEmitSubmitFragment struct {
	test.NoBeam
	r *test.Reporter
	click bool
	fail bool
	done atomic.Int32
}

func (f *sysEmitSubmitFragment) attrs() []doors.Attr {
	if !f.click {
		return nil
	}
	return []doors.Attr{doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.done.Add(1)
			return false
		},
	}}
}

func (f *sysEmitSubmitFragment) value() string {
	if f.fail {
		return "x"
	}
	return "1"
}

//line sys_emit.gox:102
func (f *sysEmitSubmitFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:104
		for i := 0; i < 2; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:108
		__e = __c.Any(f.r); if __e != nil { return }
//line sys_emit.gox:109
		__e = (doors.ASubmit[sysEmitForm]{
		On: func(ctx context.Context, r doors.RequestForm[sysEmitForm]) bool {
			time.Sleep(200 * time.Millisecond)
			f.done.Add(1)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("form"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line sys_emit.gox:116
					__e = __c.Set("type", "text"); if __e != nil { return }
//line sys_emit.gox:116
					__e = __c.Set("name", "n"); if __e != nil { return }
//line sys_emit.gox:116
					__e = __c.Set("value", f.value()); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line sys_emit.gox:117
					__e = __c.Set("id", "target"); if __e != nil { return }
//line sys_emit.gox:117
					__e = __c.Modify(doors.A(ctx, f.attrs()...)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("go"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line sys_emit.gox:119
		__e = (doors.AHook[int]{
		Name: "count",
		On: func(ctx context.Context, r doors.RequestHook[int]) (any, bool) {
			f.r.Update(ctx, 0, fmt.Sprint(r.Data()))
			f.r.Update(ctx, 1, fmt.Sprint(f.done.Load()))
			return nil, true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("let n = -1\n\t\ttry {\n\t\t\tn = await $sys.emit(\n\t\t\t\tdocument.getElementById(\"target\"),\n\t\t\t\tnew MouseEvent(\"click\", { bubbles: true, cancelable: true }),\n\t\t\t)\n\t\t} catch {\n\t\t}\n\t\tawait $hook(\"count\", n)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:137
}

type sysEmitUnrelatedFragment struct {
	test.NoBeam
	r *test.Reporter
	done atomic.Int32
}

func (f *sysEmitUnrelatedFragment) hit(slot int, delay time.Duration) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			time.Sleep(delay)
			f.done.Add(1)
			f.r.Update(ctx, slot, "hit")
			return false
		},
	}
}

//line sys_emit.gox:156
func (f *sysEmitUnrelatedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:158
		for i := 0; i < 4; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:162
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:163
			__e = __c.Set("id", "target"); if __e != nil { return }
//line sys_emit.gox:163
			__e = __c.Modify(doors.A(ctx, f.hit(0, 300*time.Millisecond))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:164
			__e = __c.Set("id", "other"); if __e != nil { return }
//line sys_emit.gox:164
			__e = __c.Modify(doors.A(ctx, f.hit(1, 0))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("other"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:165
		__e = (doors.AHook[int]{
		Name: "count",
		On: func(ctx context.Context, r doors.RequestHook[int]) (any, bool) {
			f.r.Update(ctx, 2, fmt.Sprint(r.Data()))
			return nil, true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("const other = document.getElementById(\"other\")\n\t\tconst p = $sys.emit(\n\t\t\tdocument.getElementById(\"target\"),\n\t\t\tnew PointerEvent(\"click\", { bubbles: true }),\n\t\t)\n\t\tother.dispatchEvent(new PointerEvent(\"click\", { bubbles: true }))\n\t\tsetTimeout(() => other.dispatchEvent(new PointerEvent(\"click\", { bubbles: true })), 50)\n\t\tawait $hook(\"count\", await p)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:181
}

type sysEmitDisabledFragment struct {
	test.NoBeam
	r *test.Reporter
	mode string
}

//line sys_emit.gox:189
func (f *sysEmitDisabledFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:191
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")

//line sys_emit.gox:194
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:195
			__e = __c.Set("inert", f.mode == "inert"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("fieldset"); if __e != nil { return }
			{
//line sys_emit.gox:196
				__e = __c.Set("disabled", f.mode == "fieldset"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line sys_emit.gox:197
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					f.r.Update(ctx, 0, "hit")
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line sys_emit.gox:202
						__e = __c.Set("id", "target"); if __e != nil { return }
//line sys_emit.gox:202
						__e = __c.Set("disabled", f.mode == "disabled"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("go"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:205
		__e = (doors.AHook[string]{
		Name: "result",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			f.r.Update(ctx, 1, r.Data())
			return nil, true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("let res = \"ok\"\n\t\ttry {\n\t\t\tawait $sys.emit(\n\t\t\t\tdocument.getElementById(\"target\"),\n\t\t\t\tnew PointerEvent(\"click\", { bubbles: true }),\n\t\t\t)\n\t\t} catch (e) {\n\t\t\tres = `${e.kind}: ${e.message}`\n\t\t}\n\t\tawait $hook(\"result\", res)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:223
}
