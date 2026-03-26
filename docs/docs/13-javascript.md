# JavaScript

Doors can work with plain browser JavaScript. You do not need a separate frontend framework, but you can integrate one when it helps.

## Start

Most pages only need this flow:

1. render script with `doors.ScriptInline`
2. pass inputs with `AData` / `data:name=(...)`
3. call Go with `$hook(...)` when needed
4. receive Go-triggered events with `$on(...)` when needed

## Scripts

The recommended way to write page JavaScript is:

- keep a `.ts` file next to the Go code that uses it
- embed that file into Go
- render it with `doors.ScriptInline`

```go
package page

import _ "embed"

//go:embed scroll.ts
var scrollScript string
```

```gox
<>
	~>doors.AData{
		Name: "id",
		Value: id,
	} ~doors.ScriptInline{
		Source: doors.SourceScriptString{
			Content: scrollScript,
			TypeScript: true,
		},
	}
</>
```

That keeps the script in a real `.ts` file while still attaching it to the current Doors subtree and context.

It also means your TypeScript language server can work on a normal `.ts` file instead of a string literal inside Go code.

`ScriptInline` is cooked through the Doors resource pipeline so helper injection, top-level `await`, lifecycle binding, and resource serving behavior stay consistent.

## Plain Script

Plain inline `<script>...</script>` also works on Doors pages.

Doors captures that script body, builds it as a resource, and keeps it connected to the Doors runtime. That is why top-level `await` works in normal Doors script blocks.

```gox
<script>
	const value = await $hook("countText", "hello")
	console.log(value)
</script>
```

## Escape

Doors only rewrites plain script tags automatically.

It leaves the tag alone if you:

- set `src`
- set a non-JavaScript `type`
- add the `escape` attribute

Use that when you want a literal browser script tag instead of Doors runtime injection.

## Helpers

Inside a processed Doors script, these helpers are available:

- `$on(name, handler)` registers a handler for `ActionEmit`
- `$data(name)` reads data exposed with `AData` or `data:name=(...)`
- `$hook(name, arg)` calls `AHook` and returns decoded JSON
- `$fetch(name, arg)` calls `AHook` or `ARawHook` and returns raw `Response`
- `$G` is a shared client-side global object
- `$sys.ready()` resolves when the Doors client runtime is ready
- `$sys.clean(fn)` registers cleanup for script removal
- `$sys.activateLinks()` rescans active-link indication for links your script created or changed
- `HookErr` is the hook error class

These helpers are available in Doors-managed scripts, including scripts rendered through `doors.ScriptInline`.

## Data

Expose values to JavaScript with `doors.AData`, `doors.ADataMap`, or the `data:name=(...)` shorthand.

```gox
<script data:settings=(map[string]any{
	"units": "metric",
	"days":  7,
})>
	const settings = await $data("settings")
	console.log(settings.units, settings.days)
</script>
```

`$data(...)` decodes payloads in three ways:

- text
- binary
- JSON

For `AData` / `ADataMap`, the payload type is chosen by Go value type:

- `[]byte` -> binary payload
- `string` -> text payload
- anything else -> JSON payload

On the JS side this means:

- binary is returned as `ArrayBuffer`
- text is returned as `string`
- JSON is returned as the decoded JS value

Most application code sees JSON because that is the common case, but the channel itself is not JSON-only.

## Hooks

Use hooks when JavaScript needs to call Go directly.

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

Use:

- `$hook(...)` when JSON input/output is enough
- `$fetch(...)` when you want the raw `Response`
- `doors.ARawHook` when the server side needs raw body or multipart handling

Hook details are covered in [11-hooks.md](/Users/alex/Lib/doors/docs/docs/11-hooks.md).

## Transfer

When `$hook(...)` or `$fetch(...)` sends data, the client chooses the request body shape automatically:

- `undefined`: no body
- `FormData`: multipart form body
- `URLSearchParams`: form-urlencoded body
- `Blob`: raw blob body
- `File`: octet-stream body
- `ReadableStream`: octet-stream body
- `ArrayBuffer` or typed arrays: octet-stream body
- anything else: JSON body

That makes hooks work well for both structured app calls and lower-level uploads.

## Errors

Hook failures are reported as `HookErr`.

The current kinds are:

- `canceled`
- `unauthorized`
- `not_found`
- `other`
- `network`
- `bad_request`
- `server`
- `capture`

`$hook(...)` and `$fetch(...)` throw these errors, so catch them in script code when failure is part of the normal flow.

```gox
<script>
	try {
		await $hook("save", {ok: true})
	} catch (err) {
		if (err instanceof HookErr && err.unauthorized()) {
			console.log("instance is gone")
		}
	}
</script>
```

## Emit

JavaScript can also receive calls from Go through `$on(...)` and `doors.ActionEmit`.

```gox
<script>
	$on("alert", (message) => {
		window.alert(message)
		return "ok"
	})
</script>
```

```go
doors.Call(ctx, doors.ActionEmit{
	Name: "alert",
	Arg:  "Hello!",
})
```

When `ActionEmit` is triggered from an `OnError` action pipeline, `$on` receives the hook error as the second argument (`(arg, err)`).

If you need a return value from the client handler, use `doors.XCall[T](...)`.

Action details are covered in [12-actions.md](/Users/alex/Lib/doors/docs/docs/12-actions.md).

## Imports

Doors prepares the page import map from the script modules used on that page.

When a `doors.ScriptModule` has a `Specifier`, Doors adds that specifier to the page import map, and user scripts can then import it normally:

```gox
<>
	~doors.ScriptModule{
		Specifier: "app",
		Source: doors.SourcePath("web/app.ts"),
	}

	<script>
		const app = await import("app")
		app.init()
	</script>
</>
```

That is how user scripts can require module code by specifier instead of hardcoding built resource URLs.

Full resource API details are covered in [14-resources.md](/Users/alex/Lib/doors/docs/docs/14-resources.md).

## Cleanup

Use `$sys.clean(...)` for work tied to the current script or subtree lifetime.

```gox
<script>
	const id = setInterval(() => {
		console.log("tick")
	}, 1000)

	$sys.clean(() => {
		clearInterval(id)
	})
</script>
```

The cleanup runs when the related Door subtree is removed or replaced.

## Rules

- Prefer `doors.ScriptInline` with embedded `.ts` files for page scripts.
- Use plain `<script>` when inline local code is genuinely simpler.
- Use hooks and data for JS-to-Go integration.
- Use `ActionEmit` for Go-to-JS integration.
- Use `$sys.clean(...)` for timers, listeners, and embedded widgets.
- Use `data:name=(...)` or `AData` for initial script data.
