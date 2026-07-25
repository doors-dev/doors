# Setter

`doors.Setter` sets attributes on every element it is attached to.

Use it when several elements should stay in sync.

It is also fine to use on a single element when that API is the simplest fit.

## Use

Use it when:

- attribute changes are the whole job
- Go already knows the new value
- the same attributes should stay in sync across several elements
- or there is only one element and this is still the simplest way to manage its attributes
- a rerender would be unnecessary work

Typical fits are `disabled`, `hidden`, `aria-*`, `data-*`, and single-purpose `class` or `style` values.

Prefer `ActionEmit` with a `$on(...)` handler when the browser should own the DOM manipulation logic, especially if the change depends on measurements, timers, third-party widgets, or other client-side decisions. See [JavaScript](./15-javascript.md) and [Actions](./12-actions.md).

## Example

```gox
<>
	~~
	var locked doors.Setter
	~~

	<header>
		<button (&locked)>Save draft</button>
	</header>

	<footer>
		<button (&locked)>Publish</button>
	</footer>

	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				doors.Call(ctx, locked.Set("disabled", true))
				return false
			},
		})>
		Lock both actions
	</button>
</>
```

Because both buttons share the setter, one call disables both.

## API

The zero value is ready to use. Attach it like any other attr, then dispatch changes with `Set(name, value)`, which returns an [Action](./12-actions.md) for `doors.Call`, `doors.XCall`, or any action-accepting API. `XCall[int]` reports the number of affected elements.

The value follows template attribute semantics:

- `nil` and `false` remove the attribute
- `true` sets it bare
- other values serialize as in rendered markup

```go
doors.Call(ctx, locked.Set("disabled", true))
doors.Call(ctx, s.Set("aria-label", "busy"))
doors.Call(ctx, s.Set("hidden", nil))
```

Setter is stateless: it changes live elements only. A rerendered element returns to its template attributes, so keep the source of truth in state and templates when the value must survive rerenders.

## Rules

- Use it for attribute changes shared across elements.
- Reuse one setter when several elements should stay in sync.
- Prefer `ActionEmit` with `$on(...)` for richer client-side DOM work.
- Use normal rendering and state when the UI itself should change.
- Use [Events](./08-events.md) for normal DOM events.

## Deprecated: AShared

`doors.AShared` (`NewAShared`, `Update`, `Enable`, `Disable`) is deprecated. Use `doors.Setter` instead: `Enable`/`Update` become `Set(name, value)`, `Disable` becomes `Set(name, nil)`.
