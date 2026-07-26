package app

import (
	"log/slog"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
	"github.com/evanw/esbuild/pkg/api"
)

// SessionTracker observes session creation and removal.
type SessionTracker interface {
	// Create runs when a session is created, before the request that
	// triggered it is served. r must not be retained past the call and its
	// body must not be read.
	Create(id string, r *http.Request)

	// Delete runs at most once per session, when it is removed.
	Delete(id string)
}

// Options is the resolved set of app options.
type Options struct {
	// Conf is the runtime configuration from doors.WithConf.
	Conf common.Conf
	// CSP is the policy from doors.WithCSP. Nil sends no header.
	CSP *common.CSP
	// ESBuild is the esbuild options provider from doors.WithESProfiles.
	ESBuild func(profile string) api.BuildOptions
	// SessionTracker is the observer from doors.WithSessionTracker.
	SessionTracker SessionTracker
	// ID is the app id from doors.WithID. Default: doors.
	ID string
	// CookieName is the app id cookie name from doors.WithIDCookie. Empty
	// sets no cookie.
	CookieName string
	// ErrorPage is the error renderer from doors.WithErrorPage.
	ErrorPage ErrorPage
	// Logger is the logger from doors.WithLogger. Nil uses slog.Default().
	Logger *slog.Logger
	// PrinterMiddleware is the printer wrapper from doors.WithPrinter. Nil
	// disables wrapping.
	PrinterMiddleware func(next gox.Printer) gox.Printer
}

type notracker struct{}

func (n notracker) Create(id string, r *http.Request) {
}

func (n notracker) Delete(id string) {
}

var _ SessionTracker = notracker{}

func (o *Options) initDefaults() {
	common.InitDefaults(&o.Conf)
	if o.ESBuild == nil {
		o.ESBuild = resources.DefaultProfile
	}
	if o.ID == "" {
		o.ID = "doors"
	}
	if o.SessionTracker == nil {
		o.SessionTracker = notracker{}
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.PrinterMiddleware == nil {
		o.PrinterMiddleware = func(next gox.Printer) gox.Printer { return next }
	}
}
