# Href and Src

These attributes control dynamic navigation and secure file access.

---

## `AHref` — Internal Links

`AHref` converts a path model into an internal hyperlink.  
It determines whether the link is **dynamic** (updates the page without reload) or **static** (loads a new page) depending on whether the target path model matches the current one.

```go
type AHref struct {
	Model           any
	Active          Active
	StopPropagation bool
	Scope           []Scope
	Indicator       []Indicator
	Before          []Action
	After           []Action
	OnError         []Action
}
```

### Example

```templ
@doors.AHref{
  Model: CatalogPath{IsMain: true},
  Active: doors.Active{
    Indicator:    doors.IndicatorOnlyClass("contrast"),
    PathMatcher:  doors.PathMatcherStarts(),
    QueryMatcher: doors.QueryMatcherOnlyIgnoreAll(),
  },
}
<a>catalog</a>
```

---

## Active Link Indication

`Active` defines how a link is highlighted when the current URL matches.  
When active indicators are configured, the implementation automatically appends a final step that compares all remaining query parameters for equality, and defaults the path matcher to a full match if none was set.

```go
type Active struct {
	PathMatcher  PathMatcher
	QueryMatcher []QueryMatcher
	Indicator    []Indicator
}
```

### PathMatcher

- `PathMatcherFull()` — full path must match (default)
- `PathMatcherStarts()` — browser path must start with `href`
- `PathMatcherParts(i ...int)` — compare specific path segments

---

## QueryMatcher

Controls how query parameters are checked when matching URLs.  
Matchers run left-to-right. Each step may compare specific keys or remove keys from further checks.  
At the end, all remaining parameters are always compared for equality (`All` is implicit).

### Primitive Matchers

- `QueryMatcherIgnoreAll()` — ignore all remaining parameters and stop comparison.
- `QueryMatcherIgnoreSome(params...)` — remove listed parameters from both URLs, continue comparison.
- `QueryMatcherSome(params...)` — compare only the listed parameters for equality at this step, then continue.
- `QueryMatcherIfPresent(params...)` — compare listed parameters only if they are present; ignore missing ones, then continue.

### Helper Matchers

Helpers compose complete matcher configurations:

| Helper | Behavior |
|--------|-----------|
| `QueryMatcherOnlyIgnoreAll()` | `[IgnoreAll]` |
| `QueryMatcherOnlyIgnoreSome(params...)` | `[IgnoreSome(params)]` + implicit comparison of remaining |
| `QueryMatcherOnlySome(params...)` | `[Some(params), IgnoreAll]` |
| `QueryMatcherOnlyIfPresent(params...)` | `[IfPresent(params), IgnoreAll]` |

---

## File `src` and `href`

Files can be served via private, session-scoped links.  
This enables secure access to protected resources such as images or downloads.

### `ASrc` — File Src

Serves a file privately from the filesystem.

```go
type ASrc struct {
	Once bool
	Name string
	Path string
}
```

Example:

```templ
@doors.ASrc{
	Path: "./images/luck.jpg",
	Name: "dedication.jpg",
	Once: true,
}
<img alt="secret to my success">
```

---

### `ARawSrc` — Custom Src Handler

Serves a resource dynamically via an inline HTTP handler.

```go
type ARawSrc struct {
	Once    bool
	Name    string
	Handler func(w http.ResponseWriter, r *http.Request)
}
```

---

### `AFileHref` — File Download Link

Same as `ASrc`, but generates a private `href`.

```go
type AFileHref struct {
	Once bool
	Name string
	Path string
}
```

Example:

```templ
@doors.AFileHref{
	Path: "./docs/passport.pdf",
}
<a target="_blank">download</a>
```

---

### `ARawFileHref` — Custom File Href Handler

Same as `ARawSrc`, but prepares an `href`.

```go
type ARawFileHref struct {
	Once    bool
	Name    string
	Handler func(w http.ResponseWriter, r *http.Request)
}
```
