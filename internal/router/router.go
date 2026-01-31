// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package router

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

func NewRouter() (router *Router) {
	conf := &common.SystemConf{}
	common.InitDefaults(conf)
	router = &Router{
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

type ErrorPageComponent = func(message string) gox.Elem

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

func (rr *Router) ResourceRegistry() *resources.Registry {
	return rr.registry
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
		Value:    s.ID(),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	}
	rr.sessions.Store(s.ID(), s)
	rr.sessionCallback.Create(s.ID(), r.Header)
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
