# Routing

In **Doors**, the URL is reactive state. Routing is what you do with it.

For each page instance, the current URL is exposed as a single `doors.Source[doors.Location]`, and routing means branching content on its value. The most common way to do that is with a path model — a struct that decodes URLs into typed values and encodes them back.

## Quick Example

The everyday shape is:

- declare a path model
- use `doors.Route(...)` to route the URL through it
- render a fallback for URLs that don't decode

```gox
type Path struct {
	// Matches /, /docs, and /docs/:ID.
	Section Section `/:" | docs/:ID?"`
	ID      *string
}

type Section int

const (
	SectionHome Section = iota
	SectionDocs
)

elem (a App) Main() {
	<!doctype html>
	<html lang="en">
		<body>
			~(doors.Route(
				// Decode the browser URL into Path.
				doors.RouteModel(Page),
				// Render when the URL does not decode into any path model.
				doors.RouteLocationDefaultComp(NotFound{}),
			))
		</body>
	</html>
}
```

`RouteModel` matches when the URL decodes into `Path`. The matched view receives `doors.Source[Path]`: a typed view of the current URL. Updating that source encodes the new `Path` back into the URL.

`RouteLocationDefaultComp` is the fallback for URLs that didn't decode. Routes are tried in order and the first match renders, so put fallbacks last.

Inside the matched app page, route on decoded fields instead of parsing strings:

```gox
elem Page(path doors.Source[Path]) {
	~(path.Route(
		// /: home page.
		doors.RouteMatch(func(p Path) bool {
			return p.Section == SectionHome
		}).Comp(Home{}),
		// /docs and /docs/:ID.
		doors.RouteMatch(func(p Path) bool {
			return p.Section == SectionDocs
		}).Bind(DocsPage),
		// Decoded but unsupported Section values.
		doors.RouteDefaultComp[Path](NotFound{}),
	))
}

elem DocsPage(p Path) {
	~(if p.ID == nil {
		~(DocsIndex{})
	} else {
		~(DocPage{ID: *p.ID})
	})
}
```

Branches that need URL params, such as `DocsPage`, receive the decoded `Path` value so they can render from `ID` directly.

## Path Models

A path model is a struct that defines how URLs decode into a typed value. **Doors** uses it to:

- match incoming URLs
- decode path segments and query into struct fields
- encode the model back into a URL for navigation, redirects, and links

### Variants

A path model uses one `int` field for its page variants. Use a named `int` type with `iota` constants for readable switches:

```go
type Path struct {
	Section Section `/:" | docs | guide"`
}

type Section int

const (
	SectionHome Section = iota
	SectionDocs
	SectionGuide
)
```

The `int` field stores the index of the matched variant. Decoding sets it to the index of the variant that matched. Encoding picks the variant at that index.

The tag key is the path prefix. The tag value is a `|`-separated list of variants under that prefix.

Use `/` for small models and single pages:

```go
type Path struct {
	Section Section `/:" | catalog | search"`
}
```

Use a longer prefix key to describe an area with several local variants or shared parameters:

```go
type Path struct {
	Section Section `/posts:" | new | archived"`
}
```

This matches `/posts`, `/posts/new`, and `/posts/archived`.

In a real app with dozens of pages, path models stay ergonomic when they are small. Put as much shared path as the area needs in the key, then keep the variants relative to it. For example, `"/post/:ID":"view | edit"` reads as "inside a post, choose view or edit". A larger app might use a deeper key such as `"/org/:OrgID/project/:ProjectID":"overview | settings"`.

Prefer `/:"catalog"` over `/catalog:""` for a single route. A segment in the key is for declaring an area, not for spelling every path segment.

Use `/` as the root prefix:

```go
type Path struct {
	Section Section `/:" | docs | guide"`
}
```

Quote the tag key when the prefix contains `:` or another character that cannot appear in an unquoted Go struct-tag key:

```go
type Path struct {
	Section Section `"/post/:ID":"view | edit"`
	ID      int
}
```

This matches `/post/42/view` and `/post/42/edit`.

Common rules:

- Write prefixes with a leading slash. Leading and trailing slashes are normalized — `/docs`, `docs`, and `/docs/` describe the same prefix or variant.
- Spaces around `|` are optional and trimmed — `" | docs | guide"` and `"|docs|guide"` are equivalent.
- An empty variant means the prefix itself. For example, `/posts:" | new"` matches `/posts` and `/posts/new`.

### Params

Use `:FieldName` to capture a path segment into the struct field with the same name. Captures can be in the prefix or in a variant.

```go
type Path struct {
	Section Section `/:"posts/:ID"`
	ID      int
}
```

Supported single-segment field types: `string`, `int`, `int64`, `uint`, `uint64`, `float64`.

### Optional

Add `?` to make the last captured segment optional. Optional single-segment captures must use pointer fields:

```go
type Path struct {
	Section Section `/:"catalog/:ID?"`
	ID      *int
}
```

Matches both `/catalog` (`ID == nil`) and `/catalog/42`.

### Tail

Use `+` on the last parameter to capture the remaining path into `[]string`:

```go
type Path struct {
	Section Section `/:"docs/:Rest+"`
	Rest    []string
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
	Section Section `/:"catalog"`
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
	Section Section `/:"search"`
	Query   url.Values
}
```

For `/search?q=doors&tag=go&tag=ui`, `Query.Get("q")` is `doors` and `Query["tag"]` is `[]string{"go", "ui"}`. Encoding uses the same `url.Values` field.

This works well for open-ended query parameters, preserving unknown keys, or pages that already parse query themselves.

Don't mix a `url.Values` field with `query` tags in the same model. Pick one style per model.

## Multiple Models

A page can route on more than one path model — list them in order of specificity:

```gox
<>
    ~(doors.Route(
        doors.RouteModel(renderHome),
        doors.RouteModel(renderPost),
        doors.RouteModel(renderCatalog),
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

The location is client state. A user can craft any URL that passes your path model decoder.

Two consequences:

- **Always check permissions while rendering, against the current location or decoded model.** A successful match means "this URL parses", not "this user is allowed to see what it points at".
- **For state the client must not control, use a separate `Source`.** Don't store auth, role, or other trust-bearing values on the route. Initialize them server-side and keep them on a session-scoped source. See [Storage & Auth](./18-storage-auth.md).

The convenience is that the URL can be read and written like normal state. The trust boundary does not change: reading it tells you what the client asked for, nothing more.

## Early Decisions

Two places commonly decide setup, redirects, or access before most UI is produced:

- **The page function passed to `doors.NewApp(...)`** runs once per page request and has access to `doors.Request` (cookies, headers, response writer). The usual job here is to bootstrap session-scoped state from cookies or headers, such as hydrating an auth source. See [Storage & Auth](./18-storage-auth.md). You can also read or change the location here.

  Mutating the location source in the app factory replaces the browser URL, and your app code observes the new location as the initial state. To read the current location before deciding whether to redirect, use `Get()` — unlike `Read`, it does not freeze the observable value for the render cycle. Place any `Sub` or `ReadAndSub` calls **after** the mutation.
- **Inside the matched component.** You have the **Doors** runtime context and the matched route value. Read or update it, derive smaller state, and decide what to render from session state.

To redirect or block a request before it reaches the **Doors** handler (system endpoints or page function), use HTTP middleware via `app.Use(...)`. The page function and component code both run after the request has already been accepted — they can change the location, but they aren't the place for HTTP-level redirects.

---

## Built on State Routing

URL routing is one application of a state primitive. `doors.Router(ctx)` returns the location source itself:

```go
loc := doors.Router(ctx) // doors.Source[doors.Location]
```

`doors.Location` contains reference types (`Segments []string` and `Query url.Values`). Don't mutate a value obtained from the router directly. Clone it first when you need a mutable copy:

```go
loc := doors.Router(ctx).Get().Clone()
```

`doors.Route(routes...)` routes the current location source:

```go
doors.Router(ctx).Route(routes...)
```

Matched routes may receive a `Source` and write back to the routed value.

### The Primitive

State routing swaps one of several views based on a reactive value. Routes are built from `RouteValue`, `RouteMatch`, and `RouteDerive`, then completed with `.Comp`, `.Beam(...)`, `.Bind(...)`, or `.Source(...)`, plus an optional `RouteDefault*` fallback.

The fragment swaps only when the active route changes. Within a matched route, the render function receives a live `Beam` or `Source` and reacts to value changes with normal state primitives. Full reference in [State](./07-state.md).

### Patterns

In a routing context, that opens up a few common patterns:

- **Match the location without a path model.** The general `RouteValue`, `RouteMatch`, and `RouteDerive` builders work directly on `Location`, and compose freely with `RouteModel` / `RouteModelBeam` in the same router. Reach for them when one ad-hoc URL doesn't deserve its own struct.
- **Take the location source directly.** `RouteLocationDefault`, `RouteLocationDefaultBeam`, and `RouteLocationDefaultBind` give the matched view a `Source[Location]`, `Beam[Location]`, or raw `Location` value. Useful as a fallback or when a page parses URLs itself.
- **Dispatch on a derived value.** Once a page has a typed `Path`, derive a beam for the variant field and route on it again. The fragment only swaps when the variant actually changes.
- **Mix URL routing with non-URL state.** Branch on a session-scoped flag, a feature toggle, an auth state — using the same `Route*` builders.
- **Custom slices of the URL.** Derive a single-purpose source or beam from `doors.Router(ctx)` (for example, one query param) without committing to a path model.
- **Compose levels.** Route on `Location` to pick a section, then route on a derived value inside it. Each level only rerenders on its own changes.

For derivation patterns (`DeriveSource`, `DeriveBeam`), the full `Route*` builder reference, the `.Comp` / `.Beam` / `.Source` chain, write-back via `RouteDerive(...).Source(set, render)`, and how routing fits in with `Bind` and `Effect`, see [State](./07-state.md).
