// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line foreign_fragments.gox:1
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
	"sync"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type foreignGate struct {
	*gate
	once sync.Once
}

func (g *foreignGate) foreignRelease() {
	g.once.Do(func() {
		close(g.proceed)
	})
}

type foreignShared struct {
	src doors.Source[string]
	ga *foreignGate
	gb *foreignGate
	fired chan string
}

func foreignNewShared() *foreignShared {
	return &foreignShared{
		src: doors.NewSource("s0"),
		ga: &foreignGate{gate: newGate()},
		gb: &foreignGate{gate: newGate()},
		fired: make(chan string, 8),
	}
}

type foreignFragment struct {
	test.Beam
	s *foreignShared
	shared doors.Door
	own doors.Door
	x doors.Door
	y doors.Door
}

func (f *foreignFragment) foreignHold(ctx context.Context, g *gate) {
	g.entered <- struct{}{}
	select {
	case <-g.proceed:
	case <-ctx.Done():
	}
}

func (f *foreignFragment) foreignRun(ctx context.Context, first string, second string) {
	f.s.src.Update(ctx, first)
	f.own.Inner(ctx, "own-"+first)
	f.foreignHold(ctx, f.s.ga.gate)
	if second != "" {
		f.s.src.Update(ctx, second)
	}
}

func (f *foreignFragment) foreignPeer(ctx context.Context, v string) {
	doors.OnBatch(ctx, func(ctx context.Context) {
		f.s.fired <- v
	}, func(ctx context.Context) {
		f.x.Inner(ctx, "x-"+v)
	}, func(ctx context.Context) {
		switch v {
		case "slow":
			release := doors.HoldSettle(ctx)
			go func() {
				defer release()
				f.foreignHold(ctx, f.s.gb.gate)
			}()
		case "block":
			f.foreignHold(ctx, f.s.gb.gate)
		}
		f.y.Inner(ctx, "y-"+v)
	})
}

//line foreign_fragments.gox:98
func (f *foreignFragment) foreignBatched(id string, first string, second string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line foreign_fragments.gox:99
		__e = (doors.AClick{
		Batch: true,
		Indicator: doors.IndicateAttrQuery("#foreign-ind", "data-busy", "true"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.foreignRun(ctx, first, second)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line foreign_fragments.gox:106
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line foreign_fragments.gox:106
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line foreign_fragments.gox:107
}

//line foreign_fragments.gox:109
func (f *foreignFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line foreign_fragments.gox:111
		path, _ := f.B.Read(ctx)
	peer := path.Vs
	v, _ := f.s.src.ReadAndSub(ctx, func(ctx context.Context, v string) bool {
		f.shared.Inner(ctx, v)
		if peer {
			f.foreignPeer(ctx, v)
		}
		return false
	})

		__e = __c.Init("div"); if __e != nil { return }
		{
//line foreign_fragments.gox:121
			__e = __c.Set("id", "foreign-ind"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ind"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line foreign_fragments.gox:122
		__e = (f.shared).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line foreign_fragments.gox:122
				__e = __c.Set("id", "foreign-shared"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line foreign_fragments.gox:122
				__e = __c.Any(v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line foreign_fragments.gox:123
		__e = (f.own).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line foreign_fragments.gox:123
				__e = __c.Set("id", "foreign-own"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("own0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line foreign_fragments.gox:124
		__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line foreign_fragments.gox:124
				__e = __c.Set("id", "foreign-x"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("x0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line foreign_fragments.gox:125
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line foreign_fragments.gox:125
				__e = __c.Set("id", "foreign-y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("y0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line foreign_fragments.gox:126
		__e = __c.Any(f.foreignBatched("foreign-batch", "s1", "")); if __e != nil { return }
//line foreign_fragments.gox:127
		__e = __c.Any(f.foreignBatched("foreign-slow", "slow", "")); if __e != nil { return }
//line foreign_fragments.gox:128
		__e = __c.Any(f.foreignBatched("foreign-block", "block", "")); if __e != nil { return }
//line foreign_fragments.gox:129
		__e = __c.Any(f.foreignBatched("foreign-twice", "s1", "s2")); if __e != nil { return }
	return })
//line foreign_fragments.gox:130
}

var _ gox.Comp = (*foreignFragment)(nil)
