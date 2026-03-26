# Core Concepts

Doors keeps a live server-side representation of interactive pages.

Routing, rendering, state, and event handling all work inside that same system.

## Loop

The usual flow is:

- Doors receives the request and decodes the URL into a path model
- Doors creates the page instance for that page and renders it
- the render produces static HTML plus any dynamic nodes
- if the page is dynamic, the browser keeps a sync connection to that instance
- user events are routed back to handlers on that instance
- handlers, beams, and doors update only the parts of the page that changed

A fully static page does not need to stay live after the initial response. Once you use dynamic features, the page becomes a live instance.

## Session

A session is the browser-level scope.

It groups related page instances under one session cookie.

In practice, that means:

- multiple tabs usually share one session
- session-scoped data and permissions can be shared across those tabs
- `doors.SessionEnd(ctx)` kills the whole session and all of its instances
- `doors.SessionExpire(ctx, d)` shortens the session lifetime from the current moment

## Instance

An instance is one live page, usually one tab.

It holds the current page state for that tab, including:

- the current path model source
- the dynamic door tree
- hook registrations
- instance-local runtime state

Hooks and other dynamic bindings live under that tree. If the subtree that created them is replaced or unmounted, they end with it.

Navigation within the same path model usually updates the current instance. Navigation to a different path model creates a different page instance.

Instances are also disposable. A session can keep several live instances, but older or less active ones may be suspended by configuration. When a suspended page becomes active again, the client reloads it.

Use `doors.InstanceEnd(ctx)` when you want to end only the current page instance.

## Context

Context is part of the Doors programming model.

Use the `ctx` that Doors gives you in:

- event handlers
- beam subscriptions
- render-time helpers such as `doors.Go`

That context carries the current dynamic tree position, instance identity, hook registration space, and render consistency state.

Do not use `context.Background()` for Doors operations such as:

- `door.Update`
- beam `Read`
- `doors.A(...)`
- session or instance control

Use the closest framework context in scope. The current subtree matters.

`ctx.Done()` is also meaningful here: it closes when the related dynamic subtree or lifecycle scope goes away, so it is the right cleanup signal for background work tied to rendered UI.

```gox
~(doors.Go(func(ctx context.Context) {
	<-ctx.Done()
}))
```

## Runtime

Rendering, beam propagation, and dynamic updates run on the framework runtime.

Keep that runtime work short and non-blocking:

- render content
- derive state
- handle events
- schedule DOM updates

Do not block render functions or beam subscriptions waiting for slow external events or for other runtime-dependent work to finish.

If work can take a long time or wait on channels, timers, pubsub, or network results:

- start a goroutine yourself, or
- use `doors.Go(...)` when the work should follow a rendered subtree lifetime

If you need to wait on `X*` completion channels from code that started with a framework context, wait outside the runtime work itself.

When you intentionally carry that framework context into your own blocking goroutine, wrap it with `doors.AllowBlocking(ctx)`.

## State

Doors state starts from a `Source` and branches into derived `Beam`s.

The important guarantee is consistency: during one render/update pass, a door subtree reads one coherent beam state. A parent and its children do not observe different values halfway through the same render.

That is why `Read(ctx)` matters. It participates in the render-consistency model. `Get()` is just an immediate value read and does not join that coordinated render view.

Another important rule: treat reference values as immutable for propagation purposes. If you keep maps, slices, pointers, or mutable structs in a source, replace them with a new value instead of mutating them in place.

## Security

Protected pages should enforce access both when the page is served and when data is changed.

In practice:

- check authentication when serving the page model
- call `doors.SessionEnd(ctx)` on logout
- check authorization during render
- keep write permission checks at the database transaction level when permissions can change dynamically

Event handlers already run inside the correct session and instance scope, so they do not need to rebuild that routing layer themselves. But they still must enforce the application rules for the action they perform.

## Data

Prefer storing small identifiers in fields or state, then loading fresh data when you render or handle an event.

That usually means:

- keep IDs, filters, selection, and UI state in sources
- derive smaller beams from that state
- query backing data when producing output

Avoid filling long-lived fields or beams with large database rows unless you intentionally want that caching behavior. Doors keeps page instances in memory, so state should stay deliberate.

If data is only needed to produce output, render it and forget it.

## DOM

Doors owns the dynamic DOM it renders.

That means manual JavaScript should cooperate with the framework, not fight it:

- prefer Doors attributes, hooks, actions, and data channels for integration
- treat Door-managed subtrees as framework-owned
- avoid ad-hoc mutation of nodes that Doors is also updating

Direct client-side DOM work is still possible, but it should have clear boundaries.
