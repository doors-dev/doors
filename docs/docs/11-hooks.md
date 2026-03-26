# Hooks

Hooks expose backend handlers and server data to client JavaScript.

They are useful when you want:

- manual event wiring from JavaScript to Go
- integration with embedded mini JavaScript apps
- direct client/server collaboration through `$hook(...)` and `$data(...)`

## Choose

Use this quick mapping:

- structured request/response: `doors.AHook[T]` + `$hook(...)`
- raw body or multipart handling: `doors.ARawHook` + `$fetch(...)`
- pass one value into script: `doors.AData` or `data:name=(...)`
- pass several values: `doors.ADataMap`

## `AHook`

`doors.AHook[T]` registers a named backend handler that client code can call through `$hook(name, arg)`.

```gox
<script
	(doors.AHook[string]{
		Name: "countText",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			return len(r.Data()), true
		},
	})>
	console.log(await $hook("countText", "hello"))
</script>
```

Use `AHook[T]` when JSON input/output is enough.

## `ARawHook`

`doors.ARawHook` gives raw request access.

Use it when the hook needs:

- raw request body access
- multipart handling
- custom decoding

## `AData`

`doors.AData` exposes server data to client JavaScript through `$data(name)`.

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

GoX also supports the attribute shorthand:

```gox
<script data:settings=(map[string]any{
	"units": "metric",
	"days":  7,
})>
	const settings = await $data("settings")
</script>
```

This is equivalent to attaching `doors.AData{Name: "settings", Value: ...}`.

## `ADataMap`

`doors.ADataMap` exposes several named values at once.

```gox
<script
	(doors.ADataMap{
		"userId": 42,
		"theme":  "dark",
	})>
	console.log(await $data("userId"), await $data("theme"))
</script>
```

## Decoding

On the client side, `$data(...)` can decode payloads as:

- text
- binary
- JSON

For server-side source values:

- `[]byte` is sent as binary
- `string` is sent as text
- other values are sent as JSON

The detailed decoding behavior belongs with the JavaScript-side API and can be expanded later, but it is useful to know here that `$data(...)` is not limited to JSON objects.

JavaScript-side details are covered in [13-javascript.md](/Users/alex/Lib/doors/docs/docs/13-javascript.md).

## Rules

- Use hooks when plain DOM event attributes are not the right fit.
- Prefer `AHook[T]` first; move to `ARawHook` only when transport/control needs it.
- Keep hook names stable and explicit, since scripts reference them by name.
- Use `AData`/`ADataMap` for script inputs instead of hardcoded literals in JS.
