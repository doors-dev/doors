# Core Concepts

**Doors** runs each interactive page as a live server-side instance, with the current URL exposed as a reactive state value.

Event handlers and dynamic fragments registered while rendering belong to that instance and are cleaned up when their part of the rendered tree goes away.

> If your page is fully static, **Doors** can serve it and be done. As soon as you use dynamic features, the page becomes a long-lived part of the app.

## Mental Model

Most apps in **Doors** are built from a few ideas working together:

- the current URL is a reactive source
- routing can turn that source into typed path models
- route values and your page state drive rendering
- rendering creates a dynamic tree of updatable parts
- registered browser events call handlers on the same live page
- handlers update state or dynamic DOM containers
- the runtime synchronizes those changes with the client

## Session

The most useful distinction to learn early is this:

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

## Doors And Hooks

A **door** is a dynamic placeholder in the rendered tree. It is the mechanism **Doors** uses to update, replace, or remove part of the page without re-rendering everything.

You do not always manipulate doors directly, but many higher-level features are built on them. Reactive rendering, partial updates, and lifecycle-bound UI all depend on the dynamic door tree created during render.

A **hook** is a server handler bound to rendered UI. When the user clicks, types, submits, or when JavaScript calls `$hook(...)`, **Doors** routes that event back to the live page instance that created the hook.

The practical rule is:

- if a subtree disappears, the hooks and dynamic bindings created inside it disappear too
- if the subtree is rendered again, new bindings are created for the new tree

That keeps behavior aligned with what is actually on screen.

## State

**Doors** has built-in reactive state primitives:

- a `Source` is a writable reactive value, either original state or a derived view
- a `Beam` is a read-only value derived from state or observed from it

A `Source` is also a `Beam`, so you can read, subscribe, and route from both writable and read-only values. The current URL is a `Source[Location]` — that is why `doors.Route(...)` with `doors.RouteModel(...)` gives the matched view a typed `Source` you can write back to.

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


```gox
<>
	~(doors.Go(func(ctx context.Context) {
		<-ctx.Done()
	}))
</>
```
> In request handlers, use the handler context with **Doors** APIs. Request values expose `Context()` for the underlying HTTP request context.

## Runtime

Rendering and state propagation happens on the **Doors** runtime.

It is completely normal to query a database or call an API while rendering.

Content inside dynamic containers (`doors.Door`) is rendered on the instance
goroutine pool by default, so separate dynamic fragments can make progress in
parallel.

For ordinary render flow, wrap independent slow fragments with
`doors.Parallel()` so they can render concurrently:

```gox
<>
	~>(doors.Parallel()) <>
		~{
			user := loadUser(ctx)
		}
		<section>~(user.Name)</section>
	</>
</>
```

Keep render work tied to producing the current page. For background loops,
timers, pubsub listeners, or other work that should continue after rendering,
start your own goroutine or use `doors.Go(...)` when it should follow the
lifetime of a rendered subtree.

## Security

**Doors** scopes handlers to the UI that produced them. A user can trigger only the handlers that were rendered for that user's live page instance, and only while the owning dynamic tree is still mounted.

That means rendering is the main permission boundary for UI actions. If a user is not allowed to delete a record, do not render the delete button for that user. The handler attached to that button cannot be triggered by a different user or by another page instance.

The URL is different. It is client-owned input: any path, query, or decoded path-model value is whatever the user sent. A successful path-model match means "this URL parses", not "this user is allowed to see what it points at".

In practice:

- check authentication before rendering protected UI
- check authorization while deciding which content and actions to render
- keep trust-bearing values in server-owned state, not in the route
- re-check permissions at the final mutation boundary if they can change independently, such as inside a database transaction

See [Storage & Auth](./18-storage-auth.md).

## DOM

When **Doors** renders a dynamic subtree, treat that subtree as runtime-managed. Direct DOM work is still possible, but it should complement the runtime instead of racing against it.
