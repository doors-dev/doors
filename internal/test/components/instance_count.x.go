// Managed by GoX v0.1.32

//line instance_count.gox:1
package components

import (
	"context"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type InstanceCountFragment struct {
	test.NoBeam
}

//line instance_count.gox:16
func (f *InstanceCountFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line instance_count.gox:17
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line instance_count.gox:17
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line instance_count.gox:18
			__e = __c.Set("id", "session-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line instance_count.gox:18
			__e = __c.Any(doors.SessionID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line instance_count.gox:19
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			go func() {
				<-time.After(300 * time.Millisecond)
				doors.InstanceEnd(ctx)
			}()
			return true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line instance_count.gox:27
				__e = __c.Set("id", "end-instance"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("end-instance"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line instance_count.gox:28
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			go func() {
				<-time.After(300 * time.Millisecond)
				doors.SessionEnd(ctx)
			}()
			return true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line instance_count.gox:36
				__e = __c.Set("id", "end-session"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("end-session"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line instance_count.gox:37
}
