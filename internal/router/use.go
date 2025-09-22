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
	"log/slog"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/resources"
)

type Use interface {
	apply(rr *Router)
}

type useFunc func(*Router)

func (a useFunc) apply(rr *Router) {
	a(rr)
}

func UseRoute(r Route) Use {
	return useFunc(func(rr *Router) {
		rr.addRoute(r)
	})
}

func UseFallback(handler http.Handler) Use {
	return useFunc(func(rr *Router) {
		rr.fallback = handler
	})
}

func UseSystemConf(conf common.SystemConf) Use {
	return useFunc(func(rr *Router) {
		common.InitDefaults(&conf)
		rr.conf = &conf
		rr.pool.Tune(conf.InstanceGoroutineLimit)
	})
}

func UseErrorPage(page ErrorPageComponent) Use {
	return useFunc(func(rr *Router) {
		rr.errPage = page
	})
}

type SessionCallback interface {
	Create(id string, header http.Header)
	Delete(id string)
}

func UseSessionCallback(hook SessionCallback) Use {
	return useFunc(func(rr *Router) {
		rr.sessionCallback = hook
	})
}

func UseESConf(profiles resources.BuildProfiles) Use {
	return useFunc(func(rr *Router) {
		rr.buildProfiles = profiles
	})
}

func UseCSP(csp *common.CSP) Use {
	return useFunc(func(rr *Router) {
		rr.csp = csp
	})
}

const issuer = "Doors5qRhZiAB4Fmhpd5Td2Rn4BwkFiqCdBMw7BzbCsp"

func UseLicense(cert string) Use {
	return useFunc(func(rr *Router) {
		lic, err := license.ReadCert(cert)
		if err != nil {
			slog.Error("license error", slog.String("error", err.Error()))
			return
		}
		if lic.GetIssuer() != issuer {
			slog.Error("license error", slog.String("error", "wrong issuer key"))
			return
		}
		rr.lisence = lic
	})
}
