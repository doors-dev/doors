# JavaScript

This page covers script resources and JavaScript runtime features in **Doors**.

If generic resource syntax is new, start with [Resources](./14-resources.md). This page focuses on what is specific to `<script>` and modules.

## Start

Most pages start with one of these:

- page-local code written directly in the template: plain inline `<script>...</script>`
- page-local code kept in a file or bytes: `<script src=(doors.ResourceLocalFS("web/app.ts")) output="inline"></script>`
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
		type="typescript"
		output="inline"></script>
</>
```

Inside `web/picker.ts`, that script can use `await $data("userId")` and `await $hook("save", value)`.

## Scripts

### Types

Use:

- no `type` for regular JavaScript
- `type="module"` for ES modules
- `type="typescript"` for TypeScript
- `type="module/typescript"` for module TypeScript

### Inline

Use inline script when the code belongs only to this page:

```gox
<script>
	await $sys.ready()
	console.log("managed inline script")
</script>
```

By default this is a managed script resource. **Doors** builds the body, turns it into a `src`-backed resource, and runs it through the managed runtime wrapper.

Use `output="inline"` when the code lives in a file or bytes but should behave the same way:

```gox
<script
	src=(doors.ResourceLocalFS("web/picker.ts"))
	type="typescript"
	output="inline"></script>
```

Inline rules:

- plain inline JavaScript is managed by default
- `output="inline"` works only with buildable `src` sources
- actual inline TypeScript bodies are not supported
- inline module bodies are not supported
- inline scripts cannot be bundled
- `output="raw"` leaves the original tag alone

### Linked `src`

Use a regular linked script when the code should not be loaded as a module:

```gox
<script
	src=(doors.ResourceLocalFS("web/app.ts"))
	type="typescript"></script>
```

Useful `output` values are:

- `output="default"` or omitted: normal built or prepared script resource
- `output="bundle"`: bundle dependencies into one output
- `output="raw"`: skip the build pipeline and leave the script raw

Common `src` shapes are:

- buildable app content: `ResourceLocalFS`, `ResourceFS`, `ResourceBytes`, `ResourceString`
- already-hosted local URL: plain string such as `"/assets/app.js"`
- external URL: `doors.ResourceExternal("https://cdn.example.com/app.js")`
- handler-backed source: `doors.ResourceHook(...)`, `doors.ResourceHandler(...)`
- proxy-backed source: `doors.ResourceProxy(...)`

In practice:

- buildable sources go through the JS pipeline unless `output="raw"` is used
- plain strings are just direct URLs
- `ResourceExternal(...)` is a direct URL plus automatic CSP source collection
- handler and proxy sources become hook-backed URLs

Raw TypeScript is not supported. If the source is TypeScript, let **Doors** build it.

### Managed runtime

Managed scripts get:

- helper variables like `$data`, `$hook`, and `$on`
- top-level `await`
- subtree cleanup with `$sys.clean(...)`
- build and resource handling through the JS pipeline

A script is managed when it is:

- a plain inline JavaScript `<script>...</script>` block with no `src`, no `output="raw"`, and no non-JavaScript `type`
- or a buildable `src` script with `output="inline"`

If you set `output="raw"` or use an unsupported non-JavaScript `type`, **Doors** leaves the tag alone and the browser handles it as a normal raw script tag.

For managed TypeScript, editor tooling is nicer if you add ambient declarations for helpers like `$data`, `$hook`, `$fetch`, `$on`, and `$sys`. TSserver may still warn about top-level `await`; that is expected for managed inline script bodies and `output="inline"` scripts.

## Modules

Use modules when you want `import`, `export`, or import-map based loading.

```gox
<script
	src=(doors.ResourceLocalFS("web/app.ts"))
	type="module"
	output="bundle"></script>
```

Use `output="bundle"` when the module has dependencies that should be bundled into one output.

Use `specifier` when the module should be registered in the page import map:

```gox
<div id="app"></div>

<script
	src=(doors.ResourceLocalFS("web/react/index.tsx"))
	type="module"
	output="bundle"
	specifier="app"></script>

<script>
	const { mount } = await import("app")
	mount(document.getElementById("app"))
</script>
```

On a regular `<script>` tag, `specifier` also forces module mode.

`specifier` matters only during the initial render, before the browser starts resolving module specifiers. In practice, modules you want available through the import map should usually be declared in the page head.

On a regular `<script>` tag:

- if the tag has only control attrs, **Doors** omits the tag and registers only the import-map entry
- if the tag also has any other attr, the tag stays in the HTML and the import-map entry is also registered

Here, control attrs means attrs that only configure the resource, such as `src`, `type`, `output`, `specifier`, `name`, `profile`, `private`, or `nocache`.

So if you want the module both in the import map and loaded by a regular `<script>` tag, add any normal attr such as `id`, `async`, or `crossorigin`.

Example that keeps the tag:

```gox
<script
	src=(doors.ResourceLocalFS("web/app.ts"))
	type="module"
	output="bundle"
	id="app-module"
	specifier="app"></script>
```

For preload:

```gox
<link
	rel="modulepreload"
	href=(doors.ResourceBytes(moduleJS))
	specifier="app">
```

Use `rel="modulepreload"` when the module should be preloaded and also registered in the import map.

`<link rel="modulepreload">` always stays rendered. With `specifier`, **Doors** also registers the import-map entry.

## Attrs

These attrs control script resource behavior:

- `output`: `default`, `inline`, `bundle`, or `raw`
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
	output="bundle"
	name="react_app.js"
	profile="react"
	private
	specifier="react_app"></script>
```

Build configuration itself is covered in [Configuration](./20-configuration.md).

## Go Bridge

### Data

Use `doors.AData` or `data:name=(...)` when the script needs values from Go at render time:

```gox
<script
	data:userId=(userID)
	data:theme=(theme)>
	const userId = await $data("userId")
	const theme = await $data("theme")
	console.log(userId, theme)
</script>
```

`$data(...)` always returns a promise. `string` becomes a JavaScript string, `[]byte` becomes an `ArrayBuffer`, and other values are decoded from JSON. In practice, that means binary data still needs `await` before you can use the `ArrayBuffer`.

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

When `$hook(...)` or `$fetch(...)` sends data, the client chooses the request body shape automatically. Hook API details are covered in [Custom Attrs](./13-custom-attrs.md).

### Client API

The main client-side helpers are:

- `$data(name: string): Promise<any>`
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

## Choose

- Page-local code written here: plain inline `<script>...</script>`
- Page-local code kept in a file or bytes: buildable `src=(...)` with `output="inline"`
- Reusable or TypeScript code: buildable `src=(...)`
- Module loading: `type="module"`
- Import by name: add `specifier="..."`
- Keep a `specifier` script tag rendered: add any normal attr such as `id`
- Preload and register a module: `rel="modulepreload"` with `specifier`
- Shared public resource URL is not wanted: use `private` or `nocache`
