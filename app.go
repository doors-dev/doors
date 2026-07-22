package doors

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/doors-dev/doors/internal/app"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/gox"
	"github.com/evanw/esbuild/pkg/api"
)

// With configures [NewApp].
type With interface {
	apply(*app.Options)
}

type withFunc func(*app.Options)

func (f withFunc) apply(o *app.Options) {
	f(o)
}

// Conf is the Doors runtime configuration.
type Conf = common.Conf

// WithConf applies runtime configuration to an app.
func WithConf(conf Conf) With {
	return withFunc(func(o *app.Options) {
		o.Conf = conf
	})
}

// CSP is the Content-Security-Policy configuration.
type CSP = common.CSP

// WithCSP enables Content-Security-Policy header generation for an app.
func WithCSP(csp CSP) With {
	return withFunc(func(o *app.Options) {
		o.CSP = &csp
	})
}

// WithID sets the stable app id used for generated names and session cookies.
func WithID(id string) With {
	if id != url.PathEscape(id) {
		panic("server ID must be URL compatible without escaping")
	}
	return withFunc(func(o *app.Options) {
		o.ID = id
	})
}

// WithIDCookie sets the name of an additional cookie that carries the server ID
// for sticky session load balancing. When empty, no additional cookie is set.
func WithIDCookie(name string) With {
	return withFunc(func(o *app.Options) {
		o.CookieName = name
	})
}

// WithESProfiles sets the esbuild options provider used for script resources.
func WithESProfiles(profile func(p string) api.BuildOptions) With {
	return withFunc(func(o *app.Options) {
		o.ESBuild = profile
	})
}

// SessionTracker observes session creation and deletion.
type SessionTracker = app.SessionTracker

// WithSessionTracker installs a session lifecycle observer.
func WithSessionTracker(t SessionTracker) With {
	return withFunc(func(o *app.Options) {
		o.SessionTracker = t
	})
}

// ErrorPage renders an app-level error response.
type ErrorPage = app.ErrorPage

// WithErrorPage installs a custom app-level error page renderer.
func WithErrorPage(ep ErrorPage) With {
	return withFunc(func(o *app.Options) {
		o.ErrorPage = ep
	})
}

// WithLogger installs the logger used by Doors internals.
// If l is nil, Doors uses slog.Default().
func WithLogger(l *slog.Logger) With {
	return withFunc(func(o *app.Options) {
		o.Logger = l
	})
}

// WithPrinter installs the printer middleware into an app.
//
// The middleware applies to every page render and door render cycle of the
// app. A nil middleware disables wrapping.
func WithPrinter(p func(next gox.Printer) gox.Printer) With {
	return withFunc(func(o *app.Options) {
		o.PrinterMiddleware = p
	})
}

// NewApp creates a Doors HTTP handler from the root page function.
//
// The page function receives the Doors runtime context and request helpers, and
// returns the component to render for the current request.
func NewApp[C gox.Comp](page func(ctx context.Context, r Request) C, options ...With) App {
	os := app.Options{}
	for _, o := range options {
		o.apply(&os)
	}
	return app.NewApp(func(ctx context.Context, w http.ResponseWriter, r *http.Request) gox.Comp {
		req := &request{
			w: w,
			r: r,
		}
		return page(ctx, req)
	}, os)
}

// Use is HTTP middleware used by [App.Use].
type Use = func(http.Handler) http.Handler

// App is a Doors application and HTTP handler.
type App interface {
	// Use appends middleware around the app handler.
	Use(middleware ...func(http.Handler) http.Handler)
	// InstanceCount returns the number of live instances across all sessions.
	InstanceCount() int
	// SessionCount returns the number of active sessions.
	SessionCount() int
	// Drain switches the app into drain mode. Already-started instances are
	// hard-reloaded on the next navigation. The callback runs at most once,
	// fired when the final live instance cleans up (or immediately if none
	// are live). Drain is one-way for the lifetime of the app.
	Drain(callback func())
	http.Handler
}
