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
	"log/slog"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/zeebo/xxh3"
)

// Reload rerenders the closest dynamic parent.
//
// The returned channel is optional to use and tracks completion. On success it
// sends nil when the rerender is scheduled and nil again when it is applied to
// the page, then closes. On failure it sends an error then closes. If ctx
// belongs to the root page render, it sends an error and closes.
//
// Do not wait on the channel during rendering. If you need to wait, use [Go] or
// your own goroutine with [DetachedContext].
//
// ctx must belong to a Doors render or handler; otherwise Reload panics.
func Reload(ctx context.Context) <-chan error {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Door().Reload(ctx)
}

// HasSession reports whether ctx carries a Doors session, meaning
// session-scoped helpers like [SessionID] and [SessionStore] accept it.
func HasSession(ctx context.Context) bool {
	_, ok := ctx.Value(common.KeySession).(core.Session)
	return ok
}

// HasInstance reports whether ctx belongs to a Doors render or handler,
// meaning instance- and door-scoped helpers accept it. Implies [HasSession].
func HasInstance(ctx context.Context) bool {
	_, ok := ctx.Value(common.KeyCore).(core.Core)
	return ok
}

// SessionExpire caps the remaining lifetime of the current session at d.
//
// It is a cap, not a keep-alive: the session still ends earlier if SessionTTL
// runs out first. Passing zero removes the cap.
//
// ctx must carry a Doors session; otherwise SessionExpire panics.
func SessionExpire(ctx context.Context, d time.Duration) {
	sess := ctx.Value(common.KeySession).(core.Session)
	sess.Expire(d)
}

// SessionEnd ends the current session and every instance in it immediately.
//
// Use it for a hard logout that must close every open page.
//
// ctx must carry a Doors session; otherwise SessionEnd panics.
func SessionEnd(ctx context.Context) {
	sess := ctx.Value(common.KeySession).(core.Session)
	sess.Kill()
}

// InstanceEnd ends the current instance immediately.
//
// Unlike [SessionEnd], the session and its other instances stay live.
//
// ctx must belong to a Doors render or handler; otherwise InstanceEnd panics.
func InstanceEnd(ctx context.Context) {
	core := ctx.Value(common.KeyCore).(core.Core)
	core.Instance().Kill()
}

// InstanceID returns the ID of the current instance.
//
// ctx must belong to a Doors render or handler; otherwise InstanceID panics.
func InstanceID(ctx context.Context) string {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().ID()
}

// SessionID returns the ID of the current session.
//
// Every instance in the same browser shares it, and it is the value of the
// session cookie.
//
// ctx must carry a Doors session; otherwise SessionID panics.
func SessionID(ctx context.Context) string {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.ID()
}

// InstanceLastSeen returns the time the current instance last saw its client on
// the live connection.
//
// Event and hook requests do not advance it, and it is the zero time during the
// initial page render.
//
// ctx must belong to a Doors render or handler; otherwise InstanceLastSeen
// panics.
func InstanceLastSeen(ctx context.Context) time.Time {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().LastSeen()
}

// SessionLastSeen returns the time of the most recent renewal of the current
// session.
//
// Renewals are throttled, so the value can lag behind the last request.
//
// ctx must carry a Doors session; otherwise SessionLastSeen panics.
func SessionLastSeen(ctx context.Context) time.Time {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.LastSeen()
}

// Store is goroutine-safe key-value storage, returned by [SessionStore] and
// [InstanceStore].
//
// Keys must be comparable; values may be of any type. A stored nil is
// indistinguishable from an absent key.
type Store = ctex.Store

// SessionStore returns the [Store] shared by every instance in the current
// session.
//
// ctx must carry a Doors session; otherwise SessionStore panics.
func SessionStore(ctx context.Context) Store {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.Store()
}

// InstanceStore returns the [Store] scoped to the current instance.
//
// Unlike [SessionStore], its contents are not visible to other instances.
//
// ctx must belong to a Doors render or handler; otherwise InstanceStore panics.
func InstanceStore(ctx context.Context) Store {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().Store()
}

// SessionContext returns a context that is canceled when the current session
// ends.
//
// It carries the session and nothing else: values of ctx are dropped, and it is
// tied to no instance or dynamic owner. Session-scoped helpers, Door methods,
// and Source or Beam updates accept it. [Reload] and instance-scoped helpers
// panic on it, and Source or Beam reads and subscriptions return false.
//
// ctx must carry a Doors session; otherwise SessionContext panics.
func SessionContext(ctx context.Context) context.Context {
	sess := ctx.Value(common.KeySession).(core.Session)
	return sess.Context()
}

// IDRand returns a cryptographically random, URL-safe identifier.
//
// Unlike [IDString] and [IDBytes], every call returns a new value.
func IDRand() string {
	return common.RandId()
}

// IDString returns a URL-safe identifier derived from the given string. Equal
// inputs always produce the same identifier.
//
// WARNING: it is not a cryptographic hash; do not use it to hide the input.
func IDString(string string) string {
	h := xxh3.New()
	h.WriteString(string)
	s := h.Sum128().Bytes()
	return common.EncodeId(s[:])
}

// IDBytes returns a URL-safe identifier derived from b. Equal inputs always
// produce the same identifier.
//
// WARNING: it is not a cryptographic hash; do not use it to hide the input.
func IDBytes(b []byte) string {
	h := xxh3.New()
	h.Write(b)
	s := h.Sum128().Bytes()
	return common.EncodeId(s[:])
}

// IDNumber returns a number unique within the current instance. Every call
// returns a new one, and the values are not sequential.
//
// ctx must belong to a [Door]; otherwise IDNumber panics.
func IDNumber(ctx context.Context) uint64 {
	core := ctx.Value(common.KeyCore).(core.Core)
	return core.Instance().NewID()
}

// InstanceContext returns a context that is detached from the current dynamic
// owner and bounded by the current instance lifetime.
//
// It keeps the values of ctx and moves Doors ownership to the root of the
// current instance, so subscriptions and hooks made with it survive rerenders
// of the caller. It is safe to use from goroutines that wait on completion
// channels.
//
// Unlike [DetachedContext], it outlives the current dynamic owner. If ctx does
// not belong to an instance, the two are equivalent.
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

// DetachedContext returns a context that is detached from the current
// render/update cycle and bounded by the lifetime of ctx.
//
// It keeps the values of ctx and the current Doors ownership. It is safe to use
// from goroutines that wait on completion channels.
//
// Unlike [InstanceContext], work started with it is cleaned up with the current
// dynamic owner. Any context is allowed.
func DetachedContext(ctx context.Context) context.Context {
	return ctex.NewFreeContext(ctx, ctx)
}

// Free calls [DetachedContext].
//
// Deprecated: use [DetachedContext].
func Free(ctx context.Context) context.Context {
	return DetachedContext(ctx)
}

// Logger returns the Doors logger for ctx.
//
// Any context is allowed; one that carries no Doors session falls back to
// [slog.Default].
func Logger(ctx context.Context) *slog.Logger {
	return common.Logger(ctx)
}
