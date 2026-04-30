# Routing

In **Doors**, the URL is reactive state. Routing is what you do with it.

For each page instance, the current URL is exposed as a single `doors.Source[doors.Location]`, and routing means branching content on its value. The most common way to do that is with a path model — a struct that decodes URLs into typed values and encodes them back.

## Quick Example

The everyday shape is:

- declare a path model
- use `doors.RouterSource(...)` to route the URL through it
- render a fallback for URLs that don't decode

```gox
type Path struct {
	Section Section `path:"/ | /docs/:ID? | /tutorial/:ID? | /license"`
	ID      *string
}

type Section int

const (
	SectionHome Section = iota
	SectionDocs
	SectionTutorial
	SectionLicense
)

elem (a App) Main() {
	<!doctype html>
	<html lang="en">
		<body>
			~(doors.RouterSource(
				doors.RouteModelSource(func(p doors.Source[Path]) gox.Comp {
					return Page(p)
				}),
				doors.RouteLocationDefaultComp(NotFound{}),
			))
		</body>
	</html>
}
```

`RouteModelSource` matches when the URL decodes into `Path`. The matched view receives `doors.Source[Path]`: a typed view of the current URL. Updating that source encodes the new `Path` back into the URL.

`RouteLocationDefaultComp` is the fallback for URLs that didn't decode. Routes are tried in order and the first match renders, so put fallbacks last.

> Use `doors.RouterBeam(...)` with `RouteModelBeam(...)` when the page only needs to read the route. Beam routes can still use `ALink` for navigation; they just cannot update the matched route value directly.

## Path Models

A path model is a struct that defines how URLs decode into a typed value. **Doors** uses it to:

- match incoming URLs
- decode path segments and query into struct fields
- encode the model back into a URL for navigation, redirects, and links

### Variants

A path model expresses its variants in one of two styles:

- **one `int` field** with patterns separated by `|`
- **one or more `bool` fields**, each with its own pattern

Both are first-class. The two styles can't be mixed in the same struct.

Common rules:

- Leading and trailing slashes are normalized — `"/docs"`, `"docs"`, and `"/docs/"` describe the same pattern.
- Spaces around `|` are optional and trimmed — `"/|/docs|/guide"` and `"/ | /docs | /guide"` are equivalent.

#### Int Multi-Variant

```go
type Path struct {
	Section Section `path:"/ | /docs | /guide"`
}

type Section int

const (
	SectionHome Section = iota
	SectionDocs
	SectionGuide
)
```

The `int` field stores the index of the matched pattern. Decoding sets it to the index of the pattern that matched. Encoding picks the pattern at that index.

A typed `int` paired with `iota` constants gives readable code and exhaustive switches. A plain `int` works too.

#### Bool Variants

```go
type Path struct {
	Home  bool `path:"/"`
	Docs  bool `path:"/docs"`
	Guide bool `path:"/guide"`
}
```

Each `bool` field is one variant. When **Doors** decodes a URL, the matched variant's field becomes `true`. When encoding from a model, **Doors** uses the variant whose marker field is `true` (if more than one is `true`, the first in field order wins).

#### When to Use Which

- **`int` multi-variant** is more compact for many variants and gives you exhaustive switches via typed constants. The ordering is implicit (struct tag order).
- **`bool` variants** read well when each variant has a clear name and you want to refer to it explicitly (`Path{Docs: true}` rather than `Path{Section: SectionDocs}`). Also natural for very small models — including a single-pattern model.

### Params

Use `:FieldName` to capture a path segment into the struct field with the same name.

```go
type Path struct {
	Post bool `path:"/posts/:ID"`
	ID   int
}
```

Supported single-segment field types: `string`, `int`, `int64`, `uint`, `uint64`, `float64`.

### Optional

Add `?` to make the last captured segment optional. Optional single-segment captures must use pointer fields:

```go
type Path struct {
	Catalog bool `path:"/catalog/:ID?"`
	ID      *int
}
```

Matches both `/catalog` (`ID == nil`) and `/catalog/42`.

### Tail

Use `+` on the last parameter to capture the remaining path into `[]string`:

```go
type Path struct {
	Docs bool `path:"/docs/:Rest+"`
	Rest []string
}
```

`/docs/guide/setup` decodes `Rest` as `[]string{"guide", "setup"}`.

Use `*` (or `+?`) to make that trailing capture optional — matches both `/docs` and `/docs/guide/setup`.

Rules:

- optional captures must be the last segment
- multi-segment captures must be the last segment
- `+` and `*` require a `[]string` field
- required single-segment captures must use non-pointer fields

### Query

Use `query:"name"` tags for query-string values.

```go
type Path struct {
	Catalog bool     `path:"/catalog"`
	Color   []string `query:"color"`
	Page    *int     `query:"page"`
}
```

Query fields don't decide which path variant matches — they're decoded after the path variant is fixed. Only fields tagged with `query` are encoded back into URLs.

The exact tag-based encoding rules come from [go-playground/form v4](https://github.com/go-playground/form/tree/v4.2.1) with the `query` tag in explicit mode.

#### Raw Query Values

When a page has many query values, keep `url.Values` directly in the model instead of tagging each one:

```go
import "net/url"

type Path struct {
	Search bool `path:"/search"`
	Query  url.Values
}
```

For `/search?q=doors&tag=go&tag=ui`, `Query.Get("q")` is `doors` and `Query["tag"]` is `[]string{"go", "ui"}`. Encoding uses the same `url.Values` field.

This works well for open-ended query parameters, preserving unknown keys, or pages that already parse query themselves.

Don't mix a `url.Values` field with `query` tags in the same model. Pick one style per model.

## Multiple Models

A page can route on more than one path model — list them in order of specificity:

```gox
<>
    ~(doors.RouterSource(
        doors.RouteModelSource(renderHome),
        doors.RouteModelSource(renderPost),
        doors.RouteModelSource(renderCatalog),
        doors.RouteLocationDefaultComp(NotFound{}),
    ))
</>
```

The first model that decodes successfully wins.

## URLs

> Usually you don't build URLs directly. See [Navigation](./09-navigation.md).

`doors.NewLocation(model)` produces a `doors.Location` from a model. It accepts a path-model struct, a `doors.Location` value, or any custom type implementing `doors.LocationEncoder` (covered in [Navigation](./09-navigation.md)).

```go
loc, err := doors.NewLocation(Path{
	Section: SectionDocs,
})
if err != nil {
	panic(err)
}

href := loc.String() // /docs
```

## Trust And Permissions

The location is client state. A user can craft any URL that pass your path model decoder.

Two consequences:

- **Always check permissions while rendering, against the current location or decoded model.** A successful match means "this URL parses", not "this user is allowed to see this".
- **For state the client must not control, use a separate `Source`.** Don't store auth, role, or other trust-bearing values on the route. Initialize them server-side and keep them on a session-scoped source. See [Storage & Auth](./18-storage-auth.md).

The convenience is that the URL can be read and written like normal state. The trust boundary does not change: reading it tells you what the client asked for, nothing more.

## Early Decisions

Two places commonly decide setup, redirects, or access before most UI is produced:

- **The page function passed to `doors.NewApp(...)`** runs once per page request and has access to `doors.Request` (cookies, headers, response writer). The usual job here is to bootstrap session-scoped state from cookies or headers, such as hydrating an auth source. See [Storage & Auth](./18-storage-auth.md). You can also read or change the location here.
- **Inside the matched component.** You have the **Doors** runtime context and the matched route value. Read or update it, derive smaller state, and decide what to render from session state.

To redirect or block a request before it reaches the page function, use HTTP middleware via `app.Use(...)`. The page function and component code both run after the request has already been accepted — they can change the location, but they aren't the place for HTTP-level redirects.

---

## Built on State Routing

URL routing is one application of a state primitive. `doors.Router(ctx)` returns the location source itself:

```go
loc := doors.Router(ctx) // doors.Source[doors.Location]
```

`doors.RouterSource(routes...)` and `doors.RouterBeam(routes...)` are shortcuts over the current location source:

```go
doors.Router(ctx).RouteSource(routes...)
doors.Router(ctx).RouteBeam(routes...)
```

`RouteSource` is the writable form: matched routes may receive a `Source` and write back to the routed value. `RouteBeam` is the read-only form: matched routes receive only a `Beam`.

### The Primitive

State routing swaps one of several views based on a reactive value. Routes are built from `RouteValue`, `RouteMatch`, and `RouteDerive`, then completed with `.Comp`, `.Beam(...)`, or `.Source(...)`, plus an optional `RouteDefault*` fallback.

The fragment swaps only when the active route changes. Within a matched route, the render function receives a live `Beam` or `Source` and reacts to value changes with normal state primitives. Full reference in [State](./07-state.md).

### Patterns

In a routing context, that opens up a few common patterns:

- **Match the location without a path model.** The general `RouteValue`, `RouteMatch`, and `RouteDerive` builders work directly on `Location`, and compose freely with `RouteModelSource` / `RouteModelBeam` in the same router. Reach for them when one ad-hoc URL doesn't deserve its own struct.
- **Take the location source directly.** `RouteLocationDefaultSource` (and `RouteLocationDefaultBeam`) gives the matched view raw `Source[Location]` / `Beam[Location]`, useful as a fallback or when a page parses URLs itself.
- **Dispatch on a derived value.** Once a page has a typed `Path`, derive a beam for the variant field and route on it again. The fragment only swaps when the variant actually changes.
- **Mix URL routing with non-URL state.** Branch on a session-scoped flag, a feature toggle, an auth state — using the same `Route*` builders.
- **Custom slices of the URL.** Derive a single-purpose source or beam from `doors.Router(ctx)` (for example, one query param) without committing to a path model.
- **Compose levels.** Route on `Location` to pick a section, then route on a derived value inside it. Each level only rerenders on its own changes.

For derivation patterns (`DeriveSource`, `DeriveBeam`), the full `Route*` builder reference, the `.Comp` / `.Beam` / `.Source` chain, write-back via `RouteDerive(...).Source(set, render)`, and how routing fits in with `Bind` and `Effect`, see [State](./07-state.md).
