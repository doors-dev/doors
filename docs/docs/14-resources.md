# Resources

Resources are how you ship JS/CSS to the page in Doors.

Use this page as a practical chooser: what to use for each scenario, and what behavior to expect.

## Choose

Pick by intent:

- page-local script near a Go component: `doors.ScriptInline`
- regular script file tag: `doors.ScriptCommon`
- ES module and import-map usage: `doors.ScriptModule`
- stylesheet link: `doors.Style`

## Inline

Use `ScriptInline` when the script belongs to one component/page and you want to keep TS/JS close to that Go code.

```gox
~doors.ScriptInline{
	Source: doors.SourceScriptString{
		Content: scrollScript,
		TypeScript: true,
	},
}
```

This is the recommended pattern for embedded `.ts` + `go:embed` workflows.

`ScriptInline` accepts buildable sources (`SourcePath`, `SourceFS`, script string/bytes). It does not accept `SourceLocal` or `SourceExternal`.

## Cooking

`ScriptInline` is intentionally "cooked" before it reaches the browser.

What happens:

1. Doors reads your inline source (path/fs/string/bytes).
2. Doors wraps that code as a runtime-entry function that receives script helpers (`$on`, `$data`, `$hook`, `$fetch`, `$G`, `$sys`, `HookErr`).
3. Doors builds it through the script resource pipeline (TypeScript supported when configured on source).
4. The page gets a `<script src="...">` resource URL, not raw inline script text.

Why this exists:

- top-level `await` works naturally in script code
- helpers are injected consistently into every Doors-managed script
- script lifecycle can be tied to the rendered subtree (`$sys.clean`)
- resource serving gives you cache/gzip/content-type behavior and consistent CSP flow
- generated content is deduplicated by build/cache keys when caching mode allows it

## Common

Use `ScriptCommon` when you want a normal `<script src="...">` style asset.

```gox
~doors.ScriptCommon{
	Source: doors.SourcePath("web/app.js"),
}
```

You can render it directly or attach it as a modifier to an existing `<script>` element.

## Module

Use `ScriptModule` for ES modules.

If `Specifier` is empty, it writes a module script tag.  
If `Specifier` is set, it registers the module in the page import map.

When used as an attribute modifier on `<script>` or `<link>`, you can have both behaviors at once:

- tag attributes are set (`type/src` for `<script>`, `rel/href` for `<link>`)
- `Specifier` is still registered in the import map

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

This is the standard way to expose built modules to user scripts by specifier.

## Style

Use `Style` for CSS assets.

```gox
~doors.Style{
	Source: doors.SourcePath("web/app.css"),
}
```

`Minify` controls CSS minification for buildable style sources.

## Sources

Script/style source options:

- file path: `doors.SourcePath("...")`
- embedded FS: `doors.SourceFS{FS, Path, Name}`
- in-memory script: `doors.SourceScriptString` / `doors.SourceScriptBytes`
- in-memory style: `doors.SourceStyleString` / `doors.SourceStyleBytes`
- already hosted URL: `doors.SourceLocal("/...")`
- external URL: `doors.SourceExternal("https://...")`

Use `SourceLocal` only when the asset is already hosted at that URL.  
Use `SourceExternal` for CDN/remote assets.

## Hosting

`HostMode` controls serving strategy for built assets:

- `doors.HostModePublic` (default): public resource path
- `doors.HostModePrivate`: instance-scoped URL with cached build reuse
- `doors.HostModeNoCache`: instance-scoped URL without build-cache reuse

For `SourceLocal` and `SourceExternal`, Doors uses the provided URL directly.

## Build

`ScriptOutput` controls JS build output for `ScriptCommon` and `ScriptModule`:

- `doors.ScriptOutputDefault`
- `doors.ScriptOutputBundle`
- `doors.ScriptOutputRaw`

Use `Raw` when the source is already built and should be served as-is.

`Profile` lets you select a named build profile for script builds.

Detailed esbuild configuration is covered in [15-esbuild.md](/Users/alex/Lib/doors/docs/docs/15-esbuild.md).

## Behavior

Important runtime behavior:

- Doors deduplicates built script/style content by cache keys (entry + format + profile + kind)
- content IDs are deterministic from final content bytes
- responses set correct content type (`application/javascript` / `text/css`)
- gzip is used when available and enabled
- external sources are added to CSP automatically (`script-src` / `style-src`)

## Patterns

Practical defaults:

- use `ScriptInline` for page-level interaction code
- use `ScriptModule` + `Specifier` for reusable modules imported by user scripts
- use `ScriptCommon` for classic script assets
- use `Style` for stylesheet assets
