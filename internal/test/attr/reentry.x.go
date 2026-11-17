// Managed by GoX v0.4.0

//line reentry.gox:1
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
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type reentryFragment struct {
	log *checkpointLog
	test.NoBeam
}

//line reentry.gox:31
func (f *reentryFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line reentry.gox:33
		rate := &doors.ScopeRate{Tick: time.Second}
	serial := &doors.ScopeSerial{}
	cpRate := &doors.ScopeRate{Tick: time.Second}
	cp := &doors.ScopeCheckpoint{}
	frRate := &doors.ScopeRate{Tick: time.Second}
	fr := &doors.ScopeFrame{}
	latest := &doors.ScopeLatest{}
	debounce := &doors.ScopeDebounce{Duration: 200 * time.Millisecond}

//line reentry.gox:42
		__e = __c.Any(f.button("rb", rate, 0)); if __e != nil { return }
//line reentry.gox:43
		__e = __c.Any(f.button("ra", serial.And(rate), 0)); if __e != nil { return }
//line reentry.gox:44
		__e = __c.Any(f.button("cb", cpRate, 0)); if __e != nil { return }
//line reentry.gox:45
		__e = __c.Any(f.button("ca", cp.Scope(true).And(cpRate), 0)); if __e != nil { return }
//line reentry.gox:46
		__e = __c.Any(f.button("fb", frRate, 0)); if __e != nil { return }
//line reentry.gox:47
		__e = __c.Any(f.button("fo", fr.Scope(false).And(frRate), 0)); if __e != nil { return }
//line reentry.gox:48
		__e = __c.Any(f.button("ff", fr.Scope(true).And(frRate), 0)); if __e != nil { return }
//line reentry.gox:49
		__e = __c.Any(f.button("la", serial.And(latest).And(debounce), 1500*time.Millisecond)); if __e != nil { return }
//line reentry.gox:50
		__e = __c.Any(f.button("lb", latest, 1500*time.Millisecond)); if __e != nil { return }
	return })
//line reentry.gox:51
}

//line reentry.gox:53
func (f *reentryFragment) button(id string, scope doors.Scopes, d time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line reentry.gox:54
			__e = __c.Set("id", id); if __e != nil { return }
//line reentry.gox:54
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope:     scope,
		Indicator: doors.IndicateClass("busy"),
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			select {
			case <-time.After(d):
				f.log.add("end:" + id)
			case <-ctx.Done():
				f.log.add("abort:" + id)
			}
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line reentry.gox:67
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line reentry.gox:68
}

type doubleEmptyFragment struct {
	log *checkpointLog
	test.NoBeam
}

//line reentry.gox:75
func (f *doubleEmptyFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line reentry.gox:77
		x := &doors.ScopeBlocking{}
	k := &doors.ScopeCheckpoint{}
	u := &doors.ScopeSerial{}

//line reentry.gox:81
		__e = __c.Any(f.button("z", x, 3000*time.Millisecond)); if __e != nil { return }
//line reentry.gox:82
		__e = __c.Any(f.button("a", k.Scope(false), 600*time.Millisecond)); if __e != nil { return }
//line reentry.gox:83
		__e = __c.Any(f.button("c", u.And(k.Scope(true)).And(x), 0)); if __e != nil { return }
//line reentry.gox:84
		__e = __c.Any(f.button("j", u.And(k.Scope(false)), 1500*time.Millisecond)); if __e != nil { return }
//line reentry.gox:85
		__e = __c.Any(f.button("c2", k.Scope(true), 0)); if __e != nil { return }
	return })
//line reentry.gox:86
}

//line reentry.gox:88
func (f *doubleEmptyFragment) button(id string, scope doors.Scopes, d time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line reentry.gox:89
			__e = __c.Set("id", id); if __e != nil { return }
//line reentry.gox:89
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			<-time.After(d)
			f.log.add("end:" + id)
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line reentry.gox:97
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line reentry.gox:98
}

type badIndicatorFragment struct {
	log *checkpointLog
	test.NoBeam
}

//line reentry.gox:105
func (f *badIndicatorFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line reentry.gox:107
		serial := &doors.ScopeSerial{}
	blocking := &doors.ScopeBlocking{}

//line reentry.gox:110
		__e = __c.Any(f.button("sx", serial, doors.IndicateClass("busy"), 400*time.Millisecond)); if __e != nil { return }
//line reentry.gox:111
		__e = __c.Any(f.button("sy", serial, doors.IndicateClassQuery("#box[", "busy"), 0)); if __e != nil { return }
//line reentry.gox:112
		__e = __c.Any(f.button("sz", serial, nil, 0)); if __e != nil { return }
//line reentry.gox:113
		__e = __c.Any(f.button("by", blocking, doors.IndicateClassQuery("#box[", "busy"), 0)); if __e != nil { return }
//line reentry.gox:114
		__e = __c.Any(f.button("bz", blocking, nil, 0)); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line reentry.gox:115
			__e = __c.Set("id", "e"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("e"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line reentry.gox:116
			__e = __c.Set("id", "f"); if __e != nil { return }
//line reentry.gox:116
			__e = __c.Set("data-keep", "orig"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("f"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line reentry.gox:117
			__e = __c.Set("id", "g"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("g"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line reentry.gox:119
		queue := &doors.ScopeSerial{}

//line reentry.gox:121
		__e = __c.Any(f.button("ea", queue, doors.IndicateClassQuery("#e", "busy"), 400*time.Millisecond)); if __e != nil { return }
//line reentry.gox:122
		__e = __c.Any(f.button("eb", nil, doors.IndicateAttrQuery("#e", "aria-busy ", "true"), 800*time.Millisecond)); if __e != nil { return }
//line reentry.gox:123
		__e = __c.Any(f.button("ec", queue, nil, 0)); if __e != nil { return }
//line reentry.gox:124
		__e = __c.Any(f.button("fd", nil, doors.IndicateAttrQuery("#f", "data-x", "1").
		And(doors.IndicateClassQuery("#f", "busy\tdone")).
		And(doors.IndicateAttrQuery("#f", "bad name", "1")).
		And(doors.IndicateAttrQuery("#f", "data-keep", "new")), 400*time.Millisecond)); if __e != nil { return }
//line reentry.gox:128
		__e = __c.Any(f.button("gd", nil, doors.IndicateClassQuery("#g", "busy\tdone\nnow"), 400*time.Millisecond)); if __e != nil { return }
	return })
//line reentry.gox:129
}

//line reentry.gox:131
func (f *badIndicatorFragment) button(id string, scope doors.Scopes, ind doors.Indicators, d time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line reentry.gox:132
			__e = __c.Set("id", id); if __e != nil { return }
//line reentry.gox:132
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope:     scope,
		Indicator: ind,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			<-time.After(d)
			f.log.add("end:" + id)
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line reentry.gox:141
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line reentry.gox:142
}

type debounceReentryFragment struct {
	log *checkpointLog
	test.NoBeam
}

//line reentry.gox:149
func (f *debounceReentryFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line reentry.gox:151
		de := &doors.ScopeDebounce{Duration: 200 * time.Millisecond}
	se := &doors.ScopeSerial{}
	la := &doors.ScopeLatest{}
	del := &doors.ScopeDebounce{Duration: 300 * time.Millisecond, Limit: 600 * time.Millisecond}
	sel := &doors.ScopeSerial{}
	lal := &doors.ScopeLatest{}

//line reentry.gox:158
		__e = __c.Any(f.button("x", se.And(de), 0)); if __e != nil { return }
//line reentry.gox:159
		__e = __c.Any(f.button("y", se.And(la), 0)); if __e != nil { return }
//line reentry.gox:160
		__e = __c.Any(f.button("z", la.And(de), 0)); if __e != nil { return }
//line reentry.gox:161
		__e = __c.Any(f.button("p", del, 4000*time.Millisecond)); if __e != nil { return }
//line reentry.gox:162
		__e = __c.Any(f.button("lx", sel.And(del), 0)); if __e != nil { return }
//line reentry.gox:163
		__e = __c.Any(f.button("ly", sel.And(lal), 0)); if __e != nil { return }
//line reentry.gox:164
		__e = __c.Any(f.button("lz", lal.And(del), 0)); if __e != nil { return }
//line reentry.gox:165
		__e = __c.Any(f.button("w", del, 0)); if __e != nil { return }
	return })
//line reentry.gox:166
}

//line reentry.gox:168
func (f *debounceReentryFragment) button(id string, scope doors.Scopes, d time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line reentry.gox:169
			__e = __c.Set("id", id); if __e != nil { return }
//line reentry.gox:169
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			<-time.After(d)
			f.log.add("end:" + id)
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line reentry.gox:177
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line reentry.gox:178
}

type doubleSettleFragment struct {
	log *checkpointLog
	test.NoBeam
}

//line reentry.gox:185
func (f *doubleSettleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line reentry.gox:187
		bl := &doors.ScopeBlocking{}
	se := &doors.ScopeSerial{}
	la := &doors.ScopeLatest{}

//line reentry.gox:191
		__e = __c.Any(f.button("p", se, 400*time.Millisecond)); if __e != nil { return }
//line reentry.gox:192
		__e = __c.Any(f.button("q", la.And(se).And(bl), 0)); if __e != nil { return }
//line reentry.gox:193
		__e = __c.Any(f.button("r", bl.And(se).And(la), 2000*time.Millisecond)); if __e != nil { return }
//line reentry.gox:194
		__e = __c.Any(f.button("s", se, 0)); if __e != nil { return }
//line reentry.gox:195
		__e = __c.Any(f.button("t", bl, 0)); if __e != nil { return }
	return })
//line reentry.gox:196
}

//line reentry.gox:198
func (f *doubleSettleFragment) button(id string, scope doors.Scopes, d time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line reentry.gox:199
			__e = __c.Set("id", id); if __e != nil { return }
//line reentry.gox:199
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			<-time.After(d)
			f.log.add("end:" + id)
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line reentry.gox:207
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line reentry.gox:208
}
