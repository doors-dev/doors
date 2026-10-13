// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line link_fragments.gox:1
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
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type linkFragment struct {
	test.Beam
	g *gate
}

//line link_fragments.gox:28
func (f *linkFragment) slow(v test.Path) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line link_fragments.gox:30
		if v.P != "" {
		f.g.entered <- struct{}{}
		select {
		case <-f.g.proceed:
		case <-ctx.Done():
		}
	}

		__e = __c.Init("span"); if __e != nil { return }
		{
//line link_fragments.gox:38
			__e = __c.Set("id", "slow"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line link_fragments.gox:38
			__e = __c.Any("slow-" + v.P); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line link_fragments.gox:39
}

//line link_fragments.gox:41
func (f *linkFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line link_fragments.gox:42
		__e = __c.Any(f.B.Bind(func(v test.Path) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line link_fragments.gox:43
				__e = __c.Set("id", "fast"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line link_fragments.gox:43
				__e = __c.Any("fast-" + v.P); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line link_fragments.gox:44
	})); if __e != nil { return }
//line link_fragments.gox:45
		__e = __c.Any(f.B.Bind(func(v test.Path) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line link_fragments.gox:46
			__e = __c.Any(f.slow(v)); if __e != nil { return }
		return })
//line link_fragments.gox:47
	})); if __e != nil { return }
//line link_fragments.gox:48
		__e = (doors.ALink{
		Model: test.Path{Vp: true, P: "batched"},
		Batch: true,
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line link_fragments.gox:51
				__e = __c.Set("id", "link-batched"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("batched"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line link_fragments.gox:52
		__e = (doors.ALink{
		Model: test.Path{Vp: true, P: "plain"},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line link_fragments.gox:54
				__e = __c.Set("id", "link-plain"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line link_fragments.gox:55
}

var _ gox.Comp = (*linkFragment)(nil)
