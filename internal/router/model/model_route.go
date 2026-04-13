// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// AnyModelRoute is the internal representation of a registered page model.
type AnyModelRoute interface {
	Adapter() path.AnyAdapter
	Handle(w http.ResponseWriter, r *http.Request, a any, sess *instance.Session, opt instance.Options) (Res, bool)
}

type Handler[M any] = func(w http.ResponseWriter, r *http.Request, source beam.Source[M], store ctex.Store) Res

// NewModelRoute creates a model-backed route from a path adapter and handler.
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
	model, ok := n.a.Decode(a)
	if !ok {
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
				entity: InstanceCreationError{},
			}, true
		}
		return Res{
			entity: newInstance{inst},
		}, true
	}
	return res, true
}

type component struct {
	comp gox.Comp
}

type newInstance struct {
	inst instance.AnyInstance
}

type reroute struct {
	model any
}

type redirect struct {
	model  any
	status int
}

// ResComp returns a [Res] that renders comp.
func ResComp(comp gox.Comp) Res {
	return Res{
		entity: component{comp},
	}
}

// ResRedirect returns a [Res] that redirects to model.
func ResRedirect(model any, status int) Res {
	return Res{
		entity: redirect{
			model:  model,
			status: status,
		},
	}
}

// ResReroute returns a [Res] that reroutes to model on the server.
func ResReroute(model any) Res {
	return Res{
		entity: reroute{
			model: model,
		},
	}
}

// Res describes the result of handling a registered model route.
type Res struct {
	entity any
}

// Entity returns the raw underlying response payload.
func (r Res) Entity() any {
	return r.entity
}

func (r Res) comp() (gox.Comp, bool) {
	if c, ok := r.entity.(component); ok {
		return c.comp, true
	}
	return nil, false
}

// Err returns the response error, if any.
func (r Res) Err() error {
	if c, ok := r.entity.(error); ok {
		return c
	}
	return nil
}

// Instance returns the created page instance, if this response rendered one.
func (r Res) Instance() (instance.AnyInstance, bool) {
	if c, ok := r.entity.(newInstance); ok {
		return c.inst, true
	}
	return nil, false
}

// Reroute returns the reroute target model, if any.
func (r Res) Reroute() (any, bool) {
	if r, ok := r.entity.(reroute); ok {
		return r.model, true
	}
	return nil, false
}

// Redirect returns the redirect target model and status, if any.
func (r Res) Redirect() (any, int, bool) {
	if r, ok := r.entity.(redirect); ok {
		return r.model, r.status, true
	}
	return nil, 0, false
}
