// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package printer

import (
	"io"

	"github.com/doors-dev/gox"
)

type defaultPrinter struct {
	w io.Writer
}

func (d defaultPrinter) Send(job gox.Job) error {
	return job.Output(d.w)
}
