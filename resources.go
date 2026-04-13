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

	"github.com/doors-dev/doors/internal/printer"
)

// Resource describes content or a URL that can be attached to HTML resource
// attrs such as `src` and `href`.
//
// Managed resources may be hosted by Doors when the tag and attrs require it.
type Resource = printer.SourceHandler

// ResourceStatic is a [Resource] whose content is known up front.
//
// Static resources can be shared through cached public URLs and mounted at
// fixed public routes such as [RouteResource].
type ResourceStatic = printer.SourceStatic

// ResourceExternal is a direct external URL.
//
// Use it when the browser should load the resource from another host without
// proxying it through Doors, while still letting Doors collect the host for
// Content-Security-Policy generation.
type ResourceExternal = printer.SourceExternal

// ResourceFS serves one file from fsys as a [ResourceStatic].
func ResourceFS(fsys fs.FS, entry string) ResourceStatic {
	return printer.SourceFS{
		FS:    fsys,
		Entry: entry,
	}
}

// ResourceLocalFS serves one local file from path as a [ResourceStatic].
func ResourceLocalFS(path string) ResourceStatic {
	return printer.SourceLocalFS(path)
}

// ResourceBytes serves in-memory bytes as a [ResourceStatic].
func ResourceBytes(content []byte) ResourceStatic {
	return printer.SourceBytes(content)
}

// ResourceString serves in-memory string content as a [ResourceStatic].
func ResourceString(content string) ResourceStatic {
	return printer.SourceString(content)
}

// ResourceHook serves content through a custom resource handler.
//
// Use it for dynamic or request-dependent content that should be exposed
// through a managed Doors resource URL.
//
// Returning true tells Doors it may remove the generated private resource after
// the request.
func ResourceHook(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool) Resource {
	return printer.SourceHook(handler)
}

// ResourceHandler serves content through a standard library-style handler.
func ResourceHandler(handler func(w http.ResponseWriter, r *http.Request)) Resource {
	return printer.SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		handler(w, r)
		return false
	})
}

// ResourceProxy serves a resource by reverse-proxying requests to url.
//
// Unlike [ResourceExternal], the browser still loads the resource from a
// Doors-managed URL.
func ResourceProxy(url string) Resource {
	return printer.SourceProxy(url)
}
