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

package doors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type joinedAttrs struct {
	attrs gox.Attrs
}

func (j joinedAttrs) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(j, cur, elem)
}

func (j joinedAttrs) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	attrs.Inherit(j.attrs)
	return nil
}

// Attr is a Doors attribute modifier that can be attached directly to an
// element or applied through a proxy component.
type Attr interface {
	gox.Modify
	gox.Proxy
}

// A combines one or more [Attr] values into a single modifier.
//
// Example:
//
//	attrs := doors.A(ctx,
//		doors.AClick{On: onClick},
//		doors.AData{Name: "user", Value: user},
//	)
func A(ctx context.Context, a ...Attr) Attr {
	attrs := gox.NewAttrs()
	for _, mod := range a {
		attrs.AddMod(mod)
	}
	attrs.ApplyMods(ctx, "")
	return joinedAttrs{attrs: attrs}
}

type eventAttr[E any] struct {
	capture   front.Capture
	onError   []Action
	before    []Action
	scope     []Scope
	indicator []Indicator
	on        func(context.Context, RequestEvent[E]) bool
}

func (p eventAttr[E]) apply(ctx context.Context, attrs gox.Attrs) error {
	c := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := c.RegisterHook(p.handle, nil)
	if !ok {
		return errors.New("door: hook registration failed")
	}
	front.AttrsAppendCapture(attrs, p.capture, front.Hook{
		OnError:  intoActions(ctx, p.onError),
		Before:   intoActions(ctx, p.before),
		Scope:    front.IntoScopeSet(c, p.scope),
		Indicate: front.IntoIndicate(p.indicator),
		Hook:     hook,
	})
	return nil
}

func (p *eventAttr[E]) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	var e E
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&e)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(400)
		return false
	}
	return p.on(ctx, &eventRequest[E]{
		request: request{
			r:   r,
			w:   w,
			ctx: ctx,
		},
		e: &e,
	})
}
