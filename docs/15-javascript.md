# JavaScript

This page covers script resources and JavaScript runtime features in **Doors**.

If generic resource syntax is new, start with [Resources](./14-resources.md). This page focuses on what is specific to `<script>` and modules.

It also covers the Go to JavaScript bridge: `doors.AData`, `doors.AHook[...]`, `doors.ARawHook`, `$data(...)`, `$hook(...)`, `$fetch(...)`, and `$on(...)`.

## Start

Most pages start with one of these:

- page-local code written directly in the template: plain inline `<script>...</script>`
- page-local code kept in a file or bytes: `<script src=(doors.ResourceLocalFS("web/app.ts")) inline></script>`
- shorthand bytes: `<script src=(appJS)></script>`
- already-hosted URL: `<script src="/assets/app.js"></script>`

Example:

```go
//go:embed web/picker.ts
var pickerTS []byte
```

```gox
<>
	~>(doors.AHook[string]{
		Name: "save",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			_ = r.Data()
			return nil, true
		},
	}) <script
		data:userId=(userID)
		src=(pickerTS)
		type="text/typescript"
		inline></script>
</>
```

Inside `web/picker.ts`, that script can use `$data("userId")` and `await $hook("save", value)`.

## Scripts

### Mental Model

There are three common script shapes in **Doors**:

- browser script: the browser just loads and runs it
- managed script: **Doors** wraps it and gives it runtime helpers like `$data`, `$hook`, `$on`, and `$sys`
- module script: the browser treats it as an ES module

Use a managed script when the code is page-local and should participate in the **Doors** runtime.

Use a module when the code should be imported, exported, or wired through the import map.

Use a raw or plain browser script when you want normal browser behavior with no **Doors** runtime wrapper.

### Types

Use:

- no `type` for regular JavaScript
- `type="module"` for ES modules
- `type="typescript"` for TypeScript
- `type="module/typescript"` for module TypeScript

### Output

Script output is controlled with boolean attrs:

- omitted: normal built or prepared script resource
- `inline`: treat a buildable `src` script like a managed script
- `bundle`: bundle dependencies into one output
- `raw`: skip script transformation and serve the source as-is

In practice:

- omitted is the normal choice for linked scripts
- `inline` is for code that should behave like page-local managed script code
- `bundle` is for modules with dependencies
- `raw` is for exact browser behavior

### Managed Scripts

A script is managed when it is:

- a plain inline JavaScript `<script>...</script>` block with no `src`, no `raw`, and no non-JavaScript `type`
- or a buildable `src` script with `inline`

Managed scripts get:

- helper variables like `$data`, `$hook`, `$fetch`, `$on`, and `$sys`
- top-level `await`
- subtree cleanup with `$sys.clean(...)`

If you want the browser to handle the script normally, use `raw` or a direct browser-usable URL.

### Inline

Use inline script when the code belongs only to this page:

```gox
<script>
	await $sys.ready()
	console.log("managed inline script")
</script>
```

By default this is a managed script resource. **Doors** builds the body, turns it into a `src`-backed resource, and runs it through the runtime wrapper.

Use `inline` when the code lives in a file or bytes but should behave the same way:

```gox
<script
	src=(doors.ResourceLocalFS("web/picker.ts"))
	type="text/typescript"
	inline></script>
```

Inline rules:

- plain inline JavaScript is managed by default
- actual inline TypeScript bodies are not supported
- inline module bodies are not supported
- inline scripts cannot be bundled
- `raw` leaves the original tag alone

### Linked `src`

Use a regular linked script when the code should not be loaded as a module:

```gox
<script
	src=(doors.ResourceLocalFS("web/app.ts"))
	type="text/typescript"></script>
```

Common `src` shapes are:

- buildable app content: `ResourceLocalFS`, `ResourceFS`, `ResourceBytes`, `ResourceString`
- already-hosted local URL: plain string such as `"/assets/app.js"`
- external URL: `doors.ResourceExternal("https://cdn.example.com/app.js")` for a direct browser URL that also participates in CSP source collection
- handler-backed source: `doors.ResourceHook(...)`, `doors.ResourceHandler(...)`
- proxy-backed source: `doors.ResourceProxy(...)`

Buildable `src` scripts go through the JS pipeline unless `raw` is used. Plain strings are just direct URLs.

Raw TypeScript is not supported. If the source is TypeScript, let **Doors** build it.

For managed TypeScript, editor tooling is nicer if you add ambient declarations for helpers like `$data`, `$hook`, `$fetch`, `$on`, and `$sys`. TSserver may still warn about top-level `await`; that is expected for managed inline script bodies and `inline` scripts.

## Modules

Use modules when you want `import`, `export`, or import-map based loading.

```gox
<script
	src=(doors.ResourceLocalFS("web/app.ts"))
	type="module"
	bundle></script>
```

Use `bundle` when the module has dependencies that should be bundled into one output.

Use `specifier` when the module should be registered in the page import map:

```gox
<div id="app"></div>

<script
	src=(doors.ResourceLocalFS("web/react/index.tsx"))
	type="module"
	bundle
	specifier="app"></script>

<script>
	const { mount } = await import("app")
	mount(document.getElementById("app"))
</script>
```

> For module scripts, `specifier` is usually the main way to wire modules together. If a module is not fully standalone, register it with a specifier and import it by that name.

On a regular `<script>` tag, `specifier` does not replace module typing. Use `type="module"` together with `specifier`.

`specifier` matters only during the initial render, before the browser starts resolving module specifiers. In practice, modules you want available through the import map should usually be declared in the page head.

### Import Without Execution

If a module should be available in the import map but should not be executed by a `<script>` tag, use `rel="modulepreload"` with `specifier`:

```gox
<link
	rel="modulepreload"
	href=(doors.ResourceLocalFS("web/app.ts"))
	specifier="app">
```

That registers `"app"` in the import map and preloads the module, but it does not execute it as a page script.

Later, load it explicitly:

```gox
<script>
	const app = await import("app")
	app.mount()
</script>
```

## Attrs

These attrs control script resource behavior:

- output attrs: `inline`, `bundle`, `raw`
- `specifier`: register a module in the import map
- `name`: readable output file name
- `profile`: named esbuild profile
- `private`: host through an instance-scoped hook URL while still using the resource pipeline
- `nocache`: host through an instance-scoped hook URL without shared resource caching

Example:

```gox
<script
	src=(doors.ResourceLocalFS("web/react/index.tsx"))
	type="module"
	bundle
	name="react_app.js"
	profile="react"
	private
	specifier="react_app"></script>
```

Plain string URLs are passed through as-is. `doors.ResourceExternal(...)` keeps the browser URL direct while also adding that host to CSP. Handler and proxy sources already produce hook-backed URLs.

Use `private` when the script should not be publicly reachable.

Use `nocache` for dynamically generated script output that should not use shared resource caching.

Build configuration itself is covered in [Configuration](./21-configuration.md).

## Go Bridge

### Data Binding

Use `doors.AData` or `data:name=(...)` when the script needs values from Go at render time:

```gox
<script
	data:userId=(userID)
	data:theme=(theme)>
	const userId = $data("userId")
	const theme = $data("theme")
	console.log(userId, theme)
</script>
```

`$data(...)` returns the decoded value directly for `string` and JSON-backed values. For `[]byte`, it returns a promise that resolves to an `ArrayBuffer`, so binary data still needs `await`.

GoX shorthand such as `data:userId=(userID)` is equivalent to attaching `doors.AData{Name: "userId", Value: userID}`.

Attach several `AData` attrs or several `data:name=(...)` entries when the script needs more than one server-provided value.

If the value is already known at render time, prefer `AData` over an extra hook call.

### Hooks

Use hooks when JavaScript is already in control and needs to call back into Go. For normal clicks, inputs, and forms, stay with [Events](./08-events.md).

There are two independent choices:

- on the Go side: `AHook[T]` or `ARawHook`
- on the JavaScript side: `$hook(...)` or `$fetch(...)`

Use:

- `AHook[T]` for typed JSON input and output
- `ARawHook` for raw request or response control
- `$hook(...)` when the script wants the helper-style result
- `$fetch(...)` when the script wants the raw `Response`

These are not a 1:1 pair. For example, `$hook(...)` can call `ARawHook`.

For `AHook[T]`, the handler receives `doors.RequestHook[T]`, so it can:

- read `r.Data()`
- read or set cookies
- schedule `r.After(...)` actions

Return `false` to keep the hook active.

Return `true` to remove it after the call.

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

### Raw Variant

Use `doors.ARawHook` when the Go side needs lower-level transport control.

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

When `$hook(...)` or `$fetch(...)` sends data, the client chooses the request body shape automatically:

- `undefined`: no body
- `FormData`: `multipart/form-data`
- `URLSearchParams`: `application/x-www-form-urlencoded`
- `Blob`: raw blob body
- `File`: raw file body
- `ReadableStream`: `application/octet-stream`
- any other value: JSON

### Client API

The main client-side helpers are:

- `$data<T = any>(name: string): T | Promise<ArrayBuffer>`
- `$hook(name: string, arg?: any): Promise<any>`
- `$fetch(name: string, arg?: any): Promise<Response>`
- `$on(name: string, handler: (arg: any, err?: HookErr) => any): void`
- `$sys.ready(): Promise<void>`
- `$sys.clean(fn: () => void | Promise<void>): void`
- `$sys.activateLinks(): void`
- `HookErr`

Manual `$hook(...)` and `$fetch(...)` calls throw `HookErr`. Catch it when failure is part of the normal flow.

### Emit and cleanup

Use `$on(name, handler)` when Go should call JavaScript through `doors.ActionEmit`:

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

The nearest matching handler wins in the Door tree. `$on(...)` handlers used by actions must stay synchronous.

Use `$sys.clean(...)` for timers, global listeners, and embedded widgets that need teardown:

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
