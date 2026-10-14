# Session & Instance

In **Doors**, a session and an instance are different lifetimes:

- a session is the shared scope for one browser session
- an instance is one live page in that session

Most app code does not need to control them directly. **Doors** manages them for you.

These utilities are for the cases where you want to:

- cap how long the current **Doors** session may live
- force-end a session or just one instance
- attach background work to the session lifetime
- attach runtime IDs to logs or traces

For session and instance storage, see [Storage & Auth](./18-storage-auth.md).

## Use

These functions need a **Doors** context, such as the `ctx` you get in:

- event handlers
- hook handlers
- `doors.Go(...)`
- other **Doors** runtime callbacks

Use the current **Doors** `ctx`, not `context.Background()`.

## Lifecycle

By default, **Doors** manages session and instance lifetime automatically.

At a high level:

- a request without a live **Doors** session cookie gets a new session cookie on that response right away, but the session itself is created lazily, when something first uses it
- rendering a page uses the session; requests fully handled by `app.Use` middleware or `doors.UseResource` do not create one
- a session is renewed on later requests and uses a timer-based lifetime that starts at creation
- a page gets its own live instance

The internal **Doors** session cookie is `HttpOnly`, `Secure`, scoped to `/`, and named from `ServerSessionCookiePrefix + WithID`. By default, the prefix is empty and the app ID defaults to `doors`, so the cookie is named `doors`.

For plain HTTP development, `ServerSessionCookieNoSecure` can omit `Secure`.

If you want browser-enforced cookie prefix rules, set `ServerSessionCookiePrefix` explicitly, for example to `__Host-` or `__Secure-`.

Instances also have their own lifecycle rules:

- a new instance must get its first client connection within `InstanceConnectTimeout`
- after that, inactive instances are cleaned up by `InstanceTTL`
- if the session reaches `SessionInstanceLimit`, older instances can be suspended

That suspension is why an older page may come back by reloading when the user returns to it.

## Expire

`doors.SessionExpire(ctx, d)` sets the maximum remaining lifetime of the current **Doors** session.

```go
doors.SessionExpire(ctx, 24*time.Hour)
```

This affects the whole session, not just the current page.

Use it when your application has its own session lifetime and you do not want the internal **Doors** session to outlive it.

One common case is login:

```go
doors.SessionExpire(ctx, sessionDuration)
```

Important detail: this is a cap, not a keepalive. The session can still end earlier because of the normal session TTL rules. The earlier limit wins.

It also does not affect only one tab. It sets the limit for the whole current **Doors** session.

If you call `doors.SessionExpire(ctx, 0)`, **Doors** removes that forced expiration cap and goes back to the normal session TTL behavior.

## End

`doors.SessionEnd(ctx)` force-ends the whole current **Doors** session.

```go
doors.SessionEnd(ctx)
```

That ends all live instances in the session, not just the page that called it.

Use it when you intentionally want a full teardown, for example:

- a hard logout that should close every open page immediately
- a security event that invalidates the whole session
- a forced account or tenant switch where all live pages must stop

With shared reactive session state, this is not the normal auth update path. For normal login, logout, and deauth flows, it is usually better to update session-scoped state and let pages react.

## Context

`doors.SessionContext(ctx)` returns a context that is canceled when the current **Doors** session ends.

```go
sessionCtx := doors.SessionContext(ctx)
```

Use it for goroutines or external work that should live for the whole browser session and stop on `SessionEnd`, session expiration, or session cleanup.

It is broader than the current instance or dynamic owner context. It is suitable for session-scoped helpers, Door methods, and Source or Beam reads, subscriptions, and updates; a subscription made with it lives until the session ends. `doors.Reload`, the package-level `doors.Deferred*` functions, and instance-scoped helpers panic on it.

For work that should stay scoped to the current dynamic owner, keep using the current `ctx`, or use `doors.DetachedContext(ctx)` from a goroutine of your own, see [Detached Context](#detached-context).

For work that should outlive the current dynamic owner but stay bounded by the current page instance, use `doors.InstanceContext(ctx)`.

Both are detached contexts: **wait on the channels of their operations only in goroutines of your own**, see [Detached Context](#detached-context).

## Detached Context

`doors.DetachedContext(ctx)` returns a context whose operations are delivered on their own. It keeps the values and the owner of `ctx`, so it is canceled when the content you called it from goes away, but its operations do not join the handler's updates, its batch, or its indicator.

```go
dctx := doors.DetachedContext(ctx)
go func() {
	// Door operations and doors.Call with dctx
}()
```

> Operations made with a detached context are not part of the current handler's work: not its batch, its indicator, or `OnSettle`. **Wait on their channels only in goroutines of your own**, never during rendering, in a handler, or in a subscription: that can block the page.

Whether a context is detached decides what the completion channels of Door operations and `doors.Call` deliver.

Every channel first sends one value once the call is scheduled: `nil`, or an error when it failed before that point — the render failed, the action could not be built, or the call was canceled or superseded by a newer operation (`context.Canceled`). If the Door is not mounted, the channel closes without a value.

What comes after that value depends on the context:

| Context | After the scheduled value |
|---|---|
| Detached | A second value: `nil` once the update is on the page or the browser has answered, an error, or `context.Canceled` when the call was canceled after it was scheduled. A destination armed with `Into` is written before it. Then the channel closes. |
| Any other | The channel closes. No result is reported, and a destination armed with `Into` stays unwritten; the closed channel is the only signal, there is no error. |

A channel that closes without a further value means the operation was accepted, but its outcome is not reported: the Door is not mounted yet, or the context is not detached.

Detached contexts:

- `doors.DetachedContext(ctx)`
- `doors.InstanceContext(ctx)`
- the `ctx` of `doors.Go(...)`
- the `ctx` passed to `doors.OnReady` and `doors.OnBatch` callbacks
- any context derived from one of these with the `context` package, such as `context.WithTimeout`

Not detached:

- handler `ctx` — event, form, and hook handlers (also when used later from a goroutine)
- render `ctx`
- beam subscription `ctx`
- the `ctx` passed to `doors.OnSettle` callbacks: it keeps their operations in place with the handler's updates, so do not block in them
- the `ctx` passed to the `ops` of `doors.OnSettle` and `doors.OnBatch`, even when the call itself was made with a detached context

If you need the result, wait for the second value in a goroutine of your own, on a detached context:

```go
go func(ctx context.Context) {
	var now time.Time
	ch := doors.Call(ctx, doors.ActionTime{}.Into(&now))
	<-ch // scheduled
	if err, ok := <-ch; ok && err == nil {
		// use now
	}
}(doors.DetachedContext(ctx))
```

Use `doors.Go(...)` for the same wait when the goroutine should follow the lifetime of a rendered subtree, and `doors.InstanceContext(ctx)` when it should outlive the content it started from but stop with the page.

## Instance End

`doors.InstanceEnd(ctx)` ends only the current live page instance.


```go
doors.InstanceEnd(ctx)
```

Use it when the current page should stop, but the rest of the session should stay active.

For example:

- a one-off detached page should go away after finishing its work
- the current tab should be discarded, but other tabs should stay live

This ends the current live instance immediately. It does not end the session and does not touch sibling instances.

## IDs

`doors.SessionID(ctx)` returns the ID of the current **Doors** session.

```go
sessionID := doors.SessionID(ctx)
```

All instances in the same session share that value.

`doors.InstanceID(ctx)` returns the ID of the current live page instance.

```go
instanceID := doors.InstanceID(ctx)
```

That value is specific to one live page.

These IDs are useful for:

- logging
- tracing
- grouping related events
- debugging multi-tab behavior

They are **Doors** runtime IDs, not user IDs or authentication tokens.

## Rules

- Use `SessionExpire` to cap the session lifetime.
- Use `SessionExpire(ctx, 0)` to remove an explicit session-expiration cap.
- Use `SessionContext` for work that should stop when the whole session ends.
- Use `DetachedContext` for goroutines of your own that keep the current dynamic owner; it is the context on which completion channels deliver results.
- Use `InstanceContext` for goroutines that should outlive the current dynamic owner but stop with the current instance.
- Only detached contexts get the result of a Door operation or `doors.Call`; see [Detached Context](#detached-context).
- Use `SessionEnd` only when you really want to end the whole **Doors** session.
- Use `InstanceEnd` when only the current live page should stop.
- Use `SessionID` and `InstanceID` for diagnostics, not as business identifiers.
