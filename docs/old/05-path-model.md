# Path Model

*doors* decodes the request URI into a struct (path model), treating the address as a serialized representation.

## Functionality

- Declare multiple path variants (patterns)
- Capture `int`, `string`, and `float64` parameters
- Handle query parameters

---

### Basic Declaration

Declare a path variant by tagging an exported boolean field.

```go
type Path struct {
	Path bool `path:"/"` // matches only root path
}
```

---

### Capture Parts

Capture path parts as struct fields using `:FieldName` syntax inside the path pattern.

```go
type Path struct {
	Path bool `path:"/:Id"` // matches /123 or /123/ where 123 is any int
	Id   int                // parameter stored here (field must be exported)
}
```

> **LIMITATIONS**  
> The path pattern is matched to a URL-decoded string.  
> You cannot match or use `/`, `:`, `+`, or `*` characters inside captured values.

---

### Path Variants

Provide multiple path variants in one model.  
In a deserialized struct, only the matched marker field will be `true`.

```go
type Path struct {
	Catalog bool `path:"/catalog"`
	Card    bool `path:"/catalog/:Id"`
	Id      int
}
```

---

### Capture To End

Capture the remaining path using `+` after the last (and only the last) part capture.  
The captured portion cannot be empty.

```go
type Path struct {
	Path bool   `path:"/info/:Rest+"` // matches /info/* where * is any non-empty path
	Rest string // or []string for multiple parts
}
```

---

### Optional Capture To End

Capture the remaining path (including possibly empty) using `*` after the last part capture.

```go
type Path struct {
	Path bool   `path:"/info/:Rest*"` // matches /info/* where * may be empty
	Rest string // or []string for multiple parts
}
```

> 

---

### Query Params

Capture query parameters using the `query` tag.  
Query parameters do not affect routing unless decoding fails.

```go
type Path struct {
	Path        bool     `path:"/catalog"`
	ColorFilter *string  `query:"color"` // ?color=red
}
```

> Decoding and encoding use [go-playground/form v4](https://github.com/go-playground/form).  
> Use pointer types to omit default values from generated URLs.

---

### Capture Any Path

A model that matches any path. Useful for fallback or 404 handling.

```go
type AnyPath struct {
	IsRoot bool   `path:"/"`
	IsPath bool   `path:"/:Path+"` // capture to end
	Path   string // or []string for parts
}
```

---

## Link Behavior

Links within the same path model type update dynamically without reloading;  
links to a different model type cause a full reload.

