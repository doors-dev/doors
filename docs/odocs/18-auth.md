# Authentication

For protected resources, control authentication and authorization directly in the [App handler](./04-router.md).

---

## Access Control

Example:

```go
router.Use(doors.ServeApp(
  func(a doors.ModelRouter[Path], r doors.RModel[Path]) doors.ModelRoute {
    c, err := r.GetCookie("session")
    if err != nil {
      // show unauthorized page in detached (no path sync) mode
      return a.Reroute(UnauthorizedPath{}, true)
    }

    session, ok := db.Get(c.Value)
    if !ok {
      return a.Reroute(UnauthorizedPath{}, true)
    }

    return a.App(&Admin{})
  },
))
```

`doors.RModel` provides access to cookies, headers, and the requested [Path Model](./05-path-model.md).

There is no need to recheck cookies or headers in event handlers, since they are already scoped to the session and page instance.

> **Important:** Never rely only on a cookie value — always validate against session storage.

---

## When to Check Authorization Besides the App Handler

If user access to certain **actions** or **views** can be revoked, you should:

- **Verify user view permissions during render** to ensure the user cannot access previously available views through dynamic navigation.  
- **Verify user write permissions in transactions** to ensure that even if permissions are revoked after rendering, data integrity and security are maintained.

These checks prevent stale or cached views from exposing restricted content after permissions change.

---

## Session Management

The framework internally manages sessions for pages, content, and event handling.  
By default, a session persists as long as at least one model instance remains active.

When implementing your own authentication, ensure your application’s session does **not outlive** the framework’s session.  
Otherwise, users may appear logged out while their open tabs still have access.

---

### 1. Control Expiration in Login Handler

```go
templ (l *loginFragment) Render() {
  @doors.ASubmit[loginData]{
    On: func(ctx context.Context, r doors.RForm[loginData]) bool {
      // validate credentials

      sessionDuration := 24 * time.Hour
      session := db.CreateSession(r.Data, sessionDuration)
      r.SetCookie(&http.Cookie{
        Name:     "session",
        Value:    session.Token,
        Expires:  time.Now().Add(sessionDuration),
        Path:     "/",
        HttpOnly: true,
      })

      // align internal session lifetime with cookie
      doors.SessionExpire(ctx, sessionDuration)

      // reload to initialize private instance
      r.After(doors.ActionOnlyLocationReload())
      return true
    },
  }
  <form>...</form>
}
```

---

### 2. End Session on Logout

Destroy all active instances when logging out.

```go
templ logout() {
  @doors.AClick{
    On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
      // end doors session
      defer doors.SessionEnd(ctx)

      // clear cookies
      r.SetCookie(&http.Cookie{
        Name:   "session",
        Path:   "/",
        MaxAge: -1,
      })

      // remove session record
      db.Sessions.Remove(h.session.Token)
      return true
    },
  }
  <button>Log Out</button>
}
```

> **Note:** Ending the session reloads all pages.  
> This might occur before cookies are cleared.  
> To ensure consistency, always use session storage as the single source of truth for authentication.
