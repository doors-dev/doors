// Managed by GoX v0.1.17+dirty

package router

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

type PathA struct {
	Path bool `path:"/a"`
}

func pageA(b doors.Source[PathA]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("A"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationAssign{Model: PathC{PathC1: true}})
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("button"); if __e != nil { return }
					{
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

func pageQuery(b doors.Source[PathQuery]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "tag"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						if ctx.Value(0).(PathQuery).Tag != nil {
							__e = __c.Any(*ctx.Value(0).(PathQuery).Tag); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "page-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						if ctx.Value(0).(PathQuery).Page != nil {
							__e = __c.Any(fmt.Sprint(*ctx.Value(0).(PathQuery).Page)); if __e != nil { return }
						}
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				tag := "next"
				page := 2

				__e = doors.ALink{
				Model: PathQuery{
					Path: true,
					Tag: &tag,
					Page: &page,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("a"); if __e != nil { return }
					{
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
}

func pageLocation(b doors.Source[doors.Location]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "instance-id"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(doors.InstanceId(ctx)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "location-string"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Any(ctx.Value(0).(doors.Location).String()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "location-path"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Any(ctx.Value(0).(doors.Location).Path()); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "tag-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Any(ctx.Value(0).(doors.Location).Query.Get("tag")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "page-query-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Any(ctx.Value(0).(doors.Location).Query.Get("page")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func pageEscaped(b doors.Source[PathEscaped]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "name-value"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Any(ctx.Value(0).(PathEscaped).Name); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				name := "next value/again"

				__e = doors.ALink{
				Model: PathEscaped{
					Path: true,
					Name: name,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("a"); if __e != nil { return }
					{
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
}

func static(path string, code int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("head"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			if code >= 0 {
				__e = __c.Any(doors.Status(code)); if __e != nil { return }
			}
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "path"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(path); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

type PathC struct {
	PathC1 bool `path:"/c1"`
	PathC2 bool `path:"/c2"`
}

func pageC(b doors.Source[PathC]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("html"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = doors.Inject(0, b).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					if ctx.Value(0).(PathC).PathC1 {
						__e = __c.Init("div"); if __e != nil { return }
						{
							__e = __c.AttrSet("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c1"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					} else  {
						__e = __c.Init("div"); if __e != nil { return }
						{
							__e = __c.AttrSet("id", "path"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("c2"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
				return })); if __e != nil { return }
				__e = doors.ALink{
				Model: PathC{
					PathC1: true,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("a"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c1"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.ALink{
				Model: PathC{
					PathC2: true,
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("a"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "c2"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("c2"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReplace{Model: PathC{PathC2: true}})
					return true
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("button"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "replace"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "marker"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(doors.IDRand()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Call(ctx, doors.ActionLocationReload{})
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("button"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "reload"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("reload"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationAssign(PathB{}))
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("button"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "assign_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("assign_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationReplace(PathB{}))
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("button"); if __e != nil { return }
					{
						__e = __c.AttrSet("id", "replace_after"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Text("replace_after"); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = doors.AClick{
				On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
					r.After(doors.ActionOnlyLocationReload())
					return false
				},
			}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("button"); if __e != nil { return }
					{
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
}
