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
	"errors"

	"github.com/doors-dev/gox"
)

// ProxyMod returns a proxy that applies mod to the first element in the
// wrapped subtree.
//
// It looks through components, containers, and [Parallel] markers to reach
// that element and leaves later siblings unchanged. It does not work in front
// of a [Door]: a Door, text, or other non-element output where that element is
// expected is an error. A nil mod renders the subtree unchanged.
func ProxyMod(mod gox.Modify) gox.Proxy {
	return gox.ProxyFunc(func(cur gox.Cursor, el gox.Elem) error {
		printer := &modPrinter{
			mod:     mod,
			printer: cur.Printer(),
		}
		cursor := gox.NewCursor(cur.Context(), printer)
		return el(cursor)
	})
}

func proxyMod(mod gox.Modify, cur gox.Cursor, elem gox.Elem) error {
	proxy := ProxyMod(mod)
	return proxy.Proxy(cur, elem)
}

type modPrinter struct {
	mod     gox.Modify
	printer gox.Printer
}

func (m *modPrinter) Send(j gox.Job) error {
	if m.mod == nil {
		return m.printer.Send(j)
	}
	if par, ok := j.(parallelJob); ok {
		mod := m.mod
		m.mod = nil
		return m.printParallel(mod, par)
	}
	comp, ok := j.(*gox.JobComp)
	if ok {
		mod := m.mod
		m.mod = nil
		return m.printComp(mod, comp)
	}
	open, ok := j.(*gox.JobHeadOpen)
	if ok {
		if open.Kind == gox.KindContainer {
			return m.printer.Send(open)
		}
		mod := m.mod
		m.mod = nil
		return m.printHead(mod, open)
	}
	m.mod = nil
	return errors.New("cannot attach an attribute modifier: unexpected job type")
}

func (m *modPrinter) printParallel(mod gox.Modify, job parallelJob) error {
	el := job.el
	job.el = gox.Elem(func(cur gox.Cursor) error {
		p := &modPrinter{mod: mod, printer: cur.Printer()}
		cur = gox.NewCursor(cur.Context(), p)
		return el(cur)
	})
	return m.printer.Send(job)
}

func (m *modPrinter) printComp(mod gox.Modify, job *gox.JobComp) error {
	ctx := job.Ctx
	comp := job.Comp
	gox.Release(job)
	return m.submitComp(mod, ctx, comp)
}

func (m *modPrinter) submitComp(mod gox.Modify, ctx context.Context, comp gox.Comp) error {
	return m.printer.Send(gox.NewJobComp(ctx, gox.Elem(func(cur gox.Cursor) error {
		el := comp.Main()
		if el == nil {
			return nil
		}
		p := &modPrinter{mod: mod, printer: cur.Printer()}
		cur = gox.NewCursor(cur.Context(), p)
		return el(cur)
	})))
}

func (m *modPrinter) printHead(mod gox.Modify, job *gox.JobHeadOpen) error {
	job.Attrs.AddMod(mod)
	return m.printer.Send(job)
}
