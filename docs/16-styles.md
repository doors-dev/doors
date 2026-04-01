# Styles

This page covers stylesheet resources in **Doors**.

If generic resource syntax is new, start with [Resources](./14-resources.md). This page focuses on what is specific to `<style>` and `<link rel="stylesheet">`.

## Start

Most pages start with one of these:

- CSS written directly in the template: plain `<style>...</style>`
- CSS kept in a file, bytes, or string: `<link rel="stylesheet" href=(...)>`
- CSS already hosted somewhere: plain `href="..."`

Examples:

```gox
<style>
	h1 {
		color: red;
	}
</style>
```

```gox
<link
	rel="stylesheet"
	href=(doors.ResourceLocalFS("web/app.css"))>
```

```gox
<link rel="stylesheet" href="/assets/app.css">
```

## Style Tags

Use plain `<style>...</style>` when the CSS belongs to one page or one component:

```gox
<style>
	h1 {
		color: red;
	}
</style>
```

By default, **Doors** does not keep that literal `<style>` tag. It collects the CSS, creates a stylesheet resource, and emits a stylesheet link in the final HTML.

Use `raw` when you want a literal browser `<style>` tag:

```gox
<style raw>
	h1 {
		color: red;
	}
</style>
```

`name`, `private`, and `nocache` also work here:

```gox
<style name="page.css" private>
	h1 {
		color: red;
	}
</style>
```

## Stylesheet Links

Use `<link rel="stylesheet">` when the CSS comes from a file, bytes, a string, a handler, or a proxy:

```gox
<link
	rel="stylesheet"
	href=(doors.ResourceLocalFS("web/app.css"))>
```

Buildable stylesheet sources are:

- `doors.ResourceLocalFS("web/app.css")`
- `doors.ResourceFS(webFS, "app.css")`
- `doors.ResourceBytes(appCSS)`
- `doors.ResourceString(appCSS)`

These go through the stylesheet pipeline and produce a stylesheet resource URL.

Shorthands work on `href=` too:

- `href=(appCSS)` is treated like `href=(doors.ResourceBytes(appCSS))`
- `href=(func(w http.ResponseWriter, r *http.Request) { ... })` is treated like `href=(doors.ResourceHandler(...))`

Modifier syntax is often convenient when the whole tag exists just to serve that stylesheet:

```gox
<link rel="stylesheet" (doors.ResourceBytes(appCSS))/>
```

Other `href` forms are:

- plain string such as `"/assets/app.css"` for an already-hosted URL
- `doors.ResourceExternal("https://cdn.example.com/app.css")` for a direct browser URL that should also participate in CSP source collection
- `doors.ResourceHandler(...)`, `doors.ResourceHook(...)`, or `doors.ResourceProxy(...)` for handler-backed and proxied stylesheet URLs

On stylesheet links, output behavior is:

- omitted: buildable sources go through the stylesheet pipeline
- `raw`: **Doors** leaves the original tag alone

Managed stylesheet output is minified by default. `raw` is mainly useful when `href` is already something the browser can use directly, or when an embedded `<style>` must stay literal.

## Attrs

These attrs control managed stylesheet behavior:

- `raw`: keep the stylesheet tag or link raw
- `name`: readable output file name
- `private`: serve the stylesheet through an instance-scoped hook URL while still using the stylesheet pipeline
- `nocache`: serve through an instance-scoped hook URL without shared resource caching

Example:

```gox
<link
	rel="stylesheet"
	href=(doors.ResourceBytes(appCSS))
	name="app.css"
	private>
```

Plain string URLs are passed through as-is. `doors.ResourceExternal(...)` keeps the browser URL direct while also adding that host to CSP. Handler and proxy sources already produce hook-backed URLs.

Use `private` when the stylesheet should not be publicly reachable.

Use `nocache` for dynamically generated styles that should not use shared resource caching.
