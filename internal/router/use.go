// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
