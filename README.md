# Doors
[![codecov](https://codecov.io/gh/doors-dev/doors/branch/main/graph/badge.svg?token=6FOBJKNHFZ)](https://codecov.io/gh/doors-dev/doors)
[![Go Report Card](https://goreportcard.com/badge/github.com/doors-dev/doors)](https://goreportcard.com/report/github.com/doors-dev/doors)
[![Go Reference](https://pkg.go.dev/badge/github.com/doors-dev/doors.svg)](https://pkg.go.dev/github.com/doors-dev/doors)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

[https://doors.dev](https://doors.dev)

Doors is a server-driven UI framework + runtime for building stateful, reactive web applications in Go.

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

    ~(s.input.Bind(s.results)) // bind results to state changes
}

elem (s Search) results(input string) {
    ~(for _, user := range Users.Search(input) {
        <card>
            ~(user.Name)
        </card>
    })
}
```

## Some highlights

- **Front-end framework capabilities in server-side Go.** Reactive state primitives, dynamic routing, composable components.
- **No public API layer.** No endpoint design needed, private temporal transport is handled under the hood.
- **Unified control flow.** No context switch between back-end/front-end.
- **Integrated web stack.** Bundle assets, build scripts, serve private files, automate CSP, and ship in one binary.

## Execution model

Go server is UI runtime: web application runs on a stateful server, while the browser acts as a remote renderer and input layer.

## Mental model

Link DOM to the data it depends on.

## Peculiarities

- Purposely build Go language extension with its own LSP, parser, and editor plugins. Adds HTML as Go expressions and `elem` primitives.
- Reactive state primitives that can be subscribed to, derived, and mutated.
- Dynamic containers that can be updated, replaced, or removed at runtime.
- Type-safe routing with URLs represented as Go structs.
- HTTP/3-ready synchronization protocol (rolling-request + streaming, events via regular post, no WebSockets/SSE).
- Custom concurrency engine that enables non-blocking event processing, parallel rendering, and tree-aware state propagation.
- Secure by default: every user can interact only with what you render to them. Means you check permissions when your render the button and that's is enough to be sure that related action wont be triggered by anyone else.

## Where Doors fits best

- SaaS products
- Business systems
- Customer portals
- Admin panels
- Internal tools
- Real-time apps with meaningful server-side workflows

## Where it is not the right fit

- Static or mostly non-interactive sites
- Client-first apps with minimal server behavior and simple routing
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

Doors is in beta. It is ready for development and can be used in production with caution, but you should expect fixes and updates as the ecosystem matures.

## Licensing

Doors is licensed under the Apache License 2.0.

See also:

- [Apache License 2.0](./LICENSE)
