// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package ctex

type ctxKey int

const (
	KeyCore ctxKey = iota
	KeySessionStore
	KeyInstanceStore
	KeyBlocking
	KeyAdapters
	keyWg
)

