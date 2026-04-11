// Managed by GoX v0.1.25

//line lifecycle.gox:1
package components

import (
	"context"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type lifecycleSessionKey struct{}
type lifecycleInstanceKey struct{}

type LifecycleFragment struct {
	test.NoBeam
}

//line lifecycle.gox:19
func (f *LifecycleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line lifecycle.gox:20
			__e = __c.AttrSet("id", "session-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line lifecycle.gox:20
			__e = __c.Any(doors.SessionId(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line lifecycle.gox:21
			__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line lifecycle.gox:21
			__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line lifecycle.gox:22
			__e = __c.AttrSet("id", "session-marker"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line lifecycle.gox:23
			__e = __c.Any(doors.SessionStore(ctx).Init(lifecycleSessionKey{}, func() any {
			return common.RandId()
		}).(string)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line lifecycle.gox:27
			__e = __c.AttrSet("id", "instance-marker"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line lifecycle.gox:28
			__e = __c.Any(doors.InstanceStore(ctx).Init(lifecycleInstanceKey{}, func() any {
			return common.RandId()
		}).(string)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line lifecycle.gox:32
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			doors.InstanceEnd(ctx)
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line lifecycle.gox:37
				__e = __c.AttrSet("id", "end-instance"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("end-instance"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line lifecycle.gox:38
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			doors.SessionEnd(ctx)
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line lifecycle.gox:43
				__e = __c.AttrSet("id", "end-session"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("end-session"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line lifecycle.gox:44
}
