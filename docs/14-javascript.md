# JavaScript

**Doors** works well with plain browser JavaScript.

Most pages stay mostly server-rendered and add JavaScript only where the browser really needs to take over: custom events, embedded widgets, small client apps, or module-based UI islands.

## Start

The usual flow is:

1. render a managed script, usually `doors.ScriptInline`
2. attach `AData`, `AHook`, or `ARawHook` to that same script
3. read initial values with `$data(...)`
4. call Go with `$hook(...)` or `$fetch(...)`
5. optionally receive Go-triggered calls with `$on(...)`

Example:

```gox
<>
	~>doors.AData{
		Name: "userId",
		Value: userID,
	} ~>doors.AHook[string]{
		Name: "save",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			return len(r.Data()), false
		},
	} ~doors.ScriptInline{
		Source: doors.SourcePath("web/picker.ts"),
	}
</>
```

Inside `web/picker.ts`, that script can use `await $data("userId")` and `await $hook("save", value)`.

## Managed

Only **Doors**-managed scripts get helper injection and runtime lifecycle behavior.

A script is managed when it is:

- rendered through `doors.ScriptInline`
- or written as a plain inline JavaScript `<script>...</script>` block with no `src`, no `escape`, and no non-JavaScript `type`

Every managed inline script is converted into a built resource and the final page gets a `<script src="...">` tag.

The original inline body is not left in the final HTML.

That conversion is what gives you:

- helper variables like `$data`, `$hook`, and `$on`
- top-level `await`
- subtree cleanup with `$sys.clean(...)`
- build/minify/resource behavior through the resource pipeline

If you set `src`, add `escape`, or use a non-JavaScript `type`, **Doors** leaves the tag alone and the browser handles it as a normal raw script tag.

That includes `type="module"` and TypeScript MIME types.

For TypeScript, use `doors.ScriptInline` or `doors.ScriptModule`, not `type="application/typescript"` on a plain inline script.

## Scope

Managed script helpers are bound to the current script element.

In practice that means:

- `$data(name)` reads data attrs from that script tag
- `$hook(name, arg)` and `$fetch(name, arg)` call hook attrs attached to that script tag
- `$on(name, handler)` registers a handler in that script's Door scope
- `$sys.clean(...)` is tied to that script's rendered subtree

So `AData`, `AHook`, and `ARawHook` usually belong on the same script element that uses them.

With `doors.ScriptInline`, the usual pattern is to proxy those attrs onto the generated script with `~>`.

## Helpers

Inside a managed script, these helpers are available:

- `$data(name)` reads a named value exposed by `AData` or `data:name=(...)`
- `$hook(name, arg)` calls `AHook` and returns decoded JSON
- `$fetch(name, arg)` calls `AHook` or `ARawHook` and returns the raw `Response`
- `$on(name, handler)` registers a handler for `ActionEmit`
- `$G` is a shared client-side object for other managed scripts on the same page
- `$sys.ready()` resolves when the **Doors** client runtime is ready
- `$sys.clean(fn)` registers cleanup for subtree removal or replacement
- `$sys.activateLinks()` rescans active-link state for links your script created or changed
- `HookErr` is the hook error class

## Data

Use `doors.AData` or `data:name=(...)` when the script needs values from Go at render time.

```gox
<script
	data:userId=(42)
	data:theme=("light")>
	const userId = await $data("userId")
	const theme = await $data("theme")
	console.log(userId, theme)
</script>
```

If the name is missing, `$data(...)` returns `undefined`.

Payload decoding is based on the Go value type:

- `string` becomes a JavaScript `string`
- `[]byte` becomes an `ArrayBuffer`
- other values become decoded JSON

If your page already has the value on the Go side, prefer `AData` over an extra hook call just to fetch it again.

## Hooks

Use hooks when JavaScript is already in control and needs to call back into Go.

For normal clicks, inputs, and forms, stay with [Events](./08-events.md).

Use `AHook[T]` with `$hook(...)` when JSON input and output are the natural fit.

Use `ARawHook` with `$fetch(...)` when you need multipart uploads, raw bodies, or full `Response` control.

```gox
<script
	(doors.AHook[string]{
		Name: "visibility",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			println(r.Data())
			return nil, false
		},
	})>
	document.addEventListener("visibilitychange", async () => {
		await $hook("visibility", document.visibilityState)
	})
</script>
```

When `$hook(...)` or `$fetch(...)` sends data, the client picks the request body shape automatically:

- `undefined`: no body
- `FormData`: multipart form body
- `URLSearchParams`: form-urlencoded body
- `Blob`, `File`, `ReadableStream`, `ArrayBuffer`, typed arrays: raw body
- anything else: JSON body

Hook API details are covered in [Custom Attrs](./13-custom-attrs.md).

## Errors

Manual `$hook(...)` and `$fetch(...)` calls throw `HookErr`.

The main kinds are:

- `canceled`: canceled by scope or request abort
- `not_found`: hook is gone or was not attached to this script
- `unauthorized`: the instance is gone
- `bad_request`: the server could not parse the body
- `network`: transport failure
- `server`: 5xx response
- `other`: other non-ok response
- `capture`: client-side helper or capture error

Catch these in script code when failure is part of the normal flow:

```gox
<script>
	try {
		await $hook("save", {ok: true})
	} catch (err) {
		if (err instanceof HookErr && err.notFound()) {
			console.log("hook is gone")
		}
	}
</script>
```

Event attrs do not use `try/catch` in user code for this; they use `OnError` actions instead.

## Emit

Use `$on(name, handler)` when Go should call JavaScript through `doors.ActionEmit`.

```gox
<script>
	$on("alert", (message, err) => {
		if (err) {
			console.log(err.kind)
			return
		}
		window.alert(message)
	})
</script>
```

Handler lookup is scoped through the Door tree.

When Go runs `ActionEmit`, **Doors** starts from the Door where that action was created and walks outward through parent Doors until it finds a matching handler.

So the nearest matching handler wins, and local handlers shadow outer ones.

`$on(...)` handlers used by actions must stay synchronous.

If they return a `Promise`, the action fails.

Action details are covered in [Actions](./12-actions.md).

## Modules

Use `doors.ScriptModule` when you want real ES modules and import-map based loading.

If `Specifier` is set, **Doors** adds that module to the page import map so managed scripts can `await import("specifier")`.

If `Specifier` is empty, `ScriptModule` just renders a normal module script tag.

```gox
<>
	<div id="react-root"></div>

	~doors.ScriptModule{
		Specifier: "react_app",
		Source: doors.SourcePath("web/react/index.tsx"),
		Output: doors.ScriptOutputBundle,
	}

	<script>
		const app = await import("react_app")
		app.init(document.getElementById("react-root"))
	</script>
</>
```

This is the same pattern used in the `imports` tests to mount React and Preact components into a **Doors** page.

Script and module builds go through the esbuild-backed resource pipeline.

Use `Profile` when you want a named build profile, and see [Resources](./15-resources.md) and [Configuration](./19-configuration.md) for the build and hosting details.

## Cleanup

Use `$sys.clean(...)` for timers, global listeners, and embedded widgets that need teardown.

```gox
<script>
	const onResize = () => {
		console.log(window.innerWidth)
	}

	window.addEventListener("resize", onResize)

	$sys.clean(() => {
		window.removeEventListener("resize", onResize)
	})
</script>
```

If your script creates or rewrites links that participate in active-link indication, call `$sys.activateLinks()` after that change.

## Rules

- Prefer `doors.ScriptInline` for real page scripts.
- Every managed inline script ends up as a `src`-backed resource.
- Attach `AData`, `AHook`, and `ARawHook` to the same script that uses them.
- Use hooks and data for JavaScript-to-Go work.
- Use `ActionEmit` and `$on(...)` for Go-to-JavaScript work.
- Use `doors.ScriptModule` for ES modules, React/Preact islands, and import-map workflows.
- Use `$sys.clean(...)` whenever the script owns listeners, timers, or mounted widgets.
