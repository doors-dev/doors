// Managed by GoX v0.1.28

//line components.gox:1
package test

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

//line components.gox:11
func Report(value string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:12
		__e = __c.Any(ReportId(0, value)); if __e != nil { return }
	return })
//line components.gox:13
}

//line components.gox:15
func ReportId(id int, value string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:16
			__e = __c.Set("id", fmt.Sprintf("report-%d", id)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line components.gox:16
			__e = __c.Any(value); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:17
}

//line components.gox:19
func Marker(id string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:20
			__e = __c.Set("id", id); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:21
}

type page interface {
	h1() string
	content() gox.Elem
	head() gox.Elem
}

//line components.gox:29
func Document(p page) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Raw("<!DOCTYPE html>"); if __e != nil { return }
		__e = __c.Init("html"); if __e != nil { return }
		{
//line components.gox:31
			__e = __c.Set("lang", "en"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("head"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
//line components.gox:33
					__e = __c.Set("charset", "UTF-8"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
//line components.gox:34
					__e = __c.Set("name", "viewport"); if __e != nil { return }
//line components.gox:34
					__e = __c.Set("content", "width=device-width, initial-scale=1.0"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("link"); if __e != nil { return }
				{
//line components.gox:36
					__e = __c.Set("rel", "icon"); if __e != nil { return }
//line components.gox:37
					__e = __c.Set("type", "image/png"); if __e != nil { return }
//line components.gox:38
					__e = __c.Set("href", doors.ResourceBytes([]byte{})); if __e != nil { return }
//line components.gox:39
					__e = __c.Set("name", "favicon.png"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:40
				__e = __c.Any(p.head()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("h1"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line components.gox:43
					__e = __c.Any(p.h1()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line components.gox:44
				__e = __c.Any(p.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:47
}

func NewReporter(size int) *Reporter {
	reports := make([]*doors.Door, size)
	for i := range size {
		reports[i] = &doors.Door{}
	}
	return &Reporter{
		reports: reports,
	}
}

type Reporter struct {
	reports []*doors.Door
}

func (r *Reporter) Update(ctx context.Context, i int, content string) {
	r.reports[i].Inner(ctx, ReportId(i, content))
}

//line components.gox:67
func (r *Reporter) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:68
		for _, report := range r.reports {
//line components.gox:69
			__e = __c.Any(report); if __e != nil { return }
		}
	return })
//line components.gox:71
}

//line components.gox:73
func Button(id string, handler func(context.Context) bool) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:74
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return handler(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:78
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:78
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line components.gox:79
}
