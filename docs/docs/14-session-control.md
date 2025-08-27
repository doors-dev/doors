# Session And Instance Management

This API provides functions for managing session and instance lifecycles and browser navigation. These utilities interact with the framework’s client-side runtime, enabling server-driven control over navigation, reloads, and termination of sessions or instances. 

---

## Navigation

### `LocationReload`

Reloads the current browser page asynchronously, creating a new instance.

```go
func LocationReload(ctx context.Context)
```

**Example**:

```go
doors.LocationReload(ctx)
```

---

### `LocationAssignRaw`

Navigates to an external or raw URL using `location.assign(url)`.

```go
func LocationAssignRaw(ctx context.Context, url string)
```

- Creates a new browser history entry.  
- Use for **external URLs** or paths outside path models.  

**Example**:

```go
doors.LocationAssignRaw(ctx, "https://example.com")
```

---

### `LocationReplaceRaw`

Replaces the current location with a raw URL using `location.replace(url)`.

```go
func LocationReplaceRaw(ctx context.Context, url string)
```

- Does **not** create a new history entry.  
- Suitable for **redirects**.  

---

### `LocationReplace`

Replaces the current location with a model-based URL.

```go
func LocationReplace(ctx context.Context, model any) error
```

- The model must have a registered **path adapter**.  
- No history entry is created.  

**Example**:

```go
err := doors.LocationReplace(ctx, CatalogPath{
    IsCat: true,
    CatId: "electronics",
})
```

---

### `LocationAssign`

Navigates to a model-based URL using `location.assign(url)`.

```go
func LocationAssign(ctx context.Context, model any) error
```

- Adds a new entry in the history stack.  
- Same-model navigation updates instance reactively; cross-model causes reload.  

**Example**:

```go
err := doors.LocationAssign(ctx, CatalogPath{
    IsItem: true,
    CatId:  "electronics",
    ItemId: 123,
})
```

---

## Session Management

### `SessionExpire`

Sets an expiration duration for the session.

```go
func SessionExpire(ctx context.Context, d time.Duration)
```

- After inactivity exceeding `d`, the session terminates automatically.  
- `d = 0` disables expiration (session ends immediately if no instances remain).  

---

### `SessionEnd`

Terminates the current session and all its instances immediately.

```go
func SessionEnd(ctx context.Context)
```

- Ensures **all instances** are destroyed.  

**Example**:

```go
doors.SessionEnd(ctx)
```

---

## Instance Management

### `InstanceEnd`

Ends the current instance (browser tab/page) without affecting others, causing a reload.

```go
func InstanceEnd(ctx context.Context)
```

**Example**:

```go
// Close current tab after action
doors.InstanceEnd(ctx)
```

---

### `InstanceId`

Returns a unique identifier for the current instance.

```go
func InstanceId(ctx context.Context) string
```

- Useful for logging and tracking.  

---

### `SessionId`

Returns a unique identifier for the current session.

```go
func SessionId(ctx context.Context) string
```

- Shared across all instances in the same browser session.  

---

## Location Modeling

### `Location`

Alias for `path.Location`. Encapsulates URL path and query parameters.

---

### `NewLocation`

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
// "/products/123?sort=price"
```

---

## Utilities


