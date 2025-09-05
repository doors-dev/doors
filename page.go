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

// Page defines the interface for page components that can be rendered with reactive data.
// Pages receive a SourceBeam containing the model data and return a templ.Component.
//
// Example:
//
//	type BlogPage struct {
//	    beam doors.SourceBeam[BlogPath]
//	}
//
//	func (p *BlogPage) Render(beam doors.SourceBeam[BlogPath]) templ.Component {
//	    p.beam = beam
//	    return common.Template(p)
//	}
//
//	func (p *BlogPage) Body() templ.Component {
//	    return doors.Sub(p.beam, func(path BlogPath) templ.Component {
//	        switch {
//	        case path.Home:
//	            return homePage()
//	        case path.Post:
//	            return postPage(path.ID)
//	        }
//	    })
//	}
type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}

// PageRoute represents a response that can be returned from page handlers.
// This includes page responses, redirects, reroutes, and static content.
type PageRoute = router.Response

// RPage provides access to request data and response control for page handlers.
// It combines basic request/response functionality with model access.
type RPage[M any] interface {
	R
	// Returns the decoded URL model
	GetModel() M
	// Access to incoming request headers
	RequestHeader() http.Header
	// Access to outgoing response headers
	ResponseHeader() http.Header
}

// PageRouter provides methods for creating different types of page responses.
// It allows rendering pages, redirecting, rerouting, and serving static content.
type PageRouter[M any] interface {
	// Serve page
	Page(page Page[M]) PageRoute
	// Serve func page
	PageFunc(pageFunc func(SourceBeam[M]) templ.Component) PageRoute
	// Serve static page
	StaticPage(content templ.Component, status int) PageRoute
	// Internal reroute to different model (detached=true disables path synchronization)
	Reroute(model any, detached bool) PageRoute
	// HTTP redirect with status
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
