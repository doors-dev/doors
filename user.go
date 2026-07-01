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
	core := ctx.Value(common.KeyCore).(core.Core)
	core.Door().Reload(ctx)
}

// XReload tracks completion of [Reload].
//
// On success the channel sends two nil values then closes: the first means the
// call was scheduled (render was completed), the second means it was applied
// to the page.
// On failure it sends an error then closes.
// If ctx belongs to the root page render, it sends an error and closes.
//
// Do not wait on it during rendering. If you need to wait, use [Go] or your
// own goroutine with [DetachedContext].
func XReload(ctx context.Context) <-chan error {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Door().XReload(ctx)
}

// SessionExpire sets the maximum lifetime of the current session.
func SessionExpire(ctx context.Context, d time.Duration) {
	sess := ctx.Value(common.KeySession).(core.Session)
	sess.Expire(d)
}

// SessionEnd immediately ends the current session and all instances.
// Use it during logout to close authorized pages and free server resources.
func SessionEnd(ctx context.Context) {
	sess := ctx.Value(common.KeySession).(core.Session)
	sess.Kill()
}

// InstanceEnd ends the current instance (tab/window) but keeps the session and
// other instances active.
func InstanceEnd(ctx context.Context) {
	core := ctx.Value(common.KeyCore).(core.Core)
	core.Instance().Kill()
}

// InstanceId returns the unique ID of the current instance.
// Useful for logging, debugging, and tracking connections.
func InstanceId(ctx context.Context) string {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().ID()
}

// SessionId returns the unique ID of the current session.
// All instances in the same browser share this ID via a session cookie.
func SessionId(ctx context.Context) string {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.ID()
}

// InstanceLastSeen returns the time of the most recent client contact with the
// current instance (its latest live keep-alive).
func InstanceLastSeen(ctx context.Context) time.Time {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().LastSeen()
}

// SessionLastSeen returns the time of the most recent client contact with the
// current session (its latest renewal).
func SessionLastSeen(ctx context.Context) time.Time {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.LastSeen()
}

// Store is goroutine-safe key-value storage used for session and instance
// data.
type Store = ctex.Store

// SessionStore returns storage shared by all instances in the current session.
func SessionStore(ctx context.Context) Store {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.Store()
}

// InstanceStore returns storage scoped to the current instance only.
func InstanceStore(ctx context.Context) Store {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().Store()
}

// SessionContext returns a context that is canceled when the current session
// ends.
//
// It is useful for goroutines or external work that should live for the whole
// browser session, not just the current instance or dynamic owner.
//
// The returned context carries the current session value. It is suitable for
// session-scoped helpers, Door methods, and Source or Beam mutations. It is not
// tied to the current instance or dynamic owner, so it must not be used with
// instance-scoped helpers, [Reload], or Source and Beam reads/subscriptions.
func SessionContext(ctx context.Context) context.Context {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.Context()
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

// InstanceContext returns a context that is detached from the current dynamic
// owner and bounded by the current instance lifetime.
//
// The returned context keeps the original context values, clears the current
// render frame, and switches Doors ownership to the root of the current
// instance. It is safe to use with X-prefixed methods from goroutines.
//
// Use it for long-running goroutines and work that should outlive the current
// dynamic owner.
func InstanceContext(ctx context.Context) context.Context {
	core, ok := ctx.Value(common.KeyCore).(core.Core)
	if !ok {
		return ctex.NewFreeContext(ctx, ctx)
	}
	ctx = context.WithValue(ctx, common.KeyCore, core.Door().RootCore())
	return ctex.NewFreeContext(ctx, core.Instance().Runtime().Context())
}

// FreeRoot calls [InstanceContext].
//
// Deprecated: use [InstanceContext].
func FreeRoot(ctx context.Context) context.Context {
	return InstanceContext(ctx)
}

// DetachedContext returns a context that is detached from the current render
// frame and bounded by the current context lifetime.
//
// The returned context keeps the original context values and Doors ownership.
// It is safe to use with X-prefixed methods from goroutines when the work
// should stay scoped to the current dynamic owner.
func DetachedContext(ctx context.Context) context.Context {
	return ctex.NewFreeContext(ctx, ctx)
}

// Free calls [DetachedContext].
//
// Deprecated: use [DetachedContext].
func Free(ctx context.Context) context.Context {
	return DetachedContext(ctx)
}
