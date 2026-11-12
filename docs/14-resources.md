# Resources

A resource is content your Go code owns, such as a file, bytes, a build, or a handler, that **Doors** serves under a generated URL. Build it with a chain of methods, then attach it to the tag as a modifier, put it in `src`, `href`, or a script `data:` attribute, or get its URL in code. On a tag, attributes are the shorter way to write the same chain.

This page covers how any resource is served. Script builds are covered in [JavaScript](./15-javascript.md), stylesheets in [Styles](./16-styles.md), and endpoints for browser code in [Custom Endpoints](./23-custom-endpoints.md).

## Hook and Hosted

Every resource URL is one of two kinds.

**Hook.** The URL belongs to the page. It is served by a hook on the closest Door around the render:

- only the browser session that rendered it can load it; without the session cookie it returns `410`
- it lives as long as the content that rendered it; once that content is cleared, or the page or session ends, it returns `404`
- every render gets a new URL

**Hosted.** The content, or its build, is kept in RAM once and served at a URL derived from its hash:

- anyone holding the URL can load it
- every page and user gets the same URL, and browsers cache it (`ServerCacheControl`, immutable by default)
- only content **Doors** can read up front can be hosted: files, bytes, strings
- it is not for large files or user-specific content

```gox
<script (doors.ResourceFS(web, "app.ts").Script())></script>
<link rel="stylesheet" (doors.ResourceFS(web, "app.css").Style())>

<img (doors.ResourceLocalFS("./private/chart.png"))>
<a (doors.ResourceBytes(reportPDF).Name("report.pdf"))>Download report</a>
```

The script and the stylesheet get public, cached URLs. The image and the report get hook URLs that only this page can load.

Tag defaults: a source with no stage takes its kind from the tag. `<script src>`, `<link rel="stylesheet">`, and `<link rel="modulepreload">` build it once and host it; any other `src` or `href` serves it as an asset through a hook, read on every request. Handlers, hooks, and proxies are always a hook.

## Two Syntaxes

### Builder

A chain starts with a source:

```go
doors.ResourceLocalFS("./private/chart.png") // file on disk
doors.ResourceFS(assets, "assets/chart.png") // embed.FS or any fs.FS
doors.ResourceBytes(reportPDF)               // bytes in memory
doors.ResourceString(htmlSnippet)            // string in memory
doors.ResourceHandler(handler)               // func(w, r), runs on every request
doors.ResourceHook(hook)                     // func(ctx, w, r) bool, gets the Door context, true stops serving
doors.ResourceProxy("https://example.com/embed")
```

The first four return `ResourceStatic`: content **Doors** reads itself, can host, and can build as a script or stylesheet. Handler, hook, and proxy return `ResourceDynamic`, which is always served through a hook and takes only `Name` and `Portable`.

A `ResourceStatic` starts as an asset, served as is. `Script()` or `Style()` switches it to a build, and `Raw()` after either serves the file without building it. Each stage has methods for what the matching attribute sets:

```go
doors.ResourceFS(assets, "sans.woff2").ContentType("font/woff2").Cache() // hosted asset
doors.ResourceBytes(reportPDF).Name("report.pdf").Portable()             // portable hook
doors.ResourceFS(web, "app.ts").Script().Module().Bundle()               // hosted ES module
doors.ResourceFS(web, "admin.css").Style().Private()                     // built once, hook URL
doors.ResourceFS(web, "vendor.js").Script().Raw()                        // hosted as is
doors.ResourceHandler(api.ServeHTTP).Name("api")                         // hook
```

Any chain value, at any stage, goes in four places:

```gox
<img (doors.ResourceFS(assets, "logo.svg").ContentType("image/svg+xml").Cache())>
<pdf-viewer src=(doors.ResourceBytes(manualPDF).Name("manual.pdf"))></pdf-viewer>
<script data:logo=(doors.ResourceFS(assets, "logo.svg").ContentType("image/svg+xml").Cache())>
	document.getElementById("logo").src = $data("logo")
</script>
```

- modifier, the default form: `<tag (chain)>` sets the attribute from the tag name, `href` on `a`, `area`, `base`, and `link`, `src` on `audio`, `embed`, `iframe`, `img`, `input`, `script`, `source`, `track`, and `video`; a module stage on `<script>` also adds `type="module"`
- `src=(chain)` or `href=(chain)`: works on any tag; required on tags outside the modifier mapping
- `data:name=(chain)` on `<script>`, inline or with `src`: `$data("name")` returns the URL string, for `fetch`, `import()`, or a library; a read or build error fails the render, like a broken `src`
- `URL(ctx)`: the same URL as a string in Go code, for any other attribute such as `<video poster>` or a `data-*` attribute read by a library; see [URL in Code](#url-in-code)

Only `src`, `href`, and script `data:` take a resource; a resource on any other attribute is a render error.

### Attributes

On a tag, attributes set the same things as the chain methods:

```gox
<img (doors.ResourceFS(assets, "logo.svg")) type="image/svg+xml" cache>
```

| Attribute | Method |
|---|---|
| `name="report.pdf"` | `.Name("report.pdf")` |
| `type="image/svg+xml"` (generic tags) | `.ContentType("image/svg+xml")` |
| `cache` | `.Cache()` |
| `private` | `.Script().Private()`, `.Style().Private()` |
| `nocache` | `.Script().NoCache()`, `.Style().NoCache()`; default on assets |
| `portable` | `.Portable()` |
| `type="module"`, `type="typescript"` | `.Script().Module()`, `.Script().TypeScript()` |
| `bundle`, `raw`, `profile="react"` | `.Bundle()`, `.Raw()`, `.Profile("react")` |

The difference: with attributes the tag decides the kind, so the same `ResourceFS(web, "app.ts")` is a script on `<script>` and an asset on `<a>`; a chain states it, so `.Script()` on `<a>` is an error. Attributes and a chain can be mixed on one tag; an attribute that contradicts the chain is an error, see [Render Errors](#render-errors).

On `src=` and `href=` values, a `[]byte` works like `ResourceBytes`, a `func(w, r)` like `ResourceHandler`, and a `func(ctx, w, r) bool` like `ResourceHook`. These shorthands are not modifiers: write `src=`/`href=`, or wrap them in the constructor.

## Combinations

Each case shows the chain first and the attribute form second; both serve the same URL.

### Read on Every Request

The default for files, bytes, and strings on generic tags: an asset without `Cache()` is a hook, read on each request. A local file is read from disk every time, so edits show up without a restart:

```gox
<img (doors.ResourceLocalFS("./data/today.png"))>
```

There is nothing to set; on `<script>` and stylesheet links use [Rebuilt Every Render](#rebuilt-every-render).

### Cached and Public

`Cache()` hosts an asset. Use it for stable, public files such as logos, icons, and fonts:

```gox
<img (doors.ResourceFS(assets, "logo.svg").ContentType("image/svg+xml").Cache())>
<img (doors.ResourceFS(assets, "logo.svg")) type="image/svg+xml" cache>
```

- set the content type whenever you cache; without it the type is sniffed from the content, and JavaScript and CSS come out as `text/plain`
- a name is appended after the hash; for clean download names, keep the default hook
- nothing follows `Cache()`, so set `Name` and `ContentType` before it

### Built Once, Served by Hook

`Private()` keeps a script or stylesheet build in the shared cache but hands out a hook URL. Use it for a bundle that should not be public:

```gox
<script (doors.ResourceFS(web, "admin.ts").Script().Bundle().Private())></script>
<script (doors.ResourceFS(web, "admin.ts")) bundle private></script>
```

On generic tags the default is already a hook, so `private` changes nothing there.

### Rebuilt Every Render

`NoCache()` builds a script or stylesheet on every render and serves it through a hook. Use it for output generated per render:

```gox
<link rel="stylesheet" (doors.ResourceString(themeCSS(user)).Style().NoCache())>
<link rel="stylesheet" (doors.ResourceString(themeCSS(user))) nocache>
```

On generic tags `nocache` is the default.

### Portable

A hook URL needs the session cookie, so it fails when something other than the page fetches it: an Android WebView hands downloads to the system download manager, which sends no cookies; apps and in-app browsers open links outside the page's cookie jar. `Portable()` gives the hook URL access of its own:

```gox
<a (doors.ResourceBytes(reportPDF).Name("report.pdf").Portable()) download>Download report</a>
<a (doors.ResourceBytes(reportPDF)) name="report.pdf" portable download>Download report</a>
```

- it is a bearer link: anyone holding it can load the resource
- it still stops when the content that rendered it is cleared, the page instance or session ends, or a one-shot hook returns `true`
- a request through it needs no cookie and creates no session; the handler still sees the page's session, the stripped `r.URL.Path`, and `Doors-Hook-Prefix` as with any hook
- it works on any hook: assets, handlers, and `Private()` or `NoCache()` scripts and styles; a script or style with neither acts as `Private()`
- `Portable()` with `Cache()` is an error

### Not Served by Doors

Write a URL directly when it already exists:

```gox
<img src="/assets/logo.png">
<a href="https://example.com/report.pdf">Open report</a>
<script (doors.ResourceExternal("https://cdn.example.com/lib.js"))></script>
```

A plain string is passed through as is. `doors.ResourceExternal` is also loaded directly by the browser, and **Doors** adds its URL to the page CSP for scripts and for stylesheet and modulepreload links rendered with the page, see [CSP](./21-configuration.md#csp). Direct URLs take none of `cache`, `private`, `nocache`, or `portable`.

## Reference

### Stages

The methods on each stage and what they return:

- source (`ResourceStatic`): `Name`, `ContentType` → `AssetResource`; `Cache` → `CacheableResource`; `Portable` → `HookResource`; `Script` → `ScriptResource`; `Style` → `StyleResource`
- asset (`AssetResource`): `Name`, `ContentType` → `AssetResource`; `Cache` → `CacheableResource`; `Portable` → `HookResource`; no `Script` or `Style`, so call those first
- `ScriptResource`: `Name`, `Bundle`, `Module`, `TypeScript`, `Profile` → `ScriptResource`; `Raw` → `RawResource`; `NoCache`, `Private`, `Portable` → `HookResource`
- `StyleResource`: `Name` → `StyleResource`; `Raw` → `RawResource`; `NoCache`, `Private`, `Portable` → `HookResource`
- `RawResource`: `Name` → `RawResource`; `NoCache`, `Private`, `Portable` → `HookResource`; `Raw()` drops `Bundle`, `TypeScript`, and `Profile` set before it, and keeps `Module`
- `ResourceDynamic`: `Name`, `Portable` → `HookResource`
- `HookResource`: `Name`, `Portable` → `HookResource`
- `CacheableResource`: final

A chain ends hosted, as a `CacheableResource` (`Cache()`, or a script, style, or raw stage left as is), or as a hook: an asset left as is (`AssetResource`), a handler source (`ResourceDynamic`), or a `HookResource`.

For signatures: `Resource` covers every `Resource*` constructor except `ResourceExternal`, `ServedResource` covers every stage, and `CacheableResource` is what `UseResource` takes.

### Attribute Details

- `name`: the file name at the end of the URL, for browser filename hints and extension-based handling; hosted URLs still contain the hash
- `type`: the response Content-Type on generic tags; an error on handlers, hooks, and proxies, which set their own; on `<script>` it selects the script kind, see [Types](./15-javascript.md#types)
- `cache`: host it; files, bytes, and strings only
- only one of `cache`, `private`, and `nocache` may be set
- handlers, hooks, and proxies take `name` and `portable`; `private` is accepted and changes nothing; `cache` and `nocache` are errors
- script and stylesheet attributes such as `inline`, `bundle`, `raw`, `profile`, and `specifier` are covered in [JavaScript](./15-javascript.md#attrs) and [Styles](./16-styles.md#attrs)

Resources work on any tag with `src` or `href`; the common ones are `a`, `img`, `iframe`, `link`, `script`, `video`, `audio`, and `source`.

### Render Errors

- a script stage on a tag that is not `<script>` or `<link rel="modulepreload">`
- a style stage on a tag that is not `<link rel="stylesheet">`
- a non-module script stage on `modulepreload`, unless the tag has `type="module"`
- `ContentType` on `<script>` or a stylesheet link
- an attribute that contradicts the chain, such as `cache` with `.NoCache()`, or `raw` with `.Bundle()`
- `portable` with `cache`
- a resource on an attribute other than `src`, `href`, or script `data:`, such as `<video poster=(chain)>`; the render panics with `resource in "poster": only src, href, and script data: attributes take a resource` and the error is logged
- a script `data:` resource whose source cannot be read or built; the render fails and the error is logged

A plain source and the asset stages are not errors on a script or stylesheet tag: the tag decides the kind, so `Name`, `Cache`, or `Portable` on `<script>` or a stylesheet link still builds a script or stylesheet.

### Handlers, Hooks, and Proxies

- always a hook; the URL stops with the content that rendered it
- the handler sees the part after the hook URL as `r.URL.Path`, including the name, and `Doors-Hook-Prefix` holds the hook URL without the name
- requests are handled concurrently
- a `ResourceHook` handler gets the Door context and stops serving once it returns `true`
- `ResourceProxy` fails with status `500` if the URL cannot be parsed

> Request body is not bounded by `ServerRequestBodyLimit` for handlers, hooks, and proxies. Apply a safe limit yourself if you read the body.

### URL in Code

`URL(ctx)` returns the URL the same chain produces on a tag or in a script `data:` attribute. Use it when Go code needs the string itself: an attribute other than `src`, `href`, or script `data:`, such as `<video poster>`, a `data-*` attribute read by a library, or a value stored outside the render.

- in code, a plain source is an asset: use `Script()` or `Style()` for the URL a script or stylesheet tag would produce
- a cached asset needs `ContentType` before `Cache()`, for example `doors.ResourceBytes(data).ContentType("application/json").Cache().URL(ctx)`
- `URL` returns an error when the source cannot be read or built, and `context.Canceled` for a hook whose content is already cleared
- `ctx` must come from a **Doors** render or handler; otherwise `URL` panics
- never return `err` as an attribute value, because it renders as the attribute text

## Recipes

### Module URL in Code

To import a module without the import map, put the chain in a `data:` attribute and call `import()` on it:

```gox
<div id="chart"></div>
<script data:chart=(doors.ResourceFS(web, "chart.ts").Script().Module().Bundle())>
	const { draw } = await import($data("chart"))
	draw(document.getElementById("chart"))
</script>
```

- `import()` needs `Module().Bundle()`: `Bundle()` alone outputs CommonJS, and `Module()` alone keeps relative imports the browser cannot resolve
- no import map is involved, so it also works in content first rendered by a later update; for a module shared by name see [Modules](./15-javascript.md#modules)
- a build error fails the render and is logged, like a broken `src`

### App-Wide Module URL

In markup, a hosted chain already reuses its build: every `data:chart=(...)` above resolves to the same URL. When Go code needs the string itself, for example to store it or pass it outside the render, build it once with a helper backed by a package-level store:

```go
var urls = doors.NewStore()

type chartKey struct{}

func ChartModule(ctx context.Context) string {
	return urls.Init(chartKey{}, func() any {
		url, err := doors.ResourceFS(web, "chart.ts").Script().Module().Bundle().URL(ctx)
		if err != nil {
			panic(err)
		}
		return url
	}).(string)
}
```

- a panic fails the render that called the helper; nothing is stored, and the next call tries again
- use it only for hosted resources: a hook URL stops with the content that created it
- the store is per process

### Served Once

A `ResourceHook` handler that returns `true` stops serving after that request. Use it for one-time downloads and other short-lived links:

```gox
<a (doors.ResourceHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Content-Type", "application/pdf")
	w.Write(reportPDF)
	return true
}).Name("report.pdf"))>
	Download once
</a>
```

### Hosted at a Fixed Path

`doors.UseResource` serves a hosted resource at a path you choose, for a font referenced from CSS or a script embedded on other sites:

```go
app.Use(doors.UseResource("/embed/widget.v3.js", doors.ResourceFS(web, "widget.ts").Script().Bundle()))
```

Put a version in the path: responses carry `ServerCacheControl`, which is `immutable` by default. See [UseResource](./04-app.md#useresource).

### Custom Endpoints

`<script data:api=(doors.ResourceHandler(mux.ServeHTTP))>` gives browser code its own HTTP endpoint for an API, server-sent events, or a WebSocket, see [Custom Endpoints](./23-custom-endpoints.md).

## Rules

- write a plain URL when it already exists; use a resource when your Go code owns the content
- host (`Cache()`, default scripts and styles) only stable, public, small content
- set the content type when you cache an asset
- use `Private()` or `NoCache()` for scripts and styles that must not be public
- use `Portable()` for links fetched outside the page, such as Android WebView downloads
- use `Name` when the URL should look like a real file
- state the kind in the chain; attributes are the shorter way to write it on a tag
- use `URL(ctx)` only when the URL leaves `src`, `href`, or script `data:`; in code, `Script()` or `Style()` for script and stylesheet URLs, and `Module().Bundle()` for `import()`
- version `UseResource` paths
