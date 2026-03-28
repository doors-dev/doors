// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/path"
	"github.com/zeebo/blake3"
)

// SessionExpire sets the maximum lifetime of the current session.
func SessionExpire(ctx context.Context, d time.Duration) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.SessionExpire(d)
}

// SessionEnd immediately ends the current session and all instances.
// Use it during logout to close authorized pages and free server resources.
func SessionEnd(ctx context.Context) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.SessionEnd()
}

// InstanceEnd ends the current instance (tab/window) but keeps the session and
// other instances active.
func InstanceEnd(ctx context.Context) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.InstanceEnd()
}

// InstanceId returns the unique ID of the current instance.
// Useful for logging, debugging, and tracking connections.
func InstanceId(ctx context.Context) string {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	return core.InstanceID()
}

// SessionId returns the unique ID of the current session.
// All instances in the same browser share this ID via a session cookie.
func SessionId(ctx context.Context) string {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	return core.SessionID()
}

// Store is goroutine-safe key-value storage used for session and instance
// state.
type Store = ctex.Store

// SessionStore returns storage shared by all instances in the current session.
func SessionStore(ctx context.Context) Store {
	return ctx.Value(ctex.KeySessionStore).(Store)
}

// InstanceStore returns storage scoped to the current instance only.
func InstanceStore(ctx context.Context) Store {
	return ctx.Value(ctex.KeyInstanceStore).(Store)
}

// Location is a parsed or generated URL path plus query string.
type Location = path.Location

// NewLocation encodes model into a [Location] using the registered adapter for
// the model's type. It returns an error if no adapter is registered or
// encoding fails.
func NewLocation(ctx context.Context, model any) (Location, error) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	location, err := core.Adapters().Encode(model)
	if err != nil {
		var l Location
		return l, err
	}
	return location, nil
}

// IDRand returns a cryptographically secure, URL-safe identifier.
func IDRand() string {
	return common.RandId()
}

// IDString returns a stable URL-safe identifier derived from string.
func IDString(string string) string {
	hasher := blake3.New()
	hasher.WriteString(string)
	hash := hasher.Sum(nil)
	return common.EncodeId(hash)
}

// IDBytes returns a stable URL-safe identifier derived from b.
func IDBytes(b []byte) string {
	hash := blake3.Sum256(b)
	return common.EncodeId(hash[:])
}

// Free returns a context that is safe to use with extended Doors operations
// that may wait for asynchronous completion, such as X-prefixed methods, and
// with long-running goroutines tied to the page runtime.
//
// The returned context keeps the original Values from ctx, but uses the root
// Doors context for framework features such as beam reads/subscriptions and
// uses the instance runtime for cancellation, deadline, and lifetime.
func Free(ctx context.Context) context.Context {
	core, ok := ctx.Value(ctex.KeyCore).(core.Core)
	if !ok {
		return ctex.FreeContext(ctx, ctx)
	}
	ctx = context.WithValue(ctx, ctex.KeyCore, core.RootCore())
	return ctex.FreeContext(ctx, core.Runtime().Context())
}
