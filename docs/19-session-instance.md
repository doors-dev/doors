# Session & Instance

In **Doors**, a session and an instance are different lifetimes:

- a session is the shared browser-level scope
- an instance is one live page, usually one tab

Most app code does not need to control them directly. **Doors** manages them for you.

These utilities are for the cases where you want to:

- cap how long the current **Doors** session may live
- force-end a session or just one instance
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

- a session is created when the request does not have a live **Doors** session cookie
- that session is renewed on later requests and uses a timer-based lifetime
- a page gets its own live instance

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

`doors.SessionId(ctx)` returns the ID of the current **Doors** session.

```go
sessionID := doors.SessionId(ctx)
```

All instances in the same session share that value.

`doors.InstanceId(ctx)` returns the ID of the current live page instance.

```go
instanceID := doors.InstanceId(ctx)
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
- Use `SessionEnd` only when you really want to end the whole **Doors** session.
- Use `InstanceEnd` when only the current live page should stop.
- Use `SessionId` and `InstanceId` for diagnostics, not as business identifiers.
