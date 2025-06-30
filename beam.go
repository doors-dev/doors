package doors

import "github.com/doors-dev/doors/internal/beam"

// Beam represents a reactive value stream that can be read, subscribed to, or watched.
//
// When used in a render cycle, it is guaranteed that a Node and all of its children
// will observe the exact same value for a given Beam. This ensures stable and predictable
// rendering behavior, even when multiple components depend on the same reactive source.
type Beam[T any] = beam.Beam[T]

// SourceBeam represents a writable, reactive data stream that allows both observation and mutation.
// It embeds Beam[T], supporting subscription, reading, and lifecycle consistency as described in Beam.
//
// Updates to a SourceBeam propagate to all derived beams. During a render cycle,
// all consumers will see a consistent view of the latest value.
type SourceBeam[T any] = beam.SourceBeam[T]

// NewSourceBeam creates a new SourceBeam with the given initial value.
//
// If distinct is true, updates or mutations will only trigger propagation when
// the value has actually changed — as determined by reflect.DeepEqual.
// This helps avoid unnecessary updates when the value remains structurally identical.
//
// Parameters:
//   - init: the initial value of the beam.
//   - distinct: if true, updates are only propagated when the value differs from the previous one.
//
// Returns:
//   - A new SourceBeam[T] instance
func NewSourceBeam[T any](init T, distinct bool) SourceBeam[T] {
	return beam.NewSourceBeam(init, distinct)
}

// NewSourceBeamExt creates a new SourceBeam with a custom update condition.
//
// Unlike NewSourceBeam, which uses == check to suppress duplicate updates,
// this version accepts an updateIf function to determine whether a new value should
// be propagated to subscribers.
//
// The updateIf function receives pointers to the new and previous values and should
// return true if the new value is considered different.
//
//
// Parameters:
//   - init: the initial value for the SourceBeam.
//   - updateIf: a custom function to determine if a new value should trigger propagation
//   or nil to tiggger every time
//
// Returns:
//   - A new SourceBeam[T] instance that uses updateIf for update comparisons.
func NewSourceBeamExt[T any](init T, updateIf func(new *T, old *T) bool) SourceBeam[T] {
	return beam.NewSourceBeamExt(init, updateIf)
}

// NewBeam derives a new Beam[T2] from an existing Beam[T] by applying a transformation function.
//
// The cast function maps values from the source beam to the derived beam. The derived beam
// will receive updates whenever the source beam updates.
//
// If distinct is true, the derived beam will only emit updates when the casted value changes,
// using == to compare with the previous value. This avoids redundant updates
// when the output is structurally identical.
//
// Parameters:
//   - source: the source Beam[T] to derive from.
//   - cast: a function that transforms the source value of type T into type T2.
//   - distinct: if true, suppresses duplicate values based on deep equality.
//
// Returns:
//   - A new Beam[T2] that tracks transformed updates from the source.
func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2, distinct bool) Beam[T2] {
	return beam.NewBeam(source, cast, distinct)
}

// NewBeamExt derives a new Beam[T2] from an existing Beam[T] using a custom projection and update comparison.
//
// The cast function transforms the source value of type T into type T2. The derived Beam will emit updates
// whenever the source changes, but only if the updateIf function returns true.
//
// The updateIf function receives pointers to the new and previous values (casted) and should return true
// to allow the new value to propagate, or false to suppress it. This allows fine-grained control over
// update emission beyond reflect.DeepEqual.
//
// Parameters:
//   - source: the source Beam[T] to derive from.
//   - cast: a function to transform T → T2.
//   - updateIf: a function to determine whether to emit an update, given the new and previous values.
//
// Returns:
//   - A new Beam[T2] that updates only when updateIf returns true.
func NewBeamExt[T any, T2 any](source Beam[T], cast func(T) T2, updateIf func(new T2, old T2) bool) Beam[T2] {
	return beam.NewBeamExt(source, cast, updateIf)
}
