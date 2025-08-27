# `href` and `src`

## `AHref` for  internal links 

Links can be of two types: one leads to a dynamic page update, the other to a new page load.  It depends on whether the Path Model type matches the current page or not.

To convert the path model to `href`, use  `AHref` attribute

```templ
type AHref struct {
  // Target path model
	Model           any
	// Active link indicator configuration
	Active          Active

	// For dynamic links
	StopPropagation bool  // stop click event propagation
	// scrolls into specified selector (after action)
  ScrollInto      string
	Indicator       []Indicator // loading indicator
	OnError         []OnError  // error action 
}
```

### Active Link Indication

Framework includes flexible tooling for active link indication via the indicator API. 

If the browser URL matches both the PathMatcher and QueryMatcher, the indication is applied.

```templ
 type Active struct {
  // Configure how to match path
	PathMatcher  PathMatcher
	// Configure how to match query params
	QueryMatcher QueryMatcher
	// indicator API, check ref/indication article for details
	Indicator    []Indicator
}
```

#### PathMatcher

* [Default] `PathMatcherFull`
  The browser path must match the path in `href` exactly.
* `PathMatcherStarts`
  The browser path must start with the path in `href` .
* `PathMatcherParts(n int)`
  *The first n parts of the browser path and  `href` path must match.*

#### QueryMatcher

* [Default] `QueryMatcherAll `
  *All query params must match between the page path and href*
* `QueryMatcherIgnore`
  *Query params do not matter*
* `QueryMatcherSome(params ...string)`
  *Values of the provided query params must match* 

### Example:

```templ
@doors.AHref{
  // link to the main variant of catalog page
  Model: CatalogPath{
    IsMain: true,
  },
  Active: doors.Active{
    // add "contrast" class to link if its active
    Indicator:    doors.IndicatorClass("contrast"),
    // browser path must start with href path
    PathMatcher:  doors.PathMatcherStarts(),
    // ignore query params for matching
    QueryMatcher: doors.QueryMatcherIgnore(),
  },
}
<a>catalog</a>
```

## File Src/Href

**You can serve any file via a secure, unique link** (session scoped). This is useful for any resources inside a private, authorized space (images, documents).

### Src From Path

```templ

type ASrc struct {
  // local path
	Path string
	// serve only once
	Once bool
	// file name, optional (default value is taken from path)
	Name string
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

This will generate a temporary source that can only be accessed once and only by the session to which the page was served.

### Raw Src

To handle the source as a raw HTTP request. Useful for proxying.

```templ
type ARawSrc struct {
  // serve only once
	Once    bool
	// file name, optional (default value is taken from path)
	Name    string
	// http request handler
	Handler func(w http.ResponseWriter, r *http.Request)
}
```

### File Href

The same as `ASrc` but prepates href.

```templ
type AFileHref struct {
	// local path
	Path string
	// serve only once
	Once bool
	// file name, optional (default value is taken from path)
	Name string
}
```

Example:

```templ
@doors.AFileHref{
	Path: "./docs/passport.pdf",
}
<a target="_blank">downlod</a>
```

### Raw file href

The same as raw src, to serve a resource via a raw http request handler

```templ
type ARawFileHref struct {
	Once    bool
	Name    string
	Handler func(w http.ResponseWriter, r *http.Request)
}
```



