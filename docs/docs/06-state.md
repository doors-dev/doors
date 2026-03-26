# State

Doors state is built around two types:

- `doors.Source[T]`: a writable `Beam`
- `doors.Beam[T]`: a read-only reactive view of state

`Source` extends `Beam`, so every `Source` is also a `Beam`.

Always start with a `Source`, then derive smaller `Beam`s from it for the specific dynamic fragments that need them.

## Model

A `Source` is the root beam in your state graph. Because it is also a `Beam`, you can read it, subscribe to it, derive from it, and update it.

A derived `Beam` is a read-only view of upstream state. You use it to project just the piece a fragment cares about.

Beams are not tied to one page instance. The same `Source` or derived `Beam` can be shared across multiple components, multiple instances, and even multiple sessions if you keep that value in shared application state.

```go
settings := doors.NewSource(Settings{
	Units: "metric",
	Days:  7,
})

units := doors.NewBeam(settings, func(s Settings) string {
	return s.Units
})

days := doors.NewBeam(settings, func(s Settings) int {
	return s.Days
})
```

This is the preferred pattern in Doors:

- keep one source of truth
- derive smaller beams from it
- subscribe only the DOM fragments that actually depend on each piece

## Purpose

Beams are synchronized with Doors rendering.

That gives you a few important properties:

- updates only propagate when the value actually changes
- derived beams can suppress updates independently
- subscriptions are attached to dynamic DOM fragments
- propagation goes top-to-bottom through the dynamic tree
- during one render cycle, a Door and all of its descendants observe the same value for the same Beam

That consistency guarantee is what makes Beam-based rendering predictable.

## Consistency

Inside Doors, rendering is split into dynamic fragments. Each fragment can subscribe to beams and render in its own dynamic container.

When a render cycle begins for a given fragment tree, Doors takes a consistent Beam snapshot for that cycle. Reads and derived reads inside that cycle observe that snapshot, even if the same code triggers more source mutations before child fragments render.

That means this kind of logic is stable and predictable:

```gox
elem demo(b doors.Source[int]) {
	~{
		b.Mutate(ctx, func(v int) int {
			return v + 1
		})
		snapshot, _ := b.Read(ctx)
	}
	Current render sees ~(snapshot)
}
```

`Read(ctx)` gives the render-consistent value for the current cycle, not an arbitrary interleaving.

Multiple mutations can happen while rendering sibling and child fragments, yet all `Read(ctx)` calls in that render still see the same stable value. The newer state is propagated afterward in the next update cycle.

### `Read(ctx)` vs `Get()`

Use:

- `Read(ctx)` when you need the value that is consistent with the current Doors render/subscription context
- `Get()` when you need the latest stored value and do not need render consistency

`Get()` is intentionally outside the render-consistency guarantees.

## Creation

### `NewSource`

Use `NewSource` when `T` is comparable and normal `==` equality is correct.

```go
count := doors.NewSource(0)
```

### `NewSourceEqual`

Use `NewSourceEqual` when you need custom equality.

```go
settings := doors.NewSourceEqual(Settings{}, func(new Settings, old Settings) bool {
	return new == old
})
```

The equality function returns `true` when the values should be treated as equal, which suppresses propagation.

### `NewBeam`

Use `NewBeam` to derive a smaller comparable value.

```go
country := doors.NewBeam(location, func(v SelectedLocation) string {
	return v.Country
})
```

### `NewBeamEqual`

Use `NewBeamEqual` when the derived value needs custom equality.

```go
filters := doors.NewBeamEqual(settings, func(s Settings) Filters {
	return s.Filters
}, func(new Filters, old Filters) bool {
	return reflect.DeepEqual(new, old)
})
```

## Reading

Reading and subscribing require a valid Doors instance context. In practice, use the `ctx` that Doors gives you in render, handlers, subscriptions, and `doors.Go`.

Updating a `Source` can be done from any context.

### `Read`

```go
value, ok := beam.Read(ctx)
```

Reads the current render-consistent value without subscribing.

### `Sub`

```go
ok := beam.Sub(ctx, func(ctx context.Context, value T) bool {
	// return true to stop
	return false
})
```

`Sub` calls the callback immediately with the current value, then again on later updates.

The subscription ends when:

- your callback returns `true`
- the dynamic parent is unmounted

For `XSub` and `XReadAndSub`, you can also stop the subscription explicitly via the returned cancel function.

### `ReadAndSub`

```go
value, ok := beam.ReadAndSub(ctx, func(ctx context.Context, value T) bool {
	return false
})
```

This returns the current value first, then subscribes to future updates. Unlike `Sub`, the callback is for later updates only.

### `XSub` and `XReadAndSub`

Use these when you need:

- a cancel function
- an `onCancel` callback

### `AddWatcher`

`AddWatcher` is the low-level subscription API behind the helpers above. Most application code should use `Sub` or `ReadAndSub`.

## Updates

### `Update`

```go
settings.Update(ctx, Settings{
	Units: "imperial",
	Days:  7,
})
```

Sets a new value.

### `Mutate`

```go
settings.Mutate(ctx, func(s Settings) Settings {
	s.Days = 14
	return s
})
```

Transforms the current value into a new one.

Use `Mutate` when the new value is naturally based on the old value.

### `XUpdate` and `XMutate`

These return a completion channel.

Most code should use `Update` / `Mutate`. Reach for `XUpdate` or `XMutate` when propagation completion matters, especially when you want backpressure.

Channel behavior:

- successful propagation sends `nil`
- an invalid or canceled propagation can send an error
- if equality suppresses the update, the channel simply closes without a value

Example: very frequent real-time updates.

If you push data as fast as it arrives, you can build up unnecessary pending work. Waiting for `XUpdate` lets the producer send the next state only after the previous propagation completed.

## Skipping

By default, `Source` propagation is allowed to skip stale in-flight updates.

That is intentional.

If a newer value arrives before an older propagation finishes, Doors does best-effort delivery toward the latest state instead of insisting that every intermediate version must be rendered.

This keeps the UI responsive and avoids wasting work on outdated states.

### `DisableSkipping`

If you need every value to propagate, call:

```go
source.DisableSkipping()
```

Use this only when the stream behaves more like a message channel than like UI state.

For normal UI state, the default skipping behavior is usually exactly what you want.

## Derivation

Derived beams are the main optimization tool in Doors.

Instead of subscribing a large fragment to a whole state object, derive smaller beams and subscribe each fragment only to the piece it needs.

```gox
type Dashboard struct {
	settings doors.Source[Settings]
	units    doors.Beam[string]
	days     doors.Beam[int]
}

elem (d *Dashboard) Main() {
	~(doors.Sub(d.units, elem(units string) {
		<span>Units: ~(units)</span>
	}))

	~(doors.Sub(d.days, elem(days int) {
		<span>Days: ~(days)</span>
	}))
}
```

If only `Units` changes:

- the `units` beam updates
- the `days` beam stays unchanged
- only the subscribed fragment for `units` rerenders

This is one of the core mechanics of Doors. Derivation lets the DOM update at the fragment level instead of forcing a larger rerender.

## `doors.Sub`

`doors.Sub` is the high-level rendering helper for a Beam.

```gox
~(doors.Sub(counter, elem(v int) {
	<span>~(v)</span>
}))
```

It creates a Door-backed fragment that subscribes to the Beam and rerenders that fragment whenever the Beam value changes.

`Sub` manages its own dynamic fragment, so it uses the default Door container behavior. In practice, that means it renders through the fallback `d0-r` container.

If the render function returns `nil`, the fragment is cleared.

## `doors.Inject`

`doors.Inject` subscribes to a Beam and places its current value into the child context.

```gox
~>(doors.Inject("settings", settings)) <section>
	~{
		s := ctx.Value("settings").(Settings)
	}
	<span>Days: ~(s.Days)</span>
</section>
```

Use `Inject` when:

- you prefer to read the Beam value from `ctx.Value(...)` inside the proxied subtree
- you want to control the underlying element basis of the dynamic fragment yourself

Unlike `Sub`, `Inject` is a proxy over the following element subtree. That means the element you wrap stays the basis of the dynamic fragment, instead of always going through a standalone `d0-r` container.

## Flow

Internally, Doors tracks Beam subscriptions per dynamic fragment tree.

The important user-visible effects are:

- parent dynamic fragments are synchronized before their children
- sibling branches can propagate in parallel
- derived beams for the same update sequence stay grouped consistently
- unmounting a dynamic parent automatically cancels subscriptions inside it

This is why `Beam` logic usually feels predictable even when the page has many independently updating parts.

## Rules

- Use the `ctx` that Doors gives you. Do not subscribe or read with `context.Background()`.
- Prefer one `Source` plus many derived `Beam`s over many unrelated mutable sources.
- A `Source` or `Beam` can be local to one instance or shared much more broadly, including across multiple instances or sessions.
- Keep Beam values small and structural. Derive IDs, filters, settings, selections.
- Use `Read(ctx)` inside Doors render/subscription code when consistency matters.
- Use `Get()` only when you explicitly want the latest stored value outside render guarantees.
- Do not mutate reference-type state in place. Return a fresh value from `Mutate` or pass a fresh value to `Update`.
- Use `DisableSkipping` only when you truly need every update delivered.

## Example

```gox
type State struct {
	Query string
	Page  int
}

type Search struct {
	state doors.Source[State]
	query doors.Beam[string]
	page  doors.Beam[int]
}

func NewSearch() *Search {
	state := doors.NewSource(State{})
	return &Search{
		state: state,
		query: doors.NewBeam(state, func(s State) string { return s.Query }),
		page:  doors.NewBeam(state, func(s State) int { return s.Page }),
	}
}

elem (s *Search) Main() {
	<input
		type="search"
		(doors.AInput{
			On: func(ctx context.Context, ev doors.RequestEvent[doors.InputEvent]) bool {
				value := ev.Event().Value
				s.state.Mutate(ctx, func(st State) State {
					st.Query = value
					st.Page = 1
					return st
				})
				return false
			},
		})/>

	~(doors.Sub(s.query, elem(q string) {
		<p>Query: ~(q)</p>
	}))

	~(doors.Sub(s.page, elem(page int) {
		<p>Page: ~(page)</p>
	}))
}
```

This keeps one source of truth while letting the query and page fragments update independently.
