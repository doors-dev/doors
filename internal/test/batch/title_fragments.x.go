// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line title_fragments.gox:1
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

package batch

import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type titleFragment struct {
	test.NoBeam
	g *gate
	a doors.Door
	b doors.Door
}

//line title_fragments.gox:32
func titleContent(title string, text string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("title"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line title_fragments.gox:33
			__e = __c.Any(title); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("span"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line title_fragments.gox:34
			__e = __c.Any(text); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line title_fragments.gox:35
}

func (f *titleFragment) retitle(ctx context.Context) bool {
	f.a.Inner(ctx, titleContent("second", "a1"))
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
	f.b.Inner(ctx, "b1")
	return false
}

//line title_fragments.gox:48
func (f *titleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line title_fragments.gox:49
		__e = (f.a).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line title_fragments.gox:49
				__e = __c.Set("id", "a"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line title_fragments.gox:49
				__e = __c.Any(titleContent("first", "a0")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line title_fragments.gox:50
		__e = (f.b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line title_fragments.gox:50
				__e = __c.Set("id", "b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("b0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line title_fragments.gox:51
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.retitle(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line title_fragments.gox:56
				__e = __c.Set("id", "title-batched"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("title-batched"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line title_fragments.gox:57
}
