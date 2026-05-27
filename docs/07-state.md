# State

In **Doors**, reactive values come in two public shapes:

- `Source[T]` is writable. It can own an original value or describe a writable derived view of another source.
- `Beam[T]` is read-only. It observes or derives a value without exposing writes.

A `Source` is also a `Beam`, so every source can be read, subscribed to, bound into UI, and routed like a read-only value.

The usual pattern is:

- keep one source of truth
- derive smaller, specific parts
- bind DOM to the smallest specific pieces

> Subscribers (incl. binds, effects) are triggered when the new value is not equal to the previous value for that source or beam.

## Original Sources

Create an original source with `doors.NewSource(...)` when normal `==` equality is enough:

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


## Derived Sources

Use `DeriveSource` when a part of the page should read and update a smaller part of a larger source value.

```go
type Settings struct {
	Units string
	Days  int
}

settings := doors.NewSource(Settings{
	Units: "metric",
	Days:  7,
})

units := doors.DeriveSource(settings,
	func(s Settings) string {
		return s.Units
	},
	func(s Settings, units string) Settings {
		s.Units = units
		return s
	},
)

days := doors.DeriveSource(settings,
	func(s Settings) int {
		return s.Days
	},
	func(s Settings, days int) Settings {
		s.Days = days
		return s
	},
)
```

`get` extracts the derived value. `set` receives the current parent value and the new derived value, then returns the next parent value.

Derived sources compose:

```go
isMetric := doors.DeriveSource(units,
	func(units string) bool {
		return units == "metric"
	},
	func(_ string, metric bool) string {
		if metric {
			return "metric"
		}
		return "imperial"
	},
)
```

Use `doors.DeriveSourceEqual(...)` when you need custom equality for the derived value.

Updating a derived source writes back through its parent. The parent source remains the single stored value; the derived source only describes how to read and replace one piece of it.

## Beams

Use a `Beam` when a part of the page only needs a smaller read-only view of state.

```go
longRange := doors.DeriveBeam(settings, func(s Settings) bool {
	return s.Days > 7
})
```

Use `doors.DeriveBeamEqual(...)` when you need custom equality for the derived value.

This is one of the main ways **Doors** keeps updates small. If only `Units` changes, a beam or source derived from `Days` can stay unchanged and the fragment using it does not need to rerender.

> Sources and beams are not limited to one page instance. They can be local to one page, shared across a session, or used even more broadly.

## Render

Four rendering strategies drive content from reactive values:

| Strategy | What it does | Works on |
| --- | --- | --- |
| `Bind` | Rerenders one fragment on every value change | `Beam` and `Source` |
| `Effect` | Rerenders the closest dynamic parent when any read value changes | `Beam` and `Source` |
| `RouteBeam` | Picks one of several read-only views based on the value | `Beam` and `Source` |
| `Route` | Picks one of several writable views based on the value | `Source` |

### Bind

`Bind` creates a dynamic fragment that rerenders whenever the value changes:

```gox
<>
	~(counter.Bind(elem(v int) {
		<span>~(v)</span>
	}))
</>
```

> Unmounting or updating a dynamic parent cancels old subscriptions inside it automatically.

Under the hood, `Bind` is the common shorthand for this pattern:

```gox
type CounterView struct {
	counter doors.Beam[int]
	body    doors.Door
}

elem (c *CounterView) Main() {
	~{
		c.counter.Sub(ctx, func(ctx context.Context, v int) bool {
			c.body.Inner(ctx, v)
			return false
		})
	}

	~>(c.body) <span></span>
}
```

Reach for the manual form when you need explicit control over the door, the subscription lifecycle, or the update strategy.

### Effect

`Effect` reads a value directly inside a dynamic subtree and rerenders that subtree when the value changes. It feels more React-like than passing values through a `Bind` callback.

```gox
<>
	~>(new(doors.Door)) <section>
		~{
			settings, _ := settingsBeam.Effect(ctx)
		}
		<span>Days: ~(settings.Days)</span>
	</section>
</>
```

Especially useful when the same DOM subtree depends on multiple reactive values:

```gox
type SearchView struct {
	query doors.Beam[string]
	page  doors.Beam[int]
}

elem (v *SearchView) Main() {
	~>(new(doors.Door)) <div>
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

> Use multiple `Effect` calls when the values come from different parts of the application, such as route state and language settings. Do not split one logical state into many tiny `Source`s just to read each field with its own `Effect`. Keep one source of truth. Derive smaller sources or beams when you want a narrower update surface.

### Routing

`RouteBeam` and `source.Route` pick one of several views based on a reactive value:

```go
beam.RouteBeam(routes...)      // gox.EditorComp
source.RouteBeam(routes...)    // gox.EditorComp
source.Route(routes...)        // gox.EditorComp
```

The routed fragment only swaps when the active route changes. Value changes that keep the same route matched do not rerender the route fragment. Instead, the route's render function receives a live `Beam` or `Source` and reacts inside with normal state primitives (`Bind`, `Effect`, derived values).

`RouteBeam` accepts only read-only routes. `source.Route` accepts both writable and read-only routes, because a source can also be passed where a beam is expected.

URL routing is just the special case where the source is `doors.Source[doors.Location]`; see [Routing](./05-routing.md) for path models and URL helpers. The mechanics below apply to any reactive value.

#### Match Builders

A route has two parts: a matcher and a render. Three matcher builders, ordered from most general to least:

**`RouteDerive(derive)`** matches and derives in one step. `derive(value)` returns `(derivedValue, ok)`. The route matches when `ok` is `true`, and the route's render receives the derived value. This is the central pattern when state is structural: match on a field's presence and route on the field itself.

```go
matchString := func(s State) (string, bool) {
	return s.Str, s.Str != ""
}
```

For non-comparable derived types such as slices, maps, or custom structs, use `RouteDeriveEqual(derive, equal)` and supply your own equality function.

**`RouteMatch(pred)`** matches without deriving. `pred(value)` returns a `bool`. The route's render receives the original value.

**`RouteValue(v)`** is sugar for `RouteMatch(== v)`. It requires `T` to be `comparable`, and is convenient for typed enums and small unions.

**`RouteDefault*`** always matches. Put it last. It renders with the original value, not the derived type of any non-default route.

#### Render Methods

The match builders are not yet full routes. Chain one of these to produce one:

| Method | Render signature | Available in |
| --- | --- | --- |
| `.Comp(comp)` | fixed `gox.Comp`, no value | `RouteBeam(...)` and `source.Route(...)` |
| `.Beam(render)` | `func(Beam[T]) gox.Elem` | `RouteBeam(...)` and `source.Route(...)` |
| `.Bind(render)` | `func(T) gox.Elem` | `RouteBeam(...)` and `source.Route(...)` |
| `.Source(render)` on `RouteMatch` | `func(Source[T]) gox.Elem` | `source.Route(...)` only |
| `.Source(set, render)` on `RouteDerive` | `set func(parent, derived) parent`, `render func(Source[derived]) gox.Elem` | `source.Route(...)` only |

The `set` parameter on `RouteDerive(...).Source(...)` writes a derived value back into the parent state. Without it, the route would have no way to map a write on `Source[Derived]` back to `Source[Parent]`.

Default routes also have a render-method shape:

- `RouteDefaultComp(comp)` — fixed component
- `RouteDefaultBeam(render)` — `func(Beam[T]) gox.Comp`
- `RouteDefaultBind(render)` — `func(T) gox.Comp`, receives the raw value
- `RouteDefault(render)` — `func(Source[T]) gox.Comp`, only inside `source.Route(...)`

#### Example

```gox
type Section int

const (
	SectionProfile Section = iota
	SectionBilling
)

type Settings struct {
	section doors.Source[Section]
}

elem (s Settings) Main() {
	<button (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestPointer) bool {
			s.section.Update(ctx, SectionProfile)
			return false
		},
	})>
		Profile
	</button>

	<button (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestPointer) bool {
			s.section.Update(ctx, SectionBilling)
			return false
		},
	})>
		Billing
	</button>

	~(s.section.RouteBeam(
		doors.RouteValue(SectionProfile).Comp(ProfilePanel{}),
		doors.RouteValue(SectionBilling).Comp(BillingPanel{}),
		doors.RouteDefaultComp(ProfilePanel{}),
	))
}
```

The fragment swaps only when `section` changes to a value matched by a different route. Updates that keep the same route active stay inside that branch and can be handled with `Bind`, `Effect`, or derived values.

## Update

Use `Update` when you already know the next value:

```go
units.Update(ctx, "imperial")
```

Use `Mutate` when the new value naturally depends on the old one:

```go
days.Mutate(ctx, func(days int) int {
	return days + 1
})
```

When several fields must change as one logical update, mutate the original source directly.

The `XUpdate` and `XMutate` variants return a completion channel. Most code does not need them.

They are useful when completion itself matters, especially for backpressure. For example, if updates arrive very quickly, waiting for `XUpdate` lets a producer send the next state only after the previous one finished propagating.

Do not wait on `XUpdate` or `XMutate` during rendering.

If you need to wait for propagation, do it in a hook, inside `doors.Go(...)`, or in your own goroutine with `doors.DetachedContext(ctx)`.

If that work should outlive the current dynamic owner, use `doors.InstanceContext(ctx)`.

> For reference types such as slices, maps, pointers, or mutable structs, replace the stored value instead of mutating it in place. Doors compares committed values; in-place mutation can make old and new state indistinguishable.

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
value := beam.Get()
```

`Get()` returns the latest stored value without using a render context.

Use it when you want the current value directly. It is available on sources and beams. Do not use it when render consistency matters.

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

`Watch` is the low-level subscription API behind the helpers above. Most app code should use `Sub` or `ReadAndSub`.

## Consistency

The most important state guarantee in **Doors** is consistency.

During one render/update cycle, a Door subtree sees one coherent view of a `Source` and all sources and beams derived from it. A parent and its children do not see different versions halfway through the same render.

In practice, this means beam-driven rendering stays predictable even when several parts of the page are updating at once.

## Skipping

By default, a `Source` is allowed to skip stale in-flight propagation.

That is usually what you want for UI state. If a newer value arrives before an older one finishes propagating, **Doors** prefers getting the UI to the latest useful state instead of insisting that every intermediate value must finish syncing.

If you need in-progress propagation to keep going even when a newer update arrives, create the source with a NoSkip constructor:

```go
source := doors.NewSourceNoSkip(0)
```

For custom equality, use:

```go
source := doors.NewSourceEqualNoSkip(init, equal)
```

NoSkip does not bypass equality. Equal updates are still suppressed; NoSkip only changes what happens after a committed update is already propagating.

Use NoSkip when an already-committed update should keep propagating even if a newer value arrives before it completes.

## Rules

- Prefer one `Source` plus derived sources and beams over many unrelated mutable sources.
- Keep state small and structural. Store IDs, filters, settings, selections, and route values.
- Use `Read(ctx)` when you need the render-consistent value.
- Use `Get()` only when you explicitly want the latest stored value outside render guarantees.
- Return a fresh value from `Mutate` or pass a fresh value to `Update` instead of mutating reference-type state in place.
- Use NoSkip constructors only when already-committed in-progress updates must finish.

## Example

```gox
type SearchState struct {
	Query string
	Page  int
}

type Search struct {
	query doors.Source[string]
	page  doors.Source[int]
}

func NewSearch(state doors.Source[SearchState]) *Search {
	return &Search{
		query: doors.DeriveSource(state,
			func(s SearchState) string { return s.Query },
			func(s SearchState, query string) SearchState {
				s.Query = query
				return s
			},
		),
		page: doors.DeriveSource(state,
			func(s SearchState) int { return s.Page },
			func(s SearchState, page int) SearchState {
				s.Page = page
				return s
			},
		),
	}
}

elem (s *Search) Main() {
	<input
		type="search"
		(doors.AInput{
			On: func(ctx context.Context, ev doors.RequestEvent[doors.InputEvent]) bool {
				value := ev.Event().Value
				s.query.Update(ctx, value)
				return false
			},
		})/>

	<button
		(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestPointer) bool {
				s.page.Mutate(ctx, func(page int) int {
					return page + 1
				})
				return false
			},
		})>
		Next page
	</button>

	~(s.query.Bind(elem(q string) {
		<p>Query: ~(q)</p>
	}))

	~(s.page.Bind(elem(page int) {
		<p>Page: ~(page)</p>
	}))
}
```

This keeps one source of truth while letting the query and page fragments update independently.
