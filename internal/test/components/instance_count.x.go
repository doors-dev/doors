// Managed by GoX v0.3.2

//line instance_count.gox:1
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
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type InstanceCountFragment struct {
	test.NoBeam
}

//line instance_count.gox:30
func (f *InstanceCountFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line instance_count.gox:31
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line instance_count.gox:31
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line instance_count.gox:32
			__e = __c.Set("id", "session-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line instance_count.gox:32
			__e = __c.Any(doors.SessionID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line instance_count.gox:33
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
//line instance_count.gox:41
				__e = __c.Set("id", "end-instance"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("end-instance"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line instance_count.gox:42
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
//line instance_count.gox:50
				__e = __c.Set("id", "end-session"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("end-session"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line instance_count.gox:51
}
