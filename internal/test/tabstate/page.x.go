// Managed by GoX v0.3.2

//line page.gox:1
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

//line page.gox:48
func (f *tabStateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:50
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
//line page.gox:59
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:59
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:60
		__e = __c.Any(n.Bind(func(v *int) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:61
				__e = __c.Set("id", "value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:61
				__e = __c.Any(show(v)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:62
	})); if __e != nil { return }
//line page.gox:63
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
//line page.gox:73
				__e = __c.Set("id", "inc"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("inc"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:74
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			n.Update(ctx, nil)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line page.gox:79
				__e = __c.Set("id", "clear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("clear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:80
		__e = (doors.ALink{
		Model: test.Path{Vs: true},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:82
				__e = __c.Set("id", "go-s"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("s"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:83
		__e = (doors.ALink{
		Model: test.Path{Vh: true},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:85
				__e = __c.Set("id", "go-h"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("h"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:86
}
