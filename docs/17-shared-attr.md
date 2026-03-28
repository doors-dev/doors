# Shared Attr

`doors.AShared` lets one Go handle control the same HTML attribute on every element where it is attached.

Use it when several elements should stay in sync.

It is also fine to use on a single element when that API is the simplest fit.

## Use

Use it when:

- one attribute is the whole job
- Go already knows the new value
- the same attribute should stay in sync across several elements
- or there is only one element and this is still the simplest way to manage that attribute
- a rerender would be unnecessary work

Typical fits are `disabled`, `hidden`, `aria-*`, `data-*`, and single-purpose `class` or `style` values.

Prefer `ActionEmit` with a `$on(...)` handler when the browser should own the DOM manipulation logic, especially if the change depends on measurements, timers, third-party widgets, several attributes, or other client-side decisions. See [JavaScript](./15-javascript.md) and [Actions](./12-actions.md).

## Example

```gox
<>
	~{
		locked := doors.NewAShared("disabled", "")
	}

	<header>
		<button (locked)>Save draft</button>
	</header>

	<footer>
		<button (locked)>Publish</button>
	</footer>

	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				locked.Disable(ctx)
				return false
			},
		})>
		Lock both actions
	</button>
</>
```

Because both buttons use the same handle, one call disables both.

## API

Create a handle with `doors.NewAShared(name, value)`.

It starts enabled.

Attach it like any other attr, then update it later with:

- `Update(ctx, value)` to change the attribute value
- `Enable(ctx)` to add the attribute
- `Disable(ctx)` to remove the attribute

## Rules

- Use it for one-attribute changes shared across elements.
- Reuse one handle when several elements should stay in sync.
- Prefer `ActionEmit` with `$on(...)` for richer client-side DOM work.
- Use normal rendering and state when the UI itself should change.
- Use [Events](./08-events.md) for normal DOM events.
