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
// Parameters:
//   - distinct: if true, updates are only propagated when the value differs from the previous one.
//
// Returns:
//   - A new SourceBeam[T] instance
func NewSourceBeam[T comparable](init T) SourceBeam[T] {
	return beam.NewSourceBeam(init)
}

// NewSourceBeamExt creates a new SourceBeam with a custom distinct condition.
//
// Unlike NewSourceBeam, which uses == check to suppress duplicate updates,
// this version accepts function to determine whether a new value should
// be propagated to subscribers.
//
// The distinct function receives pointers to the new and previous values and should
// return true if the new value is considered different. 
//
//
// Parameters:
//   - init: the initial value for the SourceBeam.
//   - distinct: a custom function to determine if a new value should trigger propagation
//   or nil to tigger every time
//
// Returns:
//   - A new SourceBeam[T] instance that uses distinct function for update comparisons.
func NewSourceBeamExt[T any](init T, distinct func(new *T, old *T) bool) SourceBeam[T] {
	return beam.NewSourceBeamExt(init, distinct)
}

// NewBeam derives a new Beam[T2] from an existing Beam[T] by applying a transformation function.
//
// The cast function maps values from the source beam to the derived beam. The derived beam
// watcher will receive updates whenever the source beam updates and old value != new value
//
//
// Parameters:
//   - source: the source Beam[T] to derive from.
//   - cast: a function that transforms the source value of type T into type T2.
//
// Returns:
//   - A new Beam[T2] that tracks transformed updates from the source.
func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2) Beam[T2] {
	return beam.NewBeam(source, cast)
}

// NewBeamExt derives a new Beam[T2] from an existing Beam[T] using a custom projection and update comparison.
//
// The cast function transforms the source value of type T into type T2. The derived Beam will emit updates
// whenever the source changes, but only if the updateIf function returns true.
//
// The distinct function receives pointers to the new and previous values (casted) and should return true
// to allow the new value to propagate, or false to suppress it. This allows fine-grained control over
// update emission beyond just ==.
//
// Parameters:
//   - source: the source Beam[T] to derive from.
//   - cast: a function to transform T → T2.
//   - distinct: a custom function to determine if a new value should trigger propagation
//   or nil to tigger every time
//
// Returns:
//   - A new Beam[T2] that updates only when updateIf returns true.
func NewBeamExt[T any, T2 any](source Beam[T], cast func(T) T2, distinct func(new *T2, old *T2) bool) Beam[T2] {
	return beam.NewBeamExt(source, cast, distinct)
}
