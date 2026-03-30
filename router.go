// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/router"
	"github.com/doors-dev/doors/internal/router/model"
	"github.com/doors-dev/gox"
)

// Router serves a Doors application over HTTP.
//
// Use [NewRouter] to create one, register page models with [UseModel], and
// then pass it to an HTTP server.
type Router interface {
	http.Handler
	// Count returns the current number of live sessions and live page
	// instances.
	Count() (sessions int, instances int)
	Use(r Use)
}

// NewRouter returns a router with the default Doors configuration.
//
// Example:
//
//	router := doors.NewRouter()
//	doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
//		return doors.ResponseComp(Page(s))
//	})
func NewRouter() Router {
	return router.NewRouter()
}

// Use configures a [Router].
type Use = router.Use

// Response describes how a matched page model should be handled.
//
// Use [ResponseComp] to render a page, [ResponseRedirect] to send an HTTP
// redirect, or [ResponseReroute] to hand the request to another registered
// model without redirecting the browser.
type Response = model.Res

// ResponseComp returns a [Response] that renders comp for the matched model.
func ResponseComp(comp gox.Comp) Response {
	return model.ResComp(comp)
}

// ResponseRedirect returns a [Response] that redirects to model.
//
// status may be 0 to let Doors choose its default redirect status.
func ResponseRedirect(m any, status int) Response {
	return model.ResRedirect(m, status)
}

// ResponseReroute returns a [Response] that resolves another registered model
// on the server without changing the current HTTP response into a redirect.
func ResponseReroute(m any) Response {
	return model.ResReroute(m)
}

// UseModel registers a page model and its handler on r.
//
// The model type M defines matching and URL generation through `path` and
// `query` struct tags. The handler receives the request metadata and the
// current route as a [Source].
//
// Example:
//
//	type BlogPath struct {
//		Home bool    `path:"/"`
//		Post bool    `path:"/posts/:ID"`
//		ID   int
//		Tag  *string `query:"tag"`
//	}
//
//	doors.UseModel(router, func(r doors.RequestModel, s doors.Source[BlogPath]) doors.Response {
//		return doors.ResponseComp(Page(s))
//	})
func UseModel[M any](r Router, handler func(r RequestModel, s Source[M]) Response) {
	r.Use(router.UseModel(func(w http.ResponseWriter, r *http.Request, source beam.Source[M], store ctex.Store) model.Res {
		req := modelRequest{
			request: request{
				r: r,
				w: w,
			},
			store: store,
		}
		return handler(&req, source)
	}))
}

// Route handles a non-page GET endpoint inside a [Router].
//
// Use [UseRoute] for endpoints such as health checks, stable file mounts, or
// other GET handlers that should live beside model-based pages.
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

// RouteResource serves one static [ResourceStatic] at a fixed public path.
//
// Use it when you want a stable URL but still want Doors to prepare the
// resource through its registry.
type RouteResource struct {
	// URL path at which the file is served.
	// Required.
	Path     string
	Resource ResourceStatic
	// ContentType header
	ContentType string
}

func (rt RouteResource) Match(r *http.Request) bool {
	if rt.Path == "/" || rt.Path == "" {
		slog.Error("ResourceFile cannot serve root path")
		return false
	}
	if !strings.HasPrefix(rt.Path, "/") {
		rt.Path = "/" + rt.Path
	}
	return r.URL.Path == rt.Path
}

func (rt RouteResource) Serve(w http.ResponseWriter, r *http.Request) {
	if rt.Resource == nil {
		slog.Error("RouteResource requires a static resource")
		w.WriteHeader(500)
		return
	}
	entry := rt.Resource.StaticEntry()
	if entry == nil {
		slog.Error("Static resource returned nil static entry")
		w.WriteHeader(500)
		return
	}
	rr := r.Context().Value(ctex.KeyRouter).(*router.Router)
	res, err := rr.ResourceRegistry().Static(entry, rt.ContentType)
	if err != nil {
		slog.Error("Resource preparing error", "error", err.Error())
		w.WriteHeader(500)
		return
	}
	res.Serve(w, r)
}

// RouteFS serves files from an [fs.FS] under Prefix.
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

// RouteDir serves files from a local directory under Prefix.
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

// RouteFile serves one local file at Path.
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

// UseRoute adds rt to r before model-based page routing.
func UseRoute(r Router, rt Route) {
	r.Use(router.UseRoute(rt))
}

// UseFallback sends unmatched requests to handler.
//
// This is useful when Doors is mounted inside a larger HTTP server.
func UseFallback(r Router, handler http.Handler) {
	r.Use(router.UseFallback(handler))
}

// SessionCallback observes session creation and removal.
type SessionCallback = router.SessionCallback

// UseSessionCallback registers session lifecycle callbacks on r.
func UseSessionCallback(r Router, callback SessionCallback) {
	r.Use(router.UseSessionCallback(callback))
}

// UseESConf configures esbuild profiles used by script and style imports.
func UseESConf(r Router, conf ESConf) {
	r.Use(router.UseESConf(conf))
}

// SystemConf configures router, instance, and transport behavior.
type SystemConf = common.SystemConf

// UseSystemConf applies conf to r after filling in Doors defaults.
func UseSystemConf(r Router, conf SystemConf) {
	r.Use(router.UseSystemConf(conf))
}

// UseErrorPage renders page when Doors hits an internal routing, instance, or
// rendering error.
func UseErrorPage(r Router, page func(l Location, err error) gox.Comp) {
	r.Use(router.UseErrorPage(page))
}

// CSP configures the Content-Security-Policy header generated by Doors.
type CSP = common.CSP

// UseCSP configures the Content-Security-Policy header for r.
func UseCSP(r Router, csp CSP) {
	r.Use(router.UseCSP(&csp))
}

// UseLicense stores a license string on the managed Doors client script.
// When omitted, the client prints the AGPL notice.
// You can purchase suitable license at https://doors.dev
// or via email sales@doors.dev
func UseLicense(r Router, license string) {
	r.Use(router.UseLicense(license))
}

// UseLicence is an alias of UseLicense.
func UseLicence(r Router, license string) {
	UseLicense(r, license)
}

// UseServerID sets the stable server identifier used in Doors-generated paths
// and cookies.
//
// id must already be URL-safe.
func UseServerID(r Router, id string) {
	r.Use(router.UseServerID(id))
}
