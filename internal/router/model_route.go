// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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

type ResponseApp[M any] struct {
	App     instance.App[M]
	Model   *M
	Adapter *path.Adapter[M]
}

type responseAnyApp interface {
	intoInstance(*instance.Session, *instance.Options) (instance.AnyInstance, bool)
	getModel() any
	getAdapter() path.AnyAdapter
}

func (pr *ResponseApp[M]) getAdapter() path.AnyAdapter {
	return pr.Adapter
}

func (pr *ResponseApp[M]) getModel() any {
	return pr.Model
}

func (pr *ResponseApp[M]) intoInstance(sess *instance.Session, opt *instance.Options) (instance.AnyInstance, bool) {
	return instance.NewInstance(sess, pr.App, pr.Adapter, pr.Model, opt)
}

func (pr *ResponseApp[M]) marker() responseMarker {
	return responseMarker{}
}

type ResponseReroute struct {
	Detached bool
	Model    any
}

func (pr *ResponseReroute) marker() responseMarker {
	return responseMarker{}
}

type modelRoute[M any] struct {
	adapter *path.Adapter[M]
	handler func(r *Request[M]) Response
}

type anyModelRoute interface {
	getName() string
	getAdapter() path.AnyAdapter
	handleLocation(http.ResponseWriter, *http.Request, *path.Location) (Response, bool)
	handleModel(http.ResponseWriter, *http.Request, any) (Response, bool)
}

func (pr *modelRoute[M]) handleModel(w http.ResponseWriter, r *http.Request, model any) (Response, bool) {
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

func (pr *modelRoute[M]) handleLocation(w http.ResponseWriter, r *http.Request, l *path.Location) (Response, bool) {
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

func (pr *modelRoute[M]) getAdapter() path.AnyAdapter {
	return pr.adapter
}

func (pr *modelRoute[M]) getName() string {
	return pr.adapter.GetName()
}

func (pr *modelRoute[M]) apply(rr *Router) {
	rr.addModelRoute(pr)
}

func UseModel[M any](handler func(r *Request[M]) Response) Use {
	adapter, err := path.NewAdapter[M]()
	if err != nil {
		log.Fatal(err.Error())
	}
	return &modelRoute[M]{
		adapter: adapter,
		handler: handler,
	}
}

type ResponseRedirect struct {
	Status int
	Model  any
}

func (pr *ResponseRedirect) marker() responseMarker {
	return responseMarker{}
}

type ResponseRawRedirect struct {
	Status int
	URL    string
}

func (pr *ResponseRawRedirect) marker() responseMarker {
	return responseMarker{}
}



