package router

import (
	"context"
	"errors"
	"fmt"
	"log"
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

func (rr *Router) serveHook(w http.ResponseWriter, r *http.Request, instanceId string, nodeId uint64, hookId uint64) {
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
	found = inst.TriggerHook(nodeId, hookId, w, r)
	if !found {
		w.WriteHeader(http.StatusForbidden)
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
	nodeId, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	hookId, err2 := strconv.ParseUint(matches[3], 10, 64)
	if err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	rr.serveHook(w, r, instanceId, nodeId, hookId)
	return true
}

func (rr *Router) servePage(w http.ResponseWriter, r *http.Request, page anyPageResponse, opt *instance.Options) {
	new, session := rr.ensureSession(r, w)
	instId := r.Header.Get("d00r")
	if instId != "" {
		inst, ok := session.GetInstance(instId)
		if !ok {
			w.WriteHeader(http.StatusGone)
			return
		}
		ok = inst.UpdatePath(page.getModel(), page.getAdapter())
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	inst, ok := page.intoInstance(session, opt)
	if !ok {
		if new {
			log.Fatalf("New session can't end")
		}
		rr.sess.Delete(session.Id())
		rr.servePage(w, r, page, opt)
		return
	}
	err := inst.Serve(w, page.getStatus())
	if err != nil {
		rr.serveError(w, r, err.Error())
	}
}

func (rr *Router) tryServePage(w http.ResponseWriter, r *http.Request) bool {
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
			rr.serveError(w, r, "probably you have infinite reroute")
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
            case *StaticPage: {
                w.WriteHeader(res.Status)
                ctx := context.WithValue(r.Context(), common.AdaptersCtxKey, rr.Adapters())
                res.Content.Render(ctx, w)
                return true
            }
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
					rr.serveError(w, r, "Adapter "+name+" not found")
					return true
				}
				location, err := adapter.EncodeAny(res.Model)
				if err != nil {
					rr.serveError(w, r, "Adapter "+name+" encoding error "+err.Error())
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
	instId := r.Header.Get("D00r")
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
	if r.URL.Path != "/"+main.HashString()+".js" {
		return false
	}
	main.Serve(w, r)
	return true
}

func (rr *Router) tryServeCss(w http.ResponseWriter, r *http.Request) bool {
	main := rr.registry.MainStyle()
	if r.URL.Path != "/"+main.HashString()+".css" {
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
	if rr.errPage != nil {
		err := rr.errPage(m).Render(r.Context(), w)
		println(err)
	}
}
