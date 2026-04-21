// Managed by GoX v0.1.28

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
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("A"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:21
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationAssign{Model: PathC{PathC1: true}})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:26
						__e = __c.Set("id", "assign"); if __e != nil { return }
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

//line page.gox:58
func pageParallel() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:61
				__e = (doors.Parallel()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line page.gox:63
						<-time.After(500 * time.Millisecond)

						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:65
							__e = __c.Set("id", "part-a"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("part-a"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:67
				__e = (doors.Parallel()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line page.gox:69
						<-time.After(500 * time.Millisecond)

						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:71
							__e = __c.Set("id", "part-b"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("part-b"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.InitContainer(); if __e != nil { return }
				{
//line page.gox:75
					<-time.After(500 * time.Millisecond)

					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:77
						__e = __c.Set("id", "part-c"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("part-c"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:81
}

func values(items ...string) url.Values {
	v := url.Values{}
	for i := 0; i + 1 < len(items); i += 2 {
		v.Add(items[i], items[i + 1])
	}
	return v
}

//line page.gox:91
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
//line page.gox:94
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:94
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:95
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:95
						__e = __c.Set("id", "tag"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:96
						if ctx.Value(0).(PathQuery).Tag != nil {
//line page.gox:97
							__e = __c.Any(*ctx.Value(0).(PathQuery).Tag); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:100
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:100
						__e = __c.Set("id", "page-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:101
						if ctx.Value(0).(PathQuery).Page != nil {
//line page.gox:102
							__e = __c.Any(fmt.Sprint(*ctx.Value(0).(PathQuery).Page)); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:107
				tag := "next"
				page := 2

//line page.gox:110
				__e = (doors.ALink{
				Model: PathQuery{
					Path: true,
					Tag: &tag,
					Page: &page,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:116
						__e = __c.Set("id", "query-next"); if __e != nil { return }
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
//line page.gox:119
}

//line page.gox:121
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
//line page.gox:124
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:124
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:125
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:125
						__e = __c.Set("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:126
						__e = __c.Any(ctx.Value(0).(doors.Location).String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:128
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:128
						__e = __c.Set("id", "location-path"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:129
						__e = __c.Any(ctx.Value(0).(doors.Location).Path()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:131
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:131
						__e = __c.Set("id", "tag-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:132
						__e = __c.Any(ctx.Value(0).(doors.Location).Query.Get("tag")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:134
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:134
						__e = __c.Set("id", "page-query-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:135
						__e = __c.Any(ctx.Value(0).(doors.Location).Query.Get("page")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:139
}

//line page.gox:141
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
//line page.gox:144
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:144
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:145
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:145
						__e = __c.Set("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:146
						__e = __c.Any(ctx.Value(0).(doors.Location).String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Set("hidden", true); if __e != nil { return }
//line page.gox:148
					__e = __c.Set("id", "active-links"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:149
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Active: doors.Active{
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:156
							__e = __c.Set("id", "active-full"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-full"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:157
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section"},
					},
					Active: doors.Active{
						PathMatcher: doors.PathMatcherStarts(),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:165
							__e = __c.Set("id", "active-starts"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-starts"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:166
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section", "fixed"},
					},
					Active: doors.Active{
						PathMatcher: doors.PathMatcherSegments(0),
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:174
							__e = __c.Set("id", "active-segments"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-segments"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:175
					__e = (doors.ALink{
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
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:188
							__e = __c.Set("id", "active-ignore-all"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-ignore-all"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:189
					__e = (doors.ALink{
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
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:206
							__e = __c.Set("id", "active-query"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-query"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:207
					__e = (doors.ALink{
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
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:220
							__e = __c.Set("id", "active-only-ignore-some"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-ignore-some"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:221
					__e = (doors.ALink{
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
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:233
							__e = __c.Set("id", "active-only-some"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-some"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:234
					__e = (doors.ALink{
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
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:246
							__e = __c.Set("id", "active-only-if-present"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-if-present"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:247
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Fragment: "details",
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherOnlyIgnoreAll(),
						FragmentMatch: true,
						Indicator: doors.IndicatorOnlyClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:257
							__e = __c.Set("id", "active-fragment"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-fragment"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:259
					__e = __c.Set("id", "nav-links"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:260
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:264
							__e = __c.Set("id", "nav-home"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-home"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:265
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section", "child"},
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:269
							__e = __c.Set("id", "nav-starts"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-starts"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:270
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "other"},
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:274
							__e = __c.Set("id", "nav-segments"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-segments"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:275
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Fragment: "details",
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:280
							__e = __c.Set("id", "nav-fragment"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-fragment"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:281
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"page", "9",
						),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:289
							__e = __c.Set("id", "nav-query"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:290
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"optional", "yes",
							"page", "9",
						),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:299
							__e = __c.Set("id", "nav-query-optional"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query-optional"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:300
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"optional", "no",
							"page", "9",
						),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:309
							__e = __c.Set("id", "nav-query-optional-miss"); if __e != nil { return }
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
//line page.gox:313
}

//line page.gox:315
func pageEscaped(b doors.Source[PathEscaped]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:318
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:318
						__e = __c.Set("id", "name-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:319
						__e = __c.Any(ctx.Value(0).(PathEscaped).Name); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:322
				name := "next value/again"

//line page.gox:324
				__e = (doors.ALink{
				Model: PathEscaped{
					Path: true,
					Name: name,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:329
						__e = __c.Set("id", "next-escaped"); if __e != nil { return }
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
//line page.gox:332
}

//line page.gox:334
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
//line page.gox:337
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:337
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:338
				__e = (doors.ALink{
				Model: PathCrossB{
					Path: true,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:342
						__e = __c.Set("id", "cross-next"); if __e != nil { return }
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
//line page.gox:345
}

//line page.gox:347
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
//line page.gox:350
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:350
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:351
					__e = __c.Set("id", "page-name"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("cross-b"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:354
}

//line page.gox:356
func pageSlow() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:359
				__e = __c.Any(gox.EditorFunc(func(cur gox.Cursor) error {
				<-time.After(1100 * time.Millisecond)
				return nil
			})); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:363
					__e = __c.Set("id", "slow-page"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("slow-page"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:366
}

//line page.gox:368
func pageError(err error) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:370
			__e = __c.Any(doors.Status(500)); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:372
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("error"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:373
					__e = __c.Set("id", "error-message"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:373
					__e = __c.Any(err.Error()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:376
}

//line page.gox:378
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
//line page.gox:381
			if code >= 0 {
//line page.gox:382
				__e = __c.Any(doors.Status(code)); if __e != nil { return }
			}
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:385
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:385
					__e = __c.Any(path); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:388
}

type PathC struct {
	PathC1 bool `path:"/c1"`
	PathC2 bool `path:"/c2"`
}

//line page.gox:395
func pageC(b doors.Source[PathC]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:398
				__e = (doors.Inject(0, b)).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line page.gox:398
					if ctx.Value(0).(PathC).PathC1 {
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:399
							__e = __c.Set("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c1"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					} else  {
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:401
							__e = __c.Set("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c2"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
				return })); if __e != nil { return }
//line page.gox:404
				__e = (doors.ALink{
				Model: PathC{
					PathC1: true,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c1"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:410
				__e = (doors.ALink{
				Model: PathC{
					PathC2: true,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:414
						__e = __c.Set("id", "c2"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c2"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:416
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReplace{Model: PathC{PathC2: true}})
					return true
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:421
						__e = __c.Set("id", "replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:422
					__e = __c.Set("id", "marker"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:422
					__e = __c.Any(doors.IDRand()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:424
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReload{})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:429
						__e = __c.Set("id", "reload"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("reload"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:431
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationAssign(PathB{}))
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:436
						__e = __c.Set("id", "assign_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("assign_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:438
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationReplace(PathB{}))
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:443
						__e = __c.Set("id", "replace_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:445
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationReload())
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:450
						__e = __c.Set("id", "reload_after"); if __e != nil { return }
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
//line page.gox:453
}
