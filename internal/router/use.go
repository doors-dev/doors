// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package router

import (
	"net/http"
	"net/url"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/router/model"
)

// Use mutates a [Router] during setup.
type Use interface {
	apply(rr *Router)
}

type useFunc func(*Router)

func (a useFunc) apply(rr *Router) {
	a(rr)
}

// UseModel registers a model-based page handler.
func UseModel[M any](handler model.Handler[M]) Use {
	adapter, err := path.NewAdapter[M]()
	if err != nil {
		panic(err)
	}
	return useFunc(func(r *Router) {
		route := model.NewModelRoute(adapter, handler)
		r.addModelRoute(route)
	})
}

// UseRoute registers a custom non-page route.
func UseRoute(r Route) Use {
	return useFunc(func(rr *Router) {
		rr.addRoute(r)
	})
}

// UseFallback registers the fallback handler for unmatched requests.
func UseFallback(handler http.Handler) Use {
	return useFunc(func(rr *Router) {
		rr.fallback = handler
	})
}

// UseSystemConf stores the router-wide system configuration.
func UseSystemConf(conf common.SystemConf) Use {
	return useFunc(func(rr *Router) {
		common.InitDefaults(&conf)
		rr.conf = &conf
	})
}

// UseErrorPage stores the page used for framework errors.
func UseErrorPage(page ErrorPageComponent) Use {
	return useFunc(func(rr *Router) {
		rr.errPage = page
	})
}

// SessionCallback observes session creation and removal.
type SessionCallback interface {
	// Create is called when a session is created.
	Create(id string, header http.Header)
	// Delete is called when a session is removed.
	Delete(id string)
}

// UseSessionCallback registers session lifecycle callbacks.
func UseSessionCallback(hook SessionCallback) Use {
	return useFunc(func(rr *Router) {
		rr.sessionCallback = hook
	})
}

// UseESConf stores the esbuild profiles used for imports.
func UseESConf(profiles resources.BuildProfiles) Use {
	return useFunc(func(rr *Router) {
		rr.buildProfiles = profiles
	})
}

// UseCSP stores the Content Security Policy used for responses.
func UseCSP(csp *common.CSP) Use {
	return useFunc(func(rr *Router) {
		rr.csp = csp
	})
}

// UseLicense stores the client license string.
func UseLicense(license string) Use {
	return useFunc(func(rr *Router) {
		rr.license = license
	})
}

// UseServerID stores the stable server identifier used in generated paths.
func UseServerID(id string) Use {
	if id != url.PathEscape(id) {
		panic("server ID must be URL compatible without escaping")
	}
	return useFunc(func(rr *Router) {
		rr.pathMaker = path.NewPathMaker(id)
	})
}
