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
// Attach it as an attribute to one or more elements, then pass the action
// returned by [Setter.Set] to [Call] or any action-accepting API.
//
// Setter is stateless: it changes live elements only. A re-rendered element
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

// Set returns an action setting attribute name on attached elements; its
// Into captures the number of affected elements (see [ActionInto]).
//
// The value follows template attribute semantics: nil and false remove the
// attribute, true sets it bare, [gox.Output] values serialize themselves,
// anything else is formatted with the fmt package. A [gox.Mutate] value, such
// as [Classes], fails the action: composing it needs the previous value, which
// only a rerender has.
func (s *Setter) Set(name string, value any) ActionInto[int] {
	return actionIntoFunc[int](func(ctx context.Context, core core.Core, gz bool) (action, error) {
		if _, ok := value.(gox.Mutate); ok {
			return action{}, fmt.Errorf("setter: attribute %s value implements gox.Mutate, which cannot compose on a live element", name)
		}
		act := actions.AttrSet{
			ID:   s.id.ID(core),
			Name: name,
		}
		switch v := value.(type) {
		case nil:
			return action{action: act}, nil
		case bool:
			if !v {
				return action{action: act}, nil
			}
			act.Value = new(string)
		case gox.Output:
			var b strings.Builder
			if err := v.Output(&b); err != nil {
				return action{}, err
			}
			str := b.String()
			act.Value = &str
		default:
			str := fmt.Sprint(v)
			act.Value = &str
		}
		return action{action: act}, nil
	})
}
