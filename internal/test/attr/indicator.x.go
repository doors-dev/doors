// Managed by GoX v0.1.28

//line indicator.gox:1
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

//line indicator.gox:16
func (f *indicatorFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line indicator.gox:17
		__e = __c.Any(f.selectors()); if __e != nil { return }
//line indicator.gox:18
		__e = __c.Any(f.restore()); if __e != nil { return }
//line indicator.gox:19
		__e = __c.Any(f.queue()); if __e != nil { return }
	return })
//line indicator.gox:20
}

// elem: extend to cover attributes and partial updates
//line indicator.gox:23
func (f *indicatorFragment) queue() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:24
			__e = __c.Set("id", "q-target"); if __e != nil { return }
//line indicator.gox:24
			__e = __c.Set("class", "base-class"); if __e != nil { return }
//line indicator.gox:24
			__e = __c.Set("data-a", "A0"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("base"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line indicator.gox:25
		__e = __c.Many(f.button("queue-1", []doors.Indicator{
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
	}), f.button("queue-2", []doors.Indicator{
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
	}), f.button("queue-3", []doors.Indicator{
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
	})); if __e != nil { return }
	return })
//line indicator.gox:81
}

//line indicator.gox:83
func (f *indicatorFragment) restore() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:84
			__e = __c.Set("id", "indicator-1"); if __e != nil { return }
//line indicator.gox:84
			__e = __c.Set("class", "class-1 class-3"); if __e != nil { return }
//line indicator.gox:84
			__e = __c.Set("data-attr1", "val-1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("content-1"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line indicator.gox:85
		__e = __c.Any(f.button("action-1", []doors.Indicator{
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
	})); if __e != nil { return }
	return })
//line indicator.gox:117
}

//line indicator.gox:119
func (f *indicatorFragment) selectors() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:120
			__e = __c.Set("id", "next"); if __e != nil { return }
//line indicator.gox:120
			__e = __c.Set("class", "block"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:121
			__e = __c.Set("id", "all-a"); if __e != nil { return }
//line indicator.gox:121
			__e = __c.Set("class", "multi keep"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("all-a"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:122
			__e = __c.Set("id", "all-b"); if __e != nil { return }
//line indicator.gox:122
			__e = __c.Set("class", "multi keep"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("all-b"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line indicator.gox:123
			__e = __c.Set("id", "parent"); if __e != nil { return }
//line indicator.gox:123
			__e = __c.Set("class", "block"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line indicator.gox:124
			__e = __c.Any(f.button("indicate-parent", doors.IndicatorOnlyAttrQueryParent(".block", "data-check", "true"))); if __e != nil { return }
//line indicator.gox:125
			__e = __c.Any(f.button("indicate-self", doors.IndicatorOnlyContent("indication"))); if __e != nil { return }
//line indicator.gox:126
			__e = __c.Any(f.button("indicate-selector", doors.IndicatorOnlyAttrQuery("#next", "data-check", "true"))); if __e != nil { return }
//line indicator.gox:127
			__e = __c.Any(f.button("indicate-self-attr", doors.IndicatorOnlyAttr("data-self", "true"))); if __e != nil { return }
//line indicator.gox:128
			__e = (doors.AClick{
			Indicator: doors.IndicatorOnlyClass("self-active"),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				<-time.After(500 * time.Millisecond)
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line indicator.gox:134
					__e = __c.Set("id", "indicate-self-class"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("indicate-self-class"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line indicator.gox:135
			__e = (doors.AClick{
			Indicator: doors.IndicatorOnlyClassRemove("remove-me"),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				<-time.After(500 * time.Millisecond)
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line indicator.gox:141
					__e = __c.Set("id", "indicate-self-class-remove"); if __e != nil { return }
//line indicator.gox:141
					__e = __c.Set("class", "remove-me keep"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("indicate-self-class-remove"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line indicator.gox:142
			__e = __c.Any(f.button("indicate-query-content", doors.IndicatorOnlyContentQuery("#next", "content"))); if __e != nil { return }
//line indicator.gox:143
			__e = __c.Any(f.button("indicate-query-class", doors.IndicatorOnlyClassQuery("#next", "query-class"))); if __e != nil { return }
//line indicator.gox:144
			__e = __c.Any(f.button("indicate-query-class-remove", doors.IndicatorOnlyClassRemoveQuery("#next", "block"))); if __e != nil { return }
//line indicator.gox:145
			__e = __c.Any(f.button("indicate-all-content", doors.IndicatorOnlyContentQueryAll(".multi", "all"))); if __e != nil { return }
//line indicator.gox:146
			__e = __c.Any(f.button("indicate-all-attr", doors.IndicatorOnlyAttrQueryAll(".multi", "data-all", "true"))); if __e != nil { return }
//line indicator.gox:147
			__e = __c.Any(f.button("indicate-all-class", doors.IndicatorOnlyClassQueryAll(".multi", "all-class"))); if __e != nil { return }
//line indicator.gox:148
			__e = __c.Any(f.button("indicate-all-class-remove", doors.IndicatorOnlyClassRemoveQueryAll(".multi", "keep"))); if __e != nil { return }
//line indicator.gox:149
			__e = __c.Any(f.button("indicate-parent-content", doors.IndicatorOnlyContentQueryParent(".block", "parent-content"))); if __e != nil { return }
//line indicator.gox:150
			__e = __c.Any(f.button("indicate-parent-class", doors.IndicatorOnlyClassQueryParent(".block", "parent-class"))); if __e != nil { return }
//line indicator.gox:151
			__e = __c.Any(f.button("indicate-parent-class-remove", doors.IndicatorOnlyClassRemoveQueryParent(".block", "block"))); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line indicator.gox:153
}

//line indicator.gox:155
func (f *indicatorFragment) button(id string, indicator []doors.Indicator) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line indicator.gox:156
			__e = __c.Set("id", id); if __e != nil { return }
//line indicator.gox:156
			__e = __c.Modify(doors.A(ctx, f.handler(indicator))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line indicator.gox:156
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line indicator.gox:157
}

func (f *indicatorFragment) handler(indicator []doors.Indicator) doors.Attr {
	return doors.AClick{
		Indicator: indicator,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			<-time.After(500 * time.Millisecond)
			return false
		},
	}
}
