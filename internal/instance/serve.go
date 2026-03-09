// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/printer"
)

func (inst *Instance[M]) Serve(w http.ResponseWriter, r *http.Request) error {
	if err := inst.init(); err != nil {
		return err
	}
	stack, err := inst.root.Render(inst.setup.comp)
	inst.setup = nil
	if err != nil {
		inst.end(common.EndCauseKilled)
		return err
	}
	static := inst.root.IsStatic()
	if err := inst.render(w, r, stack, static); err != nil {
		inst.end(common.EndCauseKilled)
		return nil
	}
	if static {
		inst.end(common.EndCauseKilled)
		return nil
	}
	inst.navigator.init()
	return nil
}

func (inst *Instance[M]) render(w http.ResponseWriter, r *http.Request, pipe door.Stack, static bool) error {
	gz := !inst.Conf().ServerDisableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	importMap, importHash := inst.importMap.generate()
	inst.renderHeaders(w, gz, importHash)
	var writer io.Writer = w
	if gz {
		wgz := gzip.NewWriter(w)
		defer wgz.Close()
		writer = wgz
	}
	pr := printer.NewPagePrinter(writer, inst.root.Context(), static, importMap)
	return pipe.Print(pr)
}

func (inst *Instance[M]) renderHeaders(w http.ResponseWriter, gz bool, importHash []byte) {
	if inst.csp != nil {
		if importHash != nil {
			inst.csp.ScriptHash(importHash)
		}
		header := inst.csp.Generate()
		w.Header().Add("Content-Security-Policy", header)
		inst.csp = nil
	}
	if gz {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.WriteHeader(inst.getStatus())
}
