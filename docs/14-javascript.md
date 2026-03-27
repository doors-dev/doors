# JavaScript

**Doors** works well with plain browser JavaScript.

Most pages stay mostly server-rendered and add JavaScript only where the browser really needs to take over: custom events, embedded widgets, small client apps, or module-based UI islands.

## Start

The usual flow is:

1. render a managed script, usually with the `<script>` resource syntax described below
2. attach `AData`, `AHook`, or `ARawHook` to that same script
3. read initial values with `$data(...)`
4. call Go with `$hook(...)` or `$fetch(...)`
5. optionally receive Go-triggered calls with `$on(...)`

Example:

```go
//go:embed web/picker.ts
var pickerTS []byte
```

```gox
<>
	~>(doors.AData{
		Name: "userId",
		Value: userID,
	}, doors.AHook[string]{
		Name: "save",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			_ = r.Data()
			return nil, true
		},
	}) <script
		src=(doors.SourceBytes(pickerTS))
		type="application/typescript"
		output="inline"></script>
</>
```

Embedding the script like this is often convenient because the app stays self-contained in one binary.

Inside `web/picker.ts`, that script can use `await $data("userId")` and `await $hook("save", value)`.

## Resource Syntax

The current tag syntax mirrors the older resource functionality still described in [Resources](./15-resources.md).

For now, the easiest way to map them is:

- plain managed inline `<script>...</script>` or buildable `src` with `output="inline"`: old `ScriptInline`
- buildable `src` with default output: old `ScriptCommon`
- buildable `src` with `type="module"`: old `ScriptModule`

Use plain inline script when the code belongs only to this page and should be managed by **Doors**:

```gox
<script>
	await $sys.ready()
	console.log("managed inline script")
</script>
```

Use a buildable `src` when the code lives in a file or in Go bytes or strings:

```gox
<script
	src=(doors.SourceLocalFS("web/picker.ts"))
	type="application/typescript"
	output="inline"></script>
```

That `output="inline"` form is the new equivalent of the old `ScriptInline` resource component.

Use the default output when you want a normal script resource:

```gox
<script
	src=(doors.SourceLocalFS("web/app.js"))></script>
```

Use `type="module"` for ES modules:

```gox
<script
	src=(doors.SourceLocalFS("web/app.ts"))
	type="module"
	output="bundle"></script>
```

Use `specifier` when the module should go into the page import map instead of rendering its own tag:

```gox
<script
	src=(doors.SourceLocalFS("web/react/index.tsx"))
	type="module"
	output="bundle"
	specifier="react_app"></script>

<script>
	const app = await import("react_app")
	app.init()
</script>
```

`specifier` and `specifieronly` affect only the initial page render.

That means import-map entries should be declared before the browser starts resolving module specifiers.

In practice, put all scripts that need to register import-map entries in the page head.

Use `specifieronly` for import-map registration without emitting the loading tag:

```gox
<script
	src=(doors.SourceLocalFS("web/react/index.tsx"))
	output="bundle"
	specifieronly="react_app"></script>
```

Available script `src` source shapes are:

- buildable local file: `doors.SourceLocalFS("web/app.ts")`
- buildable embedded file: `doors.SourceFS{FS: webFS, Entry: "app.ts"}`
- buildable bytes or string from Go: `doors.SourceBytes(appJS)` / `doors.SourceString(appJS)`
- already hosted local URL: plain string such as `"/assets/app.js"`
- external URL: `doors.SourceExternal("https://cdn.example.com/app.js")`
- handler-backed source: `doors.SourceHook(...)`

The main script attributes are:

- `output="default"` or omitted: build a normal script resource
- `output="inline"`: build a managed inline-style script resource
- `output="bundle"`: bundle dependencies into the output
- `output="raw"`: skip the managed or build pipeline and leave the script raw
- `type="module"`: module build or module URL behavior
- `specifier` / `specifieronly`: add the built or direct path to the page import map and force module mode
- `profile`: choose a named esbuild profile
- `private` / `nocache`: change hosting mode the same way the old resource API did

For import-map driven modules, it is recommended to place those script declarations in the HTML head so the browser sees the full import map during the initial page load.

## Managed

Only **Doors**-managed scripts get helper injection and runtime lifecycle behavior.

A script is managed when it is:

- written as a plain inline JavaScript `<script>...</script>` block with no `src`, no `output="raw"`, and no non-JavaScript `type`
- or written as `<script src=(buildable source) output="inline"></script>`

Every managed inline script is converted into a built resource and the final page gets a `<script src="...">` tag.

The original inline body is not left in the final HTML.

Managed inline scripts are not executed as raw top-level browser source.

Instead, **Doors** wraps the script body in an anonymous async function and runs it through the client runtime with the framework helpers already in scope.

That is what gives you:

- helper variables like `$data`, `$hook`, and `$on`
- top-level `await`
- subtree cleanup with `$sys.clean(...)`
- build/minify/resource behavior through the resource pipeline

If you set `output="raw"` or use an unsupported non-JavaScript `type`, **Doors** leaves the tag alone and the browser handles it as a normal raw script tag.

For TypeScript and modules, use a buildable `src` form from the syntax above, not a plain inline script body.

## Build

Buildable JavaScript in **Doors** goes through esbuild before the browser sees it.

In practice:

- a managed plain inline `<script>...</script>` is converted into a built resource and goes through esbuild
- `<script src=(buildable source) output="inline">` goes through the managed inline build path
- `<script src=(buildable source)>` and `<script src=(buildable source) type="module">` go through the normal resource build path
- `output="bundle"` enables dependency bundling
- `output="raw"` skips that build step
- plain string URLs, `SourceExternal`, and handler-backed sources do not go through esbuild

Buildable sources here means the source comes from your app code, such as `SourceLocalFS`, `SourceFS`, `SourceString`, or `SourceBytes`.

That build step is what gives **Doors**:

- TypeScript support for buildable script sources
- minification by default
- bundling when `output="bundle"` is used
- JSX handling when your esbuild config enables it

If you need named build profiles, use the `profile` attribute.

Plain managed inline `<script>...</script>` uses the default profile.

Build settings are configured at the router level with `doors.UseESConf(...)`.

See [Resources](./15-resources.md) for the serving side and [Configuration](./19-configuration.md) for the esbuild config itself.

### TypeScript

When you edit a `.ts` file used by a buildable managed script such as `<script src=(doors.SourceLocalFS("web/picker.ts")) output="inline"></script>`, a small ambient declaration file can make editor tooling much nicer.

For example:

```ts
declare const $on: (name: string, handler: (arg: any) => any) => void;
declare const $data: (name: string) => any;
declare const $hook: (name: string, arg: any) => Promise<any>;
declare const $fetch: (name: string, arg: any) => Promise<Response>;
declare const $G: { [key: string]: any };
declare const $sys: {
	ready: () => Promise<undefined>,
	clean: (handler: () => void | Promise<void>) => void,
	activateLinks: () => void,
};
declare const HookErr: new (...args: any[]) => Error;
```

TSserver may still warn about top-level `await`.

That warning is expected for managed `output="inline"` scripts and plain managed inline script bodies, because **Doors** wraps them in an anonymous async function before they run in the browser.

## Scope

Managed script helpers are bound to the current script element.

In practice that means:

- `$data(name)` reads data attrs from that script tag
- `$hook(name, arg)` and `$fetch(name, arg)` call hook attrs attached to that script tag
- `$on(name, handler)` registers a handler in that script's Door scope
- `$sys.clean(...)` is tied to that script's rendered subtree

So `AData`, `AHook`, and `ARawHook` usually belong on the same script element that uses them.

With managed script syntax, the usual pattern is to proxy those attrs onto the script with `~>`.

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
		await $hook("save", "hello")
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

Use script resource syntax with `type="module"` when you want real ES modules and import-map based loading.

If `specifier` is set, **Doors** adds that module to the page import map so managed scripts can `await import("specifier")`.

If `specifier` is empty, the script renders a normal module tag.

```gox
<>
	<div id="react-root"></div>

	<script
		src=(doors.SourceLocalFS("web/react/index.tsx"))
		type="module"
		output="bundle"
		specifier="react_app"></script>

	<script>
		const app = await import("react_app")
		app.init(document.getElementById("react-root"))
	</script>
</>
```

This is the same pattern used in the imports tests to mount React and Preact components into a **Doors** page.

`specifier` works only during the initial page render, so modules you want available through the import map should be declared in the page head.

Module builds do not wrap the module with `$data`, `$hook`, or the other managed-script helpers.

Those helpers belong to managed inline scripts.

Script and module builds go through the esbuild-backed resource pipeline.

Use `profile` when you want a named build profile, and see [Resources](./15-resources.md) and [Configuration](./19-configuration.md) for the build and hosting details.

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

- Prefer managed inline script syntax for real page scripts.
- Every managed inline script ends up as a `src`-backed resource.
- Attach `AData`, `AHook`, and `ARawHook` to the same script that uses them.
- Use hooks and data for JavaScript-to-Go work.
- Use `ActionEmit` and `$on(...)` for Go-to-JavaScript work.
- Use `type="module"` plus `specifier` when you want ES modules, React/Preact islands, and import-map workflows.
- Use `$sys.clean(...)` whenever the script owns listeners, timers, or mounted widgets.
