// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package common

import "io"

func NewJsonWriter(w io.Writer) io.Writer {
	return jsonWriter{w: w}
}

type jsonWriter struct {
	w io.Writer
}

func (j jsonWriter) Write(b []byte) (n int, err error) {
	adj := 0
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
		adj = 1
	}
	n, err = j.w.Write(b)
	if err != nil {
		return
	}
	n += adj
	return
}
