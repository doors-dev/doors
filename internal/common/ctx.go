// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package common

import (
	"context"
	"log/slog"
)

type ctxKey int

const (
	CtxKeyInstance ctxKey = iota
	CtxKeyDoor
	CtxKeyThread
	CtxKeyRenderMap
	CtxKeyBlocking
	CtxKeyAdapters
	CtxKeySessionStore
	CtxKeyInstanceStore
	CtxKeyParent
	CtxKeyAttrs
	CtxStorageKeyStatus
)

func LogCanceled(ctx context.Context, action string)  {
	ch := make(chan error, 1)
	if ctx.Err() != nil {
		ch <- context.Canceled
		slog.Warn("Tried to perfrom " + action + " from the canceled context")
	}
}
