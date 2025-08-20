# Session and Instance Storage

The **storage API** provides mechanisms to persist key-value pairs either at the **session scope** (shared across all instances/tabs of a session) or the **instance scope** (isolated to the current browser tab/page). Both storage layers are thread-safe and allow concurrent access from multiple goroutines.

---

## Session Storage

Session-scoped storage persists for the lifetime of a session and is shared across all browser tabs or instances belonging to that session.

### `SessionSave`

Stores a key-value pair in session storage.

```go
func SessionSave(ctx context.Context, key any, value any) bool
```

- **key**: any type, identifier for the stored value.  
- **value**: any type, value to persist.  
- Returns **true** if the value was successfully stored, otherwise **false**.  

**Example**:

```go
// Store user preferences globally across session
type Preferences struct {
    Theme    string
    Language string
}

saved := doors.SessionSave(ctx, "prefs", Preferences{
    Theme:    "dark",
    Language: "en",
})
```

---

### `SessionLoad`

Retrieves a value from session storage by its key.

```go
func SessionLoad(ctx context.Context, key any) any
```

- Returns the stored value, or **nil** if no value exists.  
- The result must be **type-asserted** to its original type.  

**Example**:

```go
// Load user preferences
prefs, ok := doors.SessionLoad(ctx, "prefs").(Preferences)
if !ok {
    return
}
applyTheme(prefs.Theme)
```

---

### `SessionRemove`

Deletes a key-value pair from session storage.

```go
func SessionRemove(ctx context.Context, key any)
```

- If the key does not exist, no action is taken.  

**Example**:

```go
// Remove user preferences
doors.SessionRemove(ctx, "prefs")
```

---

## Instance Storage

Instance-scoped storage persists only for the lifetime of the current **instance** (a single browser tab or page). Each instance has its own isolated storage.

### `InstanceSave`

Stores a key-value pair in instance storage.

```go
func InstanceSave(ctx context.Context, key any, value any) bool
```

- **key**: any type, identifier for the stored value.  
- **value**: any type, value to persist.  
- Returns **true** if the value was successfully stored, otherwise **false**.  

**Example**:

```go
// Store preferences only for this tab
type Preferences struct {
    Theme    string
    Language string
}

saved := doors.InstanceSave(ctx, "prefs", Preferences{
    Theme:    "dark",
    Language: "en",
})
```

---

### `InstanceLoad`

Retrieves a value from instance storage by its key.

```go
func InstanceLoad(ctx context.Context, key any) any
```

- Returns the stored value, or **nil** if the key does not exist.  
- The result must be **type-asserted** to its original type.  

**Example**:

```go
// Load preferences specific to this tab
if prefs, ok := doors.InstanceLoad(ctx, "prefs").(Preferences); ok {
    applyTheme(prefs.Theme)
}
```

---

### `InstanceRemove`

Deletes a key-value pair from instance storage.

```go
func InstanceRemove(ctx context.Context, key any)
```

- If the key does not exist, no action is taken.  

**Example**:

```go
// Remove preferences for this tab
doors.InstanceRemove(ctx, "prefs")
```

---

## Notes

- **Thread-safety**: Both storage layers are safe for concurrent use across goroutines.  
- **Persistence**:  
  - Session storage persists for the session’s lifetime and synchronizes across all instances
  - Instance storage persists only for the current browser tab/page and is isolated from others.  
- **Type safety**: Values retrieved must be explicitly type-asserted.  