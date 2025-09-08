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
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/router"
)

// Router represents the main HTTP router that handles all requests.
// It implements http.Handler and provides configuration through Use().
type Router interface {
	http.Handler
	Use(...Use)
}

// NewRouter creates a new router instance with default configuration.
// The router handles page routing, static files, hooks, and framework resources.
func NewRouter() Router {
	return router.NewRouter()
}

// Use represents a router modification that can be used to configure routing behavior.
type Use = router.Use

// UsePage registers a page handler for a path model type M.
// The model defines path/query patterns via struct tags.
//
// Example:
//
//	type BlogPath struct {
//	    Home bool   `path:"/"`                    // Match root path
//	    Post bool   `path:"/post/:ID"`           // Match /post/123, capture ID
//	    List bool   `path:"/posts"`              // Match /posts
//	    ID   int                                  // Captured from :ID parameter
//	    Tag  *string `query:"tag"`               // Query parameter ?tag=golang
//	}
//
//	router.Use(UsePage(func(p PageRouter[BlogPath], r RPage[BlogPath]) PageRoute {
//	   return p.Page(&blog{})
//	}))
func UsePage[M any](handler func(p PageRouter[M], r RPage[M]) PageRoute) Use {
	return router.UsePage(func(r *router.Request[M]) router.Response {
		pr := &pageRequest[M]{
			r: r,
		}
		return handler(pr, pr)
	})
}

type Route = router.Route

type responseWriter struct {
	w           http.ResponseWriter
	setHeaders  func(h http.Header)
	wroteHeader bool
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	if code >= 200 && code < 300 {
		w.setHeaders(w.w.Header())
	}
	w.w.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.setHeaders(w.w.Header())
	}
	return w.w.Write(b)
}

func normalizePrefix(prefix string) string {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return prefix
}
func serveFS(prefix string, fs http.FileSystem, cacheControl string, w http.ResponseWriter, r *http.Request) {
	rw := &responseWriter{
		w: w,
		setHeaders: func(h http.Header) {
			if cacheControl == "" {
				return
			}
			h.Set("Cache-Control", cacheControl)
		},
	}
	http.StripPrefix(normalizePrefix(prefix), http.FileServer(fs)).ServeHTTP(rw, r)
}

// RouteFS serves files from an fs.FS under a URL prefix.
// The prefix must not be root ("/").
type RouteFS struct {
	// URL prefix under which files are served.
	// Required.
	Prefix string
	// Filesystem to serve files from.
	// Required.
	FS fs.FS
	// Optional Cache-Control header applied to responses.
	// Optional.
	CacheControl string
}

func (rt RouteFS) Match(r *http.Request) bool {
	if rt.Prefix == "/" || rt.Prefix == "" {
		slog.Error("RouteFS can serve root prefix!")
		return false
	}
	return strings.HasPrefix(r.URL.Path, normalizePrefix(rt.Prefix))
}

func (rt RouteFS) Serve(w http.ResponseWriter, r *http.Request) {
	httpFS := http.FS(rt.FS)
	serveFS(rt.Prefix, httpFS, rt.CacheControl, w, r)
}

// RouteDir serves files from a local directory under a URL prefix.
// The prefix must not be root ("/").
type RouteDir struct {
	// URL prefix under which files are served.
	// Required.
	Prefix string
	// Filesystem directory path to serve.
	// Required.
	DirPath string
	// Cache-Control header applied to responses.
	// Optional.
	CacheControl string
}

func (rt RouteDir) Match(r *http.Request) bool {
	if rt.Prefix == "/" || rt.Prefix == "" {
		slog.Error("RouteDir cannot serve root prefix")
		return false
	}
	return strings.HasPrefix(r.URL.Path, normalizePrefix(rt.Prefix))
}
func (rt RouteDir) Serve(w http.ResponseWriter, r *http.Request) {
	httpFS := http.Dir(rt.DirPath)
	serveFS(rt.Prefix, httpFS, rt.CacheControl, w, r)
}

// RouteFile serves a single file at a fixed URL path.
// The path must not be root ("/").
type RouteFile struct {
	// URL path at which the file is served.
	// Required.
	Path string
	// Filesystem path to the file to serve.
	// Reuired.
	FilePath string
	// Cache-Control header applied to the response.
	// Optional.
	CacheControl string
}

func (rt RouteFile) Match(r *http.Request) bool {
	if rt.Path == "/" || rt.Path == "" {
		slog.Error("RouteFile cannot serve root path")
		return false
	}
	if !strings.HasPrefix(rt.Path, "/") {
		rt.Path = "/" + rt.Path
	}
	return r.URL.Path == rt.Path
}
func (rt RouteFile) Serve(w http.ResponseWriter, r *http.Request) {
	rw := &responseWriter{
		w: w,
		setHeaders: func(h http.Header) {
			if rt.CacheControl == "" {
				return
			}
			h.Set("Cache-Control", rt.CacheControl)
		},
	}
	http.ServeFile(rw, r, rt.FilePath)
}

// UseRoute adds a custom Route to the router.
// A Route must implement:
//   - Match(*http.Request) bool: whether the route handles the request
//   - Serve(http.ResponseWriter, *http.Request): serve the matched request
func UseRoute(r Route) Use {
	return router.UseRoute(r)
}

// UseFallback sets a fallback handler for requests that don't match any routes.
// This is useful for integrating with other HTTP handlers or serving custom 404 pages.
func UseFallback(handler http.Handler) Use {
	return router.UseFallback(handler)
}

type SessionCallback = router.SessionCallback

// UseSessionCallback registers callbacks for session lifecycle events.
// The Create callback is called when a new session is created.
// The Delete callback is called when a session is removed.
func UseSessionCallback(callback SessionCallback) Use {
	return router.UseSessionCallback(callback)
}

// UseESConf configures esbuild profiles for JavaScript/TypeScript processing.
// Different profiles can be used for development vs production builds.
func UseESConf(conf ESConf) Use {
	return router.UseESConf(conf)
}

// SystemConf contains system-wide configuration options for the framework.
type SystemConf = common.SystemConf

// UseSystemConf applies system-wide configuration including timeouts,
// limits, and other framework behavior settings.
func UseSystemConf(conf SystemConf) Use {
	return router.UseSystemConf(conf)
}

// UseErrorPage sets a custom error page component for handling internal errors.
// The component receives the error message as a parameter.
func UseErrorPage(page func(message string) templ.Component) Use {
	return router.UseErrorPage(page)
}

// CSP represents Content Security Policy configuration.
type CSP = common.CSP

// UseCSP configures Content Security Policy headers for enhanced security.
// This helps prevent XSS attacks and other security vulnerabilities.
func UseCSP(csp CSP) Use {
	return router.UseCSP(&csp)
}
