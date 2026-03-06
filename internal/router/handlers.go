// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/router/model"
	"github.com/mr-tron/base58"
)

var instanceRegexp = regexp.MustCompile(`^([0-9a-zA-Z]+)$`)
var hookRegexp = regexp.MustCompile(`^([0-9a-zA-Z]+)/(\d+)/(\d+)(/[^/]+)?`)
var importRegexp = regexp.MustCompile(`^r/([0-9a-zA-Z]+)\.([^/]+)`)

func ResourcePath(r *resources.Resource, ext string) string {
	return fmt.Sprint("/~0/r/" + r.HashString() + "." + ext)
}

func (rr *Router) serveHook(w http.ResponseWriter, r *http.Request, instanceID string, doorID uint64, hookID uint64, track uint64) {
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
	found = inst.TriggerHook(doorID, hookID, w, r, track)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (rr *Router) tryServeHook(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" && r.Method != "GET" {
		return false
	}
	matches := hookRegexp.FindStringSubmatch(r.URL.Path)
	if len(matches) == 0 {
		return false
	}
	instanceID := matches[1]
	doorId, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	hookId, err2 := strconv.ParseUint(matches[3], 10, 64)
	if err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	track := uint64(0)
	trackStr := r.URL.Query().Get("t")
	if trackStr != "" {
		track, err = strconv.ParseUint(trackStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
	}
	rr.serveHook(w, r, instanceID, doorId, hookId, track)
	return true
}

func (rr *Router) restorePath(w http.ResponseWriter, r *http.Request, instId string) {
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
	ok = inst.RestorePath(r)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rr *Router) serveModelRedirect(w http.ResponseWriter, r *http.Request, s *instance.Session, l path.Location, model any, status int) {
	name := path.GetAdapterName(model)
	adapter, ok := rr.modelAdapters[name]
	if !ok {
		err := errors.New("Adapter " + name + " not found")
		slog.Error("page routing error", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		rr.serveError(w, r, s, l, err)
		return
	}
	location, err := adapter.EncodeAny(model)
	if err != nil {
		err := errors.New("Adapter " + name + " encoding error: " + err.Error())
		slog.Error("page routing error", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
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
		slog.Error("instance serve error", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		rr.serveError(w, r, s, l, err)
	}
}

func (rr *Router) tryServePage(w http.ResponseWriter, r *http.Request) bool {
	instId := r.Header.Get("D0-r")
	if instId != "" {
		rr.restorePath(w, r, instId)
		return true
	}
	loc := path.NewRequestLocation(r)
	var a any = loc
	ses := rr.ensureSession(w, r)
	var counter = 0
	opt := instance.Options{
		Rerouted: false,
	}
main:
	counter += 1
	if counter > 64 {
		slog.Error("page routing error", slog.String("path", r.URL.Path), slog.String("error", "reroute loop"))
		rr.serveError(w, r, ses, loc, errors.New("rerouting loop"))
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

func (rr *Router) servePut(w http.ResponseWriter, r *http.Request, instanceId string) {
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

func (rr *Router) tryServePut(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "PUT" {
		return false
	}
	matches := instanceRegexp.FindStringSubmatch(r.URL.Path)
	if len(matches) == 0 {
		return false
	}
	instanceId := matches[1]
	rr.servePut(w, r, instanceId)
	return true
}

func (rr *Router) tryServeAssets(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" {
		return false
	}
	script := rr.registry.MainScript()
	if r.URL.Path == script.HashString()+".js" {
		script.Serve(w, r)
		return true
	}
	style := rr.registry.MainStyle()
	if r.URL.Path == style.HashString()+".css" {
		style.Serve(w, r)
		return true
	}
	return false
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

func (rr *Router) tryServeResource(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" {
		return false
	}
	matches := importRegexp.FindStringSubmatch(r.URL.Path)
	if len(matches) == 0 {
		return false
	}
	hashStr := matches[1]
	hash, err := base58.Decode(hashStr)
	if err != nil {
		w.WriteHeader(400)
		return true
	}
	rr.registry.Serve(hash, w, r)
	return true

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
	url := r.URL.Path
	if !strings.HasPrefix(url, "/~0/") {
		return false
	}
	r.URL.Path = strings.TrimPrefix(url, "/~0/")
	if rr.tryServeAssets(w, r) {
		return true
	}
	if rr.tryServeResource(w, r) {
		return true
	}
	if rr.tryServeHook(w, r) {
		return true
	}
	if rr.tryServePut(w, r) {
		return true
	}
	r.URL.Path = url
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
			return reflect.DeepEqual(a, b)
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
