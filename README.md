[![codecov](https://codecov.io/gh/doors-dev/doors/branch/main/graph/badge.svg?token=6FOBJKNHFZ)](https://codecov.io/gh/doors-dev/doors)
[![Go Report Card](https://goreportcard.com/badge/github.com/doors-dev/doors)](https://goreportcard.com/report/github.com/doors-dev/doors)
[![Go Reference](https://pkg.go.dev/badge/github.com/doors-dev/doors.svg)](https://pkg.go.dev/github.com/doors-dev/doors)

[doors.dev](https://doors.dev)

# Doors

Doors is a server-driven UI runtime for building reactive web applications in Go.

In Doors, the server owns the interaction flow and the browser acts as a renderer and input layer. You build interactive UI in Go, keep state and capabilities on the server, and let the runtime synchronize updates back to the page.

## Example

```gox
type Search struct {
    input doors.Source[string] // reactive state
}

elem (s Search) Main() {
    <input
        (doors.AInput{
            On: func(ctx context.Context, r doors.RequestInput) bool {
                s.input.Update(ctx, r.Event().Value) // update state
                return false
            },
        })
        type="text"
        placeholder="search">

    ~(doors.Sub(s.input, s.results)) // subscribe results to state changes
}

elem (s Search) results(input string) {
    ~(for _, user := range Users.Search(input) {
        <card>
            ~(user.Name)
        </card>
    })
}
```

## What it includes

- Reactive web applications written in Go
- A stack with no public or hand-written API
- JavaScript as an option, not a requirement
- Asset serving and delivery that fit a single-binary deployment model
- Real-time capabilities out of the box
- A Go language extension with first-class HTML templating and its own language server

## Core model

Doors applications run on a stateful Go server. The browser acts as a remote renderer and input layer.

That means you can keep:

- event handling
- permissions
- business rules
- data access
- UI rendering

in one execution flow instead of splitting them between browser code, handlers, API contracts, and state reassembly.

### Your Go server is the UI runtime

In Doors, the web app runs as a stateful Go process. The browser acts as a remote renderer and input layer, while application flow, state, and side effects stay on the server.

### Less drift between UI and backend logic

Because interactions are handled in Go, you do not have to split core behavior across separate frontend and backend implementations. That reduces duplication and makes changes more consistent.

### Smaller exposed surface area

Doors does not turn every UI action into a public API endpoint. Session-scoped communication happens internally, so the client only gets rendered output and user-specific interaction paths.

### Designed as a cohesive stack

From template syntax and concurrency engine to synchronization protocol — each layer was designed together to **max out**.

## Key ideas

- `gox` for writing HTML-like UI directly as Go expressions
- Reactive state primitives that can be subscribed to, derived, and mutated
- Dynamic containers that can be updated, replaced, or removed at runtime
- Type-safe routing with URLs represented as Go structs
- Real-time client sync without making WebSockets or SSE your app architecture
- Secure by default — every user can interact only with what you render to them

## Where Doors fits best

- SaaS products
- Business systems
- Customer portals
- Admin panels
- Internal tools
- Real-time apps with meaningful server-side workflows

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

- [Officiean Website](https://doors.dev)
- [Documentation](https://doors.dev/docs/)
- [Tutorial](https://doors.dev/tutorial/)
- [API Reference](https://pkg.go.dev/github.com/doors-dev/doors)
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
