# Resources

Resources are how **Doors** serves JS and CSS to the browser.

Use them when the asset belongs to your app and **Doors** should build it, host it, or give the browser a URL for it.

Most pages end up in one of four patterns:

- page-local JavaScript near the Go code: `doors.ScriptInline`
- a regular script or module entry: `doors.ScriptCommon` or `doors.ScriptModule`
- a stylesheet: `doors.Style`
- a private file or handler-backed URL on one element: `ASrc`, `ARawSrc`, `AFileHref`, `ARawFileHref`

## Script

Use script resources when the browser should load JavaScript through **Doors**.

### Inline

Use `ScriptInline` when the script belongs to one page or component.

This is the normal choice for page-local `.ts` or `.js` kept next to the Go code that renders it.

```gox
<>
	~doors.ScriptInline{
		Source: doors.SourceScriptString{
			Content: scrollScript,
			TypeScript: true,
		},
	}
</>
```

`ScriptInline` accepts only buildable sources:

- `SourcePath`
- `SourceFS`
- `SourceScriptString`
- `SourceScriptBytes`

It does not accept `SourceLocal` or `SourceExternal`.

`ScriptInline` always ends up as a `src`-backed script resource.

The browser does not receive the original inline body directly.

That is what lets **Doors** inject the managed-script runtime wrapper used by [JavaScript](./14-javascript.md).

### Common

Use `ScriptCommon` when you want a regular script asset without module behavior.

```gox
<>
	~doors.ScriptCommon{
		Source: doors.SourcePath("web/app.js"),
	}
</>
```

It can:

- render its own `<script src="...">` tag
- be attached as a modifier to an existing `<script>` element

Use this when the browser should just load a script file and you do not need import-map registration.

### Module

Use `ScriptModule` for ES modules.

If `Specifier` is empty, it renders a normal module script tag.

If `Specifier` is set, **Doors** adds that module to the page import map.

```gox
<>
	~doors.ScriptModule{
		Specifier: "app",
		Source: doors.SourcePath("web/app.ts"),
		Output: doors.ScriptOutputBundle,
	}

	<script>
		const app = await import("app")
		app.init()
	</script>
</>
```

This is the pattern used in the `imports` tests for module loading and React or Preact mounting.

When used as a modifier:

- on `<script>`, it sets `type="module"` and `src`
- on `<link>`, it sets `rel="modulepreload"` and `href`

That makes it the right tool for import-map entries, modulepreload tags, and client-side islands.

### Build

`ScriptOutput` matters for `ScriptCommon` and `ScriptModule`:

- `doors.ScriptOutputDefault`: normal build pipeline
- `doors.ScriptOutputBundle`: bundle dependencies into the output
- `doors.ScriptOutputRaw`: skip the build step and serve the source as-is

Use `Bundle` when the module should carry its dependency graph with it.

Use `Raw` when the source is already built and **Doors** should just serve it.

`Profile` selects a named esbuild build profile.

That is especially useful for module-heavy code such as TSX bundles.

Detailed build configuration is covered in [Configuration](./19-configuration.md).

## Style

Use `Style` for CSS assets.

```gox
<>
	~doors.Style{
		Source: doors.SourcePath("web/app.css"),
	}
</>
```

It renders a stylesheet link and serves the CSS through the same resource pipeline.

`Minify` matters only for buildable style sources.

Style uses the same source types, host modes, and serving behavior described below.

## Source

Script and style resources share the same source and hosting rules.

### Types

Choose the source type by where the content already lives:

- file on disk: `doors.SourcePath("...")`
- embedded filesystem: `doors.SourceFS{FS, Path, Name}`
- script bytes or string already in Go: `doors.SourceScriptBytes` / `doors.SourceScriptString`
- style bytes or string already in Go: `doors.SourceStyleBytes` / `doors.SourceStyleString`
- already hosted local URL: `doors.SourceLocal("/...")`
- external URL: `doors.SourceExternal("https://...")`

Use `SourceLocal` or `SourceExternal` only when the asset is already hosted and should not be built by **Doors**.

For those direct-URL sources:

- build output settings do not apply
- bundling does not apply
- **Doors** just uses the URL you gave it

For `SourceExternal`, **Doors** also adds the source to CSP automatically.

### Host

`HostMode` controls how built resources are exposed:

- `doors.HostModePublic`: public resource URL
- `doors.HostModePrivate`: instance-scoped hook URL, but build-cache reuse is still allowed
- `doors.HostModeNoCache`: instance-scoped hook URL with no build-cache reuse

In practice:

- `Public` gives the resource a public path
- `Private` keeps the URL instance-scoped without forcing a rebuild each time
- `NoCache` is for disposable per-instance resources

For `SourceLocal` and `SourceExternal`, `HostMode` has nothing to host, so the given URL is used directly.

### Behavior

For buildable sources, **Doors** runs the content through the resource registry.

Important user-facing behavior:

- identical build inputs can reuse the same built content when cache mode allows it
- content type is set correctly for JS and CSS
- gzip is used when the client accepts it and the server allows it
- cache-control headers come from server config for cached resources

Public, private, and no-cache modes change how the browser reaches the resource, not what the resource is.

## Private URL

Use this when you do not want the shared resource registry at all and instead need a private Door-scoped URL on one element.

Good fits are:

- protected downloads
- per-user files
- generated images, scripts, or styles
- small custom handlers that should disappear with the current Door

These attrs create hook-backed URLs under the **Doors** system path.

That means:

- the URL belongs to the Door that created it
- it stops working when that Door is gone
- the instance lifetime is still the outer limit
- `Once: true` makes it single-use

### Src

Use `doors.ASrc` when an element needs a private `src` served from a file path.

```gox
<>
	~>doors.ASrc{
		Path: "./private/chart.js",
	} <script></script>
</>
```

Use `doors.ARawSrc` when the response should come from your own handler instead of a file.

```gox
<>
	~>doors.ARawSrc{
		Name: "chart.js",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/javascript")
			w.Write(chartBytes)
		},
	} <script></script>
</>
```

This is useful for `<script>`, `<img>`, `<iframe>`, or any other element where `src` is the right browser contract.

### Download

Use `doors.AFileHref` when an element should get a private `href` from a file path.

```gox
<>
	~>doors.AFileHref{
		Path: "./private/report.pdf",
		Name: "report.pdf",
	} <a target="_blank">Download report</a>
</>
```

Use `doors.ARawFileHref` when you want to write the response yourself.

```gox
<>
	~>doors.ARawFileHref{
		Name: "report.pdf",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			w.Write(reportBytes)
		},
	} <a target="_blank">Download report</a>
</>
```

These are a good match for downloads, private stylesheets, private icons, or any other case where the browser expects `href`.

### Handler

The file and raw variants follow one simple split:

- `ASrc` / `AFileHref`: serve an existing file from disk
- `ARawSrc` / `ARawFileHref`: write the response yourself

`Name` is optional, but useful when you want the generated URL to carry a readable filename or extension.

That can help browser behavior and make the URL easier to inspect.

If the content is stable and should be shared across users or instances, prefer the normal resource types above.

If it is private, short-lived, or request-specific, prefer these attr-based URLs.

For normal page links such as `doors.AHref`, use [Navigation](./09-navigation.md).
