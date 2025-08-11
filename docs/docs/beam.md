# Beam / SourceBeam

`Beam` represents a reactive changing value stream that can be read, subscribed to, watched, or derived. 

It is guaranteed that a Node and all of its children Nodes will observe the exact same value for a given Beam during the render cycle. 

`SourceBeam` is the original Beam, which, in addition to its core functionality, includes the ability to update values and propagate changes to all subscribers and derived beams. It serves as the root of a reactive value chain. 

## Constructor Methods

### `func NewSourceBeam[T comparable](init T) SourceBeam[T]`

Creates a `SourceBeam` with the initial value `init`. This `SourceBeam` will use == to check if the value has indeed been updated, and subscribers and derived Beams must be notified of the new value.

### `func NewSourceBeamExt[T any](init T, distinct func(new T, old T) bool) SourceBeam[T]`

Creates a `SourceBeam` with an initial value `init`. This `SourceBeam` will use a custom distinct function to verify if the value has indeed been updated, and subscribers and derived Beams must be notified of the new value.

### `func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2) Beam[T2] `

Derives a new Beam[T2] from an existing Beam[T] by applying a cast function. This `Beam` will use == to check if the value has indeed been updated, and subscribers and derived Beams must be notified of the new value.

### `func NewBeamExt[T any, T2 any](source Beam[T], cast func(T) T2, distinct func(new T2, old T2) bool) Beam[T2]`

Derives a new Beam[T2] from an existing Beam[T] by applying a cast function. This `Beam` will use a custom distinct function to verify if the value has indeed been updated, and subscribers and derived Beams must be notified of the new value.

## Beam  API

### `Sub(ctx context.Context, onValue func(context.Context, T) bool) bool`

Subscribes to the value stream. Invokes `onValue` immediately with the current value and again on every update. Continues until the context is canceled or `onValue` returns `true`. Returns `true` if the subscription was established; `false` if the context was already canceled. 

### `SubExt(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (Cancel, bool)`

Extended form of `Sub`. Behaves the same, and additionally accepts `onCancel` (called when the subscription ends due to context cancellation) and returns a `Cancel` function for manual termination. Returns the `Cancel` function and a boolean indicating whether the subscription was established. 

### `ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)`

Returns the current value, then subscribes to future updates (invoking `onValue` on each update). Returns the initial value and a boolean: `true` if the value is valid and the subscription was established; `false` if the context was canceled. 

### `ReadAndSubExt(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (T, Cancel, bool)`

An extended form of `ReadAndSub`. Also accepts `onCancel` and returns a `Cancel` function. Returns the initial value, the `Cancel` function, and a success boolean. If the boolean is `false`, the value is undefined, and no subscription was established. 

### `Read(ctx context.Context) (T, bool)`

Reads the current value without subscribing. Returns the value and a boolean: `true` if the value is valid; `false` if the context was canceled (value undefined). 

### `AddWatcher(ctx context.Context, w Watcher[T]) (Cancel, bool)`

Attaches a watcher for full lifecycle control. Watchers receive separate callbacks for initialization, updates, and cancellation. Returns a `Cancel` function and a boolean indicating whether the watcher was added. 

>#### Interface `Watcher[T any]`
>
>Defines hooks for observing and reacting to the lifecycle of a Beam value stream. Implementers can perform custom logic during initialization, on each update, and when canceled.
>
>### `Cancel()`
>
>Called when the watcher is terminated due to context cancellation.
>
>### `Init(ctx context.Context, value *T, seq uint) bool`
>
>Called with the initial value.
>
>- **`seq`**: Sequence number of the update.
>- Called in the same goroutine where the watcher was added.
>- **Return value**: Return `true` to stop receiving updates after this call.
>
>### `Update(ctx context.Context, value *T, seq uint) bool`
>
>Called for each subsequent update to the value.
>
>- **`seq`**: Increments with each update.
>- **Return value**: Return `true` to stop receiving further updates.

## SourceBeam API

### `Update(ctx context.Context, value T) bool`

Sets a new value and propagates it to all subscribers and derived beams.

- The update is applied only if it passes the source's **distinct check** .
- **Returns**:
  - `true` if the context is valid and the update was accepted.
  - `false` if the context was canceled before the update.

### `Mutate(ctx context.Context, f func(*T) bool) bool`

Modifies the current value using the provided function.

- The function receives a copy of the current value and returns `true` to indicate that changes to the copy should be applied to the Beam.
- The mutation is applied only if:
  - The function returns `true`.
  - The resulting value passes the **distinct check**.
- **Returns**:
  - `true` if the context is valid and the mutation was applied.
  - `false` if the context was canceled or the mutation function returned `false`.

### `Latest() T`

Returns the most recently set or mutated value **without** requiring a context.

- Not affected by context cancellation, unlike `Read`.
- **Warning**: `Latest()` does **not** participate in render cycle consistency guarantees. 
  - During rendering, use `Read()` to ensure consistent values across the component tree.

## Extra `SourceBeam` API

```
XMutate(ctx context.Context, f func(*T) bool) (<-chan error, bool)
XUpdate(ctx context.Context, value T) (<-chan error, bool)
```

Do the same, but returns a channel that signals when the update has been fully propagated to all subscribers.

* `nil` on successful propagation and non-nil `error` if the operation failed
* closes immediately if the mutation was not applied or if there are no active subscribers
