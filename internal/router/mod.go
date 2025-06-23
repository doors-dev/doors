package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/common"
)

type Mod interface {
	apply(rr *Router)
}

type anyMod func(*Router)

func (a anyMod) apply(rr *Router) {
	a(rr)
}

func ServeRaw(path string, handler func(w http.ResponseWriter, r *http.Request)) Mod {
	return anyMod(func(rr *Router) {
		rr.dirs = append(rr.dirs, &static{
			path:    path,
			handler: http.HandlerFunc(handler),
		})
	})
}

func ServeFile(path string, localPath string) Mod {
	return anyMod(func(rr *Router) {
		rr.dirs = append(rr.dirs, &static{
			path: path,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, localPath)
			}),
		})
	})
}

func ServeDirPath(prefix string, localPath string) Mod {
	return ServeDir(prefix, http.Dir(localPath))
}

func ServeDir(prefix string, root http.FileSystem) Mod {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return anyMod(func(rr *Router) {
		rr.dirs = append(rr.dirs, &static{
			prefix:  true,
			path:    prefix,
			handler: http.StripPrefix(prefix, http.FileServer(root)),
		})
	})
}

func SetFallback(handler http.Handler) Mod {
	return anyMod(func(rr *Router) {
		rr.fallback = handler
	})
}

func SetInstanceLimit(n int) Mod {
	if n < 1 {
		common.BadPanic("At least 1")
	}
	return anyMod(func(rr *Router) {
		rr.instLimit = n
	})
}

func SetGoroutineLimit(n int) Mod {
	if n < 1 {
		common.BadPanic("At least 1")
	}
	return anyMod(func(rr *Router) {
		rr.pool.Tune(n)
	})
}

func SetInstanceTTL(duration time.Duration) Mod {
	return anyMod(func(rr *Router) {
		rr.instanceTTL = duration
	})
}

func SetErrorPage(page ErrorPageComponent) Mod {
	return anyMod(func(rr *Router) {
		rr.errPage = page
	})
}

func SetSessionHooks(create func(id string), delete func(id string)) Mod {
	return anyMod(func(rr *Router) {
		rr.sessionHooks = &sessionHooks{
			create: create,
			delete: delete,
		}
	})
}

func SetSessionExpire(d time.Duration) Mod {
	return anyMod(func(rr *Router) {
		rr.sessionExpire = d
	})
}

func SetSessionCookieExpire(d time.Duration) Mod {
	return anyMod(func(rr *Router) {
		rr.sessionCookieExpire = d
	})
}

func SetGzip(enable bool) Mod {
	return anyMod(func(rr *Router) {
		rr.registry.Gzip = enable
	})
}

func SetBuildProfiles(profiles resources.BuildProfiles) Mod {
	return anyMod(func(rr *Router) {
		rr.registry.Profiles = profiles
	})
}

func SetCSP(csp *common.CSP) Mod {
	return anyMod(func(rr *Router) {
		rr.csp = csp
	})
}
