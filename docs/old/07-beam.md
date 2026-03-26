# Beam and SourceBeam

Reactive value streams for rendering and state flow.  
A **Beam[T]** is a read-only stream. A **SourceBeam[T]** is the root stream you update. During a render cycle, a Door and all children see the same value for a given Beam.

---

## Constructors

### Create a Source Beam

```go
// Uses == for equality
NewSourceBeam[T comparable](init T) SourceBeam[T]

// Uses custom equality function
NewSourceBeamEqual[T any](init T, equal func(new T, old T) bool) SourceBeam[T]
```

- `NewSourceBeam` suppresses updates when `new == old`.
- `NewSourceBeamEqual` uses the supplied equality function. If `equal` is `nil`, every update propagates.

### Create a Derived Beam

```go
// Uses == on derived type
NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2) Beam[T2]

// Uses custom equality function
NewBeamEqual[T any, T2 any](source Beam[T], cast func(T) T2, equal func(new T2, old T2) bool) Beam[T2]
```

- Derived beams transform data from another beam.
- `NewBeamEqual` suppresses updates when equality function returns true.

---

## Beam API (Read-only)

```go
Sub(ctx, onValue) bool
XSub(ctx, onValue, onCancel) (Cancel, bool)

ReadAndSub(ctx, onValue) (T, bool)
XReadAndSub(ctx, onValue, onCancel) (T, Cancel, bool)

Read(ctx) (T, bool)

AddWatcher(ctx, w Watcher[T]) (Cancel, bool)
```

- `X*` variants include `onCancel` and return a cancel handle.
- `Sub` calls `onValue` immediately, then again on updates until `onValue` returns true, parent dynamic contauner (Door) unmounted or manually canceled.

### Watcher Interface

Watcher interface for low level beam access

#### Interface `Watcher[T any]`

Defines hooks for observing and reacting to the lifecycle of a **beam** value stream. Implementers can perform custom logic during initialization, on each update, and when canceled.

- `Cancel()`

Called when the watcher is terminated due to context cancellation.

- `Init(ctx context.Context, value *T, seq uint) bool`

Called with the initial value.

- **`seq`**: Sequence number of the update.
- Called in the same goroutine where the watcher was added.
- **Return value**: Return `true` to stop receiving updates after this call.

- `Update(ctx context.Context, value *T, seq uint) bool`

Called for each subsequent update to the value.

- **`seq`**: Increments with each update.
- **Return value**: Return `true` to stop receiving further updates.

---

## SourceBeam API (Writable)

```go
Update(ctx, value T)
XUpdate(ctx, value T) <-chan error

Mutate(ctx, f func(T) T)
XMutate(ctx, f func(T) T) <-chan error

Latest() T
DisableSkipping()
```

- `Update` and `Mutate` send updates if equality check indicates a change.
- `XUpdate` and `XMutate` signal completion through a channel:  
  - Yields `nil` on success or skip.  
  - Yields an `error` on cancellation or ended instance.  
  - Closes empty if equality suppressed update.
- `Latest` returns current value without render synchronization.  
- `DisableSkipping` disables coalescing; all values propagate.

---

## Helper Components

### `Sub`

Reactive render helper:

```templ
@doors.Sub(beam, func(v T) templ.Component { ... })
```

Re-renders when the beam updates.

### `Inject`

Injects current value into the context for child components and re-renders on updates.

---

## Design Notes

- Keep source beams minimal. Store references or identifiers, not large structs.
- Use derived beams for transformations or projections.
- Use `New*Equal` when equality by `==` is not meaningful.
