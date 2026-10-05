// Managed by GoX v0.3.2

//line indicator.gox:1
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

type indicatorFragment struct {
	test.NoBeam
}

//line indicator.gox:30
func (f *indicatorFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line indicator.gox:31
		__e = __c.Any(f.selectors()); if __e != nil { return }
//line indicator.gox:32
		__e = __c.Any(f.restore()); if __e != nil { return }
//line indicator.gox:33
		__e = __c.Any(f.queue()); if __e != nil { return }
//line indicator.gox:34
		__e = __c.Any(f.values()); if __e != nil { return }
	return })
//line indicator.gox:35
}

//line indicator.gox:37
func (f *indicatorFragment) values() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:38
			__e = __c.Set("id", "v-target"); if __e != nil { return }
//line indicator.gox:38
			__e = __c.Set("data-remove", "keep"); if __e != nil { return }
//line indicator.gox:38
			__e = __c.Set("data-off", "keep"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("values"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line indicator.gox:39
		__e = __c.Any(f.button("values-1", doors.Join[doors.Indicators](
		doors.IndicateAttrQuery("#v-target", "data-remove", nil),
		doors.IndicateAttrQuery("#v-target", "data-off", false),
		doors.IndicateAttrQuery("#v-target", "data-bare", true),
		doors.IndicateAttrQuery("#v-target", "data-num", 42),
	))); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:45
			__e = __c.Set("id", "v2-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("values"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line indicator.gox:46
		__e = __c.Any(f.button("values-set", doors.IndicateAttrQuery("#v2-target", "data-x", "X"))); if __e != nil { return }
//line indicator.gox:47
		__e = __c.Any(f.button("values-change", doors.IndicateAttrQuery("#v2-target", "data-x", "Y"))); if __e != nil { return }
//line indicator.gox:48
		__e = __c.Any(f.button("values-remove", doors.IndicateAttrQuery("#v2-target", "data-x", nil))); if __e != nil { return }
	return })
//line indicator.gox:49
}

// elem: extend to cover attributes and partial updates
//line indicator.gox:52
func (f *indicatorFragment) queue() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:53
			__e = __c.Set("id", "q-target"); if __e != nil { return }
//line indicator.gox:53
			__e = __c.Set("class", "base-class"); if __e != nil { return }
//line indicator.gox:53
			__e = __c.Set("data-a", "A0"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("base"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line indicator.gox:54
		__e = __c.Many(f.button("queue-1", doors.Join[doors.Indicators](
		doors.IndicatorAttr{
			Selector: doors.SelectorQuery("#q-target"),
			Name: "data-a",
			Value: "A1",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#q-target"),
			Class: "class-1",
		},
		doors.IndicatorContent{
			Selector: doors.SelectorQuery("#q-target"),
			Content: "first",
		},
	)), f.button("queue-2", doors.Join[doors.Indicators](
		// Partial update: does NOT touch data-a, so when this applies
		// data-a should restore to original (A0).
		doors.IndicatorAttr{
			Selector: doors.SelectorQuery("#q-target"),
			Name: "data-b",
			Value: "B2",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#q-target"),
			Class: "class-2",
		},
		doors.IndicatorContent{
			Selector: doors.SelectorQuery("#q-target"),
			Content: "second",
		},
	)), f.button("queue-3", doors.Join[doors.Indicators](
		// Partial update: does NOT touch data-a, so when this applies
		// data-a should restore to original (A0).
		doors.IndicatorAttr{
			Selector: doors.SelectorQuery("#q-target"),
			Name: "data-b",
			Value: "B2",
		},
		doors.IndicatorAttr{
			Selector: doors.SelectorQuery("#q-target"),
			Name: "data-a",
			Value: "A3",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#q-target"),
			Class: "class-2",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#q-target"),
			Class: "class-3",
		},
		doors.IndicatorContent{
			Selector: doors.SelectorQuery("#q-target"),
			Content: "second",
		},
	))); if __e != nil { return }
	return })
//line indicator.gox:110
}

//line indicator.gox:112
func (f *indicatorFragment) restore() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:113
			__e = __c.Set("id", "indicator-1"); if __e != nil { return }
//line indicator.gox:113
			__e = __c.Set("class", "class-1 class-3"); if __e != nil { return }
//line indicator.gox:113
			__e = __c.Set("data-attr1", "val-1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("content-1"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line indicator.gox:114
		__e = __c.Any(f.button("action-1", doors.Join[doors.Indicators](
		doors.IndicatorAttr{
			Selector: doors.SelectorQuery("#indicator-1"),
			Name: "data-attr1",
			Value: "val-other",
		},
		doors.IndicatorAttr{
			Selector: doors.SelectorQuery("#indicator-1"),
			Name: "data-attr2",
			Value: "val-2",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#indicator-1"),
			Class: "class-1",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#indicator-1"),
			Class: "class-1",
		},
		doors.IndicatorClassRemove{
			Selector: doors.SelectorQuery("#indicator-1"),
			Class: "class-3",
		},
		doors.IndicatorClass{
			Selector: doors.SelectorQuery("#indicator-1"),
			Class: "class-2",
		},
		doors.IndicatorContent{
			Selector: doors.SelectorQuery("#indicator-1"),
			Content: "indication",
		},
	))); if __e != nil { return }
	return })
//line indicator.gox:146
}

//line indicator.gox:148
func (f *indicatorFragment) selectors() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:149
			__e = __c.Set("id", "next"); if __e != nil { return }
//line indicator.gox:149
			__e = __c.Set("class", "block"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:150
			__e = __c.Set("id", "all-a"); if __e != nil { return }
//line indicator.gox:150
			__e = __c.Set("class", "multi keep"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("all-a"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:151
			__e = __c.Set("id", "all-b"); if __e != nil { return }
//line indicator.gox:151
			__e = __c.Set("class", "multi keep"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("all-b"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:152
			__e = __c.Set("id", "parent"); if __e != nil { return }
//line indicator.gox:152
			__e = __c.Set("class", "block"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line indicator.gox:153
			__e = __c.Any(f.button("indicate-parent", doors.IndicateAttrQueryParent(".block", "data-check", "true"))); if __e != nil { return }
//line indicator.gox:154
			__e = __c.Any(f.button("indicate-self", doors.IndicateContent("indication"))); if __e != nil { return }
//line indicator.gox:155
			__e = __c.Any(f.button("indicate-selector", doors.IndicateAttrQuery("#next", "data-check", "true"))); if __e != nil { return }
//line indicator.gox:156
			__e = __c.Any(f.button("indicate-self-attr", doors.IndicateAttr("data-self", "true"))); if __e != nil { return }
//line indicator.gox:157
			__e = (doors.AClick{
			Indicator: doors.IndicateClass("self-active"),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				<-time.After(500 * time.Millisecond)
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line indicator.gox:163
					__e = __c.Set("id", "indicate-self-class"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("indicate-self-class"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line indicator.gox:164
			__e = (doors.AClick{
			Indicator: doors.IndicateClassRemove("remove-me"),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				<-time.After(500 * time.Millisecond)
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line indicator.gox:170
					__e = __c.Set("id", "indicate-self-class-remove"); if __e != nil { return }
//line indicator.gox:170
					__e = __c.Set("class", "remove-me keep"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("indicate-self-class-remove"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line indicator.gox:171
			__e = __c.Any(f.button("indicate-query-content", doors.IndicateContentQuery("#next", "content"))); if __e != nil { return }
//line indicator.gox:172
			__e = __c.Any(f.button("indicate-query-class", doors.IndicateClassQuery("#next", "query-class"))); if __e != nil { return }
//line indicator.gox:173
			__e = __c.Any(f.button("indicate-query-class-remove", doors.IndicateClassRemoveQuery("#next", "block"))); if __e != nil { return }
//line indicator.gox:174
			__e = __c.Any(f.button("indicate-all-content", doors.IndicateContentQueryAll(".multi", "all"))); if __e != nil { return }
//line indicator.gox:175
			__e = __c.Any(f.button("indicate-all-attr", doors.IndicateAttrQueryAll(".multi", "data-all", "true"))); if __e != nil { return }
//line indicator.gox:176
			__e = __c.Any(f.button("indicate-all-class", doors.IndicateClassQueryAll(".multi", "all-class"))); if __e != nil { return }
//line indicator.gox:177
			__e = __c.Any(f.button("indicate-all-class-remove", doors.IndicateClassRemoveQueryAll(".multi", "keep"))); if __e != nil { return }
//line indicator.gox:178
			__e = __c.Any(f.button("indicate-parent-content", doors.IndicateContentQueryParent(".block", "parent-content"))); if __e != nil { return }
//line indicator.gox:179
			__e = __c.Any(f.button("indicate-parent-class", doors.IndicateClassQueryParent(".block", "parent-class"))); if __e != nil { return }
//line indicator.gox:180
			__e = __c.Any(f.button("indicate-parent-class-remove", doors.IndicateClassRemoveQueryParent(".block", "block"))); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line indicator.gox:182
}

//line indicator.gox:184
func (f *indicatorFragment) button(id string, indicator doors.Indicators) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line indicator.gox:185
			__e = __c.Set("id", id); if __e != nil { return }
//line indicator.gox:185
			__e = __c.Modify(doors.A(ctx, f.handler(indicator))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line indicator.gox:185
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line indicator.gox:186
}

func (f *indicatorFragment) handler(indicator doors.Indicators) doors.Attr {
	return doors.AClick{
		Indicator: indicator,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			<-time.After(500 * time.Millisecond)
			return false
		},
	}
}
