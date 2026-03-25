# Routing (Model)

Routing in **Doors** is declarative and type-safe. 

Path is matched and decoded into a struct value using the pattern, provided in struct tags.

```go
type Path struct {
    Home bool `path:"/"` 
}
```

This struct will match root path


## Variants 

You can provide as many path variants as you need. Field with matched variant will have `true` value


```go
type Path struct {
    Home bool `path:"/"` // first path variant
    Docs bool `path:"/docs"` // alternative path variant
}
```

This struct will match root and `/docs` path. Matched field will receive value `true` on decoding, and symmetricaly, on encoding variant form field with `true` value will be used.

> Trailing and leading slashes can be omitted, this makes no effect - `/docs`, `docs`, `/docs/` are the same.

## Parameters

To capture path segments as struct fields using :FieldName syntax inside the path variant:


```go
type Path struct {
    Catalog bool `path:"/:Cat/:ID"`
    Cat     string
    ID      int
}
```

You can capture `string`, `int`, `uint` and `float64` type fields.

> Capture fields must be exported!

### Optional Parameters

To make last parameter optional, put "?" after it:

```go
type Path struct {
    Catalog bool `path:"/:Cat/:ID?"`
    Cat     string
    ID      *int
}
```

Optional parameter must have reference type (`*string`, `*int`, etc). 

> Only last segment can be optional and only parameter segment can be optional. 

### Capture All

To make last parameter capture all remaining segments, put "+" after it:

```go
type Path struct {
    Docs bool `path:"/docs/:Rest+"`
    Rest []string
}
```
> Capture all field must be of type `[]string`.

This will require at least 1 segment after `/docs`. 

If you want the tail be complitely optional, add "?":

```go
type Path struct {
    Docs bool `path:"/docs/:Rest+?"`
    Rest []string
}
```

Now `Rest` field can be zero length and `/docs` will also match.

## Query
