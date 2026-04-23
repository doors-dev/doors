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
	return doors.ResponseComp(App{path: s})
})
```

When a request matches this model, **Doors** decodes the URL into `Path` and passes it to your handler as `doors.Source[Path]`.

That source becomes the route state for the current page instance. If the user navigates within the same model type, the page can react to the updated model instead of reloading.

Inside the component body, bind the source and render whichever view matches the current route:

```go
type App struct {
	path doors.Source[Path]
}

elem (a App) Main() {
	<!doctype html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<title>Hello Doors!</title>
		</head>
		<body>
			~(a.path.Bind(func(p Path) gox.Elem {
				if p.Post {
					return Post(p.ID)
				}
				return Home()
			}))
		</body>
	</html>
}
```

This keeps the page reactive: when the path model changes, the subscribed part of the page updates automatically.

Usually you do not use a `Source` directly for subscriptions. Instead, derive a smaller piece of state from it. For more advanced patterns, see [State](./07-state.md) and [Navigation](./09-navigation.md).


### Access the Path in the Handler

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

## Catch-All Location

`doors.Location` is the special catch-all path model.

If you register `doors.Source[doors.Location]`, that handler matches every URL:

```go
doors.UseModel(router, func(r doors.RequestModel, s doors.Source[doors.Location]) doors.Response {
	return doors.ResponseComp(Page(s))
})
```

For example, a request like `/any/deep/path?tag=hello&page=7` arrives as `doors.Location` with:

- `Segments` as `[]string{"any", "deep", "path"}`
- `Path()` as `/any/deep/path`
- `Query.Get("tag")` as `hello`
- `Query.Get("page")` as `7`

Use `doors.Location` when you want the raw path and query instead of a decoded struct model.

This is useful for bigger apps where one central route parser is easier to maintain than a very large path model struct. Keep `doors.Source[doors.Location]` as the route source, then derive page-specific values from it the same way you would derive smaller pieces of normal state.


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

With tag-based query fields, only fields tagged with `query` are encoded back into generated URLs.

For the exact tag-based query encoding and decoding rules, **Doors** uses [go-playground/form v4](https://github.com/go-playground/form/tree/v4.2.1) with the `query` tag in explicit mode.

### Raw Query Values

When a page has many query values, it can be more convenient to keep `url.Values` directly in the path model instead of tagging each query field.

```go
import "net/url"

type Path struct {
	Search bool `path:"/search"`
	Query  url.Values
}
```

**Doors** stores the whole query string in the exported `url.Values` field.

For `/search?q=doors&tag=go&tag=ui`, `Query.Get("q")` is `doors` and `Query["tag"]` is `[]string{"go", "ui"}`.

When **Doors** builds a URL from the model, it uses that same `url.Values` field as the generated query string.

This also works well when the page accepts open-ended query parameters, needs to preserve unknown parameters, or already has its own query parsing layer.

Do not mix a `url.Values` field with `query` tags in the same path model. Use one query style per model.

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

> Usually you don't use direct URLs. Please refer to [Navigation](./09-navigation.md).

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
