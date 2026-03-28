# Path Model

In **Doors**, page routing starts from a struct.

That struct is your path model. **Doors** uses it to:

- match incoming page URLs
- decode path and query values
- give your page a typed route value
- build URLs again for links, redirects, and navigation

This is what makes routing in **Doors** feel like part of your app state instead of a separate string-based system.

## UseModel

Register a page model with `doors.UseModel`:

```go
type Path struct {
	Home bool `path:"/"`
	Post bool `path:"/posts/:ID"`
	ID   int
}

doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	return doors.ResponseComp(Page(s))
})
```

When a request matches this model, **Doors** decodes the URL into `Path` and passes it to your handler as `doors.Source[Path]`.

That source becomes the route state for the current page instance. If the user navigates within the same model type, the page can react to the updated model instead of doing a full reload.

Inside the `UseModel` handler, read the current model with `s.Get()`:

```go
doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	path := s.Get()
	if path.Legacy {
		return doors.ResponseRedirect(Path{Home: true}, http.StatusMovedPermanently)
	}
	return doors.ResponseComp(Page(s))
})
```

Use `Get()` here because this handler does not run with a **Doors** render/runtime `ctx`. The `Read(ctx)`-style methods are for places that do have that runtime context.

## Location

`doors.Location` is the special catch-all model.

If you register `doors.Source[doors.Location]`, that handler matches every URL:

```go
doors.UseModel(router, func(r doors.RequestModel, s doors.Source[doors.Location]) doors.Response {
	return doors.ResponseComp(Page(s))
})
```

This is useful when you want the raw path and query instead of a decoded struct model.

For example, a request like `/any/deep/path?tag=hello&page=7` arrives as `doors.Location` with:

- `Path()` as `/any/deep/path`
- `Query.Get("tag")` as `hello`
- `Query.Get("page")` as `7`

## Variants

A path model can contain multiple page variants. Each variant is an exported `bool` field tagged with `path:"..."`.

```go
type Path struct {
	Home  bool `path:"/"`
	Docs  bool `path:"/docs"`
	Guide bool `path:"/guide"`
}
```

When **Doors** decodes a URL, the matched variant field becomes `true`.

When **Doors** encodes a URL from a model value, it uses the variant whose marker field is `true`.

Leading and trailing slashes are normalized, so `"/docs"`, `"docs"`, and `"/docs/"` describe the same route pattern.

## Params

Use `:FieldName` to capture a path segment into a struct field with the same name.

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

## Optional

Add `?` to make the last captured segment optional.

Optional single-segment captures must use pointer fields:

```go
type Path struct {
	Catalog bool `path:"/catalog/:ID?"`
	ID      *int
}
```

This matches both `/catalog` and `/catalog/42`.

## Tail

Use `+` on the last parameter to capture the remaining path into `[]string`.

```go
type Path struct {
	Docs bool `path:"/docs/:Rest+"`
	Rest []string
}
```

This matches `/docs/guide/setup` and decodes `Rest` as `[]string{"guide", "setup"}`.

Use `*` or `+?` to make that trailing capture optional:

```go
type Path struct {
	Docs bool `path:"/docs/:Rest*"`
	Rest []string
}
```

This matches both `/docs` and `/docs/guide/setup`.

Rules to keep in mind:

- optional captures must be the last segment
- multi-segment captures must be the last segment
- `+` and `*` require a `[]string` field
- required single-segment captures must use non-pointer fields

## Query

Use `query:"name"` tags for query-string values.

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

Query fields do not decide which path variant matches. They are decoded after a path variant has matched.

Only fields tagged with `query` are encoded back into generated URLs.

For the exact query encoding and decoding rules, **Doors** uses [go-playground/form v4](https://github.com/go-playground/form/tree/v4.2.1) with the `query` tag in explicit mode.

## Response

Your `doors.UseModel` handler returns a `doors.Response`, which lets you decide what should happen for the matched model.

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

The common response helpers are:

- `doors.ResponseComp(comp)` to render a page
- `doors.ResponseRedirect(model, status)` to send an HTTP redirect
- `doors.ResponseReroute(model)` to internally hand off to another registered model without an HTTP redirect

## Request

`doors.RequestModel` gives you request/session access while deciding the response:

- `SessionStore()`
- `RequestHeader()`
- `ResponseHeader()`
- `SetCookie(...)`
- `GetCookie(...)`

This is the right place for auth checks, redirects, response headers, and similar gatekeeping before the page is rendered.

It is also the usual place to initialize shared session state from cookies or headers. For that pattern, see [Storage & Auth](./18-storage-auth.md).

## URLs

Use `doors.NewLocation(ctx, model)` when you need a URL from a registered model value.

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

`doors.Location` gives you:

- `Path()` for the path only
- `String()` for path plus query string

`NewLocation` works for model types that were registered with `doors.UseModel`.
