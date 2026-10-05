// Managed by GoX v0.3.2

//line page.gox:1
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

//line page.gox:31
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
//line page.gox:34
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("A"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:35
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationAssign{Model: PathC{PathC1: true}})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:40
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
//line page.gox:43
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

//line page.gox:72
func pageParallel() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:75
				__e = (doors.Parallel()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line page.gox:77
						<-time.After(500 * time.Millisecond)

						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:79
							__e = __c.Set("id", "part-a"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("part-a"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:81
				__e = (doors.Parallel()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitContainer(); if __e != nil { return }
					{
//line page.gox:83
						<-time.After(500 * time.Millisecond)

						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:85
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
//line page.gox:89
					<-time.After(500 * time.Millisecond)

					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:91
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
//line page.gox:95
}

func values(items ...string) url.Values {
	v := url.Values{}
	for i := 0; i + 1 < len(items); i += 2 {
		v.Add(items[i], items[i + 1])
	}
	return v
}

//line page.gox:105
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
//line page.gox:108
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:108
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:109
				__e = __c.Any(b.Bind(func(path PathQuery) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:110
						__e = __c.Set("id", "tag"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:111
						if path.Tag != nil {
//line page.gox:112
							__e = __c.Any(*path.Tag); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:115
						__e = __c.Set("id", "page-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:116
						if path.Page != nil {
//line page.gox:117
							__e = __c.Any(fmt.Sprint(*path.Page)); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:120
			})); if __e != nil { return }
//line page.gox:123
				tag := "next"
			page := 2

//line page.gox:126
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
//line page.gox:132
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
//line page.gox:135
}

//line page.gox:137
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
//line page.gox:140
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:140
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:141
				__e = __c.Any(b.Bind(func(location doors.Location) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:142
						__e = __c.Set("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:143
						__e = __c.Any(location.String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:145
						__e = __c.Set("id", "location-path"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:146
						__e = __c.Any(location.Path()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:148
						__e = __c.Set("id", "tag-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:149
						__e = __c.Any(location.Query.Get("tag")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:151
						__e = __c.Set("id", "page-query-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:152
						__e = __c.Any(location.Query.Get("page")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:154
			})); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:157
}

//line page.gox:159
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
//line page.gox:162
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:162
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:163
				__e = __c.Any(b.Bind(func(location doors.Location) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:164
						__e = __c.Set("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:165
						__e = __c.Any(location.String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:167
			})); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Set("hidden", true); if __e != nil { return }
//line page.gox:168
					__e = __c.Set("id", "active-links"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:169
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Active: doors.Active{
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:176
							__e = __c.Set("id", "active-full"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-full"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:177
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section"},
					},
					Active: doors.Active{
						PathMatcher: doors.PathMatcherStarts(),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:185
							__e = __c.Set("id", "active-starts"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-starts"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:186
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section", "fixed"},
					},
					Active: doors.Active{
						PathMatcher: doors.PathMatcherSegments(0),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:194
							__e = __c.Set("id", "active-segments"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-segments"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:195
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
						),
					},
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherIgnoreAll(),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:206
							__e = __c.Set("id", "active-ignore-all"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-ignore-all"); if __e != nil { return }
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
						QueryMatcher: doors.QueryMatcherIgnoreSome("page").And(doors.QueryMatcherSome("mode")).And(doors.QueryMatcherIfPresent("optional")),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:220
							__e = __c.Set("id", "active-query"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-query"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:221
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
						QueryMatcher: doors.QueryMatcherIgnoreSome("page"),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:234
							__e = __c.Set("id", "active-only-ignore-some"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-ignore-some"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:235
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"mode", "view",
							"page", "1",
						),
					},
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherSome("mode").And(doors.QueryMatcherIgnoreAll()),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:247
							__e = __c.Set("id", "active-only-some"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-some"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:248
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
						Query: values(
							"optional", "yes",
							"page", "1",
						),
					},
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherIfPresent("optional").And(doors.QueryMatcherIgnoreAll()),
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:260
							__e = __c.Set("id", "active-only-if-present"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("active-only-if-present"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:261
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Fragment: "details",
					Active: doors.Active{
						QueryMatcher: doors.QueryMatcherIgnoreAll(),
						FragmentMatch: true,
						Indicator: doors.IndicateClass("active"),
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:271
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
//line page.gox:273
					__e = __c.Set("id", "nav-links"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:274
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:278
							__e = __c.Set("id", "nav-home"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-home"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:279
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "section", "child"},
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:283
							__e = __c.Set("id", "nav-starts"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-starts"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:284
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active", "other"},
					},
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:288
							__e = __c.Set("id", "nav-segments"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-segments"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:289
					__e = (doors.ALink{
					Model: doors.Location{
						Segments: []string{"active"},
					},
					Fragment: "details",
				}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("a"); if __e != nil { return }
						{
//line page.gox:294
							__e = __c.Set("id", "nav-fragment"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-fragment"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:295
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
//line page.gox:303
							__e = __c.Set("id", "nav-query"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:304
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
//line page.gox:313
							__e = __c.Set("id", "nav-query-optional"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("nav-query-optional"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
//line page.gox:314
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
//line page.gox:323
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
//line page.gox:327
}

//line page.gox:329
func pageEscaped(b doors.Source[PathEscaped]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:332
				__e = __c.Any(b.Bind(func(path PathEscaped) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:333
						__e = __c.Set("id", "name-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:334
						__e = __c.Any(path.Name); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:336
			})); if __e != nil { return }
//line page.gox:338
				name := "next value/again"

//line page.gox:340
				__e = (doors.ALink{
				Model: PathEscaped{
					Path: true,
					Name: name,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:345
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
//line page.gox:348
}

//line page.gox:350
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
//line page.gox:353
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:353
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:354
				__e = (doors.ALink{
				Model: PathCrossB{
					Path: true,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:358
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
//line page.gox:361
}

//line page.gox:363
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
//line page.gox:366
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:366
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:367
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
//line page.gox:370
}

//line page.gox:372
func pageSlow() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:375
				__e = __c.Any(gox.Elem(func(cur gox.Cursor) error {
				<-time.After(1100 * time.Millisecond)
				return nil
			})); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:379
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
//line page.gox:382
}

//line page.gox:384
func pageError(err error) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:386
			__e = __c.Any(doors.Status(500)); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:388
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("error"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:389
					__e = __c.Set("id", "error-message"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:389
					__e = __c.Any(err.Error()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:392
}

//line page.gox:394
func plainErrorPage(err error) gox.Elem {
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
//line page.gox:397
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("error"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:398
					__e = __c.Set("id", "error-message"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:398
					__e = __c.Any(err.Error()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:401
}

//line page.gox:403
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
//line page.gox:406
			if code >= 0 {
//line page.gox:407
				__e = __c.Any(doors.Status(code)); if __e != nil { return }
			}
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:410
					__e = __c.Set("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:410
					__e = __c.Any(path); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:413
}

type PathC struct {
	PathC1 bool `path:"/c1"`
	PathC2 bool `path:"/c2"`
}

//line page.gox:420
func pageC(b doors.Source[PathC]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:423
				__e = __c.Any(b.Bind(func(path PathC) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line page.gox:424
					if path.PathC1 {
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:425
							__e = __c.Set("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c1"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					} else  {
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:427
							__e = __c.Set("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c2"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
				return })
//line page.gox:429
			})); if __e != nil { return }
//line page.gox:431
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
//line page.gox:437
				__e = (doors.ALink{
				Model: PathC{
					PathC2: true,
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:441
						__e = __c.Set("id", "c2"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c2"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:443
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReplace{Model: PathC{PathC2: true}})
					return true
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:448
						__e = __c.Set("id", "replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:449
					__e = __c.Set("id", "marker"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:449
					__e = __c.Any(doors.IDRand()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:451
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReload{})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:456
						__e = __c.Set("id", "reload"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("reload"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:458
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionLocationAssign{Model: PathB{}})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:463
						__e = __c.Set("id", "assign_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("assign_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:465
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionLocationReplace{Model: PathB{}})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:470
						__e = __c.Set("id", "replace_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:472
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionLocationReload{})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:477
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
//line page.gox:480
}

//line page.gox:482
func routerBeamDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:485
				__e = __c.Any(doors.Route(
				doors.RouteModelBeam(beamCrossAContent),
				doors.RouteModelBeam(beamCrossBContent),
				doors.RouteDefaultComp[doors.Location](routeDefault404()),
			)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:492
}

//line page.gox:494
func beamCrossAContent(b doors.Beam[PathCrossA]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:495
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:495
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:496
		__e = __c.Any(b.Bind(func(path PathCrossA) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:497
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("cross-a"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:498
				__e = __c.Set("id", "route-model"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:498
				__e = __c.Any(fmt.Sprint(path.Path)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:499
	})); if __e != nil { return }
//line page.gox:500
		__e = (doors.ALink{
		Model: PathCrossB{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:504
				__e = __c.Set("id", "beam-cross-next"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("beam-cross-next"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:505
}

//line page.gox:507
func beamCrossBContent(b doors.Beam[PathCrossB]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:508
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:508
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:509
		__e = __c.Any(b.Bind(func(path PathCrossB) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:510
				__e = __c.Set("id", "page-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("cross-b"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:511
				__e = __c.Set("id", "route-model"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:511
				__e = __c.Any(fmt.Sprint(path.Path)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:512
	})); if __e != nil { return }
//line page.gox:513
		__e = (doors.ALink{
		Model: PathCrossA{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:517
				__e = __c.Set("id", "beam-cross-prev"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("beam-cross-prev"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:518
}

//line page.gox:520
func routeDefault404() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:521
		__e = __c.Any(doors.Status(404)); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:522
			__e = __c.Set("id", "route-name"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("default"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:523
}

//line page.gox:525
func routerLensCrossDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:528
				__e = __c.Any(doors.Route(
				doors.RouteModel(crossAContent),
				doors.RouteModel(crossBContent),
				doors.RouteDefaultComp[doors.Location](routeDefault404()),
			)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:535
}

//line page.gox:537
func crossAContent(l doors.Source[PathCrossA]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:538
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:538
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:539
		__e = (doors.ALink{
		Model: PathCrossB{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:543
				__e = __c.Set("id", "cross-next"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("cross-next"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:544
}

//line page.gox:546
func crossBContent(l doors.Source[PathCrossB]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:547
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:547
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:548
			__e = __c.Set("id", "page-name"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("cross-b"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:549
}

type CustomRoute struct {
	ID string
	Tab string
}

func (r CustomRoute) Encode() (doors.Location, error) {
	query := url.Values{}
	if r.Tab != "" {
		query.Set("tab", r.Tab)
	}
	return doors.Location{
		Segments: []string{"custom", r.ID},
		Query: query,
	}, nil
}

func deriveCustomRoute(l doors.Location) (CustomRoute, bool) {
	if len(l.Segments) != 2 || l.Segments[0] != "custom" {
		return CustomRoute{}, false
	}
	return CustomRoute{
		ID: l.Segments[1],
		Tab: l.Query.Get("tab"),
	}, true
}

func setCustomRoute(l doors.Location, r CustomRoute) doors.Location {
	next, err := r.Encode()
	if err != nil {
		return l
	}
	return next
}

//line page.gox:585
func routerCombinedLensDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:588
				__e = __c.Any(doors.Route(
				doors.RouteModel(combinedAContent),
				doors.RouteDerive(deriveCustomRoute).Source(setCustomRoute, combinedCustomContent),
				doors.RouteModel(combinedQueryContent),
				doors.RouteMatch(func(l doors.Location) bool {
					return len(l.Segments) == 1 && l.Segments[0] == "raw"
				}).Source(combinedRawContent),
				doors.RouteDefaultComp[doors.Location](routeDefault404()),
			)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:599
}

//line page.gox:601
func combinedAContent(l doors.Source[PathCrossA]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:602
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:602
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:603
			__e = __c.Set("id", "route-name"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("model-a"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:604
		__e = (doors.ALink{
		Model: CustomRoute{
			ID: "hello world/one",
			Tab: "details",
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:609
				__e = __c.Set("id", "model-to-custom"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("model-to-custom"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:611
		tag := "from-model"
	page := 3

//line page.gox:614
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
//line page.gox:620
				__e = __c.Set("id", "model-to-query"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("model-to-query"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:621
		__e = (doors.ALink{
		Model: doors.Location{
			Segments: []string{"raw"},
			Query: values("from", "model"),
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:626
				__e = __c.Set("id", "model-to-raw"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("model-to-raw"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:627
}

//line page.gox:629
func combinedCustomContent(l doors.Source[CustomRoute]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:630
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:630
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:631
		__e = __c.Any(l.Bind(func(route CustomRoute) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:632
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:633
				__e = __c.Set("id", "custom-id"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:633
				__e = __c.Any(route.ID); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:634
				__e = __c.Set("id", "custom-tab"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:634
				__e = __c.Any(route.Tab); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:635
	})); if __e != nil { return }
//line page.gox:636
		__e = (doors.ALink{
		Model: CustomRoute{
			ID: "next/child",
			Tab: "again",
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:641
				__e = __c.Set("id", "custom-next"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-next"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:643
		tag := "from-custom"
	page := 5

//line page.gox:646
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
//line page.gox:652
				__e = __c.Set("id", "custom-to-query"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-to-query"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:653
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			l.Update(ctx, CustomRoute{
				ID: "lens write/value",
				Tab: "written",
			})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line page.gox:661
				__e = __c.Set("id", "custom-lens-write"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-lens-write"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:662
}

//line page.gox:664
func combinedQueryContent(l doors.Source[PathQuery]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:665
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:665
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:666
		__e = __c.Any(l.Bind(func(path PathQuery) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:667
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("query"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:668
				__e = __c.Set("id", "tag"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:669
				if path.Tag != nil {
//line page.gox:670
					__e = __c.Any(*path.Tag); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:673
				__e = __c.Set("id", "page-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:674
				if path.Page != nil {
//line page.gox:675
					__e = __c.Any(fmt.Sprint(*path.Page)); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:678
	})); if __e != nil { return }
//line page.gox:679
		__e = (doors.ALink{
		Model: doors.Location{
			Segments: []string{"raw"},
			Query: values("from", "query"),
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:684
				__e = __c.Set("id", "query-to-raw"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("query-to-raw"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:685
}

//line page.gox:687
func combinedRawContent(l doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:688
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:688
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:689
		__e = __c.Any(l.Bind(func(location doors.Location) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:690
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:691
				__e = __c.Set("id", "raw-from"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:691
				__e = __c.Any(location.Query.Get("from")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:692
	})); if __e != nil { return }
//line page.gox:693
		__e = (doors.ALink{
		Model: PathCrossA{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:697
				__e = __c.Set("id", "raw-to-model"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw-to-model"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:698
}

//line page.gox:700
func routerCombinedBeamDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:703
				__e = __c.Any(doors.Route(
				doors.RouteModelBeam(beamCrossAContent),
				doors.RouteDerive(deriveCustomRoute).Beam(combinedCustomBeamContent),
				doors.RouteModelBeam(combinedQueryBeamContent),
				doors.RouteDefaultComp[doors.Location](routeDefault404()),
			)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:711
}

//line page.gox:713
func combinedCustomBeamContent(b doors.Beam[CustomRoute]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:714
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:714
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:715
		__e = __c.Any(b.Bind(func(route CustomRoute) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:716
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-beam"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:717
				__e = __c.Set("id", "custom-id"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:717
				__e = __c.Any(route.ID); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:718
				__e = __c.Set("id", "custom-tab"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:718
				__e = __c.Any(route.Tab); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:719
	})); if __e != nil { return }
//line page.gox:720
		__e = (doors.ALink{
		Model: CustomRoute{
			ID: "beam next",
			Tab: "read",
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:725
				__e = __c.Set("id", "beam-custom-next"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("beam-custom-next"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:726
		__e = (doors.ALink{
		Model: PathCrossA{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:730
				__e = __c.Set("id", "beam-custom-to-model"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("beam-custom-to-model"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:731
}

//line page.gox:733
func combinedQueryBeamContent(b doors.Beam[PathQuery]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:734
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:734
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:735
		__e = __c.Any(b.Bind(func(path PathQuery) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:736
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("query-beam"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:737
				__e = __c.Set("id", "tag"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:738
				if path.Tag != nil {
//line page.gox:739
					__e = __c.Any(*path.Tag); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:742
	})); if __e != nil { return }
	return })
//line page.gox:743
}

//line page.gox:745
func routeBindDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:748
				__e = __c.Any(doors.Route(
				doors.RouteDerive(deriveCustomRoute).Bind(func(route CustomRoute) gox.Elem {
					return gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:750
							__e = __c.Set("id", "route-name"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("derive-bind"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:751
							__e = __c.Set("id", "custom-id"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line page.gox:751
							__e = __c.Any(route.ID); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:752
							__e = __c.Set("id", "custom-tab"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line page.gox:752
							__e = __c.Any(route.Tab); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })
//line page.gox:753
				}),
				doors.RouteMatch(func(l doors.Location) bool {
					return len(l.Segments) == 1 && l.Segments[0] == "cross-a"
				}).Bind(func(loc doors.Location) gox.Elem {
					return gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:757
							__e = __c.Set("id", "route-name"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("match-bind"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:758
							__e = __c.Set("id", "match-path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line page.gox:758
							__e = __c.Any(loc.Path()); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })
//line page.gox:759
				}),
				doors.RouteDefaultBind(func(loc doors.Location) gox.Elem {
					return gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
//line page.gox:761
						__e = __c.Any(doors.Status(404)); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:762
							__e = __c.Set("id", "route-name"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("default-bind"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:763
							__e = __c.Set("id", "default-path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line page.gox:763
							__e = __c.Any(loc.Path()); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })
//line page.gox:764
				}),
			)); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:766
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:766
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:767
				__e = (doors.ALink{
				Model: CustomRoute{
					ID: "bind-test",
					Tab: "active",
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:772
						__e = __c.Set("id", "to-derive"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-derive"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:773
				__e = (doors.ALink{
				Model: doors.Location{
					Segments: []string{"cross-a"},
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:777
						__e = __c.Set("id", "to-match"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-match"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:778
				__e = (doors.ALink{
				Model: doors.Location{
					Segments: []string{"unknown"},
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:782
						__e = __c.Set("id", "to-default"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-default"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:785
}

//line page.gox:787
func pageHistoryReplace(b doors.Source[doors.Location]) gox.Elem {
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
//line page.gox:790
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:790
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:791
				__e = __c.Any(b.Bind(func(l doors.Location) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:792
						__e = __c.Set("id", "path"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:792
						__e = __c.Any(l.String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:793
			})); if __e != nil { return }
//line page.gox:795
				locC1, _ := doors.NewLocation(PathC{PathC1: true})
			locC2, _ := doors.NewLocation(PathC{PathC2: true})

//line page.gox:798
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					b.Update(ctx, locC1)
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:803
						__e = __c.Set("id", "soft-assign"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("soft-assign"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:804
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					b.Update(doors.HistoryReplaceContext(ctx), locC2)
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:809
						__e = __c.Set("id", "soft-replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("soft-replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:812
}

type ReplaceStep int

const (
	ReplaceA ReplaceStep = iota
	ReplaceB
	ReplaceC
)

type PathReplace struct {
	Step ReplaceStep `/:"rep-a | rep-b | rep-c"`
}

//line page.gox:826
func pageReplaceModel(s doors.Source[PathReplace]) gox.Elem {
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
//line page.gox:829
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:829
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:830
				__e = __c.Any(s.Bind(func(p PathReplace) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:831
						__e = __c.Set("id", "step"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:831
						__e = __c.Any(fmt.Sprint(int(p.Step))); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:832
			})); if __e != nil { return }
//line page.gox:833
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					s.Update(ctx, PathReplace{Step: ReplaceB})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:838
						__e = __c.Set("id", "pm-push"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("pm-push"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:839
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					s.Update(doors.HistoryReplaceContext(ctx), PathReplace{Step: ReplaceC})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:844
						__e = __c.Set("id", "pm-replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("pm-replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:845
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					s.Mutate(doors.HistoryReplaceContext(ctx), func(p PathReplace) PathReplace {
						p.Step = ReplaceC
						return p
					})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:853
						__e = __c.Set("id", "pm-replace-mutate"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("pm-replace-mutate"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:856
}

type LinkStep int

const (
	LinkA LinkStep = iota
	LinkB
	LinkC
)

type PathLink struct {
	Step LinkStep `/:"la | lb | lc"`
}

//line page.gox:870
func pageLinkNav(s doors.Source[PathLink]) gox.Elem {
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
//line page.gox:873
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:873
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:874
				__e = __c.Any(s.Bind(func(p PathLink) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:875
						__e = __c.Set("id", "step"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:875
						__e = __c.Any(fmt.Sprint(int(p.Step))); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:876
			})); if __e != nil { return }
//line page.gox:877
				__e = (doors.ALink{
				Model: PathLink{Step: LinkB},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:879
						__e = __c.Set("id", "push-b"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("push-b"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:880
				__e = (doors.ALink{
				Model: PathLink{Step: LinkC},
				HistoryReplace: true,
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:883
						__e = __c.Set("id", "replace-c"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace-c"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:884
				__e = (doors.ALink{
				Model: PathLink{Step: LinkA},
				HistoryReplace: true,
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:887
						__e = __c.Set("id", "replace-self"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace-self"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:888
				__e = (doors.ALink{
				Model: PathLink{Step: LinkA},
				Fragment: "sec",
				HistoryReplace: true,
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:892
						__e = __c.Set("id", "replace-hash"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace-hash"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:893
					__e = __c.Set("id", "sec"); if __e != nil { return }
//line page.gox:893
					__e = __c.Set("style", "margin-top: 3000px"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("sec"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:896
}

//line page.gox:898
func pageLastSeen(b doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:900
		result := doors.NewSource("")

		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:904
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:904
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:905
				__e = __c.Any(result.Bind(func(s string) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:906
						__e = __c.Set("id", "result"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:906
						__e = __c.Any(s); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:907
			})); if __e != nil { return }
//line page.gox:908
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					now := time.Now()
					inst := doors.InstanceLastSeen(ctx)
					sess := doors.SessionLastSeen(ctx)
					ok := !inst.IsZero() && !sess.IsZero() && !inst.After(now) && !sess.After(now) && now.Sub(inst) < time.Minute && now.Sub(sess) < time.Minute
					if ok {
						result.Update(ctx, "ok")
					} else {
						result.Update(ctx, fmt.Sprintf("bad inst=%v sess=%v", inst, sess))
					}
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:921
						__e = __c.Set("id", "check"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("check"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:924
}
