// Managed by GoX v0.1.28

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
		__e = (new(doors.Door)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Set("hidden", true); if __e != nil { return }
//line components.gox:13
				__e = __c.Set("id", "head-anchor"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:15
				path, _ := b.Effect(ctx)

//line components.gox:17
				if path.Vh {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("home"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:19
						__e = __c.Set("name", "description"); if __e != nil { return }
//line components.gox:19
						__e = __c.Set("content", "Welcome to the home page"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:20
						__e = __c.Set("name", "keywords"); if __e != nil { return }
//line components.gox:20
						__e = __c.Set("content", "home, main, index"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:21
						__e = __c.Set("property", "og:title"); if __e != nil { return }
//line components.gox:21
						__e = __c.Set("content", "Home Page"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
//line components.gox:22
				} else if path.Vs {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("s"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:24
						__e = __c.Set("name", "description"); if __e != nil { return }
//line components.gox:24
						__e = __c.Set("content", "String page description"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:25
						__e = __c.Set("name", "category"); if __e != nil { return }
//line components.gox:25
						__e = __c.Set("content", "text-content"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				} else  {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line components.gox:27
						__e = __c.Any(path.P); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:28
						__e = __c.Set("name", "description"); if __e != nil { return }
//line components.gox:28
						__e = __c.Set("content", "Page for parameter: " + path.P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:29
						__e = __c.Set("name", "keywords"); if __e != nil { return }
//line components.gox:29
						__e = __c.Set("content", "param, " + path.P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:30
						__e = __c.Set("property", "og:title"); if __e != nil { return }
//line components.gox:30
						__e = __c.Set("content", "Param: " + path.P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:31
						__e = __c.Set("name", "author"); if __e != nil { return }
//line components.gox:31
						__e = __c.Set("content", "Parameter Author"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line components.gox:34
}

type LinksFragment struct {
	test.Beam
	Param string
}

//line components.gox:41
func (f *LinksFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:42
		__e = __c.Any(head(f.B)); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Set("hidden", true); if __e != nil { return }
//line components.gox:43
			__e = __c.Set("id", "active-links"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line components.gox:44
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				Indicator: doors.IndicatorOnlyAttr("aria-current", "page"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:51
					__e = __c.Set("id", "active-default"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-default"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:52
			__e = (doors.ALink{
			Model: test.Path{
				Vp: true,
				P: f.Param,
			},
			Active: doors.Active{
				PathMatcher: doors.PathMatcherStarts(),
				Indicator: doors.IndicatorOnlyAttr("data-active", "starts"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:61
					__e = __c.Set("id", "active-starts"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-starts"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:62
			__e = (doors.ALink{
			Model: test.Path{
				Vp: true,
				P: f.Param,
			},
			Active: doors.Active{
				PathMatcher: doors.PathMatcherSegments(0),
				Indicator: doors.IndicatorOnlyAttr("data-active", "segments"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:71
					__e = __c.Set("id", "active-segments"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-segments"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:72
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Fragment: "details",
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlyIgnoreAll(),
				FragmentMatch: true,
				Indicator: doors.IndicatorOnlyAttr("data-active", "fragment"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:82
					__e = __c.Set("id", "active-fragment"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-fragment"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:83
			__e = (doors.ALink{
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
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:95
					__e = __c.Set("id", "active-query"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:96
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlyIgnoreSome("page"),
				Indicator: doors.IndicatorOnlyAttr("data-active", "only-ignore-some"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:104
					__e = __c.Set("id", "active-query-only-ignore-some"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-ignore-some"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:105
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlySome("mode"),
				Indicator: doors.IndicatorOnlyAttr("data-active", "only-some"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:113
					__e = __c.Set("id", "active-query-only-some"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-some"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:114
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherOnlyIfPresent("optional"),
				Indicator: doors.IndicatorOnlyAttr("data-active", "only-if-present"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:122
					__e = __c.Set("id", "active-query-only-if-present"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-if-present"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:125
		__e = (doors.ALink{
		Model: test.Path{
			Vh: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:129
				__e = __c.Set("id", "home"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("home"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:131
		__e = (doors.ALink{
		Model: test.Path{
			Vp: true,
			P: f.Param,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:136
				__e = __c.Set("id", "param"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("param"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:138
		__e = (doors.ALink{
		Model: test.Path{
			Vs: true,
			P: f.Param,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:143
				__e = __c.Set("id", "string"); if __e != nil { return }
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
//line components.gox:146
			__e = __c.Set("id", "action-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("action-target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:147
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyLocationRawAssign(test.Host + "/s"))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:152
				__e = __c.Set("id", "raw-assign"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw-assign"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:153
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyIndicate(
				doors.IndicatorOnlyAttrQuery("#action-target", "data-indicated", "true"),
				200 * time.Millisecond,
			))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:161
				__e = __c.Set("id", "action-indicate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("action-indicate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:162
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyScroll("#scroll-target"))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:167
				__e = __c.Set("id", "action-scroll"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("action-scroll"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:168
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyEmit("alert", "Hello!"))
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
//line components.gox:181
			__e = __c.Set("style", "height: 1800px"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:182
			__e = __c.Set("id", "scroll-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("scroll-target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:183
}

type ProxyFragment struct {
	test.NoBeam
	r *test.Reporter
}

type ProxyClassComponent struct {
	test.NoBeam
}

//line components.gox:194
func (ProxyClassComponent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:195
			__e = __c.Set("id", "proxy-class-component"); if __e != nil { return }
//line components.gox:195
			__e = __c.Set("class", "base-component"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-class-component"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:196
}

func proxyContainerClass() doors.Classes {
	return doors.Class("proxy-container", "proxy-skip").Filter("proxy-skip")
}

func classAttrValue() doors.Classes {
	return doors.Class("class-attr", "class-skip").Filter("class-skip").Add("class-added")
}

func classModifier() doors.Classes {
	return doors.Class("class-mod").Filter("class-skip")
}

func proxyDirectMod() gox.Modify {
	return gox.ModifyFunc(func(_ context.Context, tag string, attrs gox.Attrs) error {
		attrs.Get("data-proxy-mod").Set(tag + ":direct")
		return nil
	})
}

//line components.gox:217
func (f *ProxyFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:218
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "literal")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:223
			__e = __c.Set("id", "proxy-literal"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-literal"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:225
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 0, "container")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:231
				__e = __c.Set("id", "proxy-container"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-container"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:234
		__e = (doors.Class("proxy-literal")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:234
			__e = __c.Set("id", "proxy-class-literal"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-class-literal"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:236
		__e = (doors.ProxyMod(proxyDirectMod())).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:236
			__e = __c.Set("id", "proxy-direct-mod"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-direct-mod"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:238
		__e = (proxyContainerClass()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:239
				__e = __c.Set("id", "proxy-class-container"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-class-container"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:240
				__e = __c.Set("id", "proxy-class-sibling"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-class-sibling"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:243
		__e = (doors.Class("proxy-component")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line components.gox:243
			__e = __c.Any(ProxyClassComponent{}); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:245
		__e = (doors.Class("proxy-parallel")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
//line components.gox:246
			__e = (doors.Parallel()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.InitContainer(); if __e != nil { return }
				{
					__e = __c.Init("button"); if __e != nil { return }
					{
//line components.gox:247
						__e = __c.Set("id", "proxy-class-parallel"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("proxy-class-parallel"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:251
			__e = __c.Set("id", "class-attr"); if __e != nil { return }
//line components.gox:251
			__e = __c.Set("class", classAttrValue()); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("class-attr"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:252
			__e = __c.Set("id", "class-mod"); if __e != nil { return }
//line components.gox:252
			__e = __c.Set("class", "base-mod class-skip"); if __e != nil { return }
//line components.gox:252
			__e = __c.Modify(classModifier()); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("class-mod"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:254
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line components.gox:255
}
