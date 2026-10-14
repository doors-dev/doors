// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line group_fragments.gox:1
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

type groupFragment struct {
	test.NoBeam
	g *gate
	fired chan struct{}
	x doors.Door
	y doors.Door
}

func newGroupFragment() *groupFragment {
	return &groupFragment{
		g: newGate(),
		fired: make(chan struct{}, 8),
	}
}

func (f *groupFragment) groupHold(ctx context.Context) {
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
}

func (f *groupFragment) groupOn(ctx context.Context) {
	f.fired <- struct{}{}
}

//line group_fragments.gox:51
func (f *groupFragment) groupSlow(text string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line group_fragments.gox:53
		f.groupHold(ctx)

//line group_fragments.gox:55
		__e = __c.Any(text); if __e != nil { return }
	return })
//line group_fragments.gox:56
}

//line group_fragments.gox:58
func (f *groupFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line group_fragments.gox:59
			__e = __c.Set("id", "ind"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ind"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line group_fragments.gox:60
		__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line group_fragments.gox:60
				__e = __c.Set("id", "x"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("x0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line group_fragments.gox:61
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line group_fragments.gox:61
				__e = __c.Set("id", "y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("y0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line group_fragments.gox:62
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			<-f.x.Inner(ctx, "x-held")
			f.groupHold(ctx)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line group_fragments.gox:69
				__e = __c.Set("id", "group-holder"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("group-holder"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line group_fragments.gox:70
		__e = (doors.AClick{
		Indicator: doors.IndicateAttrQuery("#ind", "data-busy", "true"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			doors.OnBatch(ctx, f.groupOn, func(ctx context.Context) {
				f.x.Inner(ctx, "x-plain")
				f.y.Inner(ctx, "y-plain")
			})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line group_fragments.gox:79
				__e = __c.Set("id", "group-plain"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("group-plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line group_fragments.gox:80
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			doors.OnSettle(ctx, func(ctx context.Context) {
				doors.OnBatch(ctx, f.groupOn, func(ctx context.Context) {
					f.x.Inner(ctx, "x-settle")
					f.y.Inner(ctx, f.groupSlow("y-settle"))
				})
			})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line group_fragments.gox:91
				__e = __c.Set("id", "group-settle"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("group-settle"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line group_fragments.gox:92
}

var _ gox.Comp = (*groupFragment)(nil)
