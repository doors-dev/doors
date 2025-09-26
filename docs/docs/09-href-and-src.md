# href and src

## `AHref` for internal links

Links can be of two types:  
- **Dynamic link**: updates the current page without reload.  
- **Static link**: loads a new page.  

The behavior depends on whether the path model type matches the current page.

To convert a path model into an `href`, use the `AHref` attribute.

```go
type AHref struct {
	// Target path model value. Required.
	Model any
	// Active link indicator configuration. Optional.
	Active Active

	// For dynamic links
	StopPropagation bool        // stop event propagation
	Before          []Action    // actions to run before the hook request
	Indicator       []Indicator // visual indicators while the request is running
	Before          []Action // actions to run after the hook request
	OnError         []Action    // actions on error (default reload)
}
```

### Active Link Indication

Active link indication is configured through the `Active` struct.
 If both `PathMatcher` and `QueryMatcher` match the current browser URL, indicators are applied.

```templ
type Active struct {
	// Path match strategy
	PathMatcher PathMatcher
	// Query param match strategy
	QueryMatcher []QueryMatcher
	// Indicators to apply when active
	Indicator []Indicator
}
```

#### PathMatcher

- **PathMatcherFull** *(default)*
   Browser path must match `href` exactly.
- **PathMatcherStarts**
   Browser path must start with `href`.
- **PathMatcherParts(n uint8)**
   First `n` path segments of browser path and `href` must match.

#### QueryMatcher

Controls how query parameters are checked when matching URLs.  
Matchers run left-to-right. Each step may compare specific keys or remove keys from further checks.  
At the end, all remaining parameters are always compared for equality (`All` is implicit).

- **QueryMatcherIgnoreAll**  
  Ignore all parameters and stop. No further comparison is performed.

- **QueryMatcherIgnoreSome(params ...string)**  
  Remove the listed parameters from both URLs, then continue.

- **QueryMatcherSome(params ...string)**  
  Compare only the listed parameters, remove them, stop on mismatch, then continue.

- **QueryMatcherIfPresent(params ...string)**  
  For the listed parameters that exist in the new URL, compare and remove them; missing ones are ignored. Stop on mismatch, then continue.


#### Helpers

Helpers wrap the base matchers to form a complete configuration:

- **QueryMatcherOnlyIgnoreAll**  
  `[IgnoreAll]` — ignore all parameters.  

- **QueryMatcherOnlyIgnoreSome(params ...string)**  
  `[IgnoreSome(params)]` + implicit final compare of all remaining.  

- **QueryMatcherOnlySome(params ...string)**  
  `[Some(params), IgnoreAll]` — match the listed ones, ignore the rest.  

- **QueryMatcherOnlyIfPresent(params ...string)**  
  `[IfPresent(params), IgnoreAll]` — match the listed ones if present, ignore the rest.  


### Example

```templ
@doors.AHref{
  Model: CatalogPath{IsMain: true},
  Active: doors.Active{
    Indicator:    doors.IndicatorOnlyClass("contrast"),
    PathMatcher:  doors.PathMatcherStarts(),
    QueryMatcher: doors.QueryMatcherIgnoreAll(),
  },
}
<a>catalog</a>
```

## File Src/Href

Files can be served via private, session-scoped links. This is useful for protected resources (images, documents) inside authorized spaces.

### Src From Path

```templ
type ASrc struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// File system path to serve.
	Path string
}

```

example:

```templ
@doors.ASrc{
	Path: "./images/luck.jpg",
	Name: "dedication.jpg",
	Once: true,
}
<img alt="secret to my success">
```

The generated `src` can only be accessed once, bound to the session that received the page.

### Raw Src

Serve a resource directly via a custom HTTP handler. Useful for proxying or dynamic data.

```templ
type ARawSrc struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// Handler for serving the resource request.
	Handler func(w http.ResponseWriter, r *http.Request)
}
```

### File Href

Same as `ASrc`, but prepares an `href` instead of a `src`.

```templ
type AFileHref struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// File system path to serve.
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

### Raw File Href

Same as `ARawSrc`, but prepares an `href`.

```templ
type ARawFileHref struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// Handler for serving the resource request.
	Handler func(w http.ResponseWriter, r *http.Request)
}
```



