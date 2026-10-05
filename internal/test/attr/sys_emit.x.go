// Managed by GoX v0.3.2

//line sys_emit.gox:1
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

package attr

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type sysEmitFragment struct {
	test.NoBeam
	r *test.Reporter
	done atomic.Int32
}

func (f *sysEmitFragment) hit(slot int) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			time.Sleep(200 * time.Millisecond)
			f.done.Add(1)
			f.r.Update(ctx, slot, fmt.Sprint(r.Event().Button))
			return false
		},
	}
}

//line sys_emit.gox:45
func (f *sysEmitFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line sys_emit.gox:47
		for i := 0; i < 4; i++ {
		f.r.Update(ctx, i, "")
	}

//line sys_emit.gox:51
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line sys_emit.gox:52
			__e = __c.Set("id", "outer"); if __e != nil { return }
//line sys_emit.gox:52
			__e = __c.Modify(doors.A(ctx, f.hit(0))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line sys_emit.gox:53
				__e = __c.Set("id", "inner"); if __e != nil { return }
//line sys_emit.gox:53
				__e = __c.Modify(doors.A(ctx, f.hit(1))); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("target"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line sys_emit.gox:55
		__e = (doors.AHook[int]{
		Name: "count",
		On: func(ctx context.Context, r doors.RequestHook[int]) (any, bool) {
			f.r.Update(ctx, 2, fmt.Sprint(r.Data()))
			f.r.Update(ctx, 3, fmt.Sprint(f.done.Load()))
			return nil, true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("const n = await $sys.emit(\n\t\t\tdocument.getElementById(\"inner\"),\n\t\t\tnew PointerEvent(\"click\", { bubbles: true, button: 2 }),\n\t\t)\n\t\tawait $hook(\"count\", n)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line sys_emit.gox:69
}
