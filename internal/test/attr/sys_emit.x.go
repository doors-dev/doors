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

type sysEmitResult struct {
	OK       int      `json:"ok"`
	Canceled int      `json:"canceled"`
	Failed   []string `json:"failed"`
	Error    string   `json:"error"`
}

func sysEmitHook(r *test.Reporter, slot int, done *atomic.Int32) doors.AHook[sysEmitResult] {
	return doors.AHook[sysEmitResult]{
		Name: "report",
		On: func(ctx context.Context, h doors.RequestHook[sysEmitResult]) (any, bool) {
			res := h.Data()
			if res.Error != "" {
				r.Update(ctx, slot, res.Error)
			} else {
				r.Update(ctx, slot, fmt.Sprintf("ok:%d canceled:%d failed:%v", res.OK, res.Canceled, res.Failed))
			}
			r.Update(ctx, slot+1, fmt.Sprint(done.Load()))
			return nil, true
		},
	}
}

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

//line sys_emit.gox:68
func (f *sysEmitFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:70
		for i := 0; i < 4; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:74
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:75
			__e = __c.Set("id", "outer"); if __e != nil { return }
//line sys_emit.gox:75
			__e = __c.Modify(doors.A(ctx, f.hit(0))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line sys_emit.gox:76
				__e = __c.Set("id", "inner"); if __e != nil { return }
//line sys_emit.gox:76
				__e = __c.Modify(doors.A(ctx, f.hit(1))); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("target"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:78
		__e = (sysEmitHook(f.r, 2, &f.done)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("const r = await $sys.emit(\n\t\t\tdocument.getElementById(\"inner\"),\n\t\t\tnew PointerEvent(\"click\", { bubbles: true, button: 2 }),\n\t\t)\n\t\tawait $hook(\"report\", { ...r, failed: r.failed.map((e) => e.kind + \": \" + e.message) })"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:85
}

type sysEmitForm struct {
	N int `form:"n"`
}

type sysEmitSubmitFragment struct {
	test.NoBeam
	r *test.Reporter
	click bool
	fail bool
	block bool
	done atomic.Int32
}

func (f *sysEmitSubmitFragment) scope() doors.Scopes {
	if !f.block {
		return nil
	}
	return &doors.ScopeBlocking{}
}

func (f *sysEmitSubmitFragment) attrs(scope doors.Scopes) []doors.Attr {
	if !f.click {
		return nil
	}
	return []doors.Attr{doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			time.Sleep(300 * time.Millisecond)
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

//line sys_emit.gox:128
func (f *sysEmitSubmitFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:130
		for i := 0; i < 2; i++ {
		f.r.Update(ctx, i, "")
	}
	scope := f.scope()

//line sys_emit.gox:135
		__e = __c.Any(f.r); if __e != nil { return }
//line sys_emit.gox:136
		__e = (doors.ASubmit[sysEmitForm]{
		Scope: scope,
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
//line sys_emit.gox:144
					__e = __c.Set("type", "text"); if __e != nil { return }
//line sys_emit.gox:144
					__e = __c.Set("name", "n"); if __e != nil { return }
//line sys_emit.gox:144
					__e = __c.Set("value", f.value()); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line sys_emit.gox:145
					__e = __c.Set("id", "target"); if __e != nil { return }
//line sys_emit.gox:145
					__e = __c.Modify(doors.A(ctx, f.attrs(scope)...)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("go"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line sys_emit.gox:147
		__e = (sysEmitHook(f.r, 0, &f.done)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("let res\n\t\ttry {\n\t\t\tconst r = await $sys.emit(\n\t\t\t\tdocument.getElementById(\"target\"),\n\t\t\t\tnew MouseEvent(\"click\", { bubbles: true, cancelable: true }),\n\t\t\t)\n\t\t\tres = { ...r, failed: r.failed.map((e) => e.kind + \": \" + e.message) }\n\t\t} catch (e) {\n\t\t\tres = { error: e.kind + \": \" + e.message }\n\t\t}\n\t\tawait $hook(\"report\", res)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:160
}

type sysEmitLabelFragment struct {
	test.NoBeam
	r *test.Reporter
	done atomic.Int32
}

//line sys_emit.gox:168
func (f *sysEmitLabelFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:170
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")

//line sys_emit.gox:173
		__e = __c.Any(f.r); if __e != nil { return }
//line sys_emit.gox:174
		__e = (doors.AClick{
		ExactTarget: true,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.done.Add(1)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("label"); if __e != nil { return }
			{
//line sys_emit.gox:180
				__e = __c.Set("id", "label"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line sys_emit.gox:181
					__e = __c.Set("type", "checkbox"); if __e != nil { return }
//line sys_emit.gox:181
					__e = __c.Set("id", "box"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("box"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line sys_emit.gox:184
		__e = (sysEmitHook(f.r, 0, &f.done)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("let res\n\t\ttry {\n\t\t\tconst r = await $sys.emit(\n\t\t\t\tdocument.getElementById(\"label\"),\n\t\t\t\tnew PointerEvent(\"click\", { bubbles: true, cancelable: true }),\n\t\t\t)\n\t\t\tres = { ...r, failed: r.failed.map((e) => e.kind + \": \" + e.message) }\n\t\t} catch (e) {\n\t\t\tres = { error: e.kind + \": \" + e.message }\n\t\t}\n\t\tawait $hook(\"report\", res)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:197
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

//line sys_emit.gox:216
func (f *sysEmitUnrelatedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:218
		for i := 0; i < 4; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:222
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:223
			__e = __c.Set("id", "target"); if __e != nil { return }
//line sys_emit.gox:223
			__e = __c.Modify(doors.A(ctx, f.hit(0, 300*time.Millisecond))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:224
			__e = __c.Set("id", "other"); if __e != nil { return }
//line sys_emit.gox:224
			__e = __c.Modify(doors.A(ctx, f.hit(1, 0))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("other"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:225
		__e = (sysEmitHook(f.r, 2, &f.done)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("const other = document.getElementById(\"other\")\n\t\tconst p = $sys.emit(\n\t\t\tdocument.getElementById(\"target\"),\n\t\t\tnew PointerEvent(\"click\", { bubbles: true }),\n\t\t)\n\t\tother.dispatchEvent(new PointerEvent(\"click\", { bubbles: true }))\n\t\tsetTimeout(() => other.dispatchEvent(new PointerEvent(\"click\", { bubbles: true })), 50)\n\t\tconst r = await p\n\t\tawait $hook(\"report\", { ...r, failed: r.failed.map((e) => e.kind + \": \" + e.message) })"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:236
}

type sysEmitTargetFragment struct {
	test.NoBeam
	r *test.Reporter
	mode string
	done atomic.Int32
}

func (f *sysEmitTargetFragment) inert() bool {
	return f.mode == "inert" || f.mode == "shadow" || f.mode == "text"
}

//line sys_emit.gox:249
func (f *sysEmitTargetFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:251
		for i := 0; i < 3; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:255
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:256
			__e = __c.Set("inert", f.inert()); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("fieldset"); if __e != nil { return }
			{
//line sys_emit.gox:257
				__e = __c.Set("disabled", f.mode == "fieldset"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line sys_emit.gox:258
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					f.done.Add(1)
					f.r.Update(ctx, 0, "hit")
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line sys_emit.gox:264
						__e = __c.Set("id", "target"); if __e != nil { return }
//line sys_emit.gox:264
						__e = __c.Set("disabled", f.mode == "disabled"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("go"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line sys_emit.gox:266
				__e = __c.Set("id", "host"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line sys_emit.gox:267
				__e = __c.Set("id", "plain"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:269
		__e = (sysEmitHook(f.r, 1, &f.done)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
//line sys_emit.gox:269
				__e = __c.Set("data:mode", f.mode); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("const mode = await $data(\"mode\")\n\t\tlet target = document.getElementById(\"target\")\n\t\tif (mode == \"shadow\") {\n\t\t\ttarget = document.createElement(\"button\")\n\t\t\tdocument.getElementById(\"host\").attachShadow({ mode: \"open\" }).appendChild(target)\n\t\t} else if (mode == \"text\") {\n\t\t\ttarget = target.firstChild\n\t\t} else if (mode == \"none\") {\n\t\t\ttarget = document.getElementById(\"plain\")\n\t\t}\n\t\tlet res\n\t\ttry {\n\t\t\tconst r = await $sys.emit(target, new PointerEvent(\"click\", { bubbles: true }))\n\t\t\tres = { ...r, failed: r.failed.map((e) => e.kind + \": \" + e.message) }\n\t\t} catch (e) {\n\t\t\tres = { error: e.kind + \": \" + e.message }\n\t\t}\n\t\tawait $hook(\"report\", res)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:289
}
