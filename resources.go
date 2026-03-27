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

type Resource = printer.Source
type ResourceStatic = printer.SourceStatic

func ResourceFS(fsys fs.FS, entry string) ResourceStatic {
	return printer.SourceFS{
		FS:    fsys,
		Entry: entry,
	}
}

func ResourceLocalFS(path string) ResourceStatic {
	return printer.SourceLocalFS(path)
}

func ResourceExternal(path string) Resource {
	return printer.SourceExternal(path)
}

func ResourceBytes(content []byte) ResourceStatic {
	return printer.SourceBytes(content)
}

func ResourceString(content string) ResourceStatic {
	return printer.SourceString(content)
}

func ResourceHook(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool) Resource {
	return printer.SourceHook(handler)
}

func ResourceHandler(handler func(w http.ResponseWriter, r *http.Request)) Resource {
	return printer.SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		handler(w, r)
		return false
	})
}

func ResourceProxy(url string) Resource {
	return printer.SourceProxy(url)
}
