// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

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
				doors.RouteBeamModel[PathCrossA](doors.RouteBeam(beamCrossAContent)),
				doors.RouteBeamModel[PathCrossB](doors.RouteBeam(beamCrossBContent)),
				doors.RouteComp[doors.Location](routeDefault404()),
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
				doors.RouteModel[PathCrossA](doors.RouteSource(crossAContent)),
				doors.RouteModel[PathCrossB](doors.RouteSource(crossBContent)),
				doors.RouteComp[doors.Location](routeDefault404()),
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

func matchCustomRoute(l doors.Location) bool {
	return len(l.Segments) == 2 && l.Segments[0] == "custom"
}

func getCustomRoute(l doors.Location) CustomRoute {
	if !matchCustomRoute(l) {
		return CustomRoute{}
	}
	return CustomRoute{
		ID: l.Segments[1],
		Tab: l.Query.Get("tab"),
	}
}

func setCustomRoute(l doors.Location, r CustomRoute) doors.Location {
	next, err := r.Encode()
	if err != nil {
		return l
	}
	return next
}

func customRouteSource(s doors.Source[doors.Location]) doors.Source[CustomRoute] {
	return doors.DeriveSource(s, getCustomRoute, setCustomRoute)
}

func customRouteBeam(b doors.Beam[doors.Location]) doors.Beam[CustomRoute] {
	return doors.DeriveBeam(b, getCustomRoute)
}

//line page.gox:597
func routerCombinedLensDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:600
				__e = __c.Any(doors.Route(
				doors.RouteModel[PathCrossA](doors.RouteSource(combinedAContent)),
				doors.RouteMatch(matchCustomRoute, doors.RouteDerive(customRouteSource, doors.RouteSource(combinedCustomContent))),
				doors.RouteModel[PathQuery](doors.RouteSource(combinedQueryContent)),
				doors.RouteMatch(
					func(l doors.Location) bool {
						return len(l.Segments) == 1 && l.Segments[0] == "raw"
					},
					doors.RouteSource(combinedRawContent),
				),
				doors.RouteComp[doors.Location](routeDefault404()),
			)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:614
}

//line page.gox:616
func combinedAContent(l doors.Source[PathCrossA]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:617
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:617
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:618
			__e = __c.Set("id", "route-name"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("model-a"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:619
		__e = (doors.ALink{
		Model: CustomRoute{
			ID: "hello world/one",
			Tab: "details",
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:624
				__e = __c.Set("id", "model-to-custom"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("model-to-custom"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:626
		tag := "from-model"
	page := 3

//line page.gox:629
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
//line page.gox:635
				__e = __c.Set("id", "model-to-query"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("model-to-query"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:636
		__e = (doors.ALink{
		Model: doors.Location{
			Segments: []string{"raw"},
			Query: values("from", "model"),
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:641
				__e = __c.Set("id", "model-to-raw"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("model-to-raw"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:642
}

//line page.gox:644
func combinedCustomContent(l doors.Source[CustomRoute]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:645
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:645
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:646
		__e = __c.Any(l.Bind(func(route CustomRoute) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:647
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:648
				__e = __c.Set("id", "custom-id"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:648
				__e = __c.Any(route.ID); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:649
				__e = __c.Set("id", "custom-tab"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:649
				__e = __c.Any(route.Tab); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:650
	})); if __e != nil { return }
//line page.gox:651
		__e = (doors.ALink{
		Model: CustomRoute{
			ID: "next/child",
			Tab: "again",
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:656
				__e = __c.Set("id", "custom-next"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-next"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:658
		tag := "from-custom"
	page := 5

//line page.gox:661
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
//line page.gox:667
				__e = __c.Set("id", "custom-to-query"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-to-query"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:668
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
//line page.gox:676
				__e = __c.Set("id", "custom-lens-write"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-lens-write"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:677
}

//line page.gox:679
func combinedQueryContent(l doors.Source[PathQuery]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:680
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:680
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:681
		__e = __c.Any(l.Bind(func(path PathQuery) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:682
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("query"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:683
				__e = __c.Set("id", "tag"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:684
				if path.Tag != nil {
//line page.gox:685
					__e = __c.Any(*path.Tag); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:688
				__e = __c.Set("id", "page-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:689
				if path.Page != nil {
//line page.gox:690
					__e = __c.Any(fmt.Sprint(*path.Page)); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:693
	})); if __e != nil { return }
//line page.gox:694
		__e = (doors.ALink{
		Model: doors.Location{
			Segments: []string{"raw"},
			Query: values("from", "query"),
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:699
				__e = __c.Set("id", "query-to-raw"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("query-to-raw"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:700
}

//line page.gox:702
func combinedRawContent(l doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:703
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:703
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:704
		__e = __c.Any(l.Bind(func(location doors.Location) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:705
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:706
				__e = __c.Set("id", "raw-from"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:706
				__e = __c.Any(location.Query.Get("from")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:707
	})); if __e != nil { return }
//line page.gox:708
		__e = (doors.ALink{
		Model: PathCrossA{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:712
				__e = __c.Set("id", "raw-to-model"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("raw-to-model"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:713
}

//line page.gox:715
func routerCombinedBeamDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:718
				__e = __c.Any(doors.Route(
				doors.RouteBeamModel[PathCrossA](doors.RouteBeam(beamCrossAContent)),
				doors.RouteBeamMatch(matchCustomRoute, doors.RouteBeamDerive(customRouteBeam, doors.RouteBeam(combinedCustomBeamContent))),
				doors.RouteBeamModel[PathQuery](doors.RouteBeam(combinedQueryBeamContent)),
				doors.RouteComp[doors.Location](routeDefault404()),
			)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:726
}

//line page.gox:728
func combinedCustomBeamContent(b doors.Beam[CustomRoute]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:729
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:729
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:730
		__e = __c.Any(b.Bind(func(route CustomRoute) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:731
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("custom-beam"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:732
				__e = __c.Set("id", "custom-id"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:732
				__e = __c.Any(route.ID); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:733
				__e = __c.Set("id", "custom-tab"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:733
				__e = __c.Any(route.Tab); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:734
	})); if __e != nil { return }
//line page.gox:735
		__e = (doors.ALink{
		Model: CustomRoute{
			ID: "beam next",
			Tab: "read",
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:740
				__e = __c.Set("id", "beam-custom-next"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("beam-custom-next"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:741
		__e = (doors.ALink{
		Model: PathCrossA{
			Path: true,
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("a"); if __e != nil { return }
			{
//line page.gox:745
				__e = __c.Set("id", "beam-custom-to-model"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("beam-custom-to-model"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:746
}

//line page.gox:748
func combinedQueryBeamContent(b doors.Beam[PathQuery]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:749
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:749
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:750
		__e = __c.Any(b.Bind(func(path PathQuery) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:751
				__e = __c.Set("id", "route-name"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("query-beam"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:752
				__e = __c.Set("id", "tag"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:753
				if path.Tag != nil {
//line page.gox:754
					__e = __c.Any(*path.Tag); if __e != nil { return }
				}
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:757
	})); if __e != nil { return }
	return })
//line page.gox:758
}

//line page.gox:760
func routeBindDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:763
				__e = __c.Any(doors.Route(
				doors.RouteBeamMatch(
					matchCustomRoute,
					doors.RouteBeamDerive(customRouteBeam, doors.RouteBind(func(route CustomRoute) gox.Elem {
						return gox.Elem(func(__c gox.Cursor) (__e error) {
							ctx := __c.Context(); _ = ctx
							__e = __c.Init("div"); if __e != nil { return }
							{
//line page.gox:767
								__e = __c.Set("id", "route-name"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
								__e = __c.Text("derive-bind"); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
							__e = __c.Init("div"); if __e != nil { return }
							{
//line page.gox:768
								__e = __c.Set("id", "custom-id"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
//line page.gox:768
								__e = __c.Any(route.ID); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
							__e = __c.Init("div"); if __e != nil { return }
							{
//line page.gox:769
								__e = __c.Set("id", "custom-tab"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
//line page.gox:769
								__e = __c.Any(route.Tab); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
						return })
//line page.gox:770
					})),
				),
				doors.RouteMatch(
					func(l doors.Location) bool {
						return len(l.Segments) == 1 && l.Segments[0] == "cross-a"
					},
					doors.RouteBind(func(loc doors.Location) gox.Elem {
						return gox.Elem(func(__c gox.Cursor) (__e error) {
							ctx := __c.Context(); _ = ctx
							__e = __c.Init("div"); if __e != nil { return }
							{
//line page.gox:777
								__e = __c.Set("id", "route-name"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
								__e = __c.Text("match-bind"); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
							__e = __c.Init("div"); if __e != nil { return }
							{
//line page.gox:778
								__e = __c.Set("id", "match-path"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
//line page.gox:778
								__e = __c.Any(loc.Path()); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
						return })
//line page.gox:779
					}),
				),
				doors.RouteBind(func(loc doors.Location) gox.Elem {
					return gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
//line page.gox:782
						__e = __c.Any(doors.Status(404)); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:783
							__e = __c.Set("id", "route-name"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("default-bind"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:784
							__e = __c.Set("id", "default-path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line page.gox:784
							__e = __c.Any(loc.Path()); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })
//line page.gox:785
				}),
			)); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:787
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:787
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:788
				__e = (doors.ALink{
				Model: CustomRoute{
					ID: "bind-test",
					Tab: "active",
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:793
						__e = __c.Set("id", "to-derive"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-derive"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:794
				__e = (doors.ALink{
				Model: doors.Location{
					Segments: []string{"cross-a"},
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:798
						__e = __c.Set("id", "to-match"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-match"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:799
				__e = (doors.ALink{
				Model: doors.Location{
					Segments: []string{"unknown"},
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:803
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
//line page.gox:806
}

//line page.gox:808
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
//line page.gox:811
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:811
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:812
				__e = __c.Any(b.Bind(func(l doors.Location) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:813
						__e = __c.Set("id", "path"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:813
						__e = __c.Any(l.String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:814
			})); if __e != nil { return }
//line page.gox:816
				locC1, _ := doors.NewLocation(PathC{PathC1: true})
			locC2, _ := doors.NewLocation(PathC{PathC2: true})

//line page.gox:819
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					b.Update(ctx, locC1)
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:824
						__e = __c.Set("id", "soft-assign"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("soft-assign"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:825
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					b.Update(doors.HistoryReplaceContext(ctx), locC2)
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:830
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
//line page.gox:833
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

//line page.gox:847
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
//line page.gox:850
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:850
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:851
				__e = __c.Any(s.Bind(func(p PathReplace) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:852
						__e = __c.Set("id", "step"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:852
						__e = __c.Any(fmt.Sprint(int(p.Step))); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:853
			})); if __e != nil { return }
//line page.gox:854
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					s.Update(ctx, PathReplace{Step: ReplaceB})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:859
						__e = __c.Set("id", "pm-push"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("pm-push"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:860
				__e = (doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					s.Update(doors.HistoryReplaceContext(ctx), PathReplace{Step: ReplaceC})
					return false
				},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("button"); if __e != nil { return }
					{
//line page.gox:865
						__e = __c.Set("id", "pm-replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("pm-replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:866
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
//line page.gox:874
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
//line page.gox:877
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

//line page.gox:891
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
//line page.gox:894
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:894
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:895
				__e = __c.Any(s.Bind(func(p PathLink) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:896
						__e = __c.Set("id", "step"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:896
						__e = __c.Any(fmt.Sprint(int(p.Step))); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:897
			})); if __e != nil { return }
//line page.gox:898
				__e = (doors.ALink{
				Model: PathLink{Step: LinkB},
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:900
						__e = __c.Set("id", "push-b"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("push-b"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:901
				__e = (doors.ALink{
				Model: PathLink{Step: LinkC},
				HistoryReplace: true,
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:904
						__e = __c.Set("id", "replace-c"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace-c"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:905
				__e = (doors.ALink{
				Model: PathLink{Step: LinkA},
				HistoryReplace: true,
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:908
						__e = __c.Set("id", "replace-self"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace-self"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:909
				__e = (doors.ALink{
				Model: PathLink{Step: LinkA},
				Fragment: "sec",
				HistoryReplace: true,
			}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:913
						__e = __c.Set("id", "replace-hash"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace-hash"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:914
					__e = __c.Set("id", "sec"); if __e != nil { return }
//line page.gox:914
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
//line page.gox:917
}

//line page.gox:919
func pageLastSeen(b doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:921
		result := doors.NewSource("")

		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:925
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:925
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:926
				__e = __c.Any(result.Bind(func(s string) gox.Elem {
				return gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line page.gox:927
						__e = __c.Set("id", "result"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line page.gox:927
						__e = __c.Any(s); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })
//line page.gox:928
			})); if __e != nil { return }
//line page.gox:929
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
//line page.gox:942
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
//line page.gox:945
}

type NestSection int

const (
	NestHome NestSection = iota
	NestSettings
	NestEdit
)

type PathNest struct {
	Section NestSection `/nest:" | settings/:Tab | edit"`
	Tab string
	Q string `query:"q,omitempty"`
}

func nestSection(section NestSection) func(PathNest) bool {
	return func(p PathNest) bool {
		return p.Section == section
	}
}

func nestTab(s doors.Source[PathNest]) doors.Source[string] {
	return doors.DeriveSource(s, func(p PathNest) string {
		return p.Tab
	}, func(p PathNest, tab string) PathNest {
		p.Tab = tab
		return p
	})
}

//line page.gox:976
func routeNestDocument() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("main"); if __e != nil { return }
				{
//line page.gox:979
					__e = __c.Set("id", "routes"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:980
					__e = __c.Any(doors.Route(
					doors.RouteModel[PathNest](
						doors.RouteMatch(nestSection(NestHome), doors.RouteBind(func(p PathNest) gox.Elem {
							return gox.Elem(func(__c gox.Cursor) (__e error) {
								ctx := __c.Context(); _ = ctx
								__e = __c.Init("div"); if __e != nil { return }
								{
//line page.gox:983
									__e = __c.Set("id", "nest-name"); if __e != nil { return }
									__e = __c.Submit(); if __e != nil { return }
//line page.gox:983
									__e = __c.Many("home", p.Q); if __e != nil { return }
								}
								__e = __c.Close(); if __e != nil { return }
							return })
//line page.gox:984
						})),
						doors.RouteMatch(
							nestSection(NestSettings),
							doors.RouteDerive(
								nestTab,
								doors.RouteValue("profile", doors.RouteComp[string](gox.Elem(func(__c gox.Cursor) (__e error) {
									ctx := __c.Context(); _ = ctx
									__e = __c.Init("div"); if __e != nil { return }
									{
//line page.gox:989
										__e = __c.Set("id", "nest-name"); if __e != nil { return }
										__e = __c.Submit(); if __e != nil { return }
										__e = __c.Text("profile"); if __e != nil { return }
									}
									__e = __c.Close(); if __e != nil { return }
//line page.gox:989
								return }))),
								doors.RouteValue("team", doors.RouteComp[string](gox.Elem(func(__c gox.Cursor) (__e error) {
									ctx := __c.Context(); _ = ctx
									__e = __c.Init("div"); if __e != nil { return }
									{
//line page.gox:990
										__e = __c.Set("id", "nest-name"); if __e != nil { return }
										__e = __c.Submit(); if __e != nil { return }
										__e = __c.Text("team"); if __e != nil { return }
									}
									__e = __c.Close(); if __e != nil { return }
//line page.gox:990
								return }))),
							),
						),
						doors.RouteMatch(nestSection(NestEdit), doors.RouteSource(nestEdit)),
					),
					doors.RouteComp[doors.Location](gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("div"); if __e != nil { return }
						{
//line page.gox:995
							__e = __c.Set("id", "nest-name"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("not found"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
//line page.gox:995
					return })),
				)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
//line page.gox:998
					__e = __c.Set("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line page.gox:998
					__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line page.gox:999
				__e = (doors.ALink{Model: PathNest{Section: NestHome}}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:999
						__e = __c.Set("id", "to-home"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-home"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:1000
				__e = (doors.ALink{Model: PathNest{Section: NestHome, Q: "x"}}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:1000
						__e = __c.Set("id", "to-home-q"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-home-q"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:1001
				__e = (doors.ALink{Model: PathNest{Section: NestSettings, Tab: "profile"}}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:1001
						__e = __c.Set("id", "to-profile"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-profile"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:1002
				__e = (doors.ALink{Model: PathNest{Section: NestSettings, Tab: "team"}}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:1002
						__e = __c.Set("id", "to-team"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-team"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:1003
				__e = (doors.ALink{Model: PathNest{Section: NestSettings, Tab: "audit"}}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:1003
						__e = __c.Set("id", "to-audit"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-audit"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line page.gox:1004
				__e = (doors.ALink{Model: PathNest{Section: NestEdit}}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("a"); if __e != nil { return }
					{
//line page.gox:1004
						__e = __c.Set("id", "to-edit"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("to-edit"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line page.gox:1007
}

//line page.gox:1009
func nestEdit(s doors.Source[PathNest]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:1010
			__e = __c.Set("id", "nest-name"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("edit"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line page.gox:1011
			__e = __c.Set("id", "nest-input"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
//line page.gox:1012
		__e = __c.Any(s.Bind(func(p PathNest) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:1013
				__e = __c.Set("id", "nest-q"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:1013
				__e = __c.Any(p.Q); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:1014
	})); if __e != nil { return }
//line page.gox:1015
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			s.Mutate(ctx, func(p PathNest) PathNest {
				p.Q = "1"
				return p
			})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line page.gox:1023
				__e = __c.Set("id", "nest-write-q"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("nest-write-q"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line page.gox:1024
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			s.Update(ctx, PathNest{Section: NestSettings, Tab: "team"})
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line page.gox:1029
				__e = __c.Set("id", "nest-write-team"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("nest-write-team"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line page.gox:1030
}
