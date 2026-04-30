package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/path"
)

const ZombieHeader = "X-Zombie"

func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), ctex.KeyApp, a))
	if a.tryServeUtility(w, r) {
		return
	}
	a.handler.ServeHTTP(w, r)
}

func (a *app) tryServeUtility(w http.ResponseWriter, r *http.Request) bool {
	match, ok := a.pathMaker.Match(r)
	if !ok {
		return false
	}
	if id, ok := match.Resource(); ok {
		if r.Method != http.MethodGet {
			return false
		}
		a.registry.Serve(id, w, r)
		return true
	}

	if hook, ok := match.Hook(); ok {
		a.serveHook(w, r, hook.Instance, hook.Hook, hook.Track)
		return true
	}
	if instanceID, ok := match.Sync(); ok {
		if r.Method != http.MethodPut {
			return false
		}
		a.serveSync(w, r, instanceID)
		return true
	}
	if match, ok := match.Undo(); ok {
		a.restoreLocation(w, r, match.Instance, match.Location)
		return true
	}
	return false
}

func (a *app) serveSync(w http.ResponseWriter, r *http.Request, instanceId string) {
	ses := a.getSession(w, r)
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

func (a *app) serveInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html")
	loc := path.NewLocationFromURL(r.URL)
	iters := 0
sess:
	iters += 1
	if iters >= 32 {
		a.serveError(w, r, errors.New("suspected instance/session destruction loop"))
		return
	}
	sess := a.ensureSession(w, r)
	inst, ok := sess.Instance(loc)
	if !ok {
		goto sess
	}
	contextWithTimeout, cancel := context.WithTimeout(r.Context(), a.conf.RequestTimeout)
	requestWithTimeout := r.WithContext(contextWithTimeout)
	err, handeled := inst.Serve(w, requestWithTimeout, a.page)
	cancel()
	if !handeled {
		goto sess
	}
	if err != nil {
		a.serveError(w, r, err)
	}
}

func (a *app) serveHook(w http.ResponseWriter, r *http.Request, instanceID string, hookID uint64, track uint64) {
	ses := a.getSession(w, r)
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

func (a *app) restoreLocation(w http.ResponseWriter, r *http.Request, instId string, l path.Location) {
	w.Header().Set("Cache-Control", "no-cache")
	ses := a.getSession(w, r)
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
		inst.Kill()
		w.WriteHeader(http.StatusGone)
		return
	}
	ok = inst.UpdateLocation(l)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *app) serveError(w http.ResponseWriter, r *http.Request, err error) {
	if a.errPage == nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	a.errPage(r, err).Render(r.Context(), w)
}
