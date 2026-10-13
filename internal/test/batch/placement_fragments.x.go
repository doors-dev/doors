// Managed by GoX v0.3.3-0.20261012182611-94d35177549e

//line placement_fragments.gox:1
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
	"strings"
	"sync"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type placementLog struct {
	mu      sync.Mutex
	events  map[string][]string
	closed  map[string]bool
	pong    string
	reached int
}

func newPlacementLog() *placementLog {
	return &placementLog{
		events: map[string][]string{},
		closed: map[string]bool{},
	}
}

func (l *placementLog) track(label string, ch <-chan error) {
	l.mu.Lock()
	l.events[label] = []string{}
	l.mu.Unlock()
	go func() {
		for err := range ch {
			entry := "nil"
			if err != nil {
				entry = err.Error()
			}
			l.mu.Lock()
			l.events[label] = append(l.events[label], entry)
			l.mu.Unlock()
		}
		l.mu.Lock()
		l.closed[label] = true
		l.mu.Unlock()
	}()
}

func (l *placementLog) state(label string) (string, bool, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	events, ok := l.events[label]
	return strings.Join(events, ","), l.closed[label], ok
}

func (l *placementLog) values() (string, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pong, l.reached
}

type placementFragment struct {
	test.NoBeam
	g     *gate
	l     *placementLog
	box   doors.Door
	kid   doors.Door
	p     doors.Door
	q     doors.Door
	x     doors.Door
	y     doors.Door
	left  doors.Door
	right doors.Door
	mv    doors.Door
	set   doors.Setter
}

func (f *placementFragment) hold(ctx context.Context) bool {
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
	return false
}

//line placement_fragments.gox:99
func placementBatchButton(id string, handler func(context.Context) bool) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line placement_fragments.gox:100
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return handler(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line placement_fragments.gox:105
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line placement_fragments.gox:105
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line placement_fragments.gox:106
}

//line placement_fragments.gox:108
func (f *placementFragment) kidBox() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line placement_fragments.gox:109
			__e = __c.Set("id", "pa-wrap"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line placement_fragments.gox:110
			__e = (f.kid).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("span"); if __e != nil { return }
				{
//line placement_fragments.gox:110
					__e = __c.Set("id", "kid"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("kid0"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line placement_fragments.gox:112
}

//line placement_fragments.gox:114
func (f *placementFragment) chain() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line placement_fragments.gox:115
		__e = (f.p).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line placement_fragments.gox:115
				__e = __c.Set("id", "p"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line placement_fragments.gox:116
				__e = (f.q).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line placement_fragments.gox:116
						__e = __c.Set("id", "q"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line placement_fragments.gox:117
						__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
							ctx := __c.Context(); _ = ctx
							__e = __c.Init("span"); if __e != nil { return }
							{
//line placement_fragments.gox:117
								__e = __c.Set("id", "x"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
								__e = __c.Text("x0"); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
						return })); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line placement_fragments.gox:120
}

//line placement_fragments.gox:122
func (f *placementFragment) leaf() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line placement_fragments.gox:123
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line placement_fragments.gox:123
				__e = __c.Set("id", "y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("y0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line placement_fragments.gox:124
}

//line placement_fragments.gox:126
func (f *placementFragment) callerInner() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line placement_fragments.gox:128
		f.l.track("emit", doors.Call(ctx, doors.ActionEmit[any]{Name: "placement_ping", Arg: "in"}))
	f.l.track("emit-detached", doors.Call(
		doors.DetachedContext(ctx),
		doors.ActionEmit[string]{Name: "placement_ping", Arg: "out"}.Into(&f.l.pong),
	))
	f.l.track("set", doors.Call(
		doors.DetachedContext(ctx),
		f.set.Set("data-mark", "on").Into(&f.l.reached),
	))

		__e = __c.Init("span"); if __e != nil { return }
		{
//line placement_fragments.gox:138
			__e = __c.Set("id", "pd-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("t0"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line placement_fragments.gox:139
		__e = (f.set).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line placement_fragments.gox:139
				__e = __c.Set("id", "pd-set"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("s0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line placement_fragments.gox:140
}

//line placement_fragments.gox:142
func (f *placementFragment) caller() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line placement_fragments.gox:143
		__e = (f.kid).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line placement_fragments.gox:143
				__e = __c.Set("id", "kid"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line placement_fragments.gox:144
				__e = __c.Any(f.callerInner()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line placement_fragments.gox:146
}

//line placement_fragments.gox:148
func (f *placementFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line placement_fragments.gox:149
			__e = __c.Set("id", "pd-log"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line placement_fragments.gox:150
			__e = __c.Set("id", "pd-ready"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("loading"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("$on(\"placement_ping\", (arg) => {\n\t\t\tconst target = document.getElementById(\"pd-target\")\n\t\t\tconst log = document.getElementById(\"pd-log\")\n\t\t\tlog.textContent += (target ? \"hit\" : \"miss\") + \":\" + arg + \";\"\n\t\t\tif (target) {\n\t\t\t\ttarget.textContent = arg\n\t\t\t}\n\t\t\treturn \"pong:\" + arg\n\t\t})\n\t\tdocument.getElementById(\"pd-ready\").textContent = \"ready\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line placement_fragments.gox:163
		__e = (f.box).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line placement_fragments.gox:163
				__e = __c.Set("id", "box"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("box0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line placement_fragments.gox:164
		__e = (f.left).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line placement_fragments.gox:164
				__e = __c.Set("id", "left"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line placement_fragments.gox:165
				__e = (f.mv).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("span"); if __e != nil { return }
					{
//line placement_fragments.gox:165
						__e = __c.Set("id", "mv"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("mv0"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line placement_fragments.gox:167
		__e = (f.right).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line placement_fragments.gox:167
				__e = __c.Set("id", "right"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("right0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line placement_fragments.gox:168
		__e = __c.Any(placementBatchButton("pa-open", func(ctx context.Context) bool {
		f.l.track("box", f.box.Inner(ctx, f.kidBox()))
		return f.hold(ctx)
	})); if __e != nil { return }
//line placement_fragments.gox:172
		__e = __c.Any(test.Button("pa-poke", func(ctx context.Context) bool {
		f.l.track("poke", f.kid.Inner(ctx, "kid1"))
		return false
	})); if __e != nil { return }
//line placement_fragments.gox:176
		__e = __c.Any(placementBatchButton("pa-open-go", func(ctx context.Context) bool {
		placed := f.box.Inner(ctx, f.kidBox())
		detached := doors.DetachedContext(ctx)
		go func() {
			for range placed {
			}
			f.l.track("poke", f.kid.Inner(detached, "kid1"))
		}()
		return f.hold(ctx)
	})); if __e != nil { return }
//line placement_fragments.gox:186
		__e = __c.Any(placementBatchButton("pb-open", func(ctx context.Context) bool {
		f.l.track("box", f.box.Inner(ctx, f.chain()))
		return f.hold(ctx)
	})); if __e != nil { return }
//line placement_fragments.gox:190
		__e = __c.Any(test.Button("pb-poke", func(ctx context.Context) bool {
		f.l.track("poke", f.x.Inner(ctx, "x1"))
		return false
	})); if __e != nil { return }
//line placement_fragments.gox:194
		__e = __c.Any(test.Button("pb-swap", func(ctx context.Context) bool {
		f.l.track("swap", f.q.Inner(ctx, f.leaf()))
		return false
	})); if __e != nil { return }
//line placement_fragments.gox:198
		__e = __c.Any(test.Button("pb-leaf", func(ctx context.Context) bool {
		f.l.track("leaf", f.y.Inner(ctx, "y1"))
		return false
	})); if __e != nil { return }
//line placement_fragments.gox:202
		__e = __c.Any(placementBatchButton("pc-move", func(ctx context.Context) bool {
		f.l.track("right", f.right.Inner(ctx, &f.mv))
		return f.hold(ctx)
	})); if __e != nil { return }
//line placement_fragments.gox:206
		__e = __c.Any(placementBatchButton("pc-swap", func(ctx context.Context) bool {
		f.l.track("left", f.left.Inner(ctx, "left1"))
		f.l.track("right", f.right.Inner(ctx, &f.mv))
		return f.hold(ctx)
	})); if __e != nil { return }
//line placement_fragments.gox:211
		__e = __c.Any(test.Button("pc-poke", func(ctx context.Context) bool {
		f.l.track("poke", f.mv.Inner(ctx, "mv1"))
		return false
	})); if __e != nil { return }
//line placement_fragments.gox:215
		__e = __c.Any(placementBatchButton("pd-open", func(ctx context.Context) bool {
		f.l.track("box", f.box.Inner(ctx, f.caller()))
		return f.hold(ctx)
	})); if __e != nil { return }
	return })
//line placement_fragments.gox:219
}

var _ gox.Comp = (*placementFragment)(nil)
