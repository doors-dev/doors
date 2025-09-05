// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/router"
)

// Router represents the main HTTP router that handles all requests.
// It implements http.Handler and provides configuration through Use().
type Router interface {
	http.Handler
	Use(...Mod)
}

// NewRouter creates a new router instance with default configuration.
// The router handles page routing, static files, hooks, and framework resources.
func NewRouter() Router {
	return router.NewRouter()
}

// Mod represents a router modification that can be used to configure routing behavior.
type Mod = router.Mod

// UsePage registers a page handler for a specific path model type.
// The model type M defines URL patterns through struct field tags, allowing the router
// to decode request URIs into structured data. Path patterns are declared using `path` tags
// on boolean fields, with parameter capture using `:FieldName` syntax.
//
// Example:
//
//	type BlogPath struct {
//	    Home bool   `path:"/"`                    // Match root path
//	    Post bool   `path:"/post/:ID"`           // Match /post/123, capture ID
//	    List bool   `path:"/posts"`              // Match /posts
//	    ID   int                                  // Captured from :ID parameter
//	    Tag  string `query:"tag"`               // Query parameter ?tag=golang
//	}
//
//	router.Use(UsePage(func(p PageRouter[BlogPath], r RPage[BlogPath]) PageRoute {
//	   return p.Page(&blog{})
//	}))
func UsePage[M any](handler func(p PageRouter[M], r RPage[M]) PageRoute) Mod {
	return router.RoutePage(func(r *router.Request[M]) router.Response {
		pr := &pageRequest[M]{
			r: r,
		}
		return handler(pr, pr)
	})
}

// Deprecated: use UsePage
func ServePage[M any](handler func(p PageRouter[M], r RPage[M]) PageRoute) Mod {
	return UsePage(handler)
}

// UseFS serves static files from an embedded filesystem at the specified URL prefix.
// This is useful for serving embedded assets using Go's embed.FS.
//
// Parameters:
//   - prefix: URL prefix (e.g., "/assets/")
//   - fs: Filesystem to serve from (typically embed.FS)
func UseFS(prefix string, fs fs.FS) Mod {
	httpFS := http.FS(fs)
	return router.ServeDir(prefix, httpFS)
}

// Deprecated: use UseFS
func ServeFS(prefix string, fs fs.FS) Mod {
	return UseFS(prefix, fs)
}

// UseDir serves static files from a local directory using os.DirFS.
// This creates a filesystem from the directory and serves it at the prefix.
//
// Parameters:
//   - prefix: URL prefix (e.g., "/public/")
//   - path: Local directory path
func UseDir(prefix string, path string) Mod {
	fs := os.DirFS(path)
	httpFS := http.FS(fs)
	return router.ServeDir(prefix, httpFS)
}

// Deprecated: use UseDir
func ServeDir(prefix string, path string) Mod {
	return UseDir(prefix, path)
}

// UseFile serves a single file at the specified URL path.
//
// Parameters:
//   - path: URL path (e.g., "/favicon.ico")
//   - localPath: Local file path
func UseFile(path string, localPath string) Mod {
	return router.ServeFile(path, localPath)
}

// Deprecated: use UseFile
func ServeFile(path string, localPath string) Mod {
	return UseFile(path, localPath)
}

// UseFallback sets a fallback handler for requests that don't match any routes.
// This is useful for integrating with other HTTP handlers or serving custom 404 pages.
func UseFallback(handler http.Handler) Mod {
	return router.ServeFallback(handler)
}

// Deprecated: use UseFallback
func ServeFallback(handler http.Handler) Mod {
	return UseFallback(handler)
}

type SessionCallback = router.SessionCallback

// UseSessionCallback registers callbacks for session lifecycle events.
// The Create callback is called when a new session is created.
// The Delete callback is called when a session is removed.
func UseSessionCallback(callback SessionCallback) Mod {
	return router.SetSessionCallback(callback)
}

// UseESConf configures esbuild profiles for JavaScript/TypeScript processing.
// Different profiles can be used for development vs production builds.
func UseESConf(conf ESConf) Mod {
	return router.SetBuildProfiles(conf)
}

// Deprecated: use UseESConf
func SetESConf(conf ESConf) Mod {
	return UseESConf(conf)
}

// SystemConf contains system-wide configuration options for the framework.
type SystemConf = common.SystemConf

// UseSystemConf applies system-wide configuration including timeouts,
// limits, and other framework behavior settings.
func UseSystemConf(conf SystemConf) Mod {
	return router.SetSystemConf(conf)
}

// Deprecated: use UseSystemConf
func SetSystemConf(conf SystemConf) Mod {
	return UseSystemConf(conf)
}

// UseErrorPage sets a custom error page component for handling internal errors.
// The component receives the error message as a parameter.
func UseErrorPage(page func(message string) templ.Component) Mod {
	return router.SetErrorPage(page)
}

// Deprecated: use SetErrorPage
func SetErrorPage(page func(message string) templ.Component) Mod {
	return UseErrorPage(page)
}

// CSP represents Content Security Policy configuration.
type CSP = common.CSP

// UseCSP configures Content Security Policy headers for enhanced security.
// This helps prevent XSS attacks and other security vulnerabilities.
func UseCSP(csp CSP) Mod {
	return router.SetCSP(&csp)
}

// Deprecated: use UseCSP
func EnableCSP(csp CSP) Mod {
	return UseCSP(csp)
}
