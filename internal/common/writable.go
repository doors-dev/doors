// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package common

import (
	"io"
)

type WritableNone struct{}

func (w WritableNone) Destroy() {

}

func (w WritableNone) Write(io.Writer) error {
	return nil
}

type Writable interface {
	Destroy()
	Write(io.Writer) error
}

type WritableRenderMap struct {
	Rm    *RenderMap
	Index uint64
}

func (wrm *WritableRenderMap) Destroy() {
	wrm.Rm.Destroy()
}
func (wrm *WritableRenderMap) Write(w io.Writer) error {
	return wrm.Rm.Render(w, wrm.Index)
}
