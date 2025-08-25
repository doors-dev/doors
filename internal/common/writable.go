// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

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
