// Managed by GoX v0.1.28

//line dyna.gox:1
package attr

import (
	"context"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type dynaFragment struct {
	test.NoBeam
	n doors.Door
	v1 string
	v2 string
}

//line dyna.gox:18
func (f *dynaFragment) content(da1 doors.AShared, da2 doors.AShared) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line dyna.gox:19
		__e = (da1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line dyna.gox:19
			__e = (da2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line dyna.gox:19
					__e = __c.Set("id", "t1"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:20
		__e = (da1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line dyna.gox:20
			__e = (da2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line dyna.gox:20
					__e = __c.Set("id", "t2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line dyna.gox:21
					__e = (da1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
//line dyna.gox:21
						__e = (da2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
							ctx := __c.Context(); _ = ctx
							__e = __c.Init("div"); if __e != nil { return }
							{
//line dyna.gox:21
								__e = __c.Set("id", "t3"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
						return })); if __e != nil { return }
					return })); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line dyna.gox:23
}

//line dyna.gox:25
func (f *dynaFragment) buttons(index string, da doors.AShared, value string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line dyna.gox:26
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			da.Enable(ctx)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:31
				__e = __c.Set("id", "enable-" + index); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("enable-"); if __e != nil { return }
//line dyna.gox:31
				__e = __c.Any(index); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:32
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			da.Disable(ctx)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:37
				__e = __c.Set("id", "disable-" + index); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("disable-"); if __e != nil { return }
//line dyna.gox:37
				__e = __c.Any(index); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:38
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			da.Update(ctx, value)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:43
				__e = __c.Set("id", "update-" + index); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-"); if __e != nil { return }
//line dyna.gox:43
				__e = __c.Any(index); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line dyna.gox:44
}

//line dyna.gox:46
func (f *dynaFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line dyna.gox:48
		da1 := doors.NewAShared("data-test1", f.v1)

//line dyna.gox:51
		da2 := doors.NewAShared("data-test2", f.v2)

//line dyna.gox:53
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line dyna.gox:54
				__e = __c.Any(f.content(da1, da2)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:56
		__e = __c.Any(f.buttons("1", da1, f.v2)); if __e != nil { return }
//line dyna.gox:57
		__e = __c.Any(f.buttons("2", da2, f.v1)); if __e != nil { return }
//line dyna.gox:58
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			da1.Update(ctx, f.v1)
			da2.Update(ctx, f.v2)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:64
				__e = __c.Set("id", "reset"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reset"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:65
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.n.Clear(ctx)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:70
				__e = __c.Set("id", "clear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("clear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:71
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.n.Update(ctx, f.content(da1, da2))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:76
				__e = __c.Set("id", "show"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("show"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line dyna.gox:77
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.n.Replace(ctx, f.content(da1, da2))
			return true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line dyna.gox:82
				__e = __c.Set("id", "replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line dyna.gox:83
}
