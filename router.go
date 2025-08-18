package doors

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/router"
)

// Include returns a component that includes the framework's main client-side script and styles.
// This should be placed in the HTML head section and is required for the framework to function.
func Include() templ.Component {
	return front.Include
}

// Mod represents a router modification that can be applied to configure routing behavior.
type Mod = router.Mod

// Page defines the interface for page components that can be rendered with reactive data.
// Pages receive a SourceBeam containing the model data and return a templ.Component.
//
// Example:
//
//	type BlogPage struct {
//	    beam doors.SourceBeam[BlogPath]
//	}
//
//	func (p *BlogPage) Render(beam doors.SourceBeam[BlogPath]) templ.Component {
//	    p.beam = beam
//	    return common.Template(p)
//	}
//
//	func (p *BlogPage) Body() templ.Component {
//	    return doors.Sub(p.beam, func(path BlogPath) templ.Component {
//	        switch {
//	        case path.Home:
//	            return homePage()
//	        case path.Post:
//	            return postPage(path.ID)
//	        }
//	    })
//	}
type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}

// PageRoute represents a response that can be returned from page handlers.
// This includes page responses, redirects, reroutes, and static content.
type PageRoute = router.Response

// RPage provides access to request data and response control for page handlers.
// It combines basic request/response functionality with model access.
type RPage[M any] interface {
	R
	// Returns the decoded URL model
	GetModel() M
	// Access to incoming request headers
	RequestHeader() http.Header
	// Access to outgoing response headers
	ResponseHeader() http.Header
}

// PageRouter provides methods for creating different types of page responses.
// It allows rendering pages, redirecting, rerouting, and serving static content.
type PageRouter[M any] interface {
	// Serve page
	Page(page Page[M]) PageRoute
	// Serve page with custom status
	PageStatus(page Page[M], status int) PageRoute
	// Serve func page
	PageFunc(pageFunc func(SourceBeam[M]) templ.Component) PageRoute
	// Serve func page and custom status
	PageFuncStatus(pageFunc func(SourceBeam[M]) templ.Component, status int) PageRoute
	// Serve static page
	StaticPage(content templ.Component) PageRoute
	// Serve static page with custom status
	StaticPageStatus(content templ.Component, status int) PageRoute
	// Internal reroute to different model (detached=true disables path synchronization)
	Reroute(model any, detached bool) PageRoute
	// HTTP redirect to model URL
	Redirect(model any) PageRoute
	// HTTP redirect with custom status
	RedirectStatus(model any, status int) PageRoute
}

type pageRequest[M any] struct {
	r *router.Request[M]
}

func (r *pageRequest[M]) GetModel() M {
	return *r.r.Model
}

func (r *pageRequest[M]) RequestHeader() http.Header {
	return r.r.R.Header
}

func (r *pageRequest[M]) ResponseHeader() http.Header {
	return r.r.W.Header()
}

func (r *pageRequest[M]) GetCookie(name string) (*http.Cookie, error) {
	return r.r.R.Cookie(name)
}

func (r *pageRequest[M]) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.r.W, cookie)
}

func (r *pageRequest[M]) Reroute(model any, detached bool) PageRoute {
	return &router.RerouteResponse{
		Detached: detached,
		Model:    model,
	}
}

func (r *pageRequest[M]) Redirect(model any) PageRoute {
	return &router.RedirectResponse{
		Model: model,
	}
}

func (r *pageRequest[M]) RedirectStatus(model any, status int) PageRoute {
	return &router.RedirectResponse{
		Model:  model,
		Status: status,
	}
}

func (r *pageRequest[M]) Page(page Page[M]) PageRoute {
	return r.PageStatus(page, 200)
}

func (r *pageRequest[M]) PageStatus(page Page[M], status int) PageRoute {
	return &router.PageResponse[M]{
		Page:    page,
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
		Status:  status,
	}
}

func (r *pageRequest[M]) StaticPageStatus(content templ.Component, status int) PageRoute {
	return &router.StaticPage{
		Content: content,
		Status:  status,
	}
}

func (r *pageRequest[M]) StaticPage(content templ.Component) PageRoute {
	return r.StaticPageStatus(content, 200)
}

type pageFunc[M any] func(SourceBeam[M]) templ.Component

func (p pageFunc[M]) Render(b SourceBeam[M]) templ.Component {
	return p(b)
}

func (r *pageRequest[M]) PageFunc(f func(SourceBeam[M]) templ.Component) PageRoute {
	return r.PageFuncStatus(f, 200)
}

func (r *pageRequest[M]) PageFuncStatus(f func(SourceBeam[M]) templ.Component, status int) PageRoute {
	return r.PageStatus(pageFunc[M](f), status)
}

// ServePage registers a page handler for a specific path model type.
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
//	router.Use(ServePage(func(p PageRouter[BlogPath], r RPage[BlogPath]) PageRoute {
//	   return p.Page(&blog{})
//	}))
func ServePage[M any](handler func(p PageRouter[M], r RPage[M]) PageRoute) Mod {
	return router.RoutePage(func(r *router.Request[M]) router.Response {
		pr := &pageRequest[M]{
			r: r,
		}
		return handler(pr, pr)
	})
}

// ServeFS serves static files from an embedded filesystem at the specified URL prefix.
// This is useful for serving embedded assets using Go's embed.FS.
//
// Parameters:
//   - prefix: URL prefix (e.g., "/assets/")
//   - fs: Filesystem to serve from (typically embed.FS)
func ServeFS(prefix string, fs fs.FS) Mod {
	httpFS := http.FS(fs)
	return router.ServeDir(prefix, httpFS)
}

// ServeDir serves static files from a local directory using os.DirFS.
// This creates a filesystem from the directory and serves it at the prefix.
//
// Parameters:
//   - prefix: URL prefix (e.g., "/public/")
//   - path: Local directory path
func ServeDir(prefix string, path string) Mod {
	fs := os.DirFS(path)
	httpFS := http.FS(fs)
	return router.ServeDir(prefix, httpFS)
}

// ServeFile serves a single file at the specified URL path.
//
// Parameters:
//   - path: URL path (e.g., "/favicon.ico")
//   - localPath: Local file path
func ServeFile(path string, localPath string) Mod {
	return router.ServeFile(path, localPath)
}

// ServeFallback sets a fallback handler for requests that don't match any routes.
// This is useful for integrating with other HTTP handlers or serving custom 404 pages.
func ServeFallback(handler http.Handler) Mod {
	return router.ServeFallback(handler)
}

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

// SetGoroutineLimitPerInstance sets the maximum number of goroutines per instance.
// This controls resource usage for concurrent operations within each page instance.
func SetGoroutineLimitPerInstance(n int) Mod {
	return router.SetGoroutineLimit(n)
}

// SetSessionHooks registers callbacks for session lifecycle events.
// The create callback is called when a new session is created.
// The delete callback is called when a session is removed.
func SetSessionHooks(create func(id string), delete func(id string)) Mod {
	return router.SetSessionHooks(create, delete)
}

// SetESConf configures esbuild profiles for JavaScript/TypeScript processing.
// Different profiles can be used for development vs production builds.
func SetESConf(p ESConf) Mod {
	return router.SetBuildProfiles(p)
}

// SystemConf contains system-wide configuration options for the framework.
type SystemConf = common.SystemConf

// SetSystemConf applies system-wide configuration including timeouts,
// limits, and other framework behavior settings.
func SetSystemConf(conf SystemConf) Mod {
	return router.SetSystemConf(conf)
}

// SetErrorPage sets a custom error page component for handling internal errors.
// The component receives the error message as a parameter.
func SetErrorPage(page func(message string) templ.Component) Mod {
	return router.SetErrorPage(page)
}

// CSP represents Content Security Policy configuration.
type CSP = common.CSP

// SetCSP configures Content Security Policy headers for enhanced security.
// This helps prevent XSS attacks and other security vulnerabilities.
func SetCSP(csp *CSP) Mod {
	return router.SetCSP(csp)
}
