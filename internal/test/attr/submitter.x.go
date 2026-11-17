// Managed by GoX v0.4.0

//line submitter.gox:1
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

type submitterFragment struct {
	test.NoBeam
	r     *test.Reporter
	count atomic.Int32
}

type submitterData struct {
	Title  string `form:"title"`
	Intent string `form:"intent"`
}

//line submitter.gox:39
func (f *submitterFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line submitter.gox:40
		__e = __c.Any(f.r); if __e != nil { return }
//line submitter.gox:41
		__e = (doors.ASubmit[submitterData]{
		Indicator: doors.IndicateAttrSubmitter("data-busy", "yes"),
		On: func(ctx context.Context, r doors.RequestForm[submitterData]) bool {
			<-time.After(400 * time.Millisecond)
			intent := r.Data().Intent
			if intent == "" {
				intent = "none"
			}
			f.r.Update(ctx, 0, intent)
			f.r.Update(ctx, 1, fmt.Sprint(f.count.Add(1)))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("form"); if __e != nil { return }
			{
//line submitter.gox:53
				__e = __c.Set("id", "sform"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line submitter.gox:54
					__e = __c.Set("type", "text"); if __e != nil { return }
//line submitter.gox:54
					__e = __c.Set("id", "title"); if __e != nil { return }
//line submitter.gox:54
					__e = __c.Set("name", "title"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line submitter.gox:55
					__e = __c.Set("id", "draft"); if __e != nil { return }
//line submitter.gox:55
					__e = __c.Set("name", "intent"); if __e != nil { return }
//line submitter.gox:55
					__e = __c.Set("value", "draft"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("Draft"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line submitter.gox:56
					__e = __c.Set("id", "publish"); if __e != nil { return }
//line submitter.gox:56
					__e = __c.Set("name", "intent"); if __e != nil { return }
//line submitter.gox:56
					__e = __c.Set("value", "publish"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("Publish"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line submitter.gox:58
}
