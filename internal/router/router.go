// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/router/model"
	"github.com/doors-dev/gox"
)

func NewRouter() (router *Router) {
	conf := &common.SystemConf{}
	common.InitDefaults(conf)
	router = &Router{
		pathMaker:       path.NewPathMaker(""),
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

type ErrorPageComponent = func(location path.Location, err error) gox.Comp

type sessionHooks struct{}

func (s sessionHooks) Create(string, http.Header) {}

func (s sessionHooks) Delete(string) {}

type Router struct {
	pathMaker       path.PathMaker
	sessions        sync.Map
	modelAdapters   path.Adapters
	modelRoutes     []model.AnyModelRoute
	routes          []Route
	fallback        http.Handler
	errPage         ErrorPageComponent
	sessionCallback SessionCallback
	registry        *resources.Registry
	csp             *common.CSP
	conf            *common.SystemConf
	buildProfiles   resources.BuildProfiles
}

func (rr *Router) Count() (int, int) {
	sessions := 0
	instances := 0
	rr.sessions.Range(func(_, v any) bool {
		sess := v.(*instance.Session)
		sessions += 1
		instances += sess.InstanceCount()
		return true
	})
	return sessions, instances
}

func (rr *Router) SessionCookie() string {
	return "d0r-" + rr.pathMaker.ID()
}

func (rr *Router) PathMaker() path.PathMaker {
	return rr.pathMaker
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
func (rr *Router) Adapters() path.Adapters {
	return rr.modelAdapters
}

func (rr *Router) BuildProfiles() resources.BuildProfiles {
	return rr.buildProfiles
}

func (rr *Router) Conf() *common.SystemConf {
	return rr.conf
}

func (rr *Router) ensureSession(w http.ResponseWriter, r *http.Request) *instance.Session {
	s := rr.getSession(w, r)
	if s != nil {
		return s
	}
	s = instance.NewSession(rr)
	rr.sessions.Store(s.ID(), s)
	rr.sessionCallback.Create(s.ID(), r.Header)
	s.Renew(w)
	return s
}

func (rr *Router) getSession(w http.ResponseWriter, r *http.Request) *instance.Session {
	c, err := r.Cookie(rr.SessionCookie())
	if err != nil {
		return nil
	}
	v, ok := rr.sessions.Load(c.Value)
	if !ok {
		return nil
	}
	sess := v.(*instance.Session)
	if !sess.Renew(w) {
		return nil
	}
	return sess
}

func (rr *Router) addModelRoute(modelRoute model.AnyModelRoute) {
	adapter := modelRoute.Adapter()
	rr.modelRoutes = append(rr.modelRoutes, modelRoute)
	rr.modelAdapters.Add(adapter)
}

func (rr *Router) addRoute(r Route) {
	rr.routes = append(rr.routes, r)
}

func (rr *Router) Use(use Use) {
	use.apply(rr)
}
