# Storage & Auth

In **Doors**, storage is a small key-value layer attached to session or instance lifetime.

That pattern is especially useful for authentication. The same idea also works for shared settings like theme or locale.

## Start

You usually access storage in one of three places:

- `r.SessionStore()` in a model handler
- `doors.SessionStore(ctx)` when you already have a **Doors** `ctx`
- `doors.InstanceStore(ctx)` for page-instance-local storage

The store API is:

- `Init(key, func() any)` creates the value once and returns it, or returns the existing value if it is already there
- `Load(key)` gets the current value for a key
- `Save(key, value)` replaces the stored value and returns the previous one
- `Remove(key)` deletes the value and returns what was stored before

For reactive session state, the usual pattern is:

- `Init` in the model handler
- `Load` later in event or hook handlers

`Save` and `Remove` are more useful when you want to replace or clear the stored object itself.

## Session

Session storage is shared by all live instances in the same **Doors** session.

Use it for state that should be shared across pages or tabs in that browser session, such as:

- authentication
- current user data
- theme
- locale

If several pages subscribe to the same source from session storage, they can all react to the same update.

## Instance

Instance storage belongs to one live page instance.

Use it for page-local state that should survive rerenders in that page but should not sync across other pages or tabs.

If the state should affect the whole logged-in browser session, use session storage instead.

## Model

The model handler is the usual place to bootstrap shared session state from cookies or headers.

For auth, a good pattern is:

- read the session cookie
- validate it against your real server-side session storage
- create or reuse one session-scoped source
- place that source on your `App`

```go
type authKey struct{}

type App struct {
	auth doors.Source[bool]
}

doors.UseModel(router, func(r doors.RequestModel, _ doors.Source[Path]) doors.Response {
	auth := r.SessionStore().Init(authKey{}, func() any {
		c, err := r.GetCookie("session")
		if err != nil {
			return doors.NewSource(false)
		}
		_, ok := driver.Sessions.Get(c.Value)
		return doors.NewSource(ok)
	}).(doors.Source[bool])

	return doors.ResponseComp(App{
		auth: auth,
	})
})
```

`Init` returns the existing value if it was already created earlier in the same session.

That means later requests do not create a new auth source. They get the same shared one back.

Do not treat the cookie by itself as proof of authentication. Keep the real session in your own database, Redis, or other server storage, usually by session ID from the cookie.

## Render

After that, render from your `App` fields instead of reaching back into storage again:

```gox
elem (a App) Main() {
	~(a.auth.Bind(elem(ok bool) {
		~(if !ok {
			<p>Please log in</p>
		} else {
			<p>Dashboard</p>
		})
	}))
}
```

Because `a.auth` is a session-scoped source, every page that subscribes to it can react when it changes.

## Update

On login or logout, update the real session storage, update the cookie, and update the same shared source.

```go
auth := doors.SessionStore(ctx).Load(authKey{}).(doors.Source[bool])
```

On login:

```go
func login(ctx context.Context, r doors.RequestForm[LoginData]) bool {
	auth := doors.SessionStore(ctx).Load(authKey{}).(doors.Source[bool])

	sessionDuration := 24 * time.Hour
	session := driver.Sessions.Add(r.Data().Login, sessionDuration)
	r.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
	})

	doors.SessionExpire(ctx, sessionDuration)
	auth.Update(ctx, true)
	return true
}
```

If your auth session has a fixed lifetime, it usually makes sense to call `doors.SessionExpire(ctx, sessionDuration)` on login too.

That keeps the internal **Doors** session from outliving the cookie or backend session and helps ensure an already-open instance does not keep handling requests after auth should be gone.

On logout or deauth:

```go
func logout(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
	auth := doors.SessionStore(ctx).Load(authKey{}).(doors.Source[bool])

	if c, err := r.GetCookie("session"); err == nil {
		driver.Sessions.Remove(c.Value)
	}

	r.SetCookie(&http.Cookie{
		Name:     "session",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	auth.Update(ctx, false)
	return true
}
```

This is the main benefit of shared reactive session state in **Doors**: open pages can react to auth changes immediately, without reload actions and without ending the whole **Doors** session.

## Rules

- If the UI should react, store a `Source` in the store.
- Put auth in session storage, not instance storage.
- Initialize shared state in the model handler, then keep it on your `App`.
- Validate the cookie against your real session storage, not against the cookie alone.
- If auth has a fixed lifetime, usually cap the **Doors** session to that same lifetime on login.
- Use `doors.SessionEnd(ctx)` only when you intentionally want to force-close the whole **Doors** session.
