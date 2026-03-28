// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/doors-dev/doors/internal/printer"
)

// Resource describes content or a URL that can be attached to HTML resource
// attributes and, when needed, hosted by Doors.
type Resource = printer.Source

// ResourceStatic is a [Resource] that can also be mounted at a stable public
// path with [RouteResource].
type ResourceStatic = printer.SourceStatic

// ResourceFS serves one file from fsys.
func ResourceFS(fsys fs.FS, entry string) ResourceStatic {
	return printer.SourceFS{
		FS:    fsys,
		Entry: entry,
	}
}

// ResourceLocalFS serves one local file from path.
func ResourceLocalFS(path string) ResourceStatic {
	return printer.SourceLocalFS(path)
}

// ResourceExternal points at an already-hosted URL.
func ResourceExternal(path string) Resource {
	return printer.SourceExternal(path)
}

// ResourceBytes serves content from content.
func ResourceBytes(content []byte) ResourceStatic {
	return printer.SourceBytes(content)
}

// ResourceString serves content from content.
func ResourceString(content string) ResourceStatic {
	return printer.SourceString(content)
}

// ResourceHook serves content through a custom handler.
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

// ResourceProxy reverse-proxies requests to url.
func ResourceProxy(url string) Resource {
	return printer.SourceProxy(url)
}
