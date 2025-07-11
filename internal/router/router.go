package router

import (
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
)

func NewRouter() *Router {
	conf := &common.SystemConf{}
	common.InitDefaults(conf)
	return &Router{
		pool:           shredder.NewPool(conf.InstanceGoroutineLimit),
		sess:           sync.Map{},
		adapters:       make(map[string]path.AnyAdapter),
		pageRoutes:     make(map[string]anyPageRoute),
		pageRouteOrder: make([]string, 0),
		fallback:       nil,
		dirs:           make([]*static, 0),
		registry:       resources.NewRegistry(),
		conf:           conf,
	}
}

type static struct {
	prefix  bool
	path    string
	handler http.Handler
}

func (d *static) tryServe(w http.ResponseWriter, r *http.Request) bool {
	if d.prefix {
		if !strings.HasPrefix(r.URL.Path, d.path) {
			return false
		}
		d.handler.ServeHTTP(w, r)
		return true
	}
	if d.path != r.URL.Path {
		return false
	}
	d.handler.ServeHTTP(w, r)
	return true
}

type ErrorPageComponent = func(message string) templ.Component

type sessionHooks struct {
	create func(string)
	delete func(string)
}

func (s *sessionHooks) onCreate(id string) {
	if s == nil {
		return
	}
	if s.create == nil {
		return
	}
	s.create(id)
}

func (s *sessionHooks) onDelete(id string) {
	if s == nil {
		return
	}
	if s.delete == nil {
		return
	}
	s.delete(id)
}

type Router struct {
	sess           sync.Map
	adapters       map[string]path.AnyAdapter
	pageRoutes     map[string]anyPageRoute
	pageRouteOrder []string
	fallback       http.Handler
	dirs           []*static
	pool           *shredder.Pool
	errPage        ErrorPageComponent
	sessionHooks   *sessionHooks
	registry       *resources.Registry
	csp            *common.CSP
	conf           *common.SystemConf
	used           atomic.Bool
}

func (rr *Router) Gzip() bool {
	return rr.registry.Gzip
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
	rr.sess.Delete(id)
	rr.sessionHooks.onDelete(id)
}
func (rr *Router) Adapters() map[string]path.AnyAdapter {
	return rr.adapters
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
	if rr.conf.SessionCookieExpiration != 0 {
		expires = time.Now().Add(rr.conf.SessionCookieExpiration)
	}
	cookie := &http.Cookie{
		Name:     "d00r",
		Value:    s.Id(),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	}
	rr.sess.Store(s.Id(), s)
	rr.sessionHooks.onCreate(s.Id())
	http.SetCookie(w, cookie)
	return true, s
}

func (rr *Router) getSession(r *http.Request) *instance.Session {
	c, err := r.Cookie("d00r")
	if err != nil {
		return nil
	}
	v, ok := rr.sess.Load(c.Value)
	if !ok {
		return nil
	}
	return v.(*instance.Session)
}

func (rr *Router) addPage(page anyPageRoute) {
	name := page.getName()
	_, has := rr.pageRoutes[name]
	if has {
		log.Fatal("Can't register same model twice ", name)
	}
	rr.pageRoutes[name] = page
	rr.adapters[name] = page.getAdapter()
	rr.pageRouteOrder = append(rr.pageRouteOrder, name)
}

func (rr *Router) Use(mods ...Mod) {
	if rr.used.Load() {
		slog.Error("IMPORTANT! Router mod used after first request is handled, ignoring mod.")
		return
	}
	for _, r := range mods {
		r.apply(rr)
	}
}
