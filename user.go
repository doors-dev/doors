// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doors

import (
	"context"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/zeebo/xxh3"
)

// Reload rerenders the closest dynamic parent.
//
// If ctx belongs to the root page render, nothing is reloaded.
func Reload(ctx context.Context) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	core.Reload(ctx)
}

// XReload tracks completion of [Reload].
//
// The channel receives nil on success or an error on failure, then closes.
// If ctx belongs to the root page render, it sends an error and closes.
//
// Do not wait on it during rendering. If you need to wait, use [Go] or your
// own goroutine with [Free].
func XReload(ctx context.Context) <-chan error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	return core.XReload(ctx)
}

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
// data.
type Store = ctex.Store

// SessionStore returns storage shared by all instances in the current session.
func SessionStore(ctx context.Context) Store {
	return ctx.Value(ctex.KeySessionStore).(ctex.Store)
}

// InstanceStore returns storage scoped to the current instance only.
func InstanceStore(ctx context.Context) Store {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	return core.Instance().Store()
}

// SessionContext returns a context that is canceled when the current session
// ends.
//
// It is useful for goroutines or external work that should live for the whole
// browser session, not just the current instance or dynamic owner.
func SessionContext(ctx context.Context) context.Context {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	return core.SessionContext()
}

// IDRand returns a cryptographically secure, URL-safe identifier.
func IDRand() string {
	return common.RandId()
}

// IDString returns a stable URL-safe identifier derived from string.
func IDString(string string) string {
	h := xxh3.New()
	h.WriteString(string)
	s := h.Sum128().Bytes()
	return common.EncodeId(s[:])
}

// IDBytes returns a stable URL-safe identifier derived from b.
func IDBytes(b []byte) string {
	h := xxh3.New()
	h.Write(b)
	s := h.Sum128().Bytes()
	return common.EncodeId(s[:])
}

// FreeRoot returns a free context that is safe to use with extended Doors
// operations that may wait, such as X-prefixed methods.
//
// The returned context keeps the original Values from ctx, switches framework
// features and lifetime to the root Doors context.
//
// Use it for long-running goroutines and work that should outlive the current
// dynamic owner.
func FreeRoot(ctx context.Context) context.Context {
	core, ok := ctx.Value(ctex.KeyCore).(core.Core)
	if !ok {
		return ctex.NewFreeContext(ctx, ctx)
	}
	ctx = context.WithValue(ctx, ctex.KeyCore, core.RootCore())
	return ctex.NewFreeContext(ctx, core.Runtime().Context())
}

// Free returns a free context that is safe to use with extended Doors
// operations that may wait, such as X-prefixed methods.
//
// The returned context keeps the original Values from ctx together with the
// current dynamic ownership and lifecycle.
//
// Use it when waiting should stay scoped to the current dynamic owner.
func Free(ctx context.Context) context.Context {
	return ctex.NewFreeContext(ctx, ctx)
}
