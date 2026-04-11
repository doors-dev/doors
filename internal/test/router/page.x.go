// Managed by GoX v0.1.25

//line page.gox:1
package router

import (
	"context"
	"fmt"
	"net/url"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

type PathA struct {
	Path bool `path:"/a"`
}

//line page.gox:17
func pageA(b doors.Source[PathA]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:20
					__e = __c.AttrSet("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("A"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:21
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationAssign{Model: PathC{PathC1: true}})
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:26
						__e = __c.AttrSet("id", "assign"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("assign"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:29
}

type PathB struct {
	Path bool `path:"/b"`
}

type PathQuery struct {
	Path bool `path:"/q"`
	Tag *string `query:"tag"`
	Page *int `query:"page"`
}

type PathEscaped struct {
	Path bool `path:"/escaped/:Name"`
	Name string
}

type PathCrossA struct {
	Path bool `path:"/cross-a"`
}

type PathCrossB struct {
	Path bool `path:"/cross-b"`
}

type PathSlow struct {
	Path bool `path:"/slow"`
}

func values(items ...string) url.Values {
	v := url.Values{}
	for i := 0; i + 1 < len(items); i += 2 {
		v.Add(items[i], items[i + 1])
	}
	return v
}

//line page.gox:66
func pageQuery(b doors.Source[PathQuery]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:69
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:69
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:70
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:70
						__e = __c.AttrSet("id", "tag"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:71
						if ctx.Value(0).(PathQuery).Tag != nil {
//line page.gox:72
							__e = __c.Any(*ctx.Value(0).(PathQuery).Tag); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:75
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:75
						__e = __c.AttrSet("id", "page-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:76
						if ctx.Value(0).(PathQuery).Page != nil {
//line page.gox:77
							__e = __c.Any(fmt.Sprint(*ctx.Value(0).(PathQuery).Page)); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:82
				tag := "next"
				page := 2

//line page.gox:85
				__e = doors.ALink{
				Model: PathQuery{
					Path: true,
					Tag: &tag,
					Page: &page,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:91
						__e = __c.AttrSet("id", "query-next"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("query-next"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:94
}

//line page.gox:96
func pageLocation(b doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:99
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:99
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:100
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:100
						__e = __c.AttrSet("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:101
						__e = __c.Any(ctx.Value(0).(doors.Location).String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:103
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:103
						__e = __c.AttrSet("id", "location-path"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:104
						__e = __c.Any(ctx.Value(0).(doors.Location).Path()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:106
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:106
						__e = __c.AttrSet("id", "tag-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:107
						__e = __c.Any(ctx.Value(0).(doors.Location).Query.Get("tag")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:109
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:109
						__e = __c.AttrSet("id", "page-query-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:110
						__e = __c.Any(ctx.Value(0).(doors.Location).Query.Get("page")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:114
}

//line page.gox:116
func pageLocationActive(b doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:119
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:119
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:120
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:120
						__e = __c.AttrSet("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:121
						__e = __c.Any(ctx.Value(0).(doors.Location).String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("hidden", true); if __e != nil { return }
//line page.gox:123
					__e = __c.AttrSet("id", "active-links"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:124
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Active: doors.Active{
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:131
							__e = __c.AttrSet("id", "active-full"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-full"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:132
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section"},
					},
					Active: doors.Active{
						PathMatcher: doors.PathMatcherStarts(),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:140
							__e = __c.AttrSet("id", "active-starts"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-starts"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:141
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section", "fixed"},
					},
					Active: doors.Active{
						PathMatcher: doors.PathMatcherSegments(0),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:149
							__e = __c.AttrSet("id", "active-segments"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-segments"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:150
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
						),
					},
					Active: doors.Active{
						QueryMatcher: []doors.QueryMatcher{
							doors.QueryMatcherIgnoreAll(),
						},
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:163
							__e = __c.AttrSet("id", "active-ignore-all"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-ignore-all"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:164
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"optional", "yes",
							"page", "1",
						),
					},
					Active: doors.Active{
						QueryMatcher: []doors.QueryMatcher{
							doors.QueryMatcherIgnoreSome("page"),
							doors.QueryMatcherSome("mode"),
							doors.QueryMatcherIfPresent("optional"),
						},
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:181
							__e = __c.AttrSet("id", "active-query"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-query"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:182
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"optional", "yes",
							"page", "1",
						),
					},
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherOnlyIgnoreSome("page"),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:195
							__e = __c.AttrSet("id", "active-only-ignore-some"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-ignore-some"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:196
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"page", "1",
						),
					},
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherOnlySome("mode"),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:208
							__e = __c.AttrSet("id", "active-only-some"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-some"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:209
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"optional", "yes",
							"page", "1",
						),
					},
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherOnlyIfPresent("optional"),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:221
							__e = __c.AttrSet("id", "active-only-if-present"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-if-present"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:222
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Fragment: "details",
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherOnlyIgnoreAll(),
						FragmentMatch: true,
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:232
							__e = __c.AttrSet("id", "active-fragment"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-fragment"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:234
					__e = __c.AttrSet("id", "nav-links"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:235
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:239
							__e = __c.AttrSet("id", "nav-home"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-home"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:240
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section", "child"},
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:244
							__e = __c.AttrSet("id", "nav-starts"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-starts"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:245
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "other"},
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:249
							__e = __c.AttrSet("id", "nav-segments"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-segments"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:250
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Fragment: "details",
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:255
							__e = __c.AttrSet("id", "nav-fragment"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-fragment"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:256
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"page", "9",
						),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:264
							__e = __c.AttrSet("id", "nav-query"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:265
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"optional", "yes",
							"page", "9",
						),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:274
							__e = __c.AttrSet("id", "nav-query-optional"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query-optional"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:275
					__e = doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"optional", "no",
							"page", "9",
						),
					},
				}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:284
							__e = __c.AttrSet("id", "nav-query-optional-miss"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query-optional-miss"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:288
}

//line page.gox:290
func pageEscaped(b doors.Source[PathEscaped]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:293
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:293
						__e = __c.AttrSet("id", "name-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:294
						__e = __c.Any(ctx.Value(0).(PathEscaped).Name); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:297
				name := "next value/again"

//line page.gox:299
				__e = doors.ALink{
				Model: PathEscaped{
					Path: true,
					Name: name,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:304
						__e = __c.AttrSet("id", "next-escaped"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("next-escaped"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:307
}

//line page.gox:309
func pageCrossA() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:312
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:312
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:313
				__e = doors.ALink{
				Model: PathCrossB{
					Path: true,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:317
						__e = __c.AttrSet("id", "cross-next"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("cross-next"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:320
}

//line page.gox:322
func pageCrossB() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:325
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:325
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:326
					__e = __c.AttrSet("id", "page-name"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("cross-b"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:329
}

//line page.gox:331
func pageSlow() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:334
				__e = __c.Any(gox.EditorFunc(func(cur gox.Cursor) error {
				<-time.After(1100 * time.Millisecond)
				return nil
			})); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:338
					__e = __c.AttrSet("id", "slow-page"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("slow-page"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:341
}

//line page.gox:343
func pageError(err error) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:345
			__e = __c.Any(doors.Status(500)); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:347
					__e = __c.AttrSet("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("error"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:348
					__e = __c.AttrSet("id", "error-message"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:348
					__e = __c.Any(err.Error()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:351
}

//line page.gox:353
func static(path string, code int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("head"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line page.gox:356
			if code >= 0 {
//line page.gox:357
				__e = __c.Any(doors.Status(code)); if __e != nil { return }
			}
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:360
					__e = __c.AttrSet("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:360
					__e = __c.Any(path); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:363
}

type PathC struct {
	PathC1 bool `path:"/c1"`
	PathC2 bool `path:"/c2"`
}

//line page.gox:370
func pageC(b doors.Source[PathC]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:373
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line page.gox:373
					if ctx.Value(0).(PathC).PathC1 {
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:374
							__e = __c.AttrSet("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c1"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					} else  {
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:376
							__e = __c.AttrSet("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c2"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
				return })); if __e != nil { return }
//line page.gox:379
				__e = doors.ALink{
				Model: PathC{
					PathC1: true,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c1"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:385
				__e = doors.ALink{
				Model: PathC{
					PathC2: true,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:389
						__e = __c.AttrSet("id", "c2"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c2"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:391
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReplace{Model: PathC{PathC2: true}})
					return true
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:396
						__e = __c.AttrSet("id", "replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:397
					__e = __c.AttrSet("id", "marker"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:397
					__e = __c.Any(doors.IDRand()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:399
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReload{})
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:404
						__e = __c.AttrSet("id", "reload"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("reload"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:406
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationAssign(PathB{}))
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:411
						__e = __c.AttrSet("id", "assign_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("assign_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:413
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationReplace(PathB{}))
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:418
						__e = __c.AttrSet("id", "replace_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:420
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationReload())
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:425
						__e = __c.AttrSet("id", "reload_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("reload_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:428
}
