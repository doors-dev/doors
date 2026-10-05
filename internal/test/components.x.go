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

package test

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

//line components.gox:25
func Report(value string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:26
		__e = __c.Any(ReportId(0, value)); if __e != nil { return }
	return })
//line components.gox:27
}

//line components.gox:29
func ReportId(id int, value string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:30
			__e = __c.Set("id", fmt.Sprintf("report-%d", id)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line components.gox:30
			__e = __c.Any(value); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:31
}

//line components.gox:33
func Marker(id string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:34
			__e = __c.Set("id", id); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:35
}

type page interface {
	h1() string
	content() gox.Elem
	head() gox.Elem
}

//line components.gox:43
func Document(p page) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Raw("<!DOCTYPE html>"); if __e != nil { return }
		__e = __c.Init("html"); if __e != nil { return }
		{
//line components.gox:45
			__e = __c.Set("lang", "en"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("head"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
//line components.gox:47
					__e = __c.Set("charset", "UTF-8"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
//line components.gox:48
					__e = __c.Set("name", "viewport"); if __e != nil { return }
//line components.gox:48
					__e = __c.Set("content", "width=device-width, initial-scale=1.0"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("link"); if __e != nil { return }
				{
//line components.gox:50
					__e = __c.Set("rel", "icon"); if __e != nil { return }
//line components.gox:51
					__e = __c.Set("type", "image/png"); if __e != nil { return }
//line components.gox:52
					__e = __c.Set("href", doors.ResourceBytes([]byte{})); if __e != nil { return }
//line components.gox:53
					__e = __c.Set("name", "favicon.png"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:54
				__e = __c.Any(p.head()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("body"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("h1"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line components.gox:57
					__e = __c.Any(p.h1()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line components.gox:58
				__e = __c.Any(p.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:61
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

//line components.gox:81
func (r *Reporter) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:82
		for _, report := range r.reports {
//line components.gox:83
			__e = __c.Any(report); if __e != nil { return }
		}
	return })
//line components.gox:85
}

//line components.gox:87
func Button(id string, handler func(context.Context) bool) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:88
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return handler(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line components.gox:92
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line components.gox:92
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line components.gox:93
}
