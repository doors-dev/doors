# Doors
[![codecov](https://codecov.io/gh/doors-dev/doors/branch/main/graph/badge.svg?token=6FOBJKNHFZ)](https://codecov.io/gh/doors-dev/doors)
[![Go Reference](https://pkg.go.dev/badge/github.com/doors-dev/doors.svg)](https://pkg.go.dev/github.com/doors-dev/doors)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#web-frameworks)

[https://doors.dev](https://doors.dev)

Doors is a server-driven UI framework for building stateful, reactive web applications in Go.

- **Server-driven reactivity.** Front-end framework capabilities in server-side Go. Build rich UIs without writing JS.
- **No public API layer.** Private transport is handled under the hood. No endpoint design required.
- **Unified control flow.** State, routing, and interaction flow stay in Go, with automatic real-time client sync.
- **Integrated web stack.** Bundle assets, build scripts, serve private files, and ship in one binary.

## Example

> Templates are [GoX](https://github.com/doors-dev/gox) — a typed Go superset with its own parser, LSP, and editor plugins.

Live search:

```gox
type Search struct {
    query doors.Source[string] // reactive state
}

elem (s *Search) Main() {
    <input
        (doors.AInput{
            Scope: &doors.ScopeDebounce{Duration: 300 * time.Millisecond}, // debounce requests
            On: func(ctx context.Context, r doors.RequestInput) bool {
                s.query.Update(ctx, r.Event().Value) // update state
                return false
            },
        })
        type="search"
        placeholder="search">

    ~(s.query.Bind(s.results)) ~// re-renders on query changes
}

elem (s *Search) results(query string) {
    ~~
    users := DB.SearchUsers(ctx, query) // hit the DB mid-render
    ~~
    <ul>
        ~(for _, u := range users {
            <li>~(u.Name)</li>
        })
    </ul>
}
```

This is the whole feature — no endpoint, no JSON contract, no fetch call, no client-side state, no JS build step. Debounce is one attribute, and the permission check happens where you render.

## Your Go server is a UI runtime

The web application runs on a stateful server, while the browser acts as a remote renderer and input layer.

- Build complete user flows within a single execution context
- Query databases and call external systems directly
- Ship real-time features without extra infrastructure
- Secure by default — every user can interact only with what you render to them

## Engineered from syntax to socket

From the template engine to the way bytes reach the browser — each layer was designed to push server-driven UI further than it's ever gone.

- Concurrency engine powering non-blocking event processing, parallel rendering, and tree-aware state propagation
- Real-time HTTP/3-ready synchronization protocol. No WebSockets/SSE
- Go language extension with its own LSP, parser, and editor plugins. HTML as Go expressions

## Built on better definitions

Composable primitives to use independently or combine into custom abstractions. Modern frontend framework capabilities without sacrificing clarity.

- Dynamic container in the DOM tree that can be updated, replaced, or removed at runtime
- Reactive state as a communication primitive, that can be passed around, subscribed to, derived and mutated
- URL as a Go struct and routing as state. Type-safe, deterministic, with unmatched freedom in route modeling

## The right environment for AI-assisted coding

A single Go execution flow removes frontend/backend drift, while server-side execution limits exposure of data, logic, and privileged operations.

- No frontend/backend split, so AI works in one Go codebase without duplicated logic or context handoff
- No API layer — no contracts to drift, no public endpoints to secure
- Only session-scoped render output is sent to the browser, so AI can work with data and server capabilities without exposing them to the client

## Comparisons

### Doors vs HTMX, Datastar (+ templ)

HTMX and Datastar enhance HTML by coordinating behavior through attributes and endpoints; templ adds typed HTML rendering in Go on top. Interaction still flows through endpoints you design, wire up, and authorize. Doors is a UI runtime: you write the interaction flow directly in Go, handlers attach to elements, state lives on the live page instance, and the endpoint layer disappears.

### Doors vs Phoenix LiveView, Blazor Server

The same model — think LiveView for Go: UI state lives on the server, and the browser shows what the server renders. The difference is in how it behaves under real conditions.

- **Concurrency.** LiveView and Blazor Server handle a page's events one at a time by default, so a slow action holds up the rest of that page until its work is moved into background tasks. In Doors, events are handled concurrently from the start, and the page keeps responding while a slow action runs.
- **Communication.** LiveView and Blazor Server run everything through one persistent WebSocket: every user action and every screen update travels over it. Doors works like a regular web app talking to an API: every user action is its own HTTP request, and screen updates arrive over a separate stream. If that stream is interrupted, the next action still goes through and brings it back right away, so the connection behind the page goes unnoticed. Over HTTP/2 and HTTP/3, actions and updates share one connection, and every action also confirms that it is healthy. No WebSockets or SSE, HTTP/3-ready.
- **Recovery.** When the connection drops, LiveView starts the page over, and anything not saved elsewhere is gone. Doors, like Blazor Server, keeps the page on the server for a configurable time and delivers the missed updates once the connection is back, without re-rendering the page.
- **Memory.** A connected Doors page with 500 interactive elements measures about 300 KB. User reports put an active LiveView page at 1–3 MB, shrinking to about 15–150 KB once idle. For Blazor Server, Microsoft's figure is about 250 KB for a minimal Hello World page, while a component vendor advises planning for 10–20 MB per user session.
- **HTTP features.** Each page comes with its own dynamic HTTP routing. Any handler, file, or generated content becomes a private URL right where it is rendered: a download, a preview image, an API for a React island, a live stream of server events. The URL works only for the user who received the page, and only while the content that created it is on the page — no routes to register, no access checks to write.

### Doors vs Livewire, Hotwire

Livewire and Hotwire are stateless: every interaction is a plain HTTP request you must authorize. Doors runs the UI as a stateful process — multi-stage flows and server-triggered updates are natural, and there is no public action surface to authorize: each handler exists only for the user it was rendered to.

### Doors vs React, Next.js, and similar stacks

Typical JavaScript stacks place the interaction model in the browser and use the server as a data service — every feature needs an API surface, client state, and synchronization between the two. Doors keeps the flow on the server in Go, with the browser focused on display and input — as if your API had exactly one consumer: the framework.

## Start a new project

For new applications, use [doors-dev/doors-starter](https://github.com/doors-dev/doors-starter), the recommended ready-to-run **Doors** project skeleton.

The [Get Started](https://doors.dev/docs/get-started/) guide walks through a minimal hello-world app by hand, useful for understanding the basic pieces.

For AI-assisted development, install the [Doors Skill](https://github.com/doors-dev/doors-skill) for Claude Code, Codex, or OpenCode.

## Where Doors fits best

- SaaS products
- Business systems
- Customer portals
- Admin panels
- Internal tools
- Real-time apps

Especially those with heavy user flows and complex business logic.

## Where it is not the right fit

- Static or mostly non-interactive sites
- Client-first apps with minimal server behavior and simple routing
- Offline-first PWAs where the browser must be the primary runtime

## Deployment

Doors is stateful, so load balancing requires sticky sessions. Doors includes the mechanics for uninterrupted rollouts: server IDs, a sticky cookie, and drain mode. The [doors-caddy](https://github.com/doors-dev/doors-caddy) plugin puts them to work with Caddy — use it, or build your own setup on the same basis.

## Status

Doors is used in production. Breaking changes are still happening, but they make life easier and are straightforward to migrate — each is documented in [RELEASE_NOTES.md](./RELEASE_NOTES.md).

## Learn more

- [Official Website](https://doors.dev)
- [Documentation](https://doors.dev/docs/)
- [Starter Project](https://github.com/doors-dev/doors-starter)
- [Tutorial](https://doors.dev/tutorial/)
- [API Reference](https://pkg.go.dev/github.com/doors-dev/doors)
- [GoX](https://github.com/doors-dev/gox)
- [AI Agent Skill](https://github.com/doors-dev/doors-skill) — install for Claude Code, Codex, or OpenCode to get Doors-aware code generation

## License

Doors is licensed under the [Apache License 2.0](./LICENSE).
