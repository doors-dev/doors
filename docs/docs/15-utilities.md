# Utilities

## `NewLocation` 

Encodes a model into a `Location` using a registered path adapter.

```go
func NewLocation(ctx context.Context, model any) (Location, error)
```

- Supports `path:"/pattern"` tags for routing variants.  
- Supports query tags via `query:"name"`.  

**Example**:

```go
loc, err := doors.NewLocation(ctx, ProductPath{
    Item: true,
    Id:   123,
    Sort: "price",
})
if err != nil {
  	return err
}
path := loc.String() // "/products/123?sort=price"

```

**Supported on static pages (`PageRouter.StaticPage`)**

## `RandId`

Generates a secure random string identifier.

```go
func RandId() string
```

- URL-safe, case-sensitive.  
- Use for tokens, sessions, or instance IDs.  

## `AllowBlocking`

Enables blocking (`Door.X...`) operations.

```go
func AllowBlocking(ctx context.Context) context.Context
```

