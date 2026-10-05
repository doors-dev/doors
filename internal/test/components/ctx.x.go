// Managed by GoX v0.3.2

//line ctx.gox:1
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

package components

import (
	"context"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type ctxTestKey struct{}

//line ctx.gox:27
func ctxValue(id string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:29
		v, ok := ctx.Value(ctxTestKey{}).(string)
	if !ok {
		v = "none"
	}

		__e = __c.Init("div"); if __e != nil { return }
		{
//line ctx.gox:34
			__e = __c.Set("id", id); if __e != nil { return }
//line ctx.gox:34
			__e = __c.Set("data-value", v); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line ctx.gox:35
}

//line ctx.gox:37
func ctxLit(id string, v string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line ctx.gox:38
			__e = __c.Set("id", id); if __e != nil { return }
//line ctx.gox:38
			__e = __c.Set("data-value", v); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line ctx.gox:39
}

func ctxHandlerValue(ctx context.Context) string {
	v, ok := ctx.Value(ctxTestKey{}).(string)
	if !ok {
		v = "none"
	}
	return v
}

type CtxFragment struct {
	test.NoBeam
	inner doors.Door
	stat doors.Door
	deep doors.Door
}

//line ctx.gox:56
func (f *CtxFragment) innerContent(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:57
		__e = __c.Any(ctxValue(label)); if __e != nil { return }
//line ctx.gox:58
		__e = (f.deep).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:59
				__e = __c.Any(ctxValue(label + "-deep")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line ctx.gox:61
}

//line ctx.gox:63
func (f *CtxFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:64
		__e = __c.Any(ctxValue("outside")); if __e != nil { return }
//line ctx.gox:65
		__e = (doors.Ctx(context.WithValue(context.Background(), ctxTestKey{}, "v1"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:66
				__e = __c.Any(ctxValue("inside")); if __e != nil { return }
//line ctx.gox:67
				__e = (doors.Ctx(context.WithValue(ctx, ctxTestKey{}, "v2"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line ctx.gox:68
						__e = __c.Any(ctxValue("override")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line ctx.gox:70
				__e = (f.inner).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line ctx.gox:71
						__e = __c.Any(f.innerContent("initial")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line ctx.gox:73
				__e = (f.stat).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line ctx.gox:74
						__e = __c.Any(ctxValue("stat-initial")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line ctx.gox:78
		canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

//line ctx.gox:81
		__e = (doors.Ctx(context.WithValue(canceledCtx, ctxTestKey{}, "vc"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:82
				__e = __c.Any(ctxValue("canceled")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line ctx.gox:84
		__e = __c.Any(test.Button("update-inner", func(ctx context.Context) bool {
		f.inner.Inner(ctx, f.innerContent("updated"))
		return true
	})); if __e != nil { return }
//line ctx.gox:88
		__e = __c.Any(test.Button("static-stat", func(ctx context.Context) bool {
		f.stat.Static(ctx, ctxValue("stat-static"))
		return true
	})); if __e != nil { return }
	return })
//line ctx.gox:92
}

type CtxRerenderFragment struct {
	test.NoBeam
	wrap doors.Door
	dyn doors.Door
	repContent doors.Door
	repContainer doors.Door
}

//line ctx.gox:102
func (f *CtxRerenderFragment) dynContent(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:103
		__e = __c.Any(ctxValue(label)); if __e != nil { return }
//line ctx.gox:104
		__e = __c.Any(test.Button("content-handler", func(ctx context.Context) bool {
		f.repContent.Inner(ctx, ctxLit("content-report", ctxHandlerValue(ctx)))
		return true
	})); if __e != nil { return }
	return })
//line ctx.gox:108
}

//line ctx.gox:110
func (f *CtxRerenderFragment) wrapContent(val string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:111
		__e = (doors.Ctx(context.WithValue(context.Background(), ctxTestKey{}, val))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:112
				__e = (f.dyn).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line ctx.gox:112
					__e = (doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.repContainer.Inner(ctx, ctxLit("container-report", ctxHandlerValue(ctx)))
				return true
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("div"); if __e != nil { return }
						{
//line ctx.gox:117
							__e = __c.Set("id", "dyn-el"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line ctx.gox:118
							__e = __c.Any(f.dynContent("dyn-" + val)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				return })); if __e != nil { return }
//line ctx.gox:120
				__e = __c.Any(test.Button("same-scope-handler", func(ctx context.Context) bool {
			f.repContent.Inner(ctx, ctxLit("same-scope-report", ctxHandlerValue(ctx)))
			return true
		})); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line ctx.gox:125
}

//line ctx.gox:127
func (f *CtxRerenderFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:128
		__e = (f.wrap).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:129
				__e = __c.Any(f.wrapContent("a")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line ctx.gox:131
		__e = __c.Any(&f.repContent); if __e != nil { return }
//line ctx.gox:132
		__e = __c.Any(&f.repContainer); if __e != nil { return }
//line ctx.gox:133
		__e = __c.Any(test.Button("rerender-wrap", func(ctx context.Context) bool {
		f.wrap.Inner(ctx, f.wrapContent("b"))
		return true
	})); if __e != nil { return }
//line ctx.gox:137
		__e = __c.Any(test.Button("update-dyn", func(ctx context.Context) bool {
		f.dyn.Inner(ctx, f.dynContent("dyn-updated"))
		return true
	})); if __e != nil { return }
	return })
//line ctx.gox:141
}
