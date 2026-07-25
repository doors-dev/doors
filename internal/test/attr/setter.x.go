// Managed by GoX v0.2.3

//line setter.gox:1
package attr

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type setterFragment struct {
	test.NoBeam
	r *test.Reporter
	s doors.Setter
	n doors.Door
}

//line setter.gox:19
func (f *setterFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line setter.gox:20
		__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line setter.gox:20
				__e = __c.Set("id", "t1"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line setter.gox:21
		__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line setter.gox:21
				__e = __c.Set("id", "t2"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line setter.gox:22
				__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line setter.gox:22
						__e = __c.Set("id", "t3"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line setter.gox:24
}

//line setter.gox:26
func (f *setterFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line setter.gox:28
		f.r.Update(ctx, 0, "")

//line setter.gox:30
		__e = __c.Any(f.r); if __e != nil { return }
//line setter.gox:31
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line setter.gox:32
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line setter.gox:34
		__e = __c.Any(test.Button("set-string", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", `a&b"c"<d>`))
		return false
	})); if __e != nil { return }
//line setter.gox:38
		__e = __c.Any(test.Button("set-bare", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", true))
		return false
	})); if __e != nil { return }
//line setter.gox:42
		__e = __c.Any(test.Button("set-int", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", 42))
		return false
	})); if __e != nil { return }
//line setter.gox:46
		__e = __c.Any(test.Button("remove-nil", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", nil))
		return false
	})); if __e != nil { return }
//line setter.gox:50
		__e = __c.Any(test.Button("remove-false", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", false))
		return false
	})); if __e != nil { return }
//line setter.gox:54
		__e = __c.Any(test.Button("xset", func(ctx context.Context) bool {
		var count int
		ch := doors.Call(ctx, f.s.Set("data-count", "x").Into(&count))
		select {
		case err := <-ch:
			if err != nil {
				f.r.Update(ctx, 0, "err")
			} else {
				f.r.Update(ctx, 0, fmt.Sprint(count))
			}
		case <-ctx.Done():
		}
		return false
	})); if __e != nil { return }
//line setter.gox:68
		__e = __c.Any(test.Button("clear", func(ctx context.Context) bool {
		f.n.Inner(ctx, nil)
		return false
	})); if __e != nil { return }
//line setter.gox:72
		__e = __c.Any(test.Button("show", func(ctx context.Context) bool {
		f.n.Inner(ctx, f.content())
		return false
	})); if __e != nil { return }
	return })
//line setter.gox:76
}
