package app

import (
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
	"github.com/evanw/esbuild/pkg/api"
)

// Middleware wraps the app handler.
type Middleware = func(http.Handler) http.Handler

// Page is the page factory called once per page instance.
type Page = instance.Page

// App is a Doors application and HTTP handler.
type App = *app

// NewApp returns an App that serves page, configured by o.
func NewApp(page Page, o Options) App {
	o.initDefaults()
	a := &app{
		page:              page,
		conf:              o.Conf,
		csp:               o.CSP,
		pathMaker:         path.NewPathMaker(o.Conf.ServerSessionCookiePrefix, o.ID, o.CookieName),
		tracker:           trackers(o.SessionTrackers),
		esProfiles:        o.ESBuild,
		errPage:           o.ErrorPage,
		logger:            o.Logger,
		printerMiddleware: o.PrinterMiddleware,
	}
	a.registry = resources.NewRegistry(a)
	a.Use()
	return a
}

// ErrorPage renders the body of an app-level error response.
type ErrorPage = func(r *http.Request, err error) gox.Elem

type app struct {
	page              Page
	conf              common.Conf
	csp               *common.CSP
	registry          resources.Registry
	pathMaker         path.PathMaker
	tracker           SessionTracker
	esProfiles        func(profile string) api.BuildOptions
	sessions          sync.Map
	use               []Middleware
	handler           http.Handler
	errPage           ErrorPage
	logger            *slog.Logger
	printerMiddleware func(next gox.Printer) gox.Printer
	instanceCount     atomic.Int64
	drainCallback     atomic.Pointer[func()]
}

func (a *app) Logger() *slog.Logger {
	return a.logger
}

func (a *app) PrinterMiddleware() func(next gox.Printer) gox.Printer {
	return a.printerMiddleware
}

func (a *app) ResourceRegistry() resources.Registry {
	return a.registry
}

func (a *app) Use(m ...Middleware) {
	a.use = append(a.use, m...)
	a.handler = http.HandlerFunc(a.serve)
	for _, v := range slices.Backward(a.use) {
		a.handler = v(a.handler)
	}
}

func (a *app) ESProfile(profile string) api.BuildOptions {
	return a.esProfiles(profile)
}

func (a *app) Conf() *common.Conf {
	return &a.conf
}

func (a App) PathMaker() path.PathMaker {
	return a.pathMaker
}

func (a App) CSP() *common.CSP {
	return a.csp
}

func (a App) Resources() resources.Registry {
	return a.registry
}

func (a App) RemoveSession(id string) {
	_, ok := a.sessions.LoadAndDelete(id)
	if !ok {
		return
	}
	a.tracker.Delete(id)
}

func (a App) InstanceCount() int {
	return int(a.instanceCount.Load())
}

func (a App) SessionCount() (n int) {
	a.sessions.Range(func(_, _ any) bool {
		n++
		return true
	})
	return
}

func (a App) Draining() bool {
	return a.drainCallback.Load() != nil
}

func (a App) InstanceCreated() {
	a.instanceCount.Add(1)
}

func (a App) InstanceDeleted() {
	n := a.instanceCount.Add(-1)
	if n != 0 {
		return
	}
	callback := a.drainCallback.Load()
	if callback != nil {
		(*callback)()
	}
}

func (a App) Drain(callback func()) {
	once := sync.OnceFunc(callback)
	if !a.drainCallback.CompareAndSwap(nil, &once) {
		a.logger.Error("Drain called more than once")
		return
	}
	if a.instanceCount.Load() == 0 {
		once()
	}
}

func (a *app) ensureSession(w http.ResponseWriter, r *http.Request) instance.Session {
	s := a.getSession(w, r)
	if s != nil {
		return s
	}
	s = instance.NewSession(a)
	a.sessions.Store(s.ID(), s)
	a.tracker.Create(s.ID(), r)
	s.Renew(w)
	return s
}

func (a *app) getSession(w http.ResponseWriter, r *http.Request) instance.Session {
	c, err := r.Cookie(a.pathMaker.SessionCookie())
	if err != nil {
		return nil
	}
	v, ok := a.sessions.Load(c.Value)
	if !ok {
		return nil
	}
	sess := v.(instance.Session)
	if !sess.Renew(w) {
		return nil
	}
	return sess
}
