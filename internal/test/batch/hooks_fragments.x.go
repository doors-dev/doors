// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line hooks_fragments.gox:1
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
package batch

import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type hooksForm struct {
	Name string `form:"name"`
}

type hooksFragment struct {
	test.NoBeam
	ga *gate
	gb *gate
	x doors.Door
	y doors.Door
}

func (f *hooksFragment) hold(ctx context.Context, g *gate) {
	g.entered <- struct{}{}
	select {
	case <-g.proceed:
	case <-ctx.Done():
	}
}

func (f *hooksFragment) xy(ctx context.Context, g *gate, tag string) {
	<-f.x.Inner(ctx, "x-"+tag)
	f.hold(ctx, g)
	f.y.Inner(ctx, "y-"+tag)
}

func (f *hooksFragment) yx(ctx context.Context, g *gate, tag string) {
	<-f.y.Inner(ctx, "y-"+tag)
	f.hold(ctx, g)
	f.x.Inner(ctx, "x-"+tag)
}

func (f *hooksFragment) busy(selector string) doors.Indicators {
	return doors.IndicateAttrQuery(selector, "data-busy", "true")
}

//line hooks_fragments.gox:60
func (f *hooksFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hooks_fragments.gox:61
			__e = __c.Set("id", "ind-a"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ind-a"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hooks_fragments.gox:62
			__e = __c.Set("id", "ind-b"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ind-b"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hooks_fragments.gox:63
			__e = __c.Set("id", "seen"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line hooks_fragments.gox:64
		__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line hooks_fragments.gox:64
				__e = __c.Set("id", "x"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("x0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:65
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line hooks_fragments.gox:65
				__e = __c.Set("id", "y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("y0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:66
		__e = (doors.AClick{
		Batch: true,
		Indicator: f.busy("#ind-a"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.xy(ctx, f.ga, "a")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line hooks_fragments.gox:73
				__e = __c.Set("id", "batch-a"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("batch-a"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:74
		__e = (doors.AClick{
		Batch: true,
		Indicator: f.busy("#ind-b"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.yx(ctx, f.gb, "b")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line hooks_fragments.gox:81
				__e = __c.Set("id", "batch-b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("batch-b"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:82
		__e = (doors.AClick{
		Indicator: f.busy("#ind-b"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.x.Inner(ctx, "x-plain")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line hooks_fragments.gox:88
				__e = __c.Set("id", "plain-b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain-b"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:89
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.x.Inner(ctx, "x-after")
			f.y.Inner(ctx, "y-after")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line hooks_fragments.gox:95
				__e = __c.Set("id", "after"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("after"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:96
		__e = (doors.AClick{
		Batch: true,
		Indicator: f.busy("#ind-a"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.x.Inner(ctx, "x-panic")
			f.hold(ctx, f.ga)
			panic("hooks panic")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line hooks_fragments.gox:104
				__e = __c.Set("id", "panic"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("panic"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:105
		__e = (doors.ASubmit[hooksForm]{
		Batch: true,
		Indicator: f.busy("#ind-a"),
		On: func(ctx context.Context, r doors.RequestForm[hooksForm]) bool {
			f.xy(ctx, f.ga, r.Data().Name)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("form"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line hooks_fragments.gox:113
					__e = __c.Set("type", "text"); if __e != nil { return }
//line hooks_fragments.gox:113
					__e = __c.Set("name", "name"); if __e != nil { return }
//line hooks_fragments.gox:113
					__e = __c.Set("value", "form"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line hooks_fragments.gox:114
					__e = __c.Set("id", "submit"); if __e != nil { return }
//line hooks_fragments.gox:114
					__e = __c.Set("type", "submit"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("submit"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line hooks_fragments.gox:116
		__e = (doors.AKeyDown{
		Batch: true,
		Indicator: f.busy("#ind-a"),
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.xy(ctx, f.ga, r.Event().Key)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line hooks_fragments.gox:123
				__e = __c.Set("type", "text"); if __e != nil { return }
//line hooks_fragments.gox:123
				__e = __c.Set("id", "key"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line hooks_fragments.gox:124
			__e = __c.Set("id", "js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("js"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line hooks_fragments.gox:125
			__e = __c.Set("id", "status"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("status"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line hooks_fragments.gox:126
		__e = (doors.AHook[string]{
		Name: "hooksJs",
		Batch: true,
		Indicator: f.busy("#ind-a"),
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			f.xy(ctx, f.ga, r.Data())
			return "done", false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line hooks_fragments.gox:134
			__e = (doors.ARawHook{
		Name: "hooksStatus",
		Batch: true,
		Indicator: f.busy("#ind-a"),
		On: func(ctx context.Context, r doors.RequestRawHook) bool {
			f.x.Inner(ctx, "x-status")
			f.hold(ctx, f.ga)
			r.ResponseWriter().WriteHeader(500)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("script"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Raw("const read = (id) => document.getElementById(id).textContent\n\t\tconst seen = document.getElementById(\"seen\")\n\t\tdocument.getElementById(\"js\").addEventListener(\"click\", async () => {\n\t\t\tconst res = await $hook(\"hooksJs\", \"js\")\n\t\t\tseen.textContent = `${res}:${read(\"x\")}:${read(\"y\")}`\n\t\t})\n\t\tdocument.getElementById(\"status\").addEventListener(\"click\", () => {\n\t\t\t$hook(\"hooksStatus\", \"status\").then(\n\t\t\t\t() => {\n\t\t\t\t\tseen.textContent = `resolved:${read(\"x\")}`\n\t\t\t\t},\n\t\t\t\t() => {\n\t\t\t\t\tseen.textContent = `rejected:${read(\"x\")}`\n\t\t\t\t},\n\t\t\t)\n\t\t})"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line hooks_fragments.gox:162
}

var _ gox.Comp = (*hooksFragment)(nil)
