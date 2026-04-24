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

package door

import (
	"context"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

// Update replaces the door's current children while keeping the same door
// container mounted.
//
// Deprecated: use [Door.Inner].
func (d *Door) Update(ctx context.Context, content any) {
	d.inner(ctx, content)
}

// Rebase replaces the rendered door with el while keeping the same Go [Door]
// handle alive for later updates.
//
// Deprecated: use [Door.Outer].
func (d *Door) Rebase(ctx context.Context, el gox.Elem) {
	d.outer(ctx, el)
}

// Replace removes the current door container and replaces it with content.
//
// Deprecated: use [Door.Static].
func (d *Door) Replace(ctx context.Context, content any) {
	d.static(ctx, content)
}

// Delete removes the door and forgets its current content.
//
// Deprecated: use [Door.Static] with nil.
func (d *Door) Delete(ctx context.Context) {
	d.static(ctx, nil)
}

// Clear removes all door content while keeping the container mounted.
// It is equivalent to Update(ctx, nil).
//
// Deprecated: use [Door.Inner] with nil.
func (d *Door) Clear(ctx context.Context) {
	d.inner(ctx, nil)
}

// XUpdate tracks completion of [Door.Update].
//
// Deprecated: use [Door.XInner].
func (d *Door) XUpdate(ctx context.Context, content any) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XUpdate")
	return d.inner(ctx, content)
}

// XRebase tracks completion of [Door.Rebase].
//
// Deprecated: use [Door.XOuter].
func (d *Door) XRebase(ctx context.Context, el gox.Elem) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XRebase")
	return d.outer(ctx, el)
}

// XReplace tracks completion of [Door.Replace].
//
// Deprecated: use [Door.XStatic].
func (d *Door) XReplace(ctx context.Context, content any) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XReplace")
	return d.static(ctx, content)
}

// XDelete tracks completion of [Door.Delete].
//
// Deprecated: use [Door.XStatic] with nil.
func (d *Door) XDelete(ctx context.Context) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XDelete")
	return d.static(ctx, nil)
}

// XClear tracks completion of [Door.Clear].
//
// Deprecated: use [Door.XInner] with nil.
func (d *Door) XClear(ctx context.Context) <-chan error {
	ctex.LogFreeWarning(ctx, "Door", "XClear")
	return d.inner(ctx, nil)
}
