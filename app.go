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

// With is a [NewApp] option.
type With interface {
	apply(*app.Options)
}

type withFunc func(*app.Options)

func (f withFunc) apply(o *app.Options) {
	f(o)
}

// Conf tunes session and page instance lifetimes, request limits, resource
// serving, and the server-to-browser update stream.
//
// Every field is optional: zero or invalid timeouts, limits, and sizes take the
// defaults documented on each field, and the flags default to off.
type Conf = common.Conf

// WithConf sets the runtime configuration. Zero and invalid fields take their
// documented defaults.
func WithConf(conf Conf) With {
	return withFunc(func(o *app.Options) {
		o.Conf = conf
	})
}

// CSP is the Content-Security-Policy Doors sends on page responses.
//
// A nil slice and an empty non-nil slice mean different things on several
// fields, as noted on each.
type CSP = common.CSP

// WithCSP makes Doors send a Content-Security-Policy header on page responses.
// Without it, no such header is sent.
func WithCSP(csp CSP) With {
	return withFunc(func(o *app.Options) {
		o.CSP = &csp
	})
}

// WithID sets the app id used in Doors-generated URLs and in the session cookie
// name. It must survive URL path escaping unchanged; otherwise WithID panics.
// Default: doors.
func WithID(id string) With {
	if id != url.PathEscape(id) {
		panic("server ID must be URL compatible without escaping")
	}
	return withFunc(func(o *app.Options) {
		o.ID = id
	})
}

// WithIDCookie makes Doors set an extra cookie named name that carries the app
// id, for sticky load balancing. It is refreshed alongside the session cookie
// and not while draining. Empty name sets no cookie.
func WithIDCookie(name string) With {
	return withFunc(func(o *app.Options) {
		o.CookieName = name
	})
}

// WithESProfiles sets the esbuild options provider for managed scripts.
//
// profile receives the profile name requested by the script, or an empty string
// when it names none. A nil provider keeps the Doors defaults.
func WithESProfiles(profile func(p string) api.BuildOptions) With {
	return withFunc(func(o *app.Options) {
		o.ESBuild = profile
	})
}

// SessionTracker observes session creation and removal.
//
// id is the Doors session id, which is also the value of the Doors session
// cookie. Both methods run inline on the goroutine that triggers the change and
// must not block. With several trackers installed, they run one after another
// in registration order.
type SessionTracker = app.SessionTracker

// WithSessionTracker installs a session lifecycle observer. Repeat it to
// install several; they run in registration order. A nil tracker installs
// nothing.
func WithSessionTracker(t SessionTracker) With {
	return withFunc(func(o *app.Options) {
		if t == nil {
			return
		}
		o.SessionTrackers = append(o.SessionTrackers, t)
	})
}

// ErrorPage renders the body of an app-level error response.
//
// It runs when Doors cannot serve a page, with the failed request and the
// error. Doors has already written status 500, so the returned element cannot
// change it.
type ErrorPage = app.ErrorPage

// WithErrorPage installs the renderer for app-level errors. Without it, Doors
// sends a plain text 500 response.
func WithErrorPage(ep ErrorPage) With {
	return withFunc(func(o *app.Options) {
		o.ErrorPage = ep
	})
}

// WithLogger sets the logger Doors writes to. If l is nil, Doors uses
// slog.Default().
func WithLogger(l *slog.Logger) With {
	return withFunc(func(o *app.Options) {
		o.Logger = l
	})
}

// WithPrinter wraps the printer that serializes HTML output.
//
// p is called once per page render and once per Door render cycle, and must
// return a non-nil printer. A nil p disables wrapping.
func WithPrinter(p func(next gox.Printer) gox.Printer) With {
	return withFunc(func(o *app.Options) {
		o.PrinterMiddleware = p
	})
}

// NewApp returns an [App] that serves page as its page factory.
//
// page runs once per page instance, on the full page load that creates it. ctx
// is the render context of that instance and r carries the HTTP request.
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

// Use is HTTP middleware for [App.Use].
type Use = func(http.Handler) http.Handler

// App is a Doors application and HTTP handler.
type App interface {
	// Use wraps the app handler in middleware. The first one registered is
	// the outermost.
	Use(middleware ...func(http.Handler) http.Handler)
	// InstanceCount returns the number of live page instances across all
	// sessions.
	InstanceCount() int
	// SessionCount returns the number of live sessions.
	SessionCount() int
	// Drain switches the app into drain mode: the app id cookie stops being
	// refreshed and in-app navigation ends the instance, so the browser
	// reloads. callback runs at most once, when the last live instance is
	// gone or immediately if none are live. Drain is one-way for the
	// lifetime of the app.
	Drain(callback func())
	http.Handler
}
