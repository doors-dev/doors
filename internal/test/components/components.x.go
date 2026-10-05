// Managed by GoX v0.3.2

//line components.gox:1
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

package components

import (
	"context"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

//line components.gox:26
func head(b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:27
		__e = (new(doors.Door)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Set("hidden", true); if __e != nil { return }
//line components.gox:27
				__e = __c.Set("id", "head-anchor"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:29
				path, _ := b.Effect(ctx)

//line components.gox:31
				if path.Vh {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("home"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:33
						__e = __c.Set("name", "description"); if __e != nil { return }
//line components.gox:33
						__e = __c.Set("content", "Welcome to the home page"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:34
						__e = __c.Set("name", "keywords"); if __e != nil { return }
//line components.gox:34
						__e = __c.Set("content", "home, main, index"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:35
						__e = __c.Set("property", "og:title"); if __e != nil { return }
//line components.gox:35
						__e = __c.Set("content", "Home Page"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
//line components.gox:36
				} else if path.Vs {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("s"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:38
						__e = __c.Set("name", "description"); if __e != nil { return }
//line components.gox:38
						__e = __c.Set("content", "String page description"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:39
						__e = __c.Set("name", "category"); if __e != nil { return }
//line components.gox:39
						__e = __c.Set("content", "text-content"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				} else  {
					__e = __c.Init("title"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line components.gox:41
						__e = __c.Any(path.P); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:42
						__e = __c.Set("name", "description"); if __e != nil { return }
//line components.gox:42
						__e = __c.Set("content", "Page for parameter: " + path.P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:43
						__e = __c.Set("name", "keywords"); if __e != nil { return }
//line components.gox:43
						__e = __c.Set("content", "param, " + path.P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:44
						__e = __c.Set("property", "og:title"); if __e != nil { return }
//line components.gox:44
						__e = __c.Set("content", "Param: " + path.P); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.InitVoid("meta"); if __e != nil { return }
					{
//line components.gox:45
						__e = __c.Set("name", "author"); if __e != nil { return }
//line components.gox:45
						__e = __c.Set("content", "Parameter Author"); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line components.gox:48
}

type LinksFragment struct {
	test.Beam
	Param string
}

//line components.gox:55
func (f *LinksFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:56
		__e = __c.Any(head(f.B)); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Set("hidden", true); if __e != nil { return }
//line components.gox:57
			__e = __c.Set("id", "active-links"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line components.gox:58
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				Indicator: doors.IndicateAttr("aria-current", "page"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:65
					__e = __c.Set("id", "active-default"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-default"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:66
			__e = (doors.ALink{
			Model: test.Path{
				Vp: true,
				P: f.Param,
			},
			Active: doors.Active{
				PathMatcher: doors.PathMatcherStarts(),
				Indicator: doors.IndicateAttr("data-active", "starts"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:75
					__e = __c.Set("id", "active-starts"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-starts"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:76
			__e = (doors.ALink{
			Model: test.Path{
				Vp: true,
				P: f.Param,
			},
			Active: doors.Active{
				PathMatcher: doors.PathMatcherSegments(0),
				Indicator: doors.IndicateAttr("data-active", "segments"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:85
					__e = __c.Set("id", "active-segments"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-segments"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:86
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Fragment: "details",
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherIgnoreAll(),
				FragmentMatch: true,
				Indicator: doors.IndicateAttr("data-active", "fragment"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:96
					__e = __c.Set("id", "active-fragment"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-fragment"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:97
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherIgnoreSome("page").And(doors.QueryMatcherSome("mode")).And(doors.QueryMatcherIfPresent("optional")),
				Indicator: doors.IndicateAttr("data-active", "query"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:105
					__e = __c.Set("id", "active-query"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:106
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherIgnoreSome("page"),
				Indicator: doors.IndicateAttr("data-active", "only-ignore-some"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:114
					__e = __c.Set("id", "active-query-only-ignore-some"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-ignore-some"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:115
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherSome("mode").And(doors.QueryMatcherIgnoreAll()),
				Indicator: doors.IndicateAttr("data-active", "only-some"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:123
					__e = __c.Set("id", "active-query-only-some"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-some"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
//line components.gox:124
			__e = (doors.ALink{
			Model: test.Path{
				Vh: true,
			},
			Active: doors.Active{
				QueryMatcher: doors.QueryMatcherIfPresent("optional").And(doors.QueryMatcherIgnoreAll()),
				Indicator: doors.IndicateAttr("data-active", "only-if-present"),
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("a"); if __e != nil { return }
				{
//line components.gox:132
					__e = __c.Set("id", "active-query-only-if-present"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("active-query-only-if-present"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:135
		__e = (doors.ALink{
		Model: test.Path{
			Vh: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:139
				__e = __c.Set("id", "home"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("home"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:141
		__e = (doors.ALink{
		Model: test.Path{
			Vp: true,
			P: f.Param,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:146
				__e = __c.Set("id", "param"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("param"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:148
		__e = (doors.ALink{
		Model: test.Path{
			Vs: true,
			P: f.Param,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line components.gox:153
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
//line components.gox:156
			__e = __c.Set("id", "action-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("action-target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:157
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionLocationRawAssign{URL: test.Host + "/s"})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:162
				__e = __c.Set("id", "raw-assign"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw-assign"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:163
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionLocationRawReplace{URL: test.Host + "/s"})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:168
				__e = __c.Set("id", "raw-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:169
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionIndicate{
				Indicator: doors.IndicateAttrQuery("#action-target", "data-indicated", "true"),
				Duration: 200 * time.Millisecond,
			})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:177
				__e = __c.Set("id", "action-indicate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("action-indicate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:178
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionScroll{Selector: "#scroll-target"})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:183
				__e = __c.Set("id", "action-scroll"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("action-scroll"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:184
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionEmit[any]{Name: "alert", Arg: "Hello!"})
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
//line components.gox:197
			__e = __c.Set("style", "height: 1800px"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:198
			__e = __c.Set("id", "scroll-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("scroll-target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:199
}

type ProxyFragment struct {
	test.NoBeam
	r *test.Reporter
}

type ProxyClassComponent struct {
	test.NoBeam
}

//line components.gox:210
func (ProxyClassComponent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:211
			__e = __c.Set("id", "proxy-class-component"); if __e != nil { return }
//line components.gox:211
			__e = __c.Set("class", "base-component"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-class-component"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:212
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

//line components.gox:233
func (f *ProxyFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:234
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
//line components.gox:239
			__e = __c.Set("id", "proxy-literal"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-literal"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:241
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
//line components.gox:247
				__e = __c.Set("id", "proxy-container"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-container"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:250
		__e = (doors.Class("proxy-literal")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:250
			__e = __c.Set("id", "proxy-class-literal"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-class-literal"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:252
		__e = (doors.ProxyMod(proxyDirectMod())).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line components.gox:252
			__e = __c.Set("id", "proxy-direct-mod"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-direct-mod"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:254
		__e = (proxyContainerClass()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:255
				__e = __c.Set("id", "proxy-class-container"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-class-container"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:256
				__e = __c.Set("id", "proxy-class-sibling"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-class-sibling"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:259
		__e = (doors.Class("proxy-component")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line components.gox:259
			__e = __c.Any(ProxyClassComponent{}); if __e != nil { return }
		return })); if __e != nil { return }
//line components.gox:261
		__e = (doors.Class("proxy-parallel")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Any(gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
//line components.gox:262
			__e = (doors.Parallel()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.InitContainer(); if __e != nil { return }
				{
					__e = __c.Init("button"); if __e != nil { return }
					{
//line components.gox:263
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
//line components.gox:267
			__e = __c.Set("id", "class-attr"); if __e != nil { return }
//line components.gox:267
			__e = __c.Set("class", classAttrValue()); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("class-attr"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:268
			__e = __c.Set("id", "class-mod"); if __e != nil { return }
//line components.gox:268
			__e = __c.Set("class", "base-mod class-skip"); if __e != nil { return }
//line components.gox:268
			__e = __c.Modify(classModifier()); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("class-mod"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line components.gox:270
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line components.gox:271
}
