// Managed by GoX v0.1.22

//line components.gox:1
package components

import (
	"context"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

//line components.gox:12
func head(b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:13
		__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.AttrSet("hidden", true); if __e != nil { return }
//line components.gox:13
				__e = __c.AttrSet("id", "head-anchor"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:14
				if ctx.Value(0).(test.Path).Vh {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("home"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:16
						__e = __c.AttrSet("name", "description"); if __e != nil { return }
//line components.gox:16
						__e = __c.AttrSet("content", "Welcome to the home page"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:17
						__e = __c.AttrSet("name", "keywords"); if __e != nil { return }
//line components.gox:17
						__e = __c.AttrSet("content", "home, main, index"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:18
						__e = __c.AttrSet("property", "og:title"); if __e != nil { return }
//line components.gox:18
						__e = __c.AttrSet("content", "Home Page"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
//line components.gox:19
				} else if ctx.Value(0).(test.Path).Vs {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("s"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:21
						__e = __c.AttrSet("name", "description"); if __e != nil { return }
//line components.gox:21
						__e = __c.AttrSet("content", "String page description"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:22
						__e = __c.AttrSet("name", "category"); if __e != nil { return }
//line components.gox:22
						__e = __c.AttrSet("content", "text-content"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				} else  {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line components.gox:24
						__e = __c.Any(ctx.Value(0).(test.Path).P); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:25
						__e = __c.AttrSet("name", "description"); if __e != nil { return }
//line components.gox:25
						__e = __c.AttrSet("content", "Page for parameter: " + ctx.Value(0).(test.Path).P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:26
						__e = __c.AttrSet("name", "keywords"); if __e != nil { return }
//line components.gox:26
						__e = __c.AttrSet("content", "param, " + ctx.Value(0).(test.Path).P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:27
						__e = __c.AttrSet("property", "og:title"); if __e != nil { return }
//line components.gox:27
						__e = __c.AttrSet("content", "Param: " + ctx.Value(0).(test.Path).P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:28
						__e = __c.AttrSet("name", "author"); if __e != nil { return }
//line components.gox:28
						__e = __c.AttrSet("content", "Parameter Author"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line components.gox:31
}

type LinksFragment struct {
	test.Beam
	Param string
}

//line components.gox:38
func (f *LinksFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:39
		__e = __c.Any(head(f.B)); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.AttrSet("hidden", true); if __e != nil { return }
//line components.gox:40
			__e = __c.AttrSet("id", "active-links"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line components.gox:41
			__e = doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				Indicator: doors.IndicatorOnlyAttr("aria-current", "page"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:48
					__e = __c.AttrSet("id", "active-default"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-default"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:49
			__e = doors.ALink{
			Model: test.Path{
				Vp: true,
				P: f.Param,
			},
			Active: doors.Active{
				PathMatcher: doors.PathMatcherStarts(),
				Indicator: doors.IndicatorOnlyAttr("data-active", "starts"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:58
					__e = __c.AttrSet("id", "active-starts"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-starts"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:59
			__e = doors.ALink{
			Model: test.Path{
				Vp: true,
				P: f.Param,
			},
			Active: doors.Active{
				PathMatcher: doors.PathMatcherSegments(0),
				Indicator: doors.IndicatorOnlyAttr("data-active", "segments"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:68
					__e = __c.AttrSet("id", "active-segments"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-segments"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:69
			__e = doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Fragment: "details",
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlyIgnoreAll(),
				FragmentMatch: true,
				Indicator: doors.IndicatorOnlyAttr("data-active", "fragment"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:79
					__e = __c.AttrSet("id", "active-fragment"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-fragment"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:80
			__e = doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: []doors.QueryMatcher{
					doors.QueryMatcherIgnoreSome("page"),
					doors.QueryMatcherSome("mode"),
					doors.QueryMatcherIfPresent("optional"),
				},
				Indicator: doors.IndicatorOnlyAttr("data-active", "query"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:92
					__e = __c.AttrSet("id", "active-query"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:93
			__e = doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlyIgnoreSome("page"),
				Indicator: doors.IndicatorOnlyAttr("data-active", "only-ignore-some"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:101
					__e = __c.AttrSet("id", "active-query-only-ignore-some"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-ignore-some"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:102
			__e = doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlySome("mode"),
				Indicator: doors.IndicatorOnlyAttr("data-active", "only-some"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:110
					__e = __c.AttrSet("id", "active-query-only-some"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-some"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:111
			__e = doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlyIfPresent("optional"),
				Indicator: doors.IndicatorOnlyAttr("data-active", "only-if-present"),
			},
		}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:119
					__e = __c.AttrSet("id", "active-query-only-if-present"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-if-present"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:122
		__e = doors.ALink{
		Model: test.Path{
			Vh: true,
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:126
				__e = __c.AttrSet("id", "home"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("home"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:128
		__e = doors.ALink{
		Model: test.Path{
			Vp: true,
			P: f.Param,
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:133
				__e = __c.AttrSet("id", "param"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("param"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:135
		__e = doors.ALink{
		Model: test.Path{
			Vs: true,
			P: f.Param,
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:140
				__e = __c.AttrSet("id", "string"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("string"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.InitVoid("br"); if __e != nil { return }
		{
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.InitVoid("br"); if __e != nil { return }
		{
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:143
			__e = __c.AttrSet("id", "action-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("action-target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:144
		__e = doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyLocationRawAssign(test.Host + "/s"))
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:149
				__e = __c.AttrSet("id", "raw-assign"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw-assign"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:150
		__e = doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyIndicate(
				doors.IndicatorOnlyAttrQuery("#action-target", "data-indicated", "true"),
				200 * time.Millisecond,
			))
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:158
				__e = __c.AttrSet("id", "action-indicate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("action-indicate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:159
		__e = doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyScroll("#scroll-target"))
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:164
				__e = __c.AttrSet("id", "action-scroll"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("action-scroll"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:165
		__e = doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyEmit("alert", "Hello!"))
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("alert"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("$on(\"alert\", (message) => {\n\t\t\talert(message)\n\t\t})"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:178
			__e = __c.AttrSet("style", "height: 1800px"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:179
			__e = __c.AttrSet("id", "scroll-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("scroll-target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:180
}

type ProxyFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line components.gox:187
func (f *ProxyFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:188
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "literal")
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:193
			__e = __c.AttrSet("id", "proxy-literal"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-literal"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:195
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "container")
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:201
				__e = __c.AttrSet("id", "proxy-container"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-container"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:204
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line components.gox:205
}
