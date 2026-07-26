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

package doors

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/printer"
)

// Resource is app-owned content that Doors serves under a generated src or
// href URL.
//
// Assign it to a src or href attribute, or attach it as a modifier to a tag
// that takes one of them; a modifier on any other tag fails the render.
//
// Unless the resource is a [ResourceStatic] served from a cache, the generated
// URL is private to the current page instance and stops working once the
// content around it is cleared.
type Resource = printer.SourceHandler

// ResourceStatic is a [Resource] whose content Doors reads up front.
//
// Unlike other resources, it accepts the cache, private, and nocache
// attributes. With cache, Doors keeps the content in RAM and serves it from a
// shared, publicly reachable URL derived from its hash. It can also be mounted
// at a fixed path with [UseResource].
type ResourceStatic = printer.SourceStatic

// ResourceExternal is a URL the browser loads directly from another host.
//
// Doors adds the host to the generated Content-Security-Policy only for script
// tags and stylesheet or modulepreload links.
type ResourceExternal = printer.SourceExternal

// ResourceFS returns a [ResourceStatic] that serves entry from fsys.
func ResourceFS(fsys fs.FS, entry string) ResourceStatic {
	return printer.SourceFS{
		FS:    fsys,
		Entry: entry,
	}
}

// ResourceLocalFS returns a [ResourceStatic] that serves the local file at
// path.
func ResourceLocalFS(path string) ResourceStatic {
	return printer.SourceLocalFS(path)
}

// ResourceBytes returns a [ResourceStatic] that serves content from memory.
func ResourceBytes(content []byte) ResourceStatic {
	return printer.SourceBytes(content)
}

// ResourceString returns a [ResourceStatic] that serves content from memory.
func ResourceString(content string) ResourceStatic {
	return printer.SourceString(content)
}

// ResourceHook returns a [Resource] served by handler.
//
// Return true from handler to stop serving the resource after that request,
// false to keep it available.
//
// WARNING: the request body is not limited by the ServerRequestBodyLimit of
// [Conf].
func ResourceHook(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool) Resource {
	return printer.SourceHook(handler)
}

// ResourceHandler returns a [Resource] served by handler.
//
// Unlike [ResourceHook], the resource keeps serving every request.
//
// WARNING: the request body is not limited by the ServerRequestBodyLimit of
// [Conf].
func ResourceHandler(handler func(w http.ResponseWriter, r *http.Request)) Resource {
	return printer.SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		handler(w, r)
		return false
	})
}

// ResourceProxy returns a [Resource] that reverse-proxies requests to url.
//
// Unlike [ResourceExternal], the browser loads it from a Doors-generated URL.
// If url cannot be parsed, requests fail with status 500.
func ResourceProxy(url string) Resource {
	return printer.SourceProxy(url)
}

// NewHook registers r on the closest dynamic parent and returns the URL path
// that serves it. The path stops working once the current content is cleared.
//
// Appending a slash and a file name to the returned path is safe.
//
// It returns false if the closest dynamic parent has already been cleared.
//
// ctx must belong to a Doors render or handler; otherwise NewHook panics.
//
// WARNING: the request body is not limited by the ServerRequestBodyLimit of
// [Conf].
func NewHook(ctx context.Context, r Resource) (string, bool) {
	core := ctx.Value(common.KeyCore).(core.Core)
	hook, ok := core.Door().RegisterHook(r.Handler(), nil)
	if !ok {
		return "", false
	}
	return core.App().PathMaker().Hook(core.Instance().ID(), hook.HookID, ""), true
}
