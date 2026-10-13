// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line limits_fragments.gox:1
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
	"fmt"
	"sync/atomic"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

const limitsManyCount = 12

const limitsSingleCount = 50

type limitsState struct {
	entered chan struct{}
	proceed chan struct{}
	released chan struct{}
	heavyEntered chan struct{}
	heavyProceed chan struct{}
	rapid atomic.Int64
}

func limitsNewState() *limitsState {
	return &limitsState{
		entered: make(chan struct{}, 64),
		proceed: make(chan struct{}),
		released: make(chan struct{}, 64),
		heavyEntered: make(chan struct{}, 64),
		heavyProceed: make(chan struct{}),
	}
}

func (s *limitsState) pause(ctx context.Context) bool {
	s.entered <- struct{}{}
	select {
	case <-s.proceed:
		return true
	case <-ctx.Done():
		s.released <- struct{}{}
		return false
	}
}

type limitsHeavy struct {
	s *limitsState
}

//line limits_fragments.gox:64
func (h limitsHeavy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line limits_fragments.gox:66
		h.s.heavyEntered <- struct{}{}
	select {
	case <-h.s.heavyProceed:
	case <-ctx.Done():
	}

		__e = __c.Init("b"); if __e != nil { return }
		{
//line limits_fragments.gox:72
			__e = __c.Set("id", "heavy"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("heavy"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line limits_fragments.gox:73
}

type limitsFragment struct {
	test.NoBeam
	s *limitsState
	many []*doors.Door
	a doors.Door
	b doors.Door
	d doors.Door
}

func limitsNewFragment(s *limitsState) *limitsFragment {
	many := make([]*doors.Door, limitsManyCount)
	for i := range many {
		many[i] = &doors.Door{}
	}
	return &limitsFragment{
		s: s,
		many: many,
	}
}

func (f *limitsFragment) volume(ctx context.Context) bool {
	for _, door := range f.many[:5] {
		door.Inner(ctx, "v1")
	}
	if !f.s.pause(ctx) {
		return false
	}
	for _, door := range f.many[5:] {
		door.Inner(ctx, "v1")
	}
	if !f.s.pause(ctx) {
		return false
	}
	for _, door := range f.many {
		door.Inner(ctx, "v2")
	}
	return false
}

func (f *limitsFragment) split(ctx context.Context, first string, second string) bool {
	f.a.Inner(ctx, first)
	if !f.s.pause(ctx) {
		return false
	}
	f.b.Inner(ctx, second)
	return false
}

func (f *limitsFragment) single(ctx context.Context) bool {
	for i := 1; i < limitsSingleCount; i++ {
		f.a.Inner(ctx, fmt.Sprintf("s%d", i))
	}
	if !f.s.pause(ctx) {
		return false
	}
	f.a.Inner(ctx, fmt.Sprintf("s%d", limitsSingleCount))
	return false
}

func (f *limitsFragment) deferred(ctx context.Context) bool {
	f.a.Inner(ctx, "e1")
	f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("i"); if __e != nil { return }
		{
//line limits_fragments.gox:136
			__e = __c.Set("id", "skeleton"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("skeleton"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line limits_fragments.gox:136
	return }))
	f.d.DeferredInner(ctx, limitsHeavy{s: f.s})
	if !f.s.pause(ctx) {
		return false
	}
	f.b.Inner(ctx, "e1")
	return false
}

func (f *limitsFragment) rapid(ctx context.Context) bool {
	n := fmt.Sprint(f.s.rapid.Add(1))
	return f.split(ctx, n, n)
}

//line limits_fragments.gox:150
func limitsButton(id string, on func(context.Context) bool) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line limits_fragments.gox:151
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return on(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line limits_fragments.gox:156
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line limits_fragments.gox:156
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line limits_fragments.gox:157
}

//line limits_fragments.gox:159
func (f *limitsFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line limits_fragments.gox:160
		for i, door := range f.many {
//line limits_fragments.gox:161
			__e = (door).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line limits_fragments.gox:161
					__e = __c.Set("id", fmt.Sprintf("m%d", i)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("m0"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
//line limits_fragments.gox:163
		__e = (f.a).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line limits_fragments.gox:163
				__e = __c.Set("id", "a"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("a0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line limits_fragments.gox:164
		__e = (f.b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line limits_fragments.gox:164
				__e = __c.Set("id", "b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("b0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line limits_fragments.gox:165
		__e = (f.d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line limits_fragments.gox:165
				__e = __c.Set("id", "d"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("d0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line limits_fragments.gox:166
		__e = __c.Any(limitsButton("volume", f.volume)); if __e != nil { return }
//line limits_fragments.gox:167
		__e = __c.Any(limitsButton("roll", func(ctx context.Context) bool {
		return f.split(ctx, "r1", "r1")
	})); if __e != nil { return }
//line limits_fragments.gox:170
		__e = __c.Any(limitsButton("single", f.single)); if __e != nil { return }
//line limits_fragments.gox:171
		__e = __c.Any(limitsButton("leave", func(ctx context.Context) bool {
		return f.split(ctx, "l1", "l1")
	})); if __e != nil { return }
//line limits_fragments.gox:174
		__e = __c.Any(limitsButton("deferred", f.deferred)); if __e != nil { return }
//line limits_fragments.gox:175
		__e = __c.Any(limitsButton("rapid", f.rapid)); if __e != nil { return }
	return })
//line limits_fragments.gox:176
}

var _ gox.Comp = (*limitsFragment)(nil)
