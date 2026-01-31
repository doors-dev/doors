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
	"errors"
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

// SessionSave stores a key/value in session-scoped storage shared by all
// instances in the session. Returns the previous value under the key.
func SessionSave(ctx context.Context, key any, value any) any {
	return ctex.StoreSwap(ctx, ctex.KeySessionStore, key, value)
}

// SessionLoad gets a value from session-scoped storage by key.
// Returns nil if absent. Callers must type-assert the result.
func SessionLoad(ctx context.Context, key any) any {
	return ctex.StoreLoad(ctx, ctex.KeySessionStore, key)
}

// SessionRemove deletes a key/value from session-scoped storage.
// Returns the removed value or nil if absent.
func SessionRemove(ctx context.Context, key any) any {
	return ctex.StoreRemove(ctx, ctex.KeySessionStore, key)
}

// InstanceSave stores a key/value in instance-scoped storage.
// Returns the previous value.
func InstanceSave(ctx context.Context, key any, value any) any {
	return ctex.StoreSwap(ctx, ctex.KeyInstanceStore, key, value)
}

// InstanceLoad gets a value from instance-scoped storage by key.
// Returns nil if absent. Callers must type-assert the result.
func InstanceLoad(ctx context.Context, key any) any {
	return ctex.StoreLoad(ctx, ctex.KeyInstanceStore, key)
}

// InstanceRemove deletes a key/value from instance-scoped storage.
// Returns the removed value or nil if absent.
func InstanceRemove(ctx context.Context, key any) any {
	return ctex.StoreRemove(ctx, ctex.KeyInstanceStore, key)
}

// Location represents a URL built from a path model: path plus query.
// Use with navigation functions or href attributes.
type Location = path.Location

// NewLocation encodes model into a Location using the registered adapter
// for the model's type.
// Returns an error if no adapter is registered or encoding fails.
func NewLocation(ctx context.Context, model any) (Location, error) {
	adapters := ctx.Value(ctex.KeyAdapters).(map[string]path.AnyAdapter)
	name := path.GetAdapterName(model)
	adapter, ok := adapters[name]
	if !ok {
		var l Location
		return l, errors.New("adapter for " + name + " is not registered")
	}
	location, err := adapter.EncodeAny(model)
	if err != nil {
		var l Location
		return l, err
	}
	return *location, nil
}

// RandId returns a cryptographically secure, URL-safe random ID.
// Suitable for sessions, instances, tokens, attributes. Case-sensitive.
func RandId() string {
	return common.RandId()
}

// HashId creates ID using provided string, hashbased.
// For the same string outputs the same result.
// Suitable for HTML attributes.
func HashId(string string) string {
	hash := blake3.Sum256(common.AsBytes(string))
	hash[0] |= 0x80
	return base58.Encode(hash[:16])
}

// AllowBlocking returns a context that suppresses warnings when used
// with blocking X* operations. Use with caution.
func AllowBlocking(ctx context.Context) context.Context {
	return ctex.SetBlockingCtx(ctx)
}
