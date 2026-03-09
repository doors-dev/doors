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
	"github.com/mr-tron/base58"
	"github.com/zeebo/blake3"
)

// SessionExpire sets the maximum lifetime of the current session.
func SessionExpire(ctx context.Context, d time.Duration) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.SessionExpire(d)
}

// SessionEnd immediately ends the current session and all instances.
// Use during logout to close authorized pages and free server resources.
func SessionEnd(ctx context.Context) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.SessionEnd()
}

// InstanceEnd ends the current instance (tab/window) but keeps the session
// and other instances active.
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

type Store = ctex.Store

// SessionStore returns session-scoped goroutine-safe
// storage
func SessionStore(ctx context.Context) Store {
	return ctx.Value(ctex.KeySessionStore).(Store)
}

// InstanceStore returns instance-scoped goroutine-safe
// storage
func InstanceStore(ctx context.Context) Store {
	return ctx.Value(ctex.KeySessionStore).(Store)
}

// Location represents a URL built from a path model: path plus query.
// Use with navigation functions or href attributes.
type Location = path.Location

// NewLocation encodes model into a Location using the registered adapter
// for the model's type.
// Returns an error if no adapter is registered or encoding fails.
func NewLocation(ctx context.Context, model any) (Location, error) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	location, err := core.Adapters().Encode(model)
	if err != nil {
		var l Location
		return l, err
	}
	return location, nil
}

// IdRand returns a cryptographically secure, URL-safe random ID.
// Suitable for sessions, instances, tokens, attributes. Case-sensitive.
func IdRand() string {
	return common.RandId()
}

// IdString creates Id using provided string, hashbased.
// For the same string outputs the same result.
// Suitable for HTML attributes.
func IdString(string string) string {
	hasher := blake3.New()
	hasher.WriteString(string)
	hash := hasher.Sum(nil)
	return common.Cut(base58.Encode(hash[:]))
}

// IdBytes creates Id using provided bytes, hashbased.
// For the same bytes outputs the same result.
// Suitable for HTML attributes.
func IdBytes(b []byte) string {
	hash := blake3.Sum256(b)
	return common.Cut(base58.Encode(hash[:]))
}

// AllowBlocking returns a context that suppresses warnings when used
// with blocking X* operations. Use with caution.
func AllowBlocking(ctx context.Context) context.Context {
	return ctex.SetBlockingCtx(ctx)
}
