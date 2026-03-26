# Resources

Resources are how **Doors** serves JS and CSS to the browser.

Use this layer when the asset belongs to your app and should be built, hosted, or referenced through **Doors** instead of being hand-written as a raw URL.

## Choose

Pick by intent:

- page-local managed script near Go code: `doors.ScriptInline`
- regular `<script src="...">` asset: `doors.ScriptCommon`
- ES module or import-map entry: `doors.ScriptModule`
- stylesheet link: `doors.Style`

## Inline

Use `ScriptInline` when the script belongs to one page or component.

This is the normal choice for page-local `.ts` or `.js` code kept next to the Go code that renders it.

```gox
~doors.ScriptInline{
	Source: doors.SourceScriptString{
		Content: scrollScript,
		TypeScript: true,
	},
}
```

`ScriptInline` accepts only buildable sources:

- `SourcePath`
- `SourceFS`
- `SourceScriptString`
- `SourceScriptBytes`

It does not accept `SourceLocal` or `SourceExternal`.

`ScriptInline` always ends up as a `src`-backed script resource.

The browser does not receive the original inline body directly.

That is what lets **Doors** inject the managed-script runtime wrapper used by [14-javascript.md](/Users/alex/Lib/doors/docs/docs/14-javascript.md).

## Common

Use `ScriptCommon` when you want a regular script asset without module behavior.

```gox
~doors.ScriptCommon{
	Source: doors.SourcePath("web/app.js"),
}
```

It can:

- render its own `<script src="...">` tag
- or be attached as a modifier to an existing `<script>` element

Use this when the browser should just load a script file and you do not need import-map registration.

## Module

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

This is the pattern used in the `imports` tests for module loading and React/Preact mounting.

When used as a modifier:

- on `<script>`, it sets `type="module"` and `src`
- on `<link>`, it sets `rel="modulepreload"` and `href`

That makes it the right tool for import-map entries, modulepreload tags, and framework islands.

## Style

Use `Style` for CSS assets.

```gox
~doors.Style{
	Source: doors.SourcePath("web/app.css"),
}
```

It renders a stylesheet link and serves the CSS through the same resource pipeline.

`Minify` matters only for buildable style sources.

## Source

Choose the source type by where the content already lives:

- file on disk: `doors.SourcePath("...")`
- embedded filesystem: `doors.SourceFS{FS, Path, Name}`
- script bytes/string already in Go: `doors.SourceScriptBytes` / `doors.SourceScriptString`
- style bytes/string already in Go: `doors.SourceStyleBytes` / `doors.SourceStyleString`
- already hosted local URL: `doors.SourceLocal("/...")`
- external URL: `doors.SourceExternal("https://...")`

Use `SourceLocal` or `SourceExternal` only when the asset is already hosted and should not be built by **Doors**.

For those direct-URL sources:

- build output settings do not apply
- bundling does not apply
- **Doors** just uses the URL you gave it

For `SourceExternal`, **Doors** also adds the source to CSP automatically.

## Host

`HostMode` controls how built resources are exposed:

- `doors.HostModePublic`: public resource URL
- `doors.HostModePrivate`: instance-scoped hook URL, but build/cache reuse is still allowed
- `doors.HostModeNoCache`: instance-scoped hook URL with no build-cache reuse

In other words:

- `Public` changes visibility by giving the resource a public path
- `Private` keeps the URL instance-scoped without forcing a rebuild each time
- `NoCache` is for disposable per-instance resources

For `SourceLocal` and `SourceExternal`, `HostMode` has nothing to host, so the given URL is used directly.

## Build

`ScriptOutput` matters for `ScriptCommon` and `ScriptModule`:

- `doors.ScriptOutputDefault`: normal build pipeline
- `doors.ScriptOutputBundle`: bundle dependencies into the output
- `doors.ScriptOutputRaw`: skip the build step and serve the source as-is

Use `Bundle` when the module should carry its dependency graph with it.

Use `Raw` when the source is already built and **Doors** should just serve it.

`Profile` selects a named esbuild build profile.

That is especially useful for module-heavy code such as TSX bundles.

Detailed build configuration is covered in [16-configuration.md](/Users/alex/Lib/doors/docs/docs/16-configuration.md).

## Behavior

For buildable sources, **Doors** runs the content through the resource registry.

Important user-facing behavior:

- identical build inputs can reuse the same built content when cache mode allows it
- content type is set correctly for JS and CSS
- gzip is used when the client accepts it and the server allows it
- cache-control headers come from server config for cached resources

Public, private, and no-cache modes change how the browser reaches the resource, not what the resource is.

## Rules

- Use `ScriptInline` for page-owned interaction code.
- Use `ScriptModule` for import-map modules, TSX bundles, and framework islands.
- Use `ScriptCommon` for regular script assets.
- Use `Style` for CSS.
- Use `SourceLocal` or `SourceExternal` only when the asset is already hosted.
- Use `Bundle` when dependencies should ship together and `Raw` when the file is already built.
- See [14-javascript.md](/Users/alex/Lib/doors/docs/docs/14-javascript.md) for managed script behavior and helper APIs.
