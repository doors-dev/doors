// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package router

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
)

func NewRouter() (router *Router) {
	conf := &common.SystemConf{}
	common.InitDefaults(conf)
	router = &Router{
		pool:            shredder.NewPool(conf.InstanceGoroutineLimit),
		sessions:        sync.Map{},
		modelAdapters:   make(map[string]path.AnyAdapter),
		modelRoutes:     make(map[string]anyModelRoute),
		fallback:        nil,
		conf:            conf,
		buildProfiles:   resources.BaseProfile{},
		sessionCallback: sessionHooks{},
	}
	router.registry = resources.NewRegistry(router)
	return
}

type Route interface {
	Match(r *http.Request) bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type ErrorPageComponent = func(message string) templ.Component

type sessionHooks struct{}

func (s sessionHooks) Create(string, http.Header) {}

func (s sessionHooks) Delete(string) {}

type Router struct {
	lisence         license.License
	sessions        sync.Map
	modelAdapters   map[string]path.AnyAdapter
	modelRoutes     map[string]anyModelRoute
	modelRouteList  []anyModelRoute
	routes          []Route
	fallback        http.Handler
	pool            *shredder.Pool
	errPage         ErrorPageComponent
	sessionCallback SessionCallback
	registry        *resources.Registry
	csp             *common.CSP
	conf            *common.SystemConf
	buildProfiles   resources.BuildProfiles
}

func (rr *Router) License() license.License {
	return rr.lisence
}

func (rr *Router) CSP() *common.CSP {
	return rr.csp
}

func (rr *Router) ImportRegistry() *resources.Registry {
	return rr.registry
}

func (rr *Router) Spawner(op shredder.OnPanic) *shredder.Spawner {
	return rr.pool.Spawner(op)
}

func (rr *Router) RemoveSession(id string) {
	rr.sessions.Delete(id)
	rr.sessionCallback.Delete(id)
}
func (rr *Router) Adapters() map[string]path.AnyAdapter {
	return rr.modelAdapters
}

func (rr *Router) BuildProfiles() resources.BuildProfiles {
	return rr.buildProfiles
}

func (rr *Router) Conf() *common.SystemConf {
	return rr.conf
}

func (rr *Router) ensureSession(r *http.Request, w http.ResponseWriter) (bool, *instance.Session) {
	s := rr.getSession(r)
	if s != nil {
		return false, s
	}
	s = instance.NewSession(rr)
	var expires time.Time
	if rr.conf.SessionTTL != 0 {
		expires = time.Now().Add(rr.conf.SessionTTL)
	}
	cookie := &http.Cookie{
		Name:     "d00r",
		Value:    s.Id(),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	}
	rr.sessions.Store(s.Id(), s)
	rr.sessionCallback.Create(s.Id(), r.Header)
	http.SetCookie(w, cookie)
	return true, s
}

func (rr *Router) getSession(r *http.Request) *instance.Session {
	c, err := r.Cookie("d00r")
	if err != nil {
		return nil
	}
	v, ok := rr.sessions.Load(c.Value)
	if !ok {
		return nil
	}
	return v.(*instance.Session)
}

func (rr *Router) addModelRoute(modelRoute anyModelRoute) {
	name := modelRoute.getName()
	_, has := rr.modelRoutes[name]
	if has {
		panic(errors.New("Can't register same model twice " + name))
	}
	rr.modelRoutes[name] = modelRoute
	rr.modelAdapters[name] = modelRoute.getAdapter()
	rr.modelRouteList = append(rr.modelRouteList, modelRoute)
}

func (rr *Router) addRoute(r Route) {
	rr.routes = append(rr.routes, r)
}

func (rr *Router) Use(use ...Use) {
	for _, r := range use {
		r.apply(rr)
	}
}
