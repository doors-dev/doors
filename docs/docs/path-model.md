# Path Model

*doors* decodes the page request URI to a struct (path model), as if the address were a serialization format.

**Functionality**

* Declare multiple path variants (patterns)
* Capture int, string, and float64 params
* Query

### Basic Declaration

You declare a path variant by tagging a struct's exported boolean-typed field. 

```go
type Path struct {
	Path bool `path:"/"`  // match only root path
}
```

### Capture Parts

To capture path parts as `struct` fields use `:FieldName` syntax inside path pattern

```go
type Path struct {
	Path bool `path:"/:Id"` // match /123 (and /123/) where 123 is any int 
	Id 	 int // param saved here, field must be exported
}
```

### Path Variants

To provide path variants... provide path variants. In a deserialized struct, only the matched marker will have a `true` value.
```go
type Path struct {
	Catalog bool `path:"/catalog"` 
	Card    bool `path:"/catalog/:Id"` 
	Id 	    int 
}
```

### Capture To End

You can also match and capture any path "to end" using `+` after the last (must be last) part capture.

```go
type Path struct {
	Path bool `path:"/info/:Rest+"` // match /info/* where * any not empty path  
	Rest string // must be string or []string for "to end" capture
}
```

> **LIMITATIONS**
>
> For ease of use, the path pattern is matched to an already URL-decoded string. That means you can't match or use "/", ":", and "+" characters inside path parts, as they will always be treated as pattern syntax. 


### Query Params

To capture query param tag target struct field with `query` tag query param name value. Query params do not affect routing.

```go
type Path struct {
	Path          bool `path:"/catalog"` 
	ColorFilter   string `query:"color"` // ?color=red 
}
```
 > Decoding and encoding are provided by the [go-playground/form v4](https://github.com/go-playground/form) library.  So refer to its documentation for all features.
 >
 > Differences:
 >
 > -  `query` tag  instead of `form`
 > - empty values removed from query string on encoding

