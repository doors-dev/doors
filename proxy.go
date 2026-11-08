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
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/gox"
)

// ProxyMod returns a proxy that applies mod to the first element in the
// wrapped subtree.
//
// It looks through components, containers, and [Parallel] markers to reach
// that element and leaves later siblings unchanged. In front of a [Door] it
// applies to the Door's frame, or to the first element of static Door content.
// If text or other non-element output comes first, the mod is skipped and an
// error is logged. A nil mod renders the subtree unchanged.
func ProxyMod(mod gox.Modify) gox.Proxy {
	return gox.ProxyFunc(func(cur gox.Cursor, el gox.Elem) error {
		if mod == nil {
			return el(cur)
		}
		return el(gox.NewCursor(cur.Context(), printer.NewModPrinter(cur.Printer(), []gox.Modify{mod})))
	})
}

func proxyMod(mod gox.Modify, cur gox.Cursor, elem gox.Elem) error {
	proxy := ProxyMod(mod)
	return proxy.Proxy(cur, elem)
}
