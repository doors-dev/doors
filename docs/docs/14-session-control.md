# Session And Instance Management

This API provides functions for managing session and instance lifecycles. 

## Session Management

### `SessionExpire`

Sets an expiration duration for the session.

```go
func SessionExpire(ctx context.Context, d time.Duration)
```

- After inactivity exceeding `d`, the session terminates automatically.  
- `d = 0` disables expiration (session ends immediately if no instances remain).  

### `SessionEnd`

Terminates the current session and all its instances immediately.

```go
func SessionEnd(ctx context.Context)
```

- Ensures **all instances** are destroyed.  

### `SessionId`

Returns a unique identifier for the current session.

```go
func SessionId(ctx context.Context) string
```

- Shared across all instances in the same browser session.  

**Example**:

```go
doors.SessionEnd(ctx)
```

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

