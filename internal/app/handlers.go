package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
)

func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess := a.ensureSession(w, r)
	r = r.WithContext(context.WithValue(r.Context(), common.KeySession, sess))
	a.handler.ServeHTTP(w, r)
}

func (a *app) serve(w http.ResponseWriter, r *http.Request) {
	if a.tryServeUtility(w, r) {
		return
	}
	if r.Method != http.MethodGet || a.pathMaker.IsSystem(r) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	sess, ok := r.Context().Value(common.KeySession).(instance.Session)
	if !ok {
		a.serveError(w, r, errors.New("Session is removed from the request context"))
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html")
	loc := path.NewLocationFromURL(r.URL)
	inst, ok := sess.Instance(loc)
	if !ok {
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}
	contextWithTimeout, cancel := context.WithTimeout(r.Context(), a.conf.RequestTimeout)
	requestWithTimeout := r.WithContext(contextWithTimeout)
	err, handeled := inst.Serve(w, requestWithTimeout, a.page)
	cancel()
	if !handeled {
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		a.serveError(w, r, err)
	}
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
	ses, ok := r.Context().Value(common.KeySession).(instance.Session)
	if !ok {
		a.Logger().Error("Session is removed from the request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	inst, found := ses.GetInstance(instanceId)
	if !found {
		w.WriteHeader(http.StatusGone)
		return
	}
	inst.Connect(w, r)
}

func (a *app) serveHook(w http.ResponseWriter, r *http.Request, instanceID string, hookID uint64, track uint64) {
	ses, ok := r.Context().Value(common.KeySession).(instance.Session)
	if !ok {
		a.Logger().Error("Session is removed from the request context")
		w.WriteHeader(http.StatusInternalServerError)
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
	ses, ok := r.Context().Value(common.KeySession).(instance.Session)
	if !ok {
		a.Logger().Error("Session is removed from the request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	inst, ok := ses.GetInstance(instId)
	if !ok {
		w.WriteHeader(http.StatusGone)
		return
	}
	if a.Draining() {
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
