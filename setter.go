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
	"fmt"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/gox"
)

var _ Attr = (*Setter)(nil)

// Setter sets attributes on the elements it is attached to.
//
// Attach it as an attribute to one or more elements, then dispatch the action
// returned by [Setter.Set] with [Call]. Prefer [ActionEmit] when the client
// should own the DOM change.
//
// Setter is stateless: it changes live elements only. A rerendered element
// returns to its template attributes.
type Setter struct {
	id front.AutoID
}

func (s *Setter) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(s, cur, elem)
}

func (s *Setter) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(common.KeyCore).(core.Core)
	front.AttrsAppendSetter(attrs, s.id.ID(core))
	front.AttrsSetParent(attrs, core.Door().ID())
	return nil
}

// Set returns an action setting attribute name to value on attached elements;
// Into captures the number of live elements it reached, 0 when none is live
// (see [ActionInto]). An [Emitter] event counts hook requests instead.
//
// value follows template attribute semantics: nil and false remove the
// attribute, true sets it bare, [gox.Output] values serialize themselves,
// anything else is formatted with the fmt package. A [gox.Mutate] value, such
// as [Classes], fails the action: composing it needs the previous value, which
// only a rerender has.
func (s *Setter) Set(name string, value any) ActionInto[int] {
	return actionIntoFunc[int](func(ctx context.Context, core core.Core, gz bool) (action, error) {
		str, err := liveAttrValue(name, value)
		if err != nil {
			return action{}, err
		}
		return action{action: actions.AttrSet{
			ID:    s.id.ID(core),
			Name:  name,
			Value: str,
		}}, nil
	})
}

func liveAttrValue(name string, value any) (*string, error) {
	if _, ok := value.(gox.Mutate); ok {
		return nil, fmt.Errorf("attribute %s value implements gox.Mutate, which cannot compose on a live element", name)
	}
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		if !v {
			return nil, nil
		}
		return new(string), nil
	case gox.Output:
		var b strings.Builder
		if err := v.Output(&b); err != nil {
			return nil, err
		}
		str := b.String()
		return &str, nil
	default:
		str := fmt.Sprint(v)
		return &str, nil
	}
}
