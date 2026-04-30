package app

import (
	"net/http"
	"slices"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
	"github.com/evanw/esbuild/pkg/api"
)

type Middleware = func(http.Handler) http.Handler

type Page = instance.Page

type App = *app

func NewApp(page Page, o Options) App {
	o.initDefaults()
	a := &app{
		page:       page,
		conf:       o.Conf,
		csp:        o.CSP,
		pathMaker:  path.NewPathMaker(o.ID),
		tracker:    o.SessionTracker,
		esProfiles: o.ESBuild,
		errPage:    o.ErrorPage,
	}
	a.registry = resources.NewRegistry(a)
	a.Use()
	return a
}

type ErrorPage = func(r *http.Request, err error) gox.Elem

type app struct {
	page       Page
	conf       common.Conf
	csp        *common.CSP
	registry   resources.Registry
	pathMaker  path.PathMaker
	tracker    SessionTracker
	esProfiles func(profile string) api.BuildOptions
	sessions   sync.Map
	use        []Middleware
	handler    http.Handler
	errPage    ErrorPage
}

func (a *app) ResourceRegistry() resources.Registry {
	return a.registry
}

func (a *app) Use(m ...Middleware) {
	a.use = append(a.use, m...)
	a.handler = http.HandlerFunc(a.serveInstance)
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
	a.sessions.Delete(id)
	a.tracker.Delete(id)
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
