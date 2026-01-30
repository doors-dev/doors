// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import "github.com/doors-dev/doors/internal/beam"

// Beam represents a reactive value stream that can be read, subscribed to, or watched.
//
// When used in a render cycle, it is guaranteed that a Door and all of its children
// will observe the exact same value for a given Beam. This ensures stable and predictable
// rendering behavior, even when multiple components depend on the same reactive source.
type Beam[T any] = beam.Beam[T]

// SourceBeam is the initial Beam (others are derived from it), which, in addition to its core
// functionality, includes the ability to update values and propagate changes to all
// subscribers and derived beams. It serves as the root of a reactive value chain.
// Updates and mutations are synchronized across all subscribers, ensuring consistent
// state during rendering cycles. During a render cycle, all consumers will see a
// consistent view of the latest value. The source maintains a sequence of values for
// synchronization purposes.
//
// IMPORTANT: For reference types (slices, maps, pointers, structs), do not modify
// the data directly. Instead, create or provide a different instance.
// Direct modification can break the consistency guarantees since subscribers may
// observe partial changes or inconsistent state.
type SourceBeam[T any] = beam.SourceBeam[T]

// NewSourceBeam creates a new SourceBeam with the given initial value.
// Updates are only propagated when the new value passes the default distinct
// function with != comparison to the old value
//
// Parameters:
//   - init: the initial value for the SourceBeam
//
// Returns:
//   - A new SourceBeam[T] instance
func NewSourceBeam[T comparable](init T) SourceBeam[T] {
	return beam.NewSourceBeam(init)
}

// NewSourceBeamEqual creates a new SourceBeam with a custom equality function.
//
// The equality function receives new and old values and should return true
// if the new value is considered different and should be propagated to subscribers.
// If equality is nil, every update will be propagated regardless of value equality.
//
// Parameters:
//   - init: the initial value for the SourceBeam
//   - equality: a function to determine if transformed values should propagate (equal values ignored)
//     or nil to always propagate
//
// Returns:
//   - A new SourceBeam[T] instance that uses the equality function for update filtering
func NewSourceBeamEqual[T any](init T, equal func(new T, old T) bool) SourceBeam[T] {
	return beam.NewSourceBeamEqual(init, equal)
}

// NewBeam derives a new Beam[T2] from an existing Beam[T] by applying a transformation function.
//
// The cast function maps values from the source beam2.to the derived beam.
// Updates are only propagated when the new value passes the default equality
// function with != comparison to the old value
//
// Parameters:
//   - source: the source Beam[T] to derive from
//   - cast: a function that transforms values from type T to type T2
//
// Returns:
//   - A new Beam[T2] that emits transformed values when they differ from the previous value
func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2) Beam[T2] {
	return beam.NewBeam(source, cast)
}

// NewBeamEqual derives a new Beam[T2] from an existing Beam[T] using custom transformation and filtering.
//
// The cast function transforms source values from type T to type T2. The equality function
// determines whether updated values should be propagated by comparing new and old values.
// If equality function is nil, every transformation will be propagated regardless of value equality.
//
// Parameters:
//   - source: the source Beam[T] to derive from
//   - cast: a function to transform T → T2
//   - equality: a function to determine if transformed values should propagate (equal values ignored)
//     or nil to always propagate
//
// Returns:
//   - A new Beam[T2] that emits transformed values filtered by the equality function
func NewBeamEqual[T any, T2 any](source Beam[T], cast func(T) T2, equal func(new T2, old T2) bool) Beam[T2] {
	return beam.NewBeamEqual(source, cast, equal)
}
