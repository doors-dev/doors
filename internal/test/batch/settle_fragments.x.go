// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line settle_fragments.gox:1
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
	"strings"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type settleRead struct {
	value string
	values []error
}

type settleOp struct {
	values []error
}

type settleFragment struct {
	test.NoBeam
	g *gate
	fired chan struct{}
	returned chan struct{}
	ops chan settleOp
	seen chan settleRead
	x doors.Door
	y doors.Door
	o doors.Door
	d doors.Door
	res doors.Door
	s doors.Setter
}

func newSettleFragment() *settleFragment {
	return &settleFragment{
		g: newGate(),
		fired: make(chan struct{}, 8),
		returned: make(chan struct{}, 8),
		ops: make(chan settleOp, 1),
		seen: make(chan settleRead, 8),
	}
}

func (f *settleFragment) settleHold(ctx context.Context) {
	f.g.entered <- struct{}{}
	select {
	case <-f.g.proceed:
	case <-ctx.Done():
	}
}

func (f *settleFragment) settleEmit(ctx context.Context, name string, arg any) (*string, <-chan error) {
	v := new(string)
	return v, doors.Call(ctx, doors.ActionEmit[string]{Name: name, Arg: arg}.Into(v))
}

func (f *settleFragment) settleRead(ctx context.Context, name string, arg any) settleRead {
	v, ch := f.settleEmit(ctx, name, arg)
	values := drain(ch)
	return settleRead{value: *v, values: values}
}

func (f *settleFragment) settleSummary(op settleOp) string {
	parts := make([]string, len(op.values))
	for i, err := range op.values {
		parts[i] = fmt.Sprint(err)
	}
	return fmt.Sprintf("%d:%s", len(op.values), strings.Join(parts, ","))
}

func (f *settleFragment) settleMark(ctx context.Context) {
	f.fired <- struct{}{}
}

//line settle_fragments.gox:91
func (f *settleFragment) settleDelivered() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line settle_fragments.gox:92
		__e = (&f.s).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line settle_fragments.gox:92
				__e = __c.Set("id", "settle-item"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("delivered"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line settle_fragments.gox:93
}

//line settle_fragments.gox:95
func (f *settleFragment) settleBatched(id string, on func(ctx context.Context)) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line settle_fragments.gox:96
		__e = (doors.AClick{
		Batch: true,
		Indicator: doors.IndicateAttrQuery("#ind-a", "data-busy", "true"),
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			on(ctx)
			f.returned <- struct{}{}
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line settle_fragments.gox:104
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line settle_fragments.gox:104
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line settle_fragments.gox:105
}

//line settle_fragments.gox:107
func (f *settleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line settle_fragments.gox:108
			__e = __c.Set("id", "ind-a"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ind-a"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line settle_fragments.gox:109
		__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line settle_fragments.gox:109
				__e = __c.Set("id", "x"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("x0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line settle_fragments.gox:110
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line settle_fragments.gox:110
				__e = __c.Set("id", "y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("y0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line settle_fragments.gox:111
		__e = (f.o).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line settle_fragments.gox:111
				__e = __c.Set("id", "o"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("o0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line settle_fragments.gox:112
		__e = (f.d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line settle_fragments.gox:112
				__e = __c.Set("id", "d"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("d0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line settle_fragments.gox:113
		__e = (f.res).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line settle_fragments.gox:113
				__e = __c.Set("id", "settle-res"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("none"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line settle_fragments.gox:114
			__e = __c.Set("id", "settle-seen"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("none"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const text = (selector) =>\n\t\t\tdocument.querySelector(selector)?.textContent ?? \"missing\"\n\t\t$on(\"settle_text\", (selector) => text(selector))\n\t\t$on(\"settle_read\", () => {\n\t\t\tconst value = text(\"#x\") + \"|\" + text(\"#y\")\n\t\t\tdocument.getElementById(\"settle-seen\").textContent = value\n\t\t\treturn value\n\t\t})\n\t\t$on(\"settle_attr\", (selector) =>\n\t\t\tdocument.querySelector(selector)?.getAttribute(\"data-settle\") ?? \"missing\")\n\t\tdocument.documentElement.setAttribute(\"data-settle-ready\", \"\")"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line settle_fragments.gox:128
		__e = __c.Any(f.settleBatched("settle-op", func(ctx context.Context) {
		<-f.x.Inner(ctx, "x1")
		doors.OnSettle(ctx, func(sctx context.Context) {
			f.settleMark(sctx)
			ch := f.o.Inner(sctx, "o1")
			go func() {
				var op settleOp
				op.values = drain(ch)
				f.res.Inner(sctx, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line settle_fragments.gox:136
						__e = __c.Set("id", "settle-res"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line settle_fragments.gox:136
						__e = __c.Any(f.settleSummary(op)); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
//line settle_fragments.gox:136
				return }))
				f.ops <- op
			}()
		})
		f.settleHold(ctx)
		f.y.Inner(ctx, "y1")
	})); if __e != nil { return }
//line settle_fragments.gox:143
		__e = __c.Any(f.settleBatched("settle-call", func(ctx context.Context) {
		<-f.x.Inner(ctx, "x1")
		doors.OnSettle(ctx, func(sctx context.Context) {
			f.settleMark(sctx)
			go func() {
				f.seen <- f.settleRead(sctx, "settle_read", nil)
			}()
		})
		f.settleHold(ctx)
		f.y.Inner(ctx, "y1")
	})); if __e != nil { return }
//line settle_fragments.gox:154
		__e = __c.Any(f.settleBatched("settle-sync", func(ctx context.Context) {
		<-f.d.Inner(ctx, f.settleDelivered())
		doors.OnSettle(ctx, func(sctx context.Context) {
			f.settleMark(sctx)
			set := doors.Call(sctx, f.s.Set("data-settle", "set"))
			v, ch := f.settleEmit(sctx, "settle_attr", "#settle-item")
			go func() {
				values := append(drain(set), drain(ch)...)
				f.seen <- settleRead{value: *v, values: values}
			}()
		})
		f.settleHold(ctx)
	})); if __e != nil { return }
//line settle_fragments.gox:167
		__e = __c.Any(f.settleBatched("settle-hold", func(ctx context.Context) {
		doors.OnSettle(ctx, f.settleMark)
		release := doors.HoldSettle(ctx)
		go func() {
			defer release()
			<-f.x.Inner(ctx, "x-hold")
			f.settleHold(ctx)
			f.y.Inner(ctx, "y-hold")
		}()
	})); if __e != nil { return }
//line settle_fragments.gox:177
		__e = __c.Any(f.settleBatched("settle-idle", func(ctx context.Context) {
		doors.OnSettle(ctx, f.settleMark)
		<-f.x.Inner(ctx, "x-idle")
		release := doors.HoldSettle(ctx)
		go func() {
			defer release()
			f.settleHold(ctx)
		}()
	})); if __e != nil { return }
//line settle_fragments.gox:186
		__e = __c.Any(test.Button("settle-after", func(ctx context.Context) bool {
		f.x.Inner(ctx, "x-after")
		f.y.Inner(ctx, "y-after")
		return false
	})); if __e != nil { return }
	return })
//line settle_fragments.gox:191
}

var _ gox.Comp = (*settleFragment)(nil)
