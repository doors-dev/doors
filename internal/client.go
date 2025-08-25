// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

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
