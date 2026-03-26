# Custom Attrs

This page covers the **Doors** attrs you use when plain HTML attrs and the built-in event attrs are not enough.

Use them for:

- passing server values into a script
- letting a script call back into Go
- keeping one HTML attribute under runtime control without rerendering the whole element

## Data

Use `doors.AData` to expose one named value to `$data(name)`.

If the script needs several values, attach several `AData` attrs or several `data:name=(...)` entries.

GoX also supports the shorthand `data:name=(...)`, which is equivalent to attaching `doors.AData{Name: ..., Value: ...}`.

For payload shape:

- `string` arrives as text
- `[]byte` arrives as binary
- other values arrive as JSON

This is the right tool for script inputs that come from the server at render time.

Example with one value:

```gox
<script
	(doors.AData{
		Name: "settings",
		Value: map[string]any{
			"units": "metric",
			"days":  7,
		},
	})>
	const settings = await $data("settings")
	console.log(settings.units, settings.days)
</script>
```

Example with several values:

```gox
<script
	data:userId=(42)
	data:theme=("light")>
	const userId = await $data("userId")
	const theme = await $data("theme")
	console.log(userId, theme)
</script>
```

## Typed

Use `doors.AHook[T]` when JavaScript should send JSON and receive JSON.

JavaScript calls it with `$hook(name, arg)`.

On the Go side, the handler receives `doors.RequestHook[T]`, so it can:

- read `r.Data()`
- read or set cookies
- schedule `r.After(...)` actions

Return `false` to keep the hook active.

Return `true` to remove it after the call.

```gox
<script
	(doors.AHook[string]{
		Name: "countText",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			return len(r.Data()), false
		},
	})
	data:text=("hello")>
	const count = await $hook("countText", await $data("text"))
	console.log(count)
</script>
```

## Raw

Use `doors.ARawHook` with `$fetch(name, arg)` when you need lower-level transport control.

That is the right fit for:

- raw request body access
- multipart handling
- uploads
- custom decoding
- custom response bodies or headers

The handler receives `doors.RequestRawHook`, which adds:

- `r.Body()` for raw request bodies
- `r.Reader()` or `r.ParseForm(...)` for multipart forms
- `r.W()` when you want to write the response yourself

## Dyn

Use `doors.NewADyn(name, value, enable)` when one HTML attribute should change later without rerendering the whole element.

Attach the returned `doors.ADyn` like any other attr, then update it with:

- `Value(ctx, value)` to change the attribute value
- `Enable(ctx, true)` to add the attribute
- `Enable(ctx, false)` to remove it

Good fits are:

- `disabled`
- `hidden`
- `class`
- `aria-*`
- `data-*`

```gox
<>
	~{
		disabled := doors.NewADyn("disabled", "", false)
	}

	<button (disabled)>Save</button>

	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				disabled.Enable(ctx, true)
				return false
			},
		})>
		Disable save
	</button>
</>
```

## Reuse

Reusing one activated hook attribute with `doors.A(ctx, ...)` across elements means those elements share the same hook instance.

Calls to that instance are serialized in **Doors**.

Reusing one `ADyn` value across several mounted elements means `Value(...)` and `Enable(...)` update every currently mounted copy.

## Rules

- Use [Events](./08-events.md) for normal DOM events.
- Prefer `AHook[T]` first; move to `ARawHook` only when transport control needs it.
- Use `AData` or `data:name=(...)` for script inputs instead of hardcoded JS literals.
- Use `ADyn` when you want to keep the same DOM node and just change one attribute.
- Use normal rendering and state when the whole UI should react, not just one attribute.
- See [JavaScript](./14-javascript.md) for `$hook`, `$fetch`, `$data`, `HookErr`, and client-side details.
