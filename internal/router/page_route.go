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
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
)

type responseMarker struct{}

type Response interface {
	marker() responseMarker
}

type Request[M any] struct {
	Model   *M
	W       http.ResponseWriter
	R       *http.Request
	Adapter *path.Adapter[M]
}

type StaticPage struct {
	Status  int
	Content templ.Component
}

func (pr *StaticPage) marker() responseMarker {
	return responseMarker{}
}

type PageResponse[M any] struct {
	Page    instance.Page[M]
	Model   *M
	Adapter *path.Adapter[M]
}

type anyPageResponse interface {
	intoInstance(*instance.Session, *instance.Options) (instance.AnyInstance, bool)
	getModel() any
	getAdapter() path.AnyAdapter
}

func (pr *PageResponse[M]) getAdapter() path.AnyAdapter {
	return pr.Adapter
}

func (pr *PageResponse[M]) getModel() any {
	return pr.Model
}

func (pr *PageResponse[M]) intoInstance(sess *instance.Session, opt *instance.Options) (instance.AnyInstance, bool) {
	return instance.NewInstance(sess, pr.Page, pr.Adapter, pr.Model, opt)
}

func (pr *PageResponse[M]) marker() responseMarker {
	return responseMarker{}
}

type RerouteResponse struct {
	Detached bool
	Model    any
}

func (pr *RerouteResponse) marker() responseMarker {
	return responseMarker{}
}

type pageRoute[M any] struct {
	adapter *path.Adapter[M]
	handler func(r *Request[M]) Response
}

type anyPageRoute interface {
	getName() string
	getAdapter() path.AnyAdapter
	handleLocation(http.ResponseWriter, *http.Request, *path.Location) (Response, bool)
	handleModel(http.ResponseWriter, *http.Request, any) (Response, bool)
}

func (pr *pageRoute[M]) handleModel(w http.ResponseWriter, r *http.Request, model any) (Response, bool) {
	m, ok := pr.adapter.GetRef(model)
	if !ok {
		return nil, false
	}
	return pr.handler(&Request[M]{
		Model:   m,
		R:       r,
		W:       w,
		Adapter: pr.adapter,
	}), true
}

func (pr *pageRoute[M]) handleLocation(w http.ResponseWriter, r *http.Request, l *path.Location) (Response, bool) {
	m, ok := pr.adapter.Decode(l)
	if !ok {
		return nil, false
	}
	return pr.handler(&Request[M]{
		Model:   m,
		R:       r,
		W:       w,
		Adapter: pr.adapter,
	}), true
}

func (pr *pageRoute[M]) getAdapter() path.AnyAdapter {
	return pr.adapter
}

func (pr *pageRoute[M]) getName() string {
	return pr.adapter.GetName()
}

func (pr *pageRoute[M]) apply(rr *Router) {
	rr.addPage(pr)
}

func RoutePage[M any](handler func(r *Request[M]) Response) Mod {
	adapter, err := path.NewAdapter[M]()
	if err != nil {
		log.Fatal(err.Error())
	}
	return &pageRoute[M]{
		adapter: adapter,
		handler: handler,
	}
}

type RedirectResponse struct {
	Status int
	Model  any
}

func (pr *RedirectResponse) marker() responseMarker {
	return responseMarker{}
}
