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

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/instance"
)

// TabState returns a writable source of the value stored under key in the
// browser tab's state. It uses == to suppress equal updates.
//
// The stored value is not available during the initial page render: the
// browser sends it only after the page has loaded, so the source reads nil
// until then. Content that depends on tab state should render a placeholder
// and let the update fill it in.
//
// Tab state is kept by the browser tab: it survives reloads and in-app
// navigation, back and forward keep the latest value, other tabs do not see
// it, and it never appears in the URL. Every server-side update is sent back
// to the browser. Values are stored as JSON. Keep them small: browsers cap
// history state size, and the whole tab state travels on every page load and
// update.
//
// Until the sync, the value is nil unless an update set it. After the sync it
// is never nil: a missing or undecodable key reads as the zero value. Updating
// with nil removes the key. A non-nil value set before the sync wins over the
// stored one.
//
// ctx must belong to a Doors render or handler; otherwise TabState panics.
func TabState[T comparable](ctx context.Context, key string) Source[*T] {
	return TabStateEqual(ctx, key, beam.DefaultEqual[T])
}

// TabStateEqual returns a writable source of the value stored under key in
// the browser tab's state. It calls equal to suppress equal updates.
//
// equal receives dereferenced values; nil and non-nil never compare equal. If
// equal is nil, every update propagates. equal runs while the source's lock is
// held: it must not panic and must not call back into any source or beam.
//
// The stored value is not available during the initial page render: the
// browser sends it only after the page has loaded, so the source reads nil
// until then. Content that depends on tab state should render a placeholder
// and let the update fill it in.
//
// Tab state is kept by the browser tab: it survives reloads and in-app
// navigation, back and forward keep the latest value, other tabs do not see
// it, and it never appears in the URL. Every server-side update is sent back
// to the browser. Values are stored as JSON. Keep them small: browsers cap
// history state size, and the whole tab state travels on every page load and
// update.
//
// Until the sync, the value is nil unless an update set it. After the sync it
// is never nil: a missing or undecodable key reads as the zero value. Updating
// with nil removes the key. A non-nil value set before the sync wins over the
// stored one.
//
// ctx must belong to a Doors render or handler; otherwise TabStateEqual panics.
func TabStateEqual[T any](ctx context.Context, key string, equal func(old T, new T) bool) Source[*T] {
	if equal == nil {
		equal = beam.NeverEqual[T]
	}
	core := ctx.Value(common.KeyCore).(core.Core)
	s := &instance.TabStateDerive[T]{
		Key:   key,
		Equal: equal,
	}
	core.Instance().TabState(s)
	return derivedSource[instance.TabState, *T]{s.Get()}
}
