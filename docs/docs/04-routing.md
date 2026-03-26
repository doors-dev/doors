# Routing

Routing in **Doors** has two layers:

1. `doors.UseModel` registers typed page routes based on a struct model.
2. `doors.UseRoute` registers custom HTTP routes such as static files or health checks.

Custom routes run first. If none of them match, Doors tries the registered model routes. If nothing matches, the router uses the fallback handler or returns `404`.

## Setup

Create a router with `doors.NewRouter()` and register routes on it:

```go
router := doors.NewRouter()

doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	return doors.ResponseComp(Page(s))
})

doors.UseRoute(router, doors.RouteFile{
	Path:     "/favicon.ico",
	FilePath: "./static/favicon.ico",
})
```

`doors.Router` implements `http.Handler`, so you can pass it directly to `http.ListenAndServe`.

## Models

`doors.UseModel` registers a typed route handler:

```go
func UseModel[M any](
	r Router,
	handler func(r RequestModel, s Source[M]) Response,
)
```

Doors builds a path adapter from `M`, decodes the incoming URL into that model, and passes it to your handler as `doors.Source[M]`.

Use the returned `doors.Response` to decide what should happen:

```go
doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	path := s.Get()

	if path.Legacy {
		return doors.ResponseRedirect(Path{Home: true}, http.StatusMovedPermanently)
	}

	if path.Dashboard && !authorized(r) {
		return doors.ResponseReroute(Path{Login: true})
	}

	return doors.ResponseComp(Page(s))
})
```

Available response helpers:

- `doors.ResponseComp(comp)` renders a page/component for the matched model.
- `doors.ResponseRedirect(model, status)` sends an HTTP redirect to the URL encoded from another model. If `status` is `0`, Doors uses `302 Found`.
- `doors.ResponseReroute(model)` internally reroutes to another registered model without sending an HTTP redirect.

### RequestModel

`doors.RequestModel` gives access to request and session state while choosing the response:

- `SessionStore()`
- `RequestHeader()`
- `ResponseHeader()`
- `SetCookie(...)`
- `GetCookie(...)`

This is the right place to handle auth checks, per-request headers, and similar gatekeeping before rendering the page.

## Paths

A path model is a struct with one or more exported `bool` fields tagged with `path:"..."`.

```go
type Path struct {
	Home bool `path:"/"`
}
```

That model matches the root path only.

### Variants

You can register multiple path variants in one model. The matched variant field becomes `true` when decoding. When encoding a URL from a model, Doors uses the variant whose marker field is `true`.

```go
type Path struct {
	Home bool `path:"/"`
	Docs bool `path:"/docs"`
}
```

Leading and trailing slashes are normalized, so `"/docs"`, `"docs"`, and `"/docs/"` describe the same route pattern.

### Params

Use `:FieldName` to capture a path segment into an exported field with the same name:

```go
type Path struct {
	Post bool `path:"/posts/:ID"`
	ID   int
}
```

Supported single-segment field types are:

- `string`
- `int`, `int64`
- `uint`, `uint64`
- `float64`

### Optional

Add `?` to make the last captured segment optional. Optional single-segment captures must use pointer fields.

```go
type Path struct {
	Catalog bool `path:"/catalog/:ID?"`
	ID      *int
}
```

This matches both `/catalog` and `/catalog/42`.

### Tail

Add `+` to the last parameter to capture the remaining path into `[]string`.

```go
type Path struct {
	Docs bool `path:"/docs/:Rest+"`
	Rest []string
}
```

This matches `/docs/guide/setup` and decodes `Rest` as `[]string{"guide", "setup"}`.

### Tail Optional

Use either `+?` or `*` to make the trailing multi-segment capture optional:

```go
type Path struct {
	Docs bool `path:"/docs/:Rest*"`
	Rest []string
}
```

This matches both `/docs` and `/docs/guide/setup`.

Rules to keep in mind:

- Optional captures must be the last segment.
- Multi-segment captures must be the last segment.
- `+`, `+?`, and `*` require a `[]string` field.
- Required single-segment captures must use non-pointer fields.

## Query

Use `query:"name"` tags for query-string values. Query fields do not affect route matching; they are decoded after a path variant matches.

```go
type Path struct {
	Catalog bool     `path:"/catalog"`
	Color   []string `query:"color"`
	Page    *int     `query:"page"`
}
```

Examples:

- `/catalog?color=black&color=yellow`
- `/catalog?page=2`

Encoding and decoding are explicit: only fields tagged with `query` are included. Pointer fields are useful when you want a parameter to be optional.

## URLs

`doors.NewLocation(ctx, model)` encodes a registered model into a `doors.Location`.

```go
loc, err := doors.NewLocation(ctx, Path{
	Catalog: true,
	Color:   []string{"black", "yellow"},
})
if err != nil {
	panic(err)
}

href := loc.String() // /catalog?color=black&color=yellow
```

`doors.Location` contains:

- `Segments []string`
- `Query url.Values`

And provides:

- `Path()` for the escaped path only
- `String()` for full path plus query string

`NewLocation` only works for model types that were registered with `doors.UseModel`.

## Custom

For non-page endpoints, implement `doors.Route`:

```go
type Route interface {
	Match(r *http.Request) bool
	Serve(w http.ResponseWriter, r *http.Request)
}
```

Then register it with `doors.UseRoute(router, route)`.

Example:

```go
type HealthRoute struct{}

func (HealthRoute) Match(r *http.Request) bool {
	return r.URL.Path == "/health"
}

func (HealthRoute) Serve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
```

## Static

Doors exposes a few ready-to-use `Route` implementations:

### RouteFS

Serves an `fs.FS` under a URL prefix.

```go
doors.UseRoute(router, doors.RouteFS{
	Prefix: "/assets",
	FS:     assetsFS,
})
```

### RouteDir

Serves a local directory under a URL prefix.

```go
doors.UseRoute(router, doors.RouteDir{
	Prefix:  "/public",
	DirPath: "./public",
})
```

### RouteFile

Serves a single local file at a fixed URL path.

```go
doors.UseRoute(router, doors.RouteFile{
	Path:     "/robots.txt",
	FilePath: "./static/robots.txt",
})
```

### RouteResource and RouteResourceFS

These register cache-aware static resources through the Doors resource registry. Use them when you want a fixed public path backed by a local file or `fs.FS`.

## Fallback

Use `doors.UseFallback(router, handler)` to delegate unmatched requests to another `http.Handler`.

```go
doors.UseFallback(router, myMux)
```
