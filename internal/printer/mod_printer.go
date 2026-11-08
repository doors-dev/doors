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

package printer

import (
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/gox"
)

type ModJob interface {
	AddMod(m gox.Modify)
}

func NewModPrinter(printer gox.Printer, mods []gox.Modify) gox.Printer {
	return &modPrinter{mods: mods, printer: printer}
}

type modPrinter struct {
	mods    []gox.Modify
	printer gox.Printer
}

func (m *modPrinter) Send(j gox.Job) error {
	if len(m.mods) == 0 {
		return m.printer.Send(j)
	}
	if open, ok := j.(*gox.JobOpen); ok && open.Kind == gox.KindContainer {
		return m.printer.Send(open)
	}
	mods := m.mods
	m.mods = nil
	switch job := j.(type) {
	case ModJob:
		for _, mod := range mods {
			job.AddMod(mod)
		}
		return m.printer.Send(j)
	case *gox.JobOpen:
		for _, mod := range mods {
			job.Attrs.AddMod(mod)
		}
		return m.printer.Send(job)
	default:
		common.Logger(j.Context()).Error("ProxyMod skipped: no element to modify")
		return m.printer.Send(j)
	}
}
