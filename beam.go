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

import "github.com/doors-dev/doors/internal/beam"

// Beam is a read-only reactive value.
//
// Use a [Beam] to read, subscribe to, or derive a smaller view of state. During
// one render/update cycle, a Door subtree observes one consistent value for the
// same beam.
type Beam[T any] = beam.Beam[T]

// Source is a writable [Beam].
//
// Create a source with [NewSource] or [NewSourceEqual], then derive smaller
// beams with [NewBeam] or [NewBeamEqual]. For reference types such as slices,
// maps, pointers, or mutable structs, replace the stored value instead of
// mutating it in place.
type Source[T any] = beam.Source[T]

// NewSource creates a [Source] that uses `==` to suppress equal updates.
//
// Example:
//
//	count := doors.NewSource(0)
func NewSource[T comparable](init T) Source[T] {
	return beam.NewSource(init)
}

// NewSourceEqual creates a [Source] with a custom equality function.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every update propagates.
func NewSourceEqual[T any](init T, equal func(new T, old T) bool) Source[T] {
	return beam.NewSourceEqual(init, equal)
}

// NewBeam derives a [Beam] from source and uses `==` to suppress equal
// derived values.
//
// Example:
//
//	fullName := doors.NewBeam(user, func(u User) string {
//		return u.FirstName + " " + u.LastName
//	})
func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2) Beam[T2] {
	return beam.NewBeam(source, cast)
}

// NewBeamEqual derives a [Beam] from source with a custom equality function.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every derived value propagates.
func NewBeamEqual[T any, T2 any](source Beam[T], cast func(T) T2, equal func(new T2, old T2) bool) Beam[T2] {
	return beam.NewBeamEqual(source, cast, equal)
}
