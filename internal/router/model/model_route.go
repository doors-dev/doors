// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package model

import (
	"net/http"
	"reflect"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

type InstanceCreationError struct{}

func (i InstanceCreationError) Error() string {
	return "instance creation error"
}

type AnyModelRoute interface {
	Adapter() path.AnyAdapter
	Handle(w http.ResponseWriter, r *http.Request, a any, sess *instance.Session, opt instance.Options) (Res, bool)
}

type Handler[M any] = func(w http.ResponseWriter, r *http.Request, source beam.Source[M], store ctex.Store) Res

func NewModelRoute[M any](a path.Adapter[M], h Handler[M]) AnyModelRoute {
	return &modelRoute[M]{a, h}
}

type modelRoute[M any] struct {
	a path.Adapter[M]
	h Handler[M]
}

func (n *modelRoute[M]) Adapter() path.AnyAdapter {
	return n.a
}

func (n *modelRoute[M]) Handle(w http.ResponseWriter, r *http.Request, a any, sess *instance.Session, opt instance.Options) (Res, bool) {
	var model *M
	if l, ok := a.(path.Location); ok {
		if m, ok := n.a.Decode(l); ok {
			model = &m
		}
	} else if m, ok := a.(M); ok {
		model = &m
	}
	if model == nil {
		return Res{}, false
	}
	source := beam.NewSourceEqual(*model, func(a, b M) bool {
		return reflect.DeepEqual(a, b)
	})
	res := n.h(w, r, source, sess.Store())
	if comp, ok := res.comp(); ok {
		inst, ok := instance.NewInstance(sess, n.a, source, comp, opt)
		if !ok {
			return Res{
				value: InstanceCreationError{},
			}, true
		}
		return Res{
			value: instValue{inst},
		}, true
	}
	return res, true
}

type compValue struct {
	comp gox.Comp
}

type instValue struct {
	inst instance.AnyInstance
}

type rerouteValue struct {
	model    any
	detached bool
}

type redirectValue struct {
	model  any
	status int
}

func ResApp(comp gox.Comp) Res {
	return Res{
		value: compValue{comp},
	}
}

func ResRedirect(model any, status int) Res {
	return Res{
		value: redirectValue{
			model:  model,
			status: status,
		},
	}
}

func ResReroute(model any, detached bool) Res {
	return Res{
		value: rerouteValue{
			model:    model,
			detached: detached,
		},
	}
}

type Res struct {
	value any
}

func (r Res) Value() any {
	return r.value
}

func (r Res) comp() (gox.Comp, bool) {
	if c, ok := r.value.(compValue); ok {
		return c.comp, true
	}
	return nil, false
}

func (r Res) Err() error {
	if c, ok := r.value.(error); ok {
		return c
	}
	return nil
}

func (r Res) Instance() (instance.AnyInstance, bool) {
	if c, ok := r.value.(instValue); ok {
		return c.inst, true
	}
	return nil, false
}

func (r Res) Reroute() (any, bool, bool) {
	if r, ok := r.value.(rerouteValue); ok {
		return r.model, r.detached, true
	}
	return nil, false, false
}

func (r Res) Redirect() (any, int, bool) {
	if r, ok := r.value.(redirectValue); ok {
		return r.model, r.status, true
	}
	return nil, 0, false
}

