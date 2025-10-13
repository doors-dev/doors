# Storage

The **storage API** provides key-value persistence mechanisms for both **session** and **instance** scopes.  
Session storage is shared across all instances of a user session, while instance storage is isolated to the current browser tab or page.

Both layers are thread-safe and accessible from any event or handler via the `context.Context`.

---

## Session Storage

Session storage persists for the lifetime of a user session and is shared across all active instances belonging to it.

### `SessionSave`

Stores a key/value pair in session-scoped storage shared by all instances in the session.  
Returns the previous value under the key.

```go
func SessionSave(ctx context.Context, key any, value any) any
```

**Example:**

```go
type Preferences struct {
  Theme    string
  Language string
}

doors.SessionSave(ctx, "prefs", Preferences{
  Theme:    "dark",
  Language: "en",
})
```

---

### `SessionLoad`

Gets a value from session-scoped storage by key.  
Returns nil if absent. Callers must type-assert the result.

```go
func SessionLoad(ctx context.Context, key any) any
```

**Example:**

```go
prefs, ok := doors.SessionLoad(ctx, "prefs").(Preferences)
if ok {
  applyTheme(prefs.Theme)
}
```

---

### `SessionRemove`

Deletes a key/value from session-scoped storage.  
Returns the removed value or nil if absent.

```go
func SessionRemove(ctx context.Context, key any) any
```

**Example:**

```go
doors.SessionRemove(ctx, "prefs")
```

---

## Instance Storage

Instance storage is isolated to a single tab or page instance.  
It persists only for the lifetime of that instance.

### `InstanceSave`

Stores a key/value pair in instance-scoped storage.  
Returns the previous value under the key.

```go
func InstanceSave(ctx context.Context, key any, value any) any
```

**Example:**

```go
doors.InstanceSave(ctx, "counter", 42)
```

---

### `InstanceLoad`

Gets a value from instance-scoped storage by key.  
Returns nil if absent. Callers must type-assert the result.

```go
func InstanceLoad(ctx context.Context, key any) any
```

**Example:**

```go
if v, ok := doors.InstanceLoad(ctx, "counter").(int); ok {
  fmt.Println("Counter:", v)
}
```

---

### `InstanceRemove`

Deletes a key/value from instance-scoped storage.  
Returns the removed value or nil if absent.

```go
func InstanceRemove(ctx context.Context, key any) any
```

**Example:**

```go
doors.InstanceRemove(ctx, "counter")
```

---

## Notes

- **Thread-safe**: Both session and instance storage support concurrent access.
- **Session scope**: Shared across all instances of the same session.
- **Instance scope**: Isolated to a single tab or page.
- **Type safety**: Returned values must be type-asserted.
