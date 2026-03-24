# Utils

Utility functions for ID generation and context control.

---

## `RandId`

```go
func RandId() string
```

Returns a **cryptographically secure**, URL-safe random ID string.  
Used for sessions, instances, tokens, or HTML attributes.  
Case-sensitive and unique across runs.

---

## `HashId`

```go
func HashId(s string) string
```

Creates a deterministic ID derived from the given string using hash-based function.  
Always produces the same result for the same input.  

Use cases:
- Stable identifiers for HTML attributes
- Reproducible keys for caching or element binding

---

## `AllowBlocking`

```go
func AllowBlocking(ctx context.Context) context.Context
```

Returns a derived context that suppresses framework warnings for **blocking X*** operations.  
Use this only when a blocking operation is intentional (for example, in controlled synchronization or debug scenarios).

