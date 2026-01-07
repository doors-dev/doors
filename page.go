// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"github.com/a-h/templ"
)

// Page defines a renderable page component that the page must implement.
// M - the path model type.
// Deprecated: Use App[M]
type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}

// PageRoute provides the response type for page handlers
// (page, redirect, reroute, or static content).
// Deprecated: Use ModelRoute
type PageRoute = ModelRoute

// RPage provides request data and response control for page handlers.
// Deprecated: Use RModel[M]
type RPage[M any] = RModel[M]

// PageRouter provides helpers to produce page responses.
// Deprecated: Use ModelRouter[M any]
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
	// Redirect issues an HTTP redirect to URL with status.
	RawRedirect(url string, status int) PageRoute
}

