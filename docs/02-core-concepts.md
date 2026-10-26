# Core Concepts

**Doors** runs each interactive page as a live server-side instance, with the current URL exposed as a reactive state value.

Event handlers and dynamic fragments registered while rendering belong to that instance and are cleaned up when their part of the rendered tree goes away.

## Mental Model

A **Doors** page lives on the server. After the first render, you update it in two ways.

**Data-driven.** A part of the page shows a piece of data. When the data changes, its content is swapped for the new one.

- the URL: pages, items, filters
- UI state: tabs, dialogs, toggles, form inputs
- user status and settings
- live state: metrics, progress, presence

**Explicit.** The change is an operation, not a new value: append, prepend, insert, remove.

- virtual and managed lists
- growing content: chats, logs, streamed text
- real-time loops: games, animations, frames paced to the connection

Data-driven covers most parts of an app. The logic lives in the state: handlers change it, and each part shows what follows from it. However complex the logic gets, the page always matches the state.

Explicit gives precise control for the cases that need it, at the cost of more code: you keep track of what is on the page and keep it in step yourself.

## Session and Instance

- a **session** represents one browser session
- an **instance** represents one live page

Multiple live pages can belong to the same session. They share the session boundary, but each one has its own instance boundary: its own render tree, handlers, subscriptions, and lifecycle.

## Path Model

A path model is a Go struct that describes a URL shape.

The current URL is always available as `Source[Location]`. When a route is easier to work with as a typed Go value, **Doors** can decode that location into a path model.

A path model describes:

- which page variants exist
- which path segments should be decoded
- which query parameters matter

The same model is also used for navigation. Links, redirects, and programmatic updates encode the struct back into a `Location`, so reading and changing the URL share one typed shape.

## Door

A **door** is a dynamic placeholder in the rendered tree. It is the mechanism **Doors** uses to update, replace, or remove part of the page without re-rendering everything.

You do not always manipulate doors directly, but many higher-level features are built on them. Reactive rendering, partial updates, and lifecycle-bound UI all depend on the dynamic door tree created during render.

## Hook

A **hook** is an HTTP handler on the server, created during render. It lives as long as the door that encloses it and answers only the session that rendered it.

Hooks are behind:

- [event handlers](./08-events.md): a button click, a form submit
- [private resources](./14-resources.md): a user's avatar, a report download
- [custom endpoints](./23-custom-endpoints.md) for scripts and embedded apps: a chart's data API, a live feed over a WebSocket

## State

**Doors** has built-in reactive state primitives:

- a `Source` is a writable reactive value, either original state or a derived view
- a `Beam` is a read-only value derived from state or observed from it

A `Source` is also a `Beam`, so you can read either one, subscribe to it, bind a part of the page to it, or match its value against patterns to pick which branch renders (routing). The current URL is a `Source` too.

The important user-facing behavior is consistency. During a render/update pass, a whole rendered branch will observe the same state.

One practical rule helps avoid many bugs: treat source values as immutable. If a source holds a slice, map, pointer, or mutable struct, replace it with a new value instead of mutating it in place.

A good default pattern is to keep identifiers and UI state in **Doors** state, then load the actual data when rendering or handling an event.

For example:

- keep `ProductID`, filters, pagination, and selection in sources
- derive smaller sources or beams from those values
- query backing data when producing output

This keeps live instances lightweight and avoids turning page memory into an accidental cache of large database records.

If data is only needed to produce output, render it and forget it.

## Context

In **Doors**, `context.Context` tells the **Doors** runtime where you are in the dynamic tree and which instance/session/lifecycle scope your code belongs to.

Use the `ctx` that **Doors** gives you in:

- event handlers
- beam subscriptions
- render-time helpers
- lifecycle-bound background work

Do not swap it for `context.Background()` when calling **Doors** APIs like beam reads, updates, hooks, links, or session/instance control. Those operations depend on the current **Doors** scope.

`ctx.Done()` is also meaningful here. It closes when the related subtree or lifecycle scope goes away, which makes it the right cleanup signal for work attached to rendered UI.

A render context may even be canceled mid-render, when the content being produced is already superseded. That is deliberate — it lets the runtime stop wasted work early. There is no need to check for cancellation in render code; it surfaces through the APIs the context is passed to — a database query, an HTTP request, `beam.Effect`. When such a call fails, return the error or nil; discarded output never reaches the page, so nothing is left torn.


```gox
<>
	~(doors.Go(func(ctx context.Context) {
		<-ctx.Done()
	}))
</>
```
> In request handlers, use the handler context with **Doors** APIs. Request values expose `Context()` for the underlying HTTP request context.

## Runtime

Rendering, state propagation, and hook execution are managed by the **Doors** runtime.

The rules:

- Load data where you render it: query the database or call an API right in the render.
- If that is too slow, render a skeleton and replace it with a [deferred operation](./06-door.md#deferred-operations).
- Most of the time, ignore the channels that Door methods and `doors.Call` return, and never wait on them in render or beam subscriptions: whether an update lands depends on the lifetime of its target, which is managed for you.
- Keep render work tied to producing the current page. Background loops, timers, pubsub listeners, and other work that should continue after rendering belong in their own goroutine.
- Independent slow fragments on the same page can render concurrently with [`doors.Parallel()`](./03-template-syntax.md#parallel).

If you do need the result, for backpressure in advanced setups or for data coming back from the browser, wait in a goroutine of your own with `doors.DetachedContext(ctx)`. Only a detached context delivers the result; on any other context the channel closes once the call is scheduled. Use `doors.Go(...)` when that work should follow the lifetime of a rendered subtree. See [Detached Context](./19-session-instance.md#detached-context), [Door](./06-door.md#completion-channels), and [Actions](./12-actions.md#direct).

To react after updates, use the callbacks instead of waiting: `doors.OnSettle` runs once the updates are queued for delivery, and `doors.OnBatch` applies a group of updates at once and runs once the browser has applied them. See [Door](./06-door.md#lifecycle-hooks).

## Security

**Doors** scopes handlers to the UI that produced them. A user can trigger only the handlers that were rendered for that user's live page instance, and only while the owning dynamic tree is still mounted.

**That means rendering is the main permission boundary for UI actions.** If a user is not allowed to delete a record, do not render the delete button for that user. The handler attached to that button cannot be triggered by a different user or by another page instance.

The URL is different. It is client-owned input: any path, query, or decoded path-model value is whatever the user sent. A successful path-model match means "this URL parses", not "this user is allowed to see what it points at".

In practice:

- check authentication before rendering protected UI
- check authorization while deciding which content and actions to render
- keep trust-bearing values in server-owned state, not in the route
- re-check permissions at the final mutation boundary if they can change independently, such as inside a database transaction

See [Storage & Auth](./18-storage-auth.md).

## DOM

**Doors** renders dynamic parts of the page into containers it manages. Your scripts can work alongside them:

- set attributes, classes, and styles
- show and hide elements
- mount islands controlled by JavaScript, like a React app

Do not move or remove the dynamic containers themselves: **Doors** updates them in place and expects to find them where it rendered them.
