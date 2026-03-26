# Core Concepts

**Doors** is easiest to understand if you stop thinking in terms of "request in, HTML out" and instead think of each interactive page as a live server-side object.

When a user opens a page, **Doors** decodes the URL into your path model, creates a page instance, renders HTML, and keeps that instance around as long as the page needs to stay interactive. Events, state changes, and partial updates all happen through that same live instance.

If your page is fully static, **Doors** can serve it and be done. As soon as you use dynamic features, the page becomes a long-lived part of the app.

## Mental Model

Most apps in **Doors** are built from a few ideas working together:

- the URL becomes a typed path model
- the path model and your page state drive rendering
- rendering creates a dynamic tree of updatable parts
- browser events route back to handlers on the same live page
- only the changed parts of the page are updated

That means routing, state, rendering, and event handling are not separate systems glued together later. They are different sides of the same runtime.

## Session

The most useful distinction to learn early is this:

- a **session** is usually the whole browser session
- an **instance** is usually one live page, often one tab

Multiple tabs usually share one session cookie, so they can also share session-level data such as authentication and permissions. Each tab still has its own page instance and its own local UI state.

This is a good default way to think about it:

- put login state, current user, and other browser-wide concerns at the session level, often as a `Source` stored in session storage
- put form state, selected rows, expanded panels, and other page-local concerns at the instance level

Useful lifecycle controls:

- `doors.SessionEnd(ctx)` force-ends the whole **Doors** session and all related instances
- `doors.SessionExpire(ctx, d)` changes the remaining session lifetime
- `doors.InstanceEnd(ctx)` ends only the current page instance

Changing the URL within the same model type usually updates the current instance. Switching to a different model type usually creates a different instance.

Instances are not meant to live forever. **Doors** can suspend older or less active instances based on configuration. When that happens, the page is restored by reloading it.

## Path Model

In **Doors**, routing starts from a struct, not a stringly-typed route table.

Your path model describes:

- which page variants exist
- which path segments should be decoded
- which query parameters matter

That gives you one typed value that can be used for matching, rendering, navigation, and redirects.

This is why the path model often becomes part of your page state instead of being treated as a separate concern. If the URL changes, your page can react to it the same way it reacts to any other state change.

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

- a `Source` is a piece of state you can update
- a `Beam` is a value derived from state or observed from it

In other words, a `Source` usually holds the state your page owns, and `Beam`s let the rest of the page react to it.

You usually keep small, durable UI values in sources: selected IDs, filters, search text, toggles, or the current path model. From there you derive whatever the page needs for rendering.

The important user-facing behavior is consistency. During a render/update pass, one part of the page does not see "old" state while another part sees "new" state halfway through the same update.

That is why `Read(ctx)` matters. It joins the coordinated render view. `Get()` is just an immediate read.

One practical rule helps avoid many bugs: treat source values as immutable. If a source holds a slice, map, pointer, or mutable struct, replace it with a new value instead of mutating it in place.

## Context

In **Doors**, `context.Context` is not just for cancellation. It also tells the framework where you are in the dynamic tree and which instance/session/lifecycle scope your code belongs to.

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

## Runtime

Rendering, event handling, and beam propagation happen on the framework runtime. That work should stay fast.

In practice, this means the code running in places like:

- page and component rendering
- beam subscriptions and reactive updates
- hook and event handlers

It is completely normal to query a database or call an API while rendering or handling an event. That is often where application data comes from.

The thing to avoid is blocking the runtime on work that is not really part of the current render or event flow, such as long-lived waits, background loops, timers, pubsub listeners, or waiting on completion channels from runtime-triggered work.

If work should continue independently of the current render/event flow, start your own goroutine or use `doors.Go(...)` when it should follow the lifetime of a rendered subtree.

## Data

A good default pattern is to keep identifiers and UI state in **Doors** state, then load the actual data when rendering or handling an event.

For example:

- keep `ProductID`, filters, pagination, and selection in sources
- derive smaller beams from those values
- query backing data when producing output

This keeps live instances lightweight and avoids turning page memory into an accidental cache of large database records.

If data is only needed to produce output, render it and forget it.

## Security

**Doors** gives you the right session and instance scope automatically, but your application still needs to enforce its own rules.

In practice:

- check authentication when choosing which page response to serve
- check authorization while rendering protected content
- keep a real server-side session store behind the cookie, and initialize shared auth state from it
- re-check write permissions where the actual mutation happens, especially at the database transaction level

Handlers already run inside the correct page/session context. That removes plumbing, not responsibility.

## DOM

When **Doors** renders a dynamic subtree, treat that subtree as framework-owned.

The safest integrations are the ones that cooperate with **Doors**:

- use **Doors** attributes, hooks, actions, and data channels
- keep custom JavaScript scoped to clear boundaries
- avoid manually mutating the same nodes that **Doors** is also updating

Direct DOM work is still possible, but it should complement the framework instead of racing against it.
