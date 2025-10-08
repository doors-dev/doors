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

// App defines a renderable app component, where Render() must output page HTML
// M - the path model type.
type App[M any] interface {
	Init(SourceBeam[M])
	Render() templ.Component
}

// ModelRoute provides the response type for page handlers
// (page, redirect, reroute, or static content).
type ModelRoute = router.Response

// RModel provides request data and response control for model handlers.
type RModel[M any] interface {
	R
	// Model returns the decoded path model.
	Model() M
	// RequestHeader returns the incoming request headers.
	RequestHeader() http.Header
	// ResponseHeader returns the outgoing response headers.
	ResponseHeader() http.Header
}

// ModelRouter provides helpers to produce page responses.
type ModelRouter[M any] interface {
	// Page renders a Page.
	App(app App[M]) ModelRoute
	// PageFunc renders a Page from a function.
	AppFunc(AppFunc func(SourceBeam[M]) templ.Component) ModelRoute
	// StaticPage returns a static page with status.
	StaticPage(content templ.Component, status int) ModelRoute
	// Reroute performs an internal reroute to model (detached=true disables path sync).
	Reroute(model any, detached bool) ModelRoute
	// Redirect issues an HTTP redirect to model with status.
	Redirect(model any, status int) ModelRoute
	// Redirect issues an HTTP redirect to URL with status.
	RawRedirect(url string, status int) ModelRoute
}

type modelRequest[M any] struct {
	r *router.Request[M]
}

func (r *modelRequest[M]) Model() M {
	return *r.r.Model
}

func (r *modelRequest[M]) Done() <-chan struct{} {
	return r.r.R.Context().Done()
}

func (r *modelRequest[M]) RequestHeader() http.Header {
	return r.r.R.Header
}

func (r *modelRequest[M]) ResponseHeader() http.Header {
	return r.r.W.Header()
}

func (r *modelRequest[M]) GetCookie(name string) (*http.Cookie, error) {
	return r.r.R.Cookie(name)
}

func (r *modelRequest[M]) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.r.W, cookie)
}

func (r *modelRequest[M]) Reroute(model any, detached bool) ModelRoute {
	return &router.ResponseReroute{
		Detached: detached,
		Model:    model,
	}
}

func (r *modelRequest[M]) Redirect(model any, status int) ModelRoute {
	return &router.ResponseRedirect{
		Model:  model,
		Status: status,
	}
}

func (r *modelRequest[M]) RawRedirect(url string, status int) ModelRoute {
	return &router.ResponseRawRedirect{
		URL:    url,
		Status: status,
	}
}

func (r *modelRequest[M]) App(app App[M]) ModelRoute {
	return &router.ResponseApp[M]{
		App:     app,
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
	}
}

func (r *modelRequest[M]) StaticPage(content templ.Component, status int) ModelRoute {
	return &router.StaticPage{
		Content: content,
		Status:  status,
	}
}

type appFunc[M any] struct {
	model  SourceBeam[M]
	render func(SourceBeam[M]) templ.Component
}

func (af *appFunc[M]) Init(model SourceBeam[M]) {
	af.model = model
}

func (af *appFunc[M]) Render() templ.Component {
	return af.render(af.model)
}

func (r *modelRequest[M]) AppFunc(f func(SourceBeam[M]) templ.Component) ModelRoute {
	return &router.ResponseApp[M]{
		App: &appFunc[M]{
			render: f,
		},
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
	}
}

func (r *modelRequest[M]) PageFunc(f func(SourceBeam[M]) templ.Component) PageRoute {
	return &router.ResponseApp[M]{
		App: &appFunc[M]{
			render: f,
		},
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
	}
}

func (r *modelRequest[M]) Page(page Page[M]) PageRoute {
	return &router.ResponseApp[M]{
		App: &appFunc[M]{
			render: func(model SourceBeam[M]) templ.Component {
				return page.Render(model)
			},
		},
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
	}
}
