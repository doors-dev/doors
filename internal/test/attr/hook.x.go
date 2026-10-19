// Managed by GoX v0.3.3

//line hook.gox:1
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

package attr

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type nullHookFragment struct {
	test.NoBeam
	r *test.Reporter
}

type nullHookData struct {
	Value string
}

//line hook.gox:37
func (f *nullHookFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line hook.gox:38
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:39
			__e = __c.Set("id", "ptr"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:40
			__e = __c.Set("id", "str"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:41
			__e = __c.Set("id", "obj"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:42
			__e = __c.Set("id", "raw"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line hook.gox:43
		__e = (doors.AHook[*string]{
		Name: "ptrHook",
		On: func(ctx context.Context, r doors.RequestHook[*string]) (any, bool) {
			f.r.Update(ctx, 0, fmt.Sprint("called:", r.Data() == nil))
			return "ptr", true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line hook.gox:49
			__e = (doors.AHook[string]{
		Name: "strHook",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			f.r.Update(ctx, 1, fmt.Sprintf("called:%q", r.Data()))
			return "str", true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line hook.gox:55
				__e = (doors.AHook[nullHookData]{
		Name: "objHook",
		On: func(ctx context.Context, r doors.RequestHook[nullHookData]) (any, bool) {
			f.r.Update(ctx, 2, fmt.Sprintf("called:%q", r.Data().Value))
			return "obj", true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line hook.gox:61
					__e = (doors.ARawHook{
		Name: "rawHook",
		On: func(ctx context.Context, r doors.RequestRawHook) bool {
			body, _ := io.ReadAll(r.Body())
			f.r.Update(ctx, 3, string(body))
			fmt.Fprint(r.ResponseWriter(), `"raw"`)
			return true
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("script"); if __e != nil { return }
						{
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Raw("const run = async (id, name) => {\n\t\t\ttry {\n\t\t\t\tdocument.getElementById(id).innerHTML = `${await $hook(name)}`\n\t\t\t} catch (e) {\n\t\t\t\tdocument.getElementById(id).innerHTML = \"err\"\n\t\t\t}\n\t\t}\n\t\tawait run(\"ptr\", \"ptrHook\")\n\t\tawait run(\"str\", \"strHook\")\n\t\tawait run(\"obj\", \"objHook\")\n\t\tawait run(\"raw\", \"rawHook\")"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				return })); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line hook.gox:82
}

type trackHookFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line hook.gox:89
func (f *trackHookFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line hook.gox:91
		path, _ := doors.NewHook(ctx, doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s|%s|%s", r.URL.Path, r.URL.RawPath, r.URL.RawQuery)
	}))

//line hook.gox:95
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:96
			__e = __c.Set("id", "track"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:97
			__e = __c.Set("id", "bare"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:98
			__e = __c.Set("id", "slash"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:99
			__e = __c.Set("id", "sub"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line hook.gox:100
			__e = __c.Set("id", "query"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line hook.gox:101
		__e = (doors.AHook[string]{
		Name: "trackHook",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			f.r.Update(ctx, 0, r.RequestHeader().Get("D0-Track"))
			return r.Data(), false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("script"); if __e != nil { return }
			{
//line hook.gox:107
				__e = __c.Set("data:hookPath", path); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw("document.getElementById(\"track\").innerHTML = await $hook(\"trackHook\", \"first\")\n\t\tdocument.getElementById(\"track\").innerHTML = await $hook(\"trackHook\", \"second\")\n\t\tconst base = await $data(\"hookPath\")\n\t\tconst get = async (id, url) => {\n\t\t\tconst r = await fetch(url)\n\t\t\tdocument.getElementById(id).innerHTML = await r.text()\n\t\t}\n\t\tawait get(\"bare\", base)\n\t\tawait get(\"slash\", `${base}/`)\n\t\tawait get(\"sub\", `${base}/a/b%2Fc?q=1`)\n\t\tawait get(\"query\", `${base}?q=2`)"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line hook.gox:120
}
