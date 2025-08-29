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
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/mr-tron/base58"
)

var instanceRegexp = regexp.MustCompile(`^/d00r/([0-9a-zA-Z]+)$`)
var hookRegexp = regexp.MustCompile(`/d00r/([0-9a-zA-Z]+)/(\d+)/(\d+)(/[^/]+)?`)
var importRegexp = regexp.MustCompile(`/d00r/r/([0-9a-zA-Z]+)\.([^/]+)`)

func ResourcePath(r *resources.Resource, ext string) string {
	return fmt.Sprint("/d00r/r/" + r.HashString() + "." + ext)
}

func (rr *Router) serveHook(w http.ResponseWriter, r *http.Request, instanceId string, doorId uint64, hookId uint64) {
	sess := rr.getSession(r)
	if sess == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	inst, found := sess.GetInstance(instanceId)
	if !found {
		w.WriteHeader(http.StatusGone)
		return
	}
	found = inst.TriggerHook(doorId, hookId, w, r)
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
	instanceId := matches[1]
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
	rr.serveHook(w, r, instanceId, doorId, hookId)
	return true
}

func (rr *Router) servePage(w http.ResponseWriter, r *http.Request, page anyPageResponse, opt *instance.Options) {
	new, session := rr.ensureSession(r, w)
	inst, ok := page.intoInstance(session, opt)
	if !ok {
		if new {
			panic("New session can't end")
		}
		rr.sess.Delete(session.Id())
		rr.servePage(w, r, page, opt)
		return
	}
	err := inst.Serve(w, r)
	if err != nil {
		slog.Error("instance serve error", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		rr.serveError(w, r, err.Error())
	}
}

func (rr *Router) restorePath(w http.ResponseWriter, r *http.Request, instId string) {
	w.Header().Set("Cache-Control", "no-cache")
	sess := rr.getSession(r)
	if sess == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	inst, ok := sess.GetInstance(instId)
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

func (rr *Router) tryServePage(w http.ResponseWriter, r *http.Request) bool {
	instId := r.Header.Get("D00r")
	if instId != "" {
		rr.restorePath(w, r, instId)
		return true
	}
	l := path.NewRequestLocation(r)
	var model any = nil
	var response Response = nil
	var page anyPageResponse = nil
	var counter = 0
	opt := &instance.Options{
		Detached: false,
		Rerouted: false,
	}
main:
	for {
		counter += 1
		if counter > 64 {
			slog.Error("page routing error", slog.String("path", r.URL.Path), slog.String("error", "reroute loop"))
			rr.serveError(w, r, "reroute loop")
			return true
		}
		if page != nil {
			break
		}
		if model != nil {
			m := model
			model = nil
			name := path.GetAdapterName(m)
			pageRoute, ok := rr.pageRoutes[name]
			if !ok {
				break
			}
			response, ok = pageRoute.handleModel(w, r, m)
			if !ok {
				panic(errors.New("model name confilct " + name))
			}
		}
		if response != nil {
			res := response
			response = nil
			switch res := res.(type) {
			case *StaticPage:
				if res.Status == 0 {
					res.Status = http.StatusOK
				}
				w.WriteHeader(res.Status)
				ctx := context.WithValue(r.Context(), common.CtxKeyAdapters, rr.Adapters())
				res.Content.Render(ctx, w)
				return true
			case *RerouteResponse:
				if res.Detached {
					opt.Detached = true
				}
				opt.Rerouted = true
				model = res.Model
			case *RedirectResponse:
				name := path.GetAdapterName(res.Model)
				adapter, ok := rr.adapters[name]
				if !ok {
					msg := "Adapter " + name + " not found"
					slog.Error("page routing error", slog.String("path", r.URL.Path), slog.String("error", msg))
					rr.serveError(w, r, msg)
					return true
				}
				location, err := adapter.EncodeAny(res.Model)
				if err != nil {
					msg := "Adapter " + name + " encoding error: " + err.Error()
					slog.Error("page routing error", slog.String("path", r.URL.Path), slog.String("error", msg))
					rr.serveError(w, r, msg)
					return true
				}
				if res.Status == 0 {
					res.Status = http.StatusFound
				}
				http.Redirect(w, r, location.String(), res.Status)
				return true
			case anyPageResponse:
				page = res
			default:
				log.Fatalf("Unsupported response type")
			}
			continue
		}
		for _, name := range rr.pageRouteOrder {
			resp, ok := rr.pageRoutes[name].handleLocation(w, r, l)
			if ok {
				response = resp
				continue main
			}
		}
		break
	}
	if page == nil {
		if instId != "" {
			w.WriteHeader(http.StatusNotFound)
			return true
		}
		return false
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html")
	rr.servePage(w, r, page, opt)
	return true
}

func (rr *Router) servePut(w http.ResponseWriter, r *http.Request, instanceId string) {
	sess := rr.getSession(r)
	if sess == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	inst, found := sess.GetInstance(instanceId)
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

func (rr *Router) tryServeJs(w http.ResponseWriter, r *http.Request) bool {
	main := rr.registry.MainScript()
	if r.URL.Path != "/"+main.HashString()+".doors.js" {
		return false
	}
	main.Serve(w, r)
	return true
}

func (rr *Router) tryServeCss(w http.ResponseWriter, r *http.Request) bool {
	main := rr.registry.MainStyle()
	if r.URL.Path != "/"+main.HashString()+".doors.css" {
		return false
	}
	main.Serve(w, r)
	return true
}

func (rr *Router) tryServeDir(w http.ResponseWriter, r *http.Request) bool {
	for _, dir := range rr.dirs {
		if dir.tryServe(w, r) {
			return true
		}
	}
	return false
}

func (rr *Router) tryServeImport(w http.ResponseWriter, r *http.Request) bool {
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

func (rr *Router) tryServeGet(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" {
		return false
	}
	if rr.tryServeJs(w, r) {
		return true
	}
	if rr.tryServeCss(w, r) {
		return true
	}
	if rr.tryServeImport(w, r) {
		return true
	}
	if rr.tryServeDir(w, r) {
		return true
	}
	if rr.tryServePage(w, r) {
		return true
	}
	return false
}

func (rr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr.used.Store(true)
	if rr.tryServeHook(w, r) {
		return
	}
	if rr.tryServePut(w, r) {
		return
	}
	if rr.tryServeGet(w, r) {
		return
	}
	if rr.fallback != nil {
		rr.fallback.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func (rr *Router) serveError(w http.ResponseWriter, r *http.Request, m string) {
	w.WriteHeader(http.StatusInternalServerError)
	if rr.errPage == nil {
		return
	}
	err := rr.errPage(m).Render(r.Context(), w)
	if err == nil {
		return
	}
	slog.Error("error page rendering error", slog.String("error", err.Error()))
}
