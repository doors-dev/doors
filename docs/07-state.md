# State

In **Doors**, state starts with one `doors.Source[T]` and branches into smaller `doors.Beam[T]` values.

- `Source` is state you can update
- `Beam` is a read-only view of state

A `Source` also implements `Beam`, so you can read, subscribe to, and derive from the same value.

State always starts with a `Source`, then is derived into smaller `Beam`s.

The pattern is:

- keep one source of truth
- derive smaller beams from it
- subscribe only the page fragments that need each piece

Subscribers are triggered when the new value is not equal to the current one.

## Source

Create one with `doors.NewSource(...)` when normal `==` equality is enough:

```go
count := doors.NewSource(0)
```

Use `doors.NewSourceEqual(...)` when you need custom equality:

```go
import "reflect"

settings := doors.NewSourceEqual(Settings{}, func(new Settings, old Settings) bool {
	return reflect.DeepEqual(new, old)
})
```

The equality function should return `true` when the values should be treated as equal, which suppresses propagation.

## Beam

Use a `Beam` when a part of the page only needs a smaller read-only view of state.

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

This is one of the main ways **Doors** keeps updates small. If only `Units` changes, the `days` beam can stay unchanged and the fragment using it does not need to rerender.

> Sources and beams are not limited to one page instance. They can be local to one page, shared across a session, or used even more broadly.

## Render

You can subscribe to a `Beam` and use its values to update rendered content through a `doors.Door`:

1. keep a `doors.Door` on the component
2. subscribe to the beam during render
3. update that door when the beam value changes
4. render the door

```gox
type CounterView struct {
	counter doors.Beam[int]
	body    doors.Door
}

elem (c *CounterView) Main() {
	~{
		c.counter.Sub(ctx, func(ctx context.Context, v int) bool {
			c.body.Update(ctx, v)
			return false
		})
	}

	~>(c.body) <span></span>
}
```

That is the core pattern: the subscription drives updates into a `Door`, and the `Door` keeps that fragment in sync.

For most app code, the helper components are easier.

### Sub

`doors.Sub` wraps that pattern for the common case:

```gox
<>
	~(doors.Sub(counter, elem(v int) {
		<span>~(v)</span>
	}))
</>
```
It creates a dynamic fragment that subscribes to the beam and rerenders that fragment when the value changes.

> Unmounting/updating a dynamic parent also cancels old subscriptions inside it automatically.

### Inject

`doors.Inject` subscribes to a beam, places the current value into the child context, and makes the following tag the dynamic container:

```gox
<>
	~>(doors.Inject("settings", settings)) <section>
		~{
			s := ctx.Value("settings").(Settings)
		}
		<span>Days: ~(s.Days)</span>
	</section>
</>
```

## Effect

`Effect` returns the current value and rerenders the closest dynamic parent when the value changes.

```gox
type CounterView struct {
	count doors.Source[int]
}

elem (c *CounterView) Main() {
	~>(&doors.Door{}) <div>
		~{
			value, _ := c.count.Effect(ctx)
		}
		<p>Count: ~(value)</p>
	</div>
}
```

That is useful when a small dynamic fragment only needs to read a value and rerender itself on changes, without writing an explicit subscription callback.

You can also read multiple values this way:

```gox
type SearchView struct {
	query doors.Beam[string]
	page  doors.Beam[int]
}

elem (v *SearchView) Main() {
	~>(&doors.Door{}) <div>
		~{
			query, _ := v.query.Effect(ctx)
			page, ok := v.page.Effect(ctx)
		}
		~(if ok {
			<p>Query: ~(query)</p>
			<p>Page: ~(page)</p>
		})
	</div>
}
```

It is enough to check only the last `ok`. `Effect` fails only when the context was already canceled, so if the last call succeeds, the earlier ones did too.

> Use multiple `Effect` calls when the values come from different parts of the application, such as route state and language settings. Do not split one logical state into many tiny `Source`s just to read each field with its own `Effect`. Keep one source of truth. Derive smaller `Beam`s only when you want a narrower update surface.

## Update

Use `Update` when you already know the next value:

```go
settings.Update(ctx, Settings{
	Units: "imperial",
	Days:  7,
})
```

Use `Mutate` when the new value naturally depends on the old one:

```go
settings.Mutate(ctx, func(s Settings) Settings {
	s.Days += 1
	return s
})
```

The `XUpdate` and `XMutate` variants return a completion channel. Most code does not need them.

They are useful when completion itself matters, especially for backpressure. For example, if updates arrive very quickly, waiting for `XUpdate` lets a producer send the next state only after the previous one finished propagating.

Do not wait on `XUpdate` or `XMutate` during rendering.

If you need to wait for propagation, do it in a hook, inside `doors.Go(...)`, or
in your own goroutine with `doors.Free(ctx)`.

If that work should outlive the current dynamic owner, use
`doors.FreeRoot(ctx)` instead.

## Read

Reading and subscribing need a valid **Doors** context, such as the `ctx` you get in render code, handlers, subscriptions, or `doors.Go(...)`.

Updating a `Source` can be done from any context.

### Read

```go
value, ok := beam.Read(ctx)
```

Use `Read(ctx)` when you want the value that is consistent with the current **Doors** render/update cycle.

### Get

```go
value := source.Get()
```

`Get()` returns the latest stored value without using a render context.

Use it when you want the current value directly. Do not use it when render consistency matters.

### Sub

```go
ok := beam.Sub(ctx, func(ctx context.Context, value T) bool {
	return false
})
```

`Sub` calls the callback immediately with the current value, then again on later updates.

The subscription ends when:

- your callback returns `true`
- the owning dynamic parent is unmounted

### ReadAndSub

```go
value, ok := beam.ReadAndSub(ctx, func(ctx context.Context, value T) bool {
	return false
})
```

This returns the current value first, then subscribes to future updates. The callback is for later updates only.


### Watcher

`AddWatcher` is the low-level subscription API behind the helpers above. Most app code should use `Sub` or `ReadAndSub`.


## Consistency

The most important state guarantee in **Doors** is consistency.

During one render/update cycle, a Door subtree sees one coherent view of a `Source` and all `Beam`s derived from it. A parent and its children do not see different versions halfway through the same render.

In practice, this means beam-driven rendering stays predictable even when several parts of the page are updating at once.

## Skipping

By default, a `Source` is allowed to skip stale in-flight updates.

That is usually what you want for UI state. If a newer value arrives before an older one finishes propagating, **Doors** prefers getting the UI to the latest useful state instead of insisting that every intermediate value must be rendered.

If you really need every value to propagate, call:

```go
source.DisableSkipping()
```

Use this only when the source behaves more like a message stream than like normal UI state.

## Rules

- Prefer one `Source` plus many derived `Beam`s over many unrelated mutable sources.
- Keep state small and structural. Store IDs, filters, settings, selections, and route values.
- Use `Read(ctx)` when you need the render-consistent value.
- Use `Get()` only when you explicitly want the latest stored value outside render guarantees.
- Return a fresh value from `Mutate` or pass a fresh value to `Update` instead of mutating reference-type state in place.
- Use `DisableSkipping` only when you truly need every update delivered.

## Example

```gox
type SearchState struct {
	Query string
	Page  int
}

type Search struct {
	state doors.Source[SearchState]
	query doors.Beam[string]
	page  doors.Beam[int]
}

func NewSearch() *Search {
	state := doors.NewSource(SearchState{})
	return &Search{
		state: state,
		query: doors.NewBeam(state, func(s SearchState) string { return s.Query }),
		page:  doors.NewBeam(state, func(s SearchState) int { return s.Page }),
	}
}

elem (s *Search) Main() {
	<input
		type="search"
		(doors.AInput{
			On: func(ctx context.Context, ev doors.RequestEvent[doors.InputEvent]) bool {
				value := ev.Event().Value
				s.state.Mutate(ctx, func(st SearchState) SearchState {
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
