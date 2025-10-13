# Session and Instance Management

Functions for controlling session and instance lifecycles.  
Use them to manage authentication scopes, resource cleanup, and user state across instances.

---

## Session Management

### `SessionExpire`

Sets the maximum lifetime of the current session.

```go
func SessionExpire(ctx context.Context, d time.Duration)
```

After inactivity exceeding `d`, the session terminates automatically.  
If `d = 0`, expiration is disabled — the session will end only when no active instances remain.

---

### `SessionEnd`

Immediately ends the current session and all its instances.  
Used during logout or when invalidating an active session.

```go
func SessionEnd(ctx context.Context)
```

This call closes all authorized pages, frees server resources, and resets the session scope.

---

### `SessionId`

Returns the unique ID of the current session.

```go
func SessionId(ctx context.Context) string
```

All instances within the same browser share this ID via a session cookie.  
Useful for grouping and tracking user activity.

---

## Instance Management

### `InstanceEnd`

Ends the current instance (tab or window) while keeping the session and other instances active.

```go
func InstanceEnd(ctx context.Context)
```

Used to close an active tab or release resources without affecting the session.  
Often invoked after operations that isolate a page or switch user context.

---

### `InstanceId`

Returns the unique ID of the current instance.

```go
func InstanceId(ctx context.Context) string
```

Useful for logging, debugging, and identifying client connections across tabs.

---

## Example

```go
func (h *Logout) Handle(ctx context.Context) {
  doors.SessionEnd(ctx) // end session for all instances
}

func (h *CloseTab) Handle(ctx context.Context) {
  doors.InstanceEnd(ctx) // end only this tab
}
```

---

## Notes

- Each **instance** represents an open browser tab or page.  
- Each **session** groups multiple instances sharing authentication.  
- Session IDs persist via cookies; instance IDs are transient.  
- All functions are context-based and safe to call within any handler.
