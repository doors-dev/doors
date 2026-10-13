// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line path_fragments.gox:1
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

type pathFragment struct {
	test.Beam
	g *gate
	done chan struct{}
	a doors.Door
	b doors.Door
}

//line path_fragments.gox:32
func pathMetaContent(desc string, text string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("meta"); if __e != nil { return }
		{
//line path_fragments.gox:33
			__e = __c.Set("name", "pathdesc"); if __e != nil { return }
//line path_fragments.gox:33
			__e = __c.Set("content", desc); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("span"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line path_fragments.gox:34
			__e = __c.Any(text); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line path_fragments.gox:35
}

func (f *pathFragment) pathHold(ctx context.Context) {
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
}

func (f *pathFragment) pathMove(ctx context.Context, p string, first string, second string) bool {
	f.B.Update(ctx, test.Path{Vp: true, P: p})
	f.a.Inner(ctx, first)
	f.pathHold(ctx)
	f.b.Inner(ctx, second)
	return false
}

func (f *pathFragment) pathOnly(ctx context.Context, p string) bool {
	f.B.Update(ctx, test.Path{Vp: true, P: p})
	f.pathHold(ctx)
	return false
}

func (f *pathFragment) pathHolder(ctx context.Context) bool {
	<-f.a.Inner(ctx, "a-held")
	f.pathHold(ctx)
	return false
}

func (f *pathFragment) pathFollow(ctx context.Context, p string) bool {
	f.B.Update(ctx, test.Path{Vp: true, P: p})
	f.a.Inner(ctx, "a-"+p)
	f.done <- struct{}{}
	return false
}

func (f *pathFragment) pathMeta(ctx context.Context) bool {
	f.a.Inner(ctx, pathMetaContent("second", "a3"))
	f.pathHold(ctx)
	f.b.Inner(ctx, "b3")
	return false
}

//line path_fragments.gox:79
func (f *pathFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line path_fragments.gox:80
		__e = (f.a).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line path_fragments.gox:80
				__e = __c.Set("id", "a"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line path_fragments.gox:80
				__e = __c.Any(pathMetaContent("first", "a0")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line path_fragments.gox:81
		__e = (f.b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line path_fragments.gox:81
				__e = __c.Set("id", "b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("b0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line path_fragments.gox:82
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.pathMove(ctx, "batched", "a1", "b1")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line path_fragments.gox:87
				__e = __c.Set("id", "path-batched"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("path-batched"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line path_fragments.gox:88
		__e = __c.Any(test.Button("path-plain", func(ctx context.Context) bool {
		return f.pathMove(ctx, "plain", "a2", "b2")
	})); if __e != nil { return }
//line path_fragments.gox:91
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.pathOnly(ctx, "only")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line path_fragments.gox:96
				__e = __c.Set("id", "path-only"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("path-only"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line path_fragments.gox:97
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.pathHolder(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line path_fragments.gox:102
				__e = __c.Set("id", "path-holder"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("path-holder"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line path_fragments.gox:103
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.pathFollow(ctx, "follow")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line path_fragments.gox:108
				__e = __c.Set("id", "path-follow"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("path-follow"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line path_fragments.gox:109
		__e = __c.Any(test.Button("path-follow-plain", func(ctx context.Context) bool {
		return f.pathFollow(ctx, "follow-plain")
	})); if __e != nil { return }
//line path_fragments.gox:112
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.pathMeta(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line path_fragments.gox:117
				__e = __c.Set("id", "path-meta"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("path-meta"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line path_fragments.gox:118
}

var _ gox.Comp = (*pathFragment)(nil)
