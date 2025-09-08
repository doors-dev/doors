// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/router"
	"net/http"
)

// Page defines a renderable page component that the application must implement.
// M - the path model type.
type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}

// PageRoute provides the response type for page handlers
// (page, redirect, reroute, or static content).
type PageRoute = router.Response

// RPage provides request data and response control for page handlers.
type RPage[M any] interface {
	R
	// GetModel returns the decoded URL model.
	GetModel() M
	// RequestHeader returns the incoming request headers.
	RequestHeader() http.Header
	// ResponseHeader returns the outgoing response headers.
	ResponseHeader() http.Header
}

// PageRouter provides helpers to produce page responses.
type PageRouter[M any] interface {
	// Page renders a Page.
	Page(page Page[M]) PageRoute
	// PageFunc renders a Page from a function.
	PageFunc(pageFunc func(SourceBeam[M]) templ.Component) PageRoute
	// StaticPage returns a static page with status.
	StaticPage(content templ.Component, status int) PageRoute
	// Reroute performs an internal reroute to model (detached=true disables path sync).
	Reroute(model any, detached bool) PageRoute
	// Redirect issues an HTTP redirect to model with status.
	Redirect(model any, status int) PageRoute
}

type pageRequest[M any] struct {
	r *router.Request[M]
}

func (r *pageRequest[M]) GetModel() M {
	return *r.r.Model
}

func (r *pageRequest[M]) Done() <-chan struct{} {
	return r.r.R.Context().Done()
}

func (r *pageRequest[M]) RequestHeader() http.Header {
	return r.r.R.Header
}

func (r *pageRequest[M]) ResponseHeader() http.Header {
	return r.r.W.Header()
}

func (r *pageRequest[M]) GetCookie(name string) (*http.Cookie, error) {
	return r.r.R.Cookie(name)
}

func (r *pageRequest[M]) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.r.W, cookie)
}

func (r *pageRequest[M]) Reroute(model any, detached bool) PageRoute {
	return &router.RerouteResponse{
		Detached: detached,
		Model:    model,
	}
}

func (r *pageRequest[M]) Redirect(model any, status int) PageRoute {
	return &router.RedirectResponse{
		Model:  model,
		Status: status,
	}
}

func (r *pageRequest[M]) Page(page Page[M]) PageRoute {
	return &router.PageResponse[M]{
		Page:    page,
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
	}
}

func (r *pageRequest[M]) StaticPage(content templ.Component, status int) PageRoute {
	return &router.StaticPage{
		Content: content,
		Status:  status,
	}
}

type pageFunc[M any] func(SourceBeam[M]) templ.Component

func (p pageFunc[M]) Render(b SourceBeam[M]) templ.Component {
	return p(b)
}

func (r *pageRequest[M]) PageFunc(f func(SourceBeam[M]) templ.Component) PageRoute {
	return r.Page(pageFunc[M](f))
}
