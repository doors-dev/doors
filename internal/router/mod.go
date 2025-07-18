package router

import (
	"errors"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/resources"
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

func ServeFallback(handler http.Handler) Mod {
	return anyMod(func(rr *Router) {
		rr.fallback = handler
	})
}



func SetSystemConf(conf common.SystemConf) Mod {
	return anyMod(func(rr *Router) {
		common.InitDefaults(&conf)
		rr.conf = &conf
		rr.registry.Gzip = !conf.ServerDisableGzip
		rr.pool.Tune(conf.InstanceGoroutineLimit)
	})
}


func SetGoroutineLimit(n int) Mod {
	if n < 1 {
		panic(errors.New("At least 1"))
	}
	return anyMod(func(rr *Router) {
		rr.pool.Tune(n)
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
