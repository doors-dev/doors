// Managed by GoX v0.2.3

//line ctx.gox:1
package components

import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type ctxTestKey struct{}

//line ctx.gox:13
func ctxValue(id string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:15
		v, ok := ctx.Value(ctxTestKey{}).(string)
	if !ok {
		v = "none"
	}

		__e = __c.Init("div"); if __e != nil { return }
		{
//line ctx.gox:20
			__e = __c.Set("id", id); if __e != nil { return }
//line ctx.gox:20
			__e = __c.Set("data-value", v); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line ctx.gox:21
}

//line ctx.gox:23
func ctxLit(id string, v string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line ctx.gox:24
			__e = __c.Set("id", id); if __e != nil { return }
//line ctx.gox:24
			__e = __c.Set("data-value", v); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line ctx.gox:25
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
	stat  doors.Door
	deep  doors.Door
}

//line ctx.gox:42
func (f *CtxFragment) innerContent(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:43
		__e = __c.Any(ctxValue(label)); if __e != nil { return }
//line ctx.gox:44
		__e = (f.deep).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:45
				__e = __c.Any(ctxValue(label + "-deep")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line ctx.gox:47
}

//line ctx.gox:49
func (f *CtxFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:50
		__e = __c.Any(ctxValue("outside")); if __e != nil { return }
//line ctx.gox:51
		__e = (doors.Ctx(context.WithValue(context.Background(), ctxTestKey{}, "v1"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:52
				__e = __c.Any(ctxValue("inside")); if __e != nil { return }
//line ctx.gox:53
				__e = (doors.Ctx(context.WithValue(ctx, ctxTestKey{}, "v2"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line ctx.gox:54
						__e = __c.Any(ctxValue("override")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line ctx.gox:56
				__e = (f.inner).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line ctx.gox:57
						__e = __c.Any(f.innerContent("initial")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line ctx.gox:59
				__e = (f.stat).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line ctx.gox:60
						__e = __c.Any(ctxValue("stat-initial")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line ctx.gox:64
		canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

//line ctx.gox:67
		__e = (doors.Ctx(context.WithValue(canceledCtx, ctxTestKey{}, "vc"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:68
				__e = __c.Any(ctxValue("canceled")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line ctx.gox:70
		__e = __c.Any(test.Button("update-inner", func(ctx context.Context) bool {
		f.inner.Inner(ctx, f.innerContent("updated"))
		return true
	})); if __e != nil { return }
//line ctx.gox:74
		__e = __c.Any(test.Button("static-stat", func(ctx context.Context) bool {
		f.stat.Static(ctx, ctxValue("stat-static"))
		return true
	})); if __e != nil { return }
	return })
//line ctx.gox:78
}

type CtxRerenderFragment struct {
	test.NoBeam
	wrap         doors.Door
	dyn          doors.Door
	repContent   doors.Door
	repContainer doors.Door
}

//line ctx.gox:88
func (f *CtxRerenderFragment) dynContent(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:89
		__e = __c.Any(ctxValue(label)); if __e != nil { return }
//line ctx.gox:90
		__e = __c.Any(test.Button("content-handler", func(ctx context.Context) bool {
		f.repContent.Inner(ctx, ctxLit("content-report", ctxHandlerValue(ctx)))
		return true
	})); if __e != nil { return }
	return })
//line ctx.gox:94
}

//line ctx.gox:96
func (f *CtxRerenderFragment) wrapContent(val string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:97
		__e = (doors.Ctx(context.WithValue(context.Background(), ctxTestKey{}, val))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:98
				__e = (f.dyn).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line ctx.gox:98
					__e = (doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.repContainer.Inner(ctx, ctxLit("container-report", ctxHandlerValue(ctx)))
				return true
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("div"); if __e != nil { return }
						{
//line ctx.gox:103
							__e = __c.Set("id", "dyn-el"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line ctx.gox:104
							__e = __c.Any(f.dynContent("dyn-" + val)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				return })); if __e != nil { return }
//line ctx.gox:106
				__e = __c.Any(test.Button("same-scope-handler", func(ctx context.Context) bool {
			f.repContent.Inner(ctx, ctxLit("same-scope-report", ctxHandlerValue(ctx)))
			return true
		})); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line ctx.gox:111
}

//line ctx.gox:113
func (f *CtxRerenderFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line ctx.gox:114
		__e = (f.wrap).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line ctx.gox:115
				__e = __c.Any(f.wrapContent("a")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line ctx.gox:117
		__e = __c.Any(&f.repContent); if __e != nil { return }
//line ctx.gox:118
		__e = __c.Any(&f.repContainer); if __e != nil { return }
//line ctx.gox:119
		__e = __c.Any(test.Button("rerender-wrap", func(ctx context.Context) bool {
		f.wrap.Inner(ctx, f.wrapContent("b"))
		return true
	})); if __e != nil { return }
//line ctx.gox:123
		__e = __c.Any(test.Button("update-dyn", func(ctx context.Context) bool {
		f.dyn.Inner(ctx, f.dynContent("dyn-updated"))
		return true
	})); if __e != nil { return }
	return })
//line ctx.gox:127
}
