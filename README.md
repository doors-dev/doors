# Doors
[![codecov](https://codecov.io/gh/doors-dev/doors/branch/main/graph/badge.svg?token=6FOBJKNHFZ)](https://codecov.io/gh/doors-dev/doors)
[![Go Reference](https://pkg.go.dev/badge/github.com/doors-dev/doors.svg)](https://pkg.go.dev/github.com/doors-dev/doors)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#web-frameworks)

[https://doors.dev](https://doors.dev)

Doors is a server-driven UI framework + runtime for building stateful, reactive web applications in Go.

**Your Go server is a UI runtime**: the web application runs on a stateful server, while the browser acts as a remote renderer and input layer. No endpoints to design, no JSON contracts, no client bundle — `go build` produces the web application as one binary.

## Example

Templates are [GoX](https://github.com/doors-dev/gox) — a typed Go superset with its own parser, LSP, and editor plugins. Live search, complete:

```gox
type Search struct {
    db    *DB                  // your data layer
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
    users := s.db.SearchUsers(ctx, query) // hit the DB mid-render
    ~~
    <ul>
        ~(for _, u := range users {
            <li>~(u.Name)</li>
        })
    </ul>
}
```

This is the whole feature — no endpoint, no JSON contract, no fetch call, no client-side state, no JS build step. Debounce is one attribute, and the permission check happens where you render.

## How it works

- **A page is a live instance.** Each open tab holds its render tree, reactive state, and handlers as Go values on the server. A production page with 500+ interactive elements costs ~300 KB of server memory per open tab.
- **No virtual DOM.** Re-rendering is explicit and targeted: state derivation replaces diffing, and only fragments bound to changed values re-render.
- **Rendering is authorization.** Every user can interact only with what you rendered to them. You check permissions when you render the button, and that is enough to guarantee the related action won't be triggered by anyone else.
- **Go concurrency is the web architecture.** Handlers can block on channels, start goroutines scoped to rendered content, and receive cancellation through `ctx.Done()` when their subtree leaves the page. External systems integrate directly into the stateful session process — no wrapping stateful work into stateless logic.
- **The UI is awaitable.** Every mutating operation — source updates, door updates, calls into the browser, server-emitted DOM events — returns a plain `<-chan error`.
- **Typed routing.** URLs are Go structs, decoded from and encoded back to the address bar. Renaming a route field is a build error, not a 404.
- **Plain-HTTP sync.** A rolling-request and streaming protocol over regular HTTP — no WebSockets, no SSE, HTTP/3-ready.
- **Integrated web stack.** Bundle assets, build scripts, serve private files, automate CSP — the whole stack ships in the binary.
- **Stateful deployments, answered.** The [doors-caddy](https://github.com/doors-dev/doors-caddy) plugin provides seamless rollouts and sticky balancing for production clusters.

## Comparisons

### Doors vs HTMX, Datastar (+ templ)

HTMX and Datastar enhance HTML by coordinating behavior through attributes and endpoints; templ adds typed HTML rendering in Go on top. Interaction still flows through endpoints you design, wire up, and authorize. Doors is a UI runtime: you write the interaction flow directly in Go, handlers attach to elements, state lives on the live page instance, and the endpoint layer disappears.

### Doors vs Phoenix LiveView, Blazor Server

The same family: the server owns UI state. Doors does it in pure Go — no BEAM or .NET runtime, typed URL structs instead of string routes, plain-HTTP sync instead of WebSockets, one static binary. And the main thing: Doors UI is non-blocking. LiveView and Blazor serialize interactions through one process per client; Doors handles events concurrently, so interactivity behaves like a JavaScript SPA.

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
- Real-time apps with meaningful server-side workflows

## Where it is not the right fit

- Static or mostly non-interactive sites
- Client-first apps with minimal server behavior and simple routing
- Offline-first PWAs where the browser must be the primary runtime

The trade: server memory per open page (~300 KB for a heavy production page) and interaction latency bounded by the network round trip.

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
