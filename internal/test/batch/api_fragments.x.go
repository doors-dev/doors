// Managed by GoX v0.3.3-0.20261012182611-94d35177549e

//line api_fragments.gox:1
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

type apiRead struct {
	value string
	err error
	ok bool
}

type apiGuard struct {
	call apiRead
	plainOk bool
	inner apiRead
}

type apiFragment struct {
	test.NoBeam
	g *gate
	start chan struct{}
	fired chan struct{}
	seen chan apiRead
	guard chan apiGuard
	a doors.Door
	b doors.Door
	d doors.Door
	s doors.Setter
}

func newAPIFragment() *apiFragment {
	return &apiFragment{
		g: newGate(),
		start: make(chan struct{}),
		fired: make(chan struct{}, 8),
		seen: make(chan apiRead, 8),
		guard: make(chan apiGuard, 1),
	}
}

func (f *apiFragment) apiHold(ctx context.Context) {
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
}

func (f *apiFragment) apiRead(ctx context.Context) apiRead {
	var v string
	err, ok := <-doors.Call(ctx, doors.ActionEmit[string]{Name: "api_read"}.Into(&v))
	return apiRead{value: v, err: err, ok: ok}
}

func (f *apiFragment) apiOn(ctx context.Context) {
	f.fired <- struct{}{}
	go func() {
		f.seen <- f.apiRead(ctx)
	}()
}

//line api_fragments.gox:80
func (f *apiFragment) apiDelivered() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line api_fragments.gox:81
		__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line api_fragments.gox:81
				__e = __c.Set("id", "api-item"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("delivered"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line api_fragments.gox:82
}

//line api_fragments.gox:84
func (f *apiFragment) apiBatched(id string, on func(ctx context.Context)) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line api_fragments.gox:85
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			on(ctx)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line api_fragments.gox:91
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line api_fragments.gox:91
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line api_fragments.gox:92
}

//line api_fragments.gox:94
func (f *apiFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line api_fragments.gox:95
		__e = (f.a).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line api_fragments.gox:95
				__e = __c.Set("id", "a"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("a0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line api_fragments.gox:96
		__e = (f.b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line api_fragments.gox:96
				__e = __c.Set("id", "b"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("b0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line api_fragments.gox:97
		__e = (f.d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line api_fragments.gox:97
				__e = __c.Set("id", "d"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("d0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line api_fragments.gox:98
			__e = __c.Set("id", "api-seen"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("none"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const text = (selector) =>\n\t\t\tdocument.querySelector(selector)?.textContent ?? \"missing\"\n\t\t$on(\"api_read\", () => text(\"#a\") + \"|\" + text(\"#b\"))\n\t\t$on(\"api_mark\", (selector) => {\n\t\t\tdocument.getElementById(\"api-seen\").textContent = text(selector)\n\t\t})\n\t\tdocument.documentElement.setAttribute(\"data-api-ready\", \"\")"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line api_fragments.gox:108
		__e = __c.Any(doors.Go(func(ctx context.Context) {
		select {
		case <-f.start:
		case <-ctx.Done():
			return
		}
		doors.OnBatch(doors.DetachedContext(ctx), f.apiOn, func(ctx context.Context) {
			f.a.Inner(ctx, "a1")
		}, func(ctx context.Context) {
			f.apiHold(ctx)
			f.b.Inner(ctx, "b1")
		})
	})); if __e != nil { return }
//line api_fragments.gox:121
		__e = __c.Any(test.Button("api-plain", func(ctx context.Context) bool {
		doors.OnBatch(ctx, f.apiOn, func(ctx context.Context) {
			f.a.Inner(ctx, "a1")
		}, func(ctx context.Context) {
			f.apiHold(ctx)
			f.b.Inner(ctx, "b1")
		})
		return false
	})); if __e != nil { return }
//line api_fragments.gox:130
		__e = __c.Any(f.apiBatched("api-join", func(ctx context.Context) {
		doors.OnBatch(ctx, f.apiOn, func(ctx context.Context) {
			f.a.Inner(ctx, "a1")
		})
		f.apiHold(ctx)
		f.b.Inner(ctx, "b1")
	})); if __e != nil { return }
//line api_fragments.gox:137
		__e = __c.Any(f.apiBatched("api-settle", func(ctx context.Context) {
		f.d.Inner(ctx, f.apiDelivered())
		doors.OnSettle(ctx, func(ctx context.Context) {
			doors.Call(ctx, f.s.Set("data-api", "set"))
			doors.Call(ctx, doors.ActionEmit[any]{Name: "api_mark", Arg: "#api-item"})
		})
		f.apiHold(ctx)
	})); if __e != nil { return }
//line api_fragments.gox:145
		__e = __c.Any(f.apiBatched("api-guard", func(ctx context.Context) {
		res := apiGuard{call: f.apiRead(ctx)}
		_, res.plainOk = <-doors.Call(ctx, doors.ActionScroll{Selector: "#a"})
		res.inner.err, res.inner.ok = <-f.a.Inner(ctx, "a1")
		f.guard <- res
		f.b.Inner(ctx, "b1")
	})); if __e != nil { return }
//line api_fragments.gox:152
		__e = __c.Any(f.apiBatched("api-detached", func(ctx context.Context) {
		f.a.Inner(ctx, "a1")
		f.seen <- f.apiRead(doors.DetachedContext(ctx))
		f.apiHold(ctx)
		f.b.Inner(ctx, "b1")
	})); if __e != nil { return }
//line api_fragments.gox:158
		__e = __c.Any(f.apiBatched("api-held", func(ctx context.Context) {
		doors.Call(ctx, doors.ActionScroll{Selector: "#a"})
		var v string
		ch := doors.Call(doors.DetachedContext(ctx), doors.ActionEmit[string]{Name: "api_read"}.Into(&v))
		go func() {
			err, ok := <-ch
			f.seen <- apiRead{value: v, err: err, ok: ok}
		}()
		f.apiHold(ctx)
		f.a.Inner(ctx, "a1")
		f.b.Inner(ctx, "b1")
	})); if __e != nil { return }
	return })
//line api_fragments.gox:170
}

var _ gox.Comp = (*apiFragment)(nil)
