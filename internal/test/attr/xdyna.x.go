// Managed by GoX v0.2.3

//line xdyna.gox:1
package attr

import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type xdynaFragment struct {
	test.NoBeam
	r  *test.Reporter
	v1 string
	v2 string
}

func (f *xdynaFragment) outcome(ctx context.Context, slot int, ch <-chan error) {
	err, ok := <-ch
	switch {
	case !ok:
		f.r.Update(ctx, slot, "noop")
	case err != nil:
		f.r.Update(ctx, slot, "err")
	default:
		f.r.Update(ctx, slot, "ok")
	}
}

//line xdyna.gox:30
func (f *xdynaFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line xdyna.gox:32
		da := doors.NewAShared("data-x", f.v1)

//line xdyna.gox:34
		__e = __c.Any(f.r); if __e != nil { return }
//line xdyna.gox:35
		__e = (da).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line xdyna.gox:35
				__e = __c.Set("id", "target"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line xdyna.gox:36
		__e = __c.Any(test.Button("x-update", func(ctx context.Context) bool {
		f.outcome(ctx, 0, da.XUpdate(ctx, f.v2))
		return false
	})); if __e != nil { return }
//line xdyna.gox:40
		__e = __c.Any(test.Button("x-update-same", func(ctx context.Context) bool {
		f.outcome(ctx, 0, da.XUpdate(ctx, f.v2))
		return false
	})); if __e != nil { return }
//line xdyna.gox:44
		__e = __c.Any(test.Button("x-disable", func(ctx context.Context) bool {
		f.outcome(ctx, 1, da.XDisable(ctx))
		return false
	})); if __e != nil { return }
//line xdyna.gox:48
		__e = __c.Any(test.Button("x-disable-again", func(ctx context.Context) bool {
		f.outcome(ctx, 1, da.XDisable(ctx))
		return false
	})); if __e != nil { return }
//line xdyna.gox:52
		__e = __c.Any(test.Button("x-enable", func(ctx context.Context) bool {
		f.outcome(ctx, 2, da.XEnable(ctx))
		return false
	})); if __e != nil { return }
	return })
//line xdyna.gox:56
}
