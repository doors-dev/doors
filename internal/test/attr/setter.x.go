// Managed by GoX v0.3.2

//line setter.gox:1
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

//line setter.gox:33
func (f *setterFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line setter.gox:34
		__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line setter.gox:34
				__e = __c.Set("id", "t1"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line setter.gox:35
		__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line setter.gox:35
				__e = __c.Set("id", "t2"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line setter.gox:36
				__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line setter.gox:36
						__e = __c.Set("id", "t3"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line setter.gox:38
}

//line setter.gox:40
func (f *setterFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line setter.gox:42
		f.r.Update(ctx, 0, "")

//line setter.gox:44
		__e = __c.Any(f.r); if __e != nil { return }
//line setter.gox:45
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line setter.gox:46
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line setter.gox:48
		__e = __c.Any(test.Button("set-string", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", `a&b"c"<d>`))
		return false
	})); if __e != nil { return }
//line setter.gox:52
		__e = __c.Any(test.Button("set-bare", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", true))
		return false
	})); if __e != nil { return }
//line setter.gox:56
		__e = __c.Any(test.Button("set-int", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", 42))
		return false
	})); if __e != nil { return }
//line setter.gox:60
		__e = __c.Any(test.Button("remove-nil", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", nil))
		return false
	})); if __e != nil { return }
//line setter.gox:64
		__e = __c.Any(test.Button("remove-false", func(ctx context.Context) bool {
		doors.Call(ctx, f.s.Set("data-test", false))
		return false
	})); if __e != nil { return }
//line setter.gox:68
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
//line setter.gox:82
		__e = __c.Any(test.Button("clear", func(ctx context.Context) bool {
		f.n.Inner(ctx, nil)
		return false
	})); if __e != nil { return }
//line setter.gox:86
		__e = __c.Any(test.Button("show", func(ctx context.Context) bool {
		f.n.Inner(ctx, f.content())
		return false
	})); if __e != nil { return }
	return })
//line setter.gox:90
}
