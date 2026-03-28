# Doors

[![Coverage](https://codecov.io/gh/doors-dev/doors/branch/main/graph/badge.svg)](https://codecov.io/gh/doors-dev/doors)

Doors is a server-driven UI runtime for building reactive web applications in Go.

In Doors, the server owns the interaction flow and the browser acts as a renderer and input layer. You build interactive UI in Go, keep state and capabilities on the server, and let the runtime synchronize updates back to the page.

## What it includes

- Reactive web applications written in Go
- UI interactions without first designing a public API
- JavaScript as an option, not a requirement
- Built-in real-time synchronization
- Asset serving and delivery that fit a single-binary deployment model
- One place for business logic, state changes, and rendering

## Core model

Doors applications run on a stateful Go server. The browser acts as a remote renderer and input layer.

That means you can keep:

- event handling
- permissions
- business rules
- data access
- UI rendering

in one execution flow instead of splitting them between browser code, handlers, API contracts, and state reassembly.

## Why it is useful

### Your Go server is the UI runtime

Doors is not an HTML enhancement layer. It is a runtime for interactive applications where the server owns the flow and the browser reflects current UI state.

### No frontend/backend drift by default

Because the interaction model stays in Go, there is less duplicated logic and less drift between client and server behavior.

### Less exposed surface area

Doors does not push every UI action into a public API. Session-scoped communication happens under the hood, which cuts boilerplate and narrows the exposed surface area.

### One stack, top to bottom

Templates, state, routing, synchronization, assets, and deployment all follow the same server-driven model.

## Key ideas

- `gox` for writing HTML-like UI directly as Go expressions
- Reactive state primitives that can be subscribed to, derived, and mutated
- Dynamic containers that can be updated, replaced, or removed at runtime
- Type-safe routing with URLs represented as Go structs
- Real-time client sync without making WebSockets or SSE your app architecture
- Server-rendered interactions scoped to what each user can actually see

## Example

```gox
type Search struct {
    input doors.Source[string]
}

elem (s Search) Main() {
    <input
        (doors.AInput{
            On: func(ctx context.Context, r doors.RequestInput) bool {
                s.input.Update(ctx, r.Event().Value)
                return false
            },
        })
        type="text"
        placeholder="search">

    ~(doors.Sub(s.input, s.results))
}

elem (s Search) results(input string) {
    ~(for _, user := range Users.Search(input) {
        <card>
            ~(user.Name)
        </card>
    })
}
```

The point is to keep the UI event, state mutation, rendering logic, and server-side behavior in the same language and runtime.

## Where Doors fits best

- SaaS products
- Business systems
- Customer portals
- Admin panels
- Internal tools
- Realtime apps with meaningful server-side workflows

## Where it is not the right fit

- Static or mostly non-interactive sites
- Client-first apps with minimal server behavior
- Offline-first PWAs where the browser must be the primary runtime

## Comparisons

### Doors vs HTMX

HTMX enhances HTML by coordinating behavior through attributes and endpoints. Doors is a UI runtime: you write the interaction flow directly in Go and let the runtime handle synchronization.

### Doors vs React, Next.js, and similar stacks

Typical JavaScript stacks place much of the interaction model in the browser while the server acts mainly as a data service. Doors keeps that flow on the server in Go, with the browser focused on display and input.

## Learn more

- [doors.dev](https://doors.dev)
- [Tutorial](https://doors.dev/tutorial/)
- [Documentation](https://doors.dev/docs/)
- [API Reference](https://docs.doors.dev)
- [GoX](https://github.com/doors-dev/gox)

## Status

Doors is in beta. It is ready for development and can be used in production with care, but the ecosystem is still maturing. Expect fixes, refinements, and some breaking changes as it evolves.

## Licensing

Doors is dual-licensed by **doors dev LLC**.

- **Open-source use:** available under **AGPL-3.0-only**
- **Commercial use:** required for proprietary / closed-source use or other use that cannot comply with AGPL-3.0-only
- Commercial licensing details: [doors.dev/license](https://doors.dev/license)

Commercial licenses are issued through an Order Form and a separate commercial agreement. The signed agreement controls.

Contact: [sales@doors.dev](mailto:sales@doors.dev)

See also:

- [AGPL-3.0 License](./LICENSE)
- [Licensing Terms](./LICENSING.md)
- [Commercial License Summary](./LICENSE-COMMERCIAL)
