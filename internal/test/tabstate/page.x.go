// Managed by GoX v0.3.2

//line page.gox:1
package tabstate

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

const (
	modePlain = "plain"
	modePre = "pre"
	modePreNil = "pre-nil"
)

type tabStateFragment struct {
	test.NoBeam
	mode string
}

func ptr(v int) *int {
	return &v
}

func show(v *int) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprint(*v)
}

//line page.gox:34
func (f *tabStateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:36
		n := doors.TabState[int](ctx, "n")
	switch f.mode {
	case modePre:
		n.Update(ctx, ptr(5))
	case modePreNil:
		n.Update(ctx, ptr(5))
		n.Update(ctx, nil)
	}

		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:45
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:45
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:46
		__e = __c.Any(n.Bind(func(v *int) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:47
				__e = __c.Set("id", "value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:47
				__e = __c.Any(show(v)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:48
	})); if __e != nil { return }
//line page.gox:49
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			n.Mutate(ctx, func(v *int) *int {
				if v == nil {
					return ptr(1)
				}
				return ptr(*v + 1)
			})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line page.gox:59
				__e = __c.Set("id", "inc"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("inc"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:60
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			n.Update(ctx, nil)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line page.gox:65
				__e = __c.Set("id", "clear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("clear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:66
		__e = (doors.ALink{
		Model: test.Path{Vs: true},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:68
				__e = __c.Set("id", "go-s"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("s"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:69
		__e = (doors.ALink{
		Model: test.Path{Vh: true},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:71
				__e = __c.Set("id", "go-h"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("h"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:72
}
