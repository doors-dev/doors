// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line batch_fragments.gox:1
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

type gate struct {
	entered chan struct{}
	proceed chan struct{}
}

func newGate() *gate {
	return &gate{
		entered: make(chan struct{}, 1),
		proceed: make(chan struct{}),
	}
}

type batchFragment struct {
	test.NoBeam
	r *test.Reporter
	g *gate
	a doors.Door
	b doors.Door
	c doors.Door
}

func (f *batchFragment) split(ctx context.Context, first string, second string) bool {
	f.a.Inner(ctx, first)
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
	f.b.Inner(ctx, second)
	return false
}

//line batch_fragments.gox:56
func (f *batchFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line batch_fragments.gox:57
		__e = __c.Any(f.r); if __e != nil { return }
//line batch_fragments.gox:58
		__e = (f.a).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line batch_fragments.gox:58
				__e = __c.Set("id", "a"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("a0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line batch_fragments.gox:59
		__e = (f.b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line batch_fragments.gox:59
				__e = __c.Set("id", "b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("b0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line batch_fragments.gox:60
		__e = (f.c).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line batch_fragments.gox:60
				__e = __c.Set("id", "c"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("c0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line batch_fragments.gox:61
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.split(ctx, "a1", "b1")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line batch_fragments.gox:66
				__e = __c.Set("id", "batched"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("batched"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line batch_fragments.gox:67
		__e = __c.Any(test.Button("plain", func(ctx context.Context) bool {
		return f.split(ctx, "a2", "b2")
	})); if __e != nil { return }
//line batch_fragments.gox:70
		__e = __c.Any(test.Button("other", func(ctx context.Context) bool {
		f.c.Inner(ctx, "c1")
		return false
	})); if __e != nil { return }
//line batch_fragments.gox:74
		__e = __c.Any(test.Button("same", func(ctx context.Context) bool {
		f.a.Inner(ctx, "a-late")
		return false
	})); if __e != nil { return }
//line batch_fragments.gox:78
		__e = __c.Any(test.Button("onbatch", func(ctx context.Context) bool {
		doors.OnBatch(ctx, func(ctx context.Context) {
			f.r.Update(ctx, 0, "applied")
		}, func(ctx context.Context) {
			f.a.Inner(ctx, "a3")
			f.b.Inner(ctx, "b3")
		})
		return false
	})); if __e != nil { return }
	return })
//line batch_fragments.gox:87
}

var _ gox.Comp = (*batchFragment)(nil)
