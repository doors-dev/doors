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
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/router/model"
)

func (rr *Router) serveHook(w http.ResponseWriter, r *http.Request, instanceID string, hookID uint64, track uint64) {
	ses := rr.getSession(w, r)
	if ses == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	inst, found := ses.GetInstance(instanceID)
	if !found {
		w.WriteHeader(http.StatusGone)
		return
	}
	found = inst.TriggerHook(hookID, w, r, track)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	}
}

const ZombieHeader = "Zombie"

func (rr *Router) restoreLocation(w http.ResponseWriter, r *http.Request, instId string, l path.Location) {
	w.Header().Set("Cache-Control", "no-cache")
	ses := rr.getSession(w, r)
	if ses == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	inst, ok := ses.GetInstance(instId)
	if !ok {
		w.WriteHeader(http.StatusGone)
		return
	}
	if w.Header().Get(ZombieHeader) != "" {
		inst.InstanceEnd()
		w.WriteHeader(http.StatusGone)
		return
	}
	ok = inst.RestorePath(l)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rr *Router) serveModelRedirect(w http.ResponseWriter, r *http.Request, s *instance.Session, l path.Location, model any, status int) {
	location, err := rr.modelAdapters.Encode(model)
	if err != nil {
		err := errors.New("adapter encoding failed: " + err.Error())
		slog.Error("page routing error", "path", r.URL.Path, "error", err)
		rr.serveError(w, r, s, l, err)
		return
	}
	if status == 0 {
		status = http.StatusFound
	}
	http.Redirect(w, r, location.String(), status)
}

func (rr *Router) serveInstance(w http.ResponseWriter, r *http.Request, s *instance.Session, l path.Location, inst instance.AnyInstance) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html")
	err := inst.Serve(w, r)
	if err != nil {
		slog.Error("instance serve error", "path", r.URL.Path, "error", err)
		rr.serveError(w, r, s, l, err)
	}
}

func (rr *Router) tryServePage(w http.ResponseWriter, r *http.Request) bool {
	ses := rr.ensureSession(w, r)
	loc, err := path.NewLocationFromURL(r.URL)
	if err != nil {
		rr.serveError(w, r, ses, loc, err)
		return true
	}
	var a any = loc
	var counter = 0
	opt := instance.Options{
		Rerouted: false,
	}
main:
	counter += 1
	if counter > 64 {
		slog.Error("page routing error", "path", r.URL.Path, "error", "routing loop detected")
		rr.serveError(w, r, ses, loc, errors.New("routing loop detected"))
		return true
	}
	for _, route := range rr.modelRoutes {
		res, handeled := route.Handle(w, r, a, ses, opt)
		if !handeled {
			continue
		}
		if res.Err() != nil {
			rr.serveError(w, r, ses, loc, res.Err())
			return true
		}
		if model, status, ok := res.Redirect(); ok {
			rr.serveModelRedirect(w, r, ses, loc, model, status)
			return true
		}
		if model, ok := res.Reroute(); ok {
			opt.Rerouted = true
			a = model
			goto main
		}
		inst, ok := res.Instance()
		if !ok {
			panic("unexpected response state")
		}
		rr.serveInstance(w, r, ses, loc, inst)
		return true
	}
	return false
}

func (rr *Router) serveSync(w http.ResponseWriter, r *http.Request, instanceId string) {
	ses := rr.getSession(w, r)
	if ses == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	inst, found := ses.GetInstance(instanceId)
	if !found {
		w.WriteHeader(http.StatusGone)
		return
	}
	inst.Connect(w, r)
}

func (rr *Router) tryServeRoute(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range rr.routes {
		if !route.Match(r) {
			continue
		}
		ctx := context.WithValue(r.Context(), ctex.KeyRouter, rr)
		route.Serve(w, r.WithContext(ctx))
		return true
	}
	return false
}

func (rr *Router) tryServeContent(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" {
		return false
	}
	if rr.tryServeRoute(w, r) {
		return true
	}
	if rr.tryServePage(w, r) {
		return true
	}
	return false
}

func (rr *Router) tryServeUtility(w http.ResponseWriter, r *http.Request) bool {
	match, ok := rr.pathMaker.Match(r)
	if !ok {
		return false
	}

	if id, ok := match.Resource(); ok {
		if r.Method != http.MethodGet {
			return false
		}
		rr.registry.Serve(id, w, r)
		return true
	}

	if hook, ok := match.Hook(); ok {
		rr.serveHook(w, r, hook.Instance, hook.Hook, hook.Track)
		return true
	}

	if instanceID, ok := match.Sync(); ok {
		if r.Method != http.MethodPut {
			return false
		}
		rr.serveSync(w, r, instanceID)
		return true
	}
	if match, ok := match.Undo(); ok {
		rr.restoreLocation(w, r, match.Instance, match.Location)
		return true
	}
	return false
}

func (rr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rr.tryServeUtility(w, r) {
		return
	}
	if rr.tryServeContent(w, r) {
		return
	}
	if rr.fallback != nil {
		rr.fallback.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func (rr *Router) serveError(w http.ResponseWriter, r *http.Request, ses *instance.Session, loc path.Location, err error) {
	if errors.Is(err, model.InstanceCreationError{}) {
		http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
		return
	}
	if rr.errPage == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	comp := rr.errPage(loc, err)
	if comp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	inst, ok := instance.NewInstance(ses,
		path.NewLocationAdapter(),
		beam.NewSourceEqual(loc, func(a, b path.Location) bool {
			return path.EqualLocation(a, b)
		}),
		comp,
		instance.Options{
			Rerouted: false,
		},
	)
	if ok == false {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	inst.SetStatus(http.StatusInternalServerError)
	rr.serveInstance(w, r, ses, loc, inst)
}
