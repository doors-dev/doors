// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package internal

import (
	"embed"
	"io/fs"
)

//go:embed client/*
var clientSrc embed.FS


var ClientSrc fs.FS
var ClientStyles []byte

func init() {
    ClientSrc, _ = fs.Sub(clientSrc, "client")
    ClientStyles, _ = fs.ReadFile(ClientSrc, "style.css")
}
