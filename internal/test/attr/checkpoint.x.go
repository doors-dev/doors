// Managed by GoX v0.4.0

//line checkpoint.gox:1
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
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type checkpointLog struct {
	mu      sync.Mutex
	entries []string
	times   []time.Time
}

func (l *checkpointLog) add(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
	l.times = append(l.times, time.Now())
}

func (l *checkpointLog) at(entry string) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	i := slices.Index(l.entries, entry)
	if i < 0 {
		return time.Time{}
	}
	return l.times[i]
}

func (l *checkpointLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.entries)
}

type checkpointFragment struct {
	log   *checkpointLog
	value atomic.Value
	test.NoBeam
}

//line checkpoint.gox:66
func (f *checkpointFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line checkpoint.gox:68
		cp := &doors.ScopeCheckpoint{}
	blocking := &doors.ScopeBlocking{}
	debounce := &doors.ScopeDebounce{Duration: 300 * time.Millisecond}
	latest := &doors.ScopeLatest{}
	fieldDebounce := &doors.ScopeDebounce{Duration: 500 * time.Millisecond}
	slowLatest := &doors.ScopeLatest{}
	runLatest := &doors.ScopeLatest{}
	cpBlock := &doors.ScopeBlocking{}
	lrLatest := &doors.ScopeLatest{}
	lrRate := &doors.ScopeRate{Tick: 300 * time.Millisecond}

//line checkpoint.gox:79
		__e = __c.Any(f.button("o1", cp.Scope(false), 600*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:80
		__e = __c.Any(f.button("o2", cp.Scope(false), 900*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:81
		__e = __c.Any(f.button("of", cp.Scope(false), 0)); if __e != nil { return }
//line checkpoint.gox:82
		__e = __c.Any(f.button("c1", cp.Scope(true), 400*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:83
		__e = __c.Any(f.button("c2", cp.Scope(true), 0)); if __e != nil { return }
//line checkpoint.gox:84
		__e = __c.Any(f.button("c3", cp.Scope(true), 0)); if __e != nil { return }
//line checkpoint.gox:85
		__e = __c.Any(f.button("cb", blocking.And(cp.Scope(true)), 400*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:86
		__e = __c.Any(f.button("od", cp.Scope(false).And(debounce), 300*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:87
		__e = __c.Any(f.button("cl", latest.And(cp.Scope(true)), 0)); if __e != nil { return }
//line checkpoint.gox:88
		__e = __c.Any(f.button("ck", cp.Scope(true).And(cpBlock), 400*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:89
		__e = __c.Any(f.button("bo", cpBlock, 1000*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:90
		__e = __c.Any(f.button("lr", cp.Scope(false).And(lrLatest).And(lrRate), 400*time.Millisecond)); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:91
			__e = __c.Set("id", "os"); if __e != nil { return }
//line checkpoint.gox:91
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope: cp.Scope(false),
		Race:  doors.RaceParallel,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:os")
			<-time.After(500 * time.Millisecond)
			f.log.add("end:os")
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("os"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line checkpoint.gox:101
			__e = __c.Set("id", "field"); if __e != nil { return }
//line checkpoint.gox:101
			__e = __c.Modify(doors.A(ctx, doors.AInput{
		Scope: cp.Scope(false).And(fieldDebounce),
		On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
			<-time.After(200 * time.Millisecond)
			f.value.Store(r.Event().Value)
			f.log.add("field:" + r.Event().Value)
			return false
		},
	})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:110
			__e = __c.Set("id", "sub"); if __e != nil { return }
//line checkpoint.gox:110
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope: cp.Scope(true),
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			value, _ := f.value.Load().(string)
			f.log.add("submit:" + value)
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("sub"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:118
			__e = __c.Set("id", "ls"); if __e != nil { return }
//line checkpoint.gox:118
			__e = __c.Modify(doors.A(ctx, doors.AClick{
		Scope: cp.Scope(false).And(slowLatest),
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:ls")
			r.After(doors.ActionEmit[any]{Name: "slow"})
			return false
		},
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ls"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line checkpoint.gox:126
		__e = __c.Any(f.button("lc", runLatest.And(cp.Scope(true)), 500*time.Millisecond)); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("$on(\"slow\", () => new Promise((resolve) => setTimeout(resolve, 600)))"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:130
			__e = __c.Set("id", "he"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("he"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:131
			__e = __c.Set("id", "ht"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ht"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line checkpoint.gox:132
		__e = (doors.ARawHook{
		Name:  "he",
		Scope: cp.Scope(false),
		On: func(ctx context.Context, r doors.RequestRawHook) bool {
			f.log.add("start:he")
			<-time.After(400 * time.Millisecond)
			f.log.add("fail:he")
			r.ResponseWriter().WriteHeader(http.StatusConflict)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line checkpoint.gox:142
			__e = (doors.ARawHook{
		Name:           "ht",
		Scope:          cp.Scope(false),
		RequestTimeout: 300 * time.Millisecond,
		On: func(ctx context.Context, r doors.RequestRawHook) bool {
			f.log.add("start:ht")
			<-time.After(800 * time.Millisecond)
			f.log.add("end:ht")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("script"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Raw("for (const name of [\"he\", \"ht\"]) {\n\t\t\tdocument.getElementById(name).addEventListener(\"click\", () => {\n\t\t\t\t$hook(name).catch(() => {})\n\t\t\t})\n\t\t}"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line checkpoint.gox:159
}

//line checkpoint.gox:161
func (f *checkpointFragment) button(id string, scope doors.Scopes, delay time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:162
			__e = __c.Set("id", id); if __e != nil { return }
//line checkpoint.gox:162
			__e = __c.Modify(doors.A(ctx, f.handler(id, scope, delay))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line checkpoint.gox:162
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line checkpoint.gox:163
}

func (f *checkpointFragment) handler(id string, scope doors.Scopes, delay time.Duration) doors.Attr {
	return doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			<-time.After(delay)
			f.log.add("end:" + id)
			return false
		},
	}
}
