# Beam / SourceBeam

`Beam` represents a reactive changing value stream that can be read, subscribed to, watched, or derived. 

It is guaranteed that a Node and all of its children Nodes will observe the exact same value for a given Beam during the render cycle. 

`SourceBeam` is the initial `Beam` (others are derived from it), which, in addition to its core functionality, includes the ability to update values and propagate changes to all subscribers and derived beams. It serves as the root of a reactive value chain. 

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
>##### `Cancel()`
>
>Called when the watcher is terminated due to context cancellation.
>
>##### `Init(ctx context.Context, value *T, seq uint) bool`
>
>Called with the initial value.
>
>- **`seq`**: Sequence number of the update.
>- Called in the same goroutine where the watcher was added.
>- **Return value**: Return `true` to stop receiving updates after this call.
>
>##### `Update(ctx context.Context, value *T, seq uint) bool`
>
>Called for each subsequent update to the value.
>
>- **`seq`**: Increments with each update.
>- **Return value**: Return `true` to stop receiving further updates.

## SourceBeam API

### `Update(ctx context.Context, value T)`

Sets a new value and propagates it to all subscribers and derived beams.

- The update is applied only if it passes the source's **distinct check** and context is valid .

### `Mutate(ctx context.Context, f func(T) T)`

Modifies the current value using the provided function.

- The function receives a copy of the current value and must return a new value.
- The mutation is applied only if the resulting value passes the **distinct check**. Returning an unchanged value (when a distinct function is set) results in no update (if distinct function not `nil`).


### `Latest() T`

Returns the most recently set or mutated value **without** requiring a context.

- Not affected by context cancellation, unlike `Read`.
- **Warning**: `Latest()` does **not** participate in render cycle consistency guarantees. During rendering, use `Read()` to ensure consistent values across the component tree.

## Extra `SourceBeam` API

```
XMutate(ctx context.Context, f func(T) T) (<-chan error, bool)
XUpdate(ctx context.Context, value T) (<-chan error, bool)
```

Do the same, and returns a channel that signals when the mutation has been fully propagated to all subscribers. This allows coordination ofdependent operations that must wait for the mutation to complete.

* Channel receives `nil` on successful propagation
* Channel receives `error` if provided context is invalid or instance ended before propagation finished
* Channel closed without any value if distinct check failed, so no update needed

## Helper Components

For simple relations between `Beam` and DOM.

### `func Sub[T any](beam Beam[T], render func(T) templ.Component) templ.Component`

Creates a reactive component that re-renders whenever a beam’s value changes. It subscribes to the beam, computes content with the provided `render` function, and displays it in a node.

```templ
templ display(n int) {
  <span>{strconv.Itoa(n)}</span>
}

templ demo(b Beam[int]) {
  @doors.Sub(b, func(v int) templ.Component {
    return display(v)
  })
}

```

### `func Inject[T any](key any, beam Beam[T]) templ.Component`

Creates a reactive component that writes the current beam value into the rendering context for its children. On every update, children are re-rendered with the latest value available via the context.

```templ
// inject beam value into context
@Inject("user", userBeam) {
  {{
    user := ctx.Value("user").(User)
  }}
  <p>{user.Name}</p>
}
```

## Best Practices

* Store origin values in beam, not business data. For example — id, but not the whole entry.
* Minimize update frequency and region by deriving into smaller state pieces. If specific element depends on a single field in **beamed** data, derive beam with this field only and subscribe element to it.

