# Doors `0.15` Release Notes — "Cadence"

Doors `0.15` rebuilds the event and lifecycle machinery: server-driven synthetic events, precise lifecycle hooks, one unified action API, optimized memory usage, and the move to GoX v0.3.0.

> Most of the changes of this release are based on development and production feedback from a social platform with live feeds, comments, notifications, and moderation plus paid professional services. The services lean on what is unique to Doors: external systems integrate directly into the stateful session process — no API endpoints, no wrapping of stateful work into stateless logic.
>
> The other driver of this release is preparation for native agentic integration — stay tuned.

## Highlights

### Emitter — synthetic events from the server

`doors.Emitter` is a zero-value attribute handle that dispatches synthetic DOM events — pointer, keyboard, focus, input, change, submit — to its attached elements from server code. One emitter can attach to several elements, and one element can carry several emitters. Emitted events bubble, so event attrs on ancestor elements run too.

Each event method returns `ActionInto[int]`; `Into` captures how many hook requests the emitted events triggered (any failure fails the call).

```go
var e doors.Emitter
// template: <button (&e) (doors.AClick{On: onClick})>Save</button>

var n int
err := <-doors.Call(ctx, e.Click(doors.PointerEmit{}).Into(&n))
// n = hook requests the emitted events triggered
```

Emit structs (`PointerEmit`, `KeyboardEmit`, `FocusEmit`, `InputEmit`, `ChangeEmit`, `SubmitEmit`) carry only fields that round-trip through dispatch and capture.

Related: the `On` callback on all event, form, and hook attrs is now optional. With nil `On` the request is still accepted (its body discarded unread) and the hook stays registered, so an attr can be attached purely for its client-side effects or as an Emitter round-trip target.

See [Element Handles](./docs/17-element-handles.md).

### Lifecycle hooks: OnReady, OnSettle, OnClean

- `doors.OnReady(ctx, on)` fires at most once when the render cycle that produced the current content completes and its page/update is enqueued for the client. Best-effort: dropped if the cycle fails or is superseded. Use it to start work that must not outrun the markup it targets — e.g. launch the goroutine that streams live updates into freshly rendered content only once that content is on its way to the client.
- `doors.OnClean(ctx, f)` fires exactly once when the content is cleared, on any teardown path. On replacement, old content's `OnClean` runs before new content's `OnReady`.
- `doors.OnSettle(ctx, on, ops...)` runs ops within the current dispatch batch and fires once when everything the batch started — door updates, beam propagation — is processed and enqueued. In a handler, the batch spans the whole handler. Two practical uses: a reliable point to collect information gathered during render, and controlled batched UI updates — issuing the next batch from the `on` callback guarantees it lands after the previous one.

`doors.HoldSettle(ctx)` keeps the current dispatch batch open after the handler returns: `OnSettle` callbacks, indicators, scopes, and the `$hook` promise wait until the returned release func is called. Use it when a handler hands work to another goroutine or subsystem — loading indicators and the client's hook promise stay pending until the work actually finishes, not until the handler returns.

Callbacks execute inline on the goroutine firing the frame (or the calling goroutine when the frame already fired); the must-not-block rule is unchanged.

`doors.Go` now starts its function only after the surrounding render cycle is enqueued for delivery, so Door updates made inside always land after their host markup.

See [Door](./docs/06-door.md).

### Unified Call and completion channels

`doors.Call(ctx, action)` returns a plain `<-chan error`: nil on success, an error on failure, closed without a value on cancel. Client results are captured by arming the action with `Into` instead of a generic call variant.

```go
var res string
ch := doors.Call(ctx, doors.ActionEmit[string]{Name: "toast", Arg: msg}.Into(&res))
// res is valid once ch delivers nil
```

`$on(...)` action handlers may now return a `Promise`: the client awaits it, the action settles when it does, and the resolved value is delivered to `Into` — the old `async actions are prohibited` error is gone.

In the same spirit, all mutating operations return their completion channel directly: door operations, `doors.Reload`, and `Source` updates now return `<-chan error`, optional to use; the `X*` variants are gone — see [Breaking Changes](#breaking-changes). The `Source` completion contract is refined: nil means propagated, `context.Canceled` means superseded by a newer update, closed without a value means suppressed (equal value or no subscribers).

### Setter — stateless attribute control

`doors.Setter` replaces the removed `AShared`: a zero-value attribute handle whose `Set(name, value)` returns an action that sets the attribute on every attached live element.

```go
var locked doors.Setter
// attach: <button (&locked)>Save</button> <button (&locked)>Publish</button>

doors.Call(ctx, locked.Set("disabled", true))
doors.Call(ctx, locked.Set("hidden", nil)) // removes the attribute
```

Values follow template attribute semantics (nil/false remove, true sets bare). Setter is stateless: a rerendered element returns to its template attributes. `Set(...).Into(&n)` captures the number of live elements reached.

### Door: Freeze and flexible Outer

- `Door.Freeze(ctx)` keeps the Door's current markup on the page while releasing hooks, subscriptions, and nested Doors on the server. Meant for feed-like content that goes final; the Door keeps its stored state and can be mounted again.
- `Door.Outer` accepts `any` renderable content instead of only `gox.Elem` — comps, strings, slices all work; nil leaves an empty live container, and a typed-nil `Elem` no longer panics.

### Context and sessions

`doors.Ctx(ctx)` propagates user `context.WithValue` values through the render subtree — visible in nested renders, event handlers, and later door updates — while cancelation, deadlines, and Doors ownership stay with the enclosing render:

```gox
~>(doors.Ctx(context.WithValue(ctx, themeKey{}, "dark"))) <>
    // subtree sees ctx.Value(themeKey{})
</>
```

- `doors.HasSession` / `doors.HasInstance` report which Doors API level a context supports.
- `Beam.Sub`, `Read`, `ReadAndSub`, and `Watch` now work outside an instance (e.g. on `SessionContext` or a background context); such subscriptions attach directly to the source and end when their context is canceled.
- `doors.Logger(ctx)` returns the configured `*slog.Logger`, falling back to `slog.Default`.
- `doors.IDNumber(ctx)` returns an instance-unique `uint64`.
- `WithSessionTracker` accumulates: repeating the option installs several observers, run in registration order.
- Sessions are created lazily on first actual touch — requests that never use the session (static resources, crawlers) allocate no session state. Cookies and expiry semantics are unchanged.

### App and platform

- `doors.WithPrinter(func(next gox.Printer) gox.Printer)` wraps the HTML output printer once per drain unit, after all framework transforms. See [Printer Middleware](./docs/22-printer-middleware.md).
- `AHook`, `ARawHook`, `ASubmit`, and `ARawSubmit` gain `RequestTimeout time.Duration`, overriding `Conf.RequestTimeout` per hook.
- `ActionLocationRawReplace{URL}` replaces the current history entry with a literal URL, complementing `ActionLocationRawAssign`.
- Exported error sentinels for `errors.Is`: `ErrPathModel`, `ErrPathEncode`, `ErrExecution` (action reached the browser and failed there), `ErrTerminated` (instance ended first). Hook registration on a released door now yields `context.Canceled`.
- `Location` marshals to JSON with `segments`/`query` tags; nil encodes as `[]`/`{}` instead of null.
- The `doors.css` resource and its head `<link>` are removed: the `d0-r` rule is applied via a constructed stylesheet in the blocking head script, so it holds at first paint and a strict `style-src` CSP no longer needs the doors resource origin.

### Lower per-instance memory

Registered hooks no longer pin their attr structs: trigger closures capture only the handler and body limit, so serialized attribute config (scopes, indicators, actions) is garbage-collectable right after render. On top of that, pooled gzip writers are returned on finalize instead of held until client ack.

Real-life data from a heavy page with 500+ interactive elements: per-instance memory dropped from ~800 KB to ~300 KB.

### GoX v0.3.0

Doors now requires GoX v0.3.0, which merges `Editor`/`EditorComp` into `Comp`/`Elem`. `Door` is a plain `gox.Comp`; template usage `~(&doors.Door{})` is unchanged. Direct-render and signature changes are listed under [Breaking Changes](#breaking-changes).

### Documentation

Godocs across the public API were rewritten as contracts. `docs/17-shared-attr.md` became [docs/17-element-handles.md](./docs/17-element-handles.md), covering Setter and Emitter together. `ALink` docs no longer claim a nil `OnError` defaults to `ActionLocationReload`; a failed navigation reverts to the previous history entry.

## Fixes

- `doors.A` on a door's container element no longer overwrites the door's parent marker, which broke the door's next update on the client.
- `Setter.Set` rejects `gox.Mutate` values (such as `doors.Class`) instead of silently replacing the attribute; plain values like `Set("class", "hl")` remain the supported form.
- The client aborts its open long-poll and report stream when the connector pauses (pagehide, hidden-tab disconnect), freeing server connections immediately and unblocking back/forward cache.
- Closed an unlocked-read race in door container tracking.

## Breaking Changes

### GoX v0.3.0 (Editor merged into Comp)

| Old | New |
|---|---|
| `github.com/doors-dev/gox` v0.2.3 | `github.com/doors-dev/gox` v0.3.0 |
| `Door.Edit(cur gox.Cursor) error` | `Door.Main() gox.Elem` (or `cur.Comp(door)`) |
| `Beam.Bind(...) gox.EditorComp` | `Beam.Bind(...) gox.Elem` |
| `Beam.RouteBeam(...) gox.EditorComp` | `Beam.RouteBeam(...) gox.Elem` |
| `Source.Route(...) gox.EditorComp` | `Source.Route(...) gox.Elem` |
| `doors.Route(...) gox.EditorComp` | `doors.Route(...) gox.Elem` |
| `doors.Go(f) gox.Editor` | `doors.Go(f) gox.Elem` |
| `doors.Status(code) gox.Editor` | `doors.Status(code) gox.Elem` |

```go
// old
return door.Edit(cur)
// new
return cur.Comp(door)
```

### X* variants merged into base methods

All return `<-chan error`; ignore the channel for fire-and-forget.

| Old | New |
|---|---|
| `Door.XInner` / `XOuter` / `XStatic` / `XReload` / `XUnmount` | `Door.Inner` / `Outer` / `Static` / `Reload` / `Unmount` |
| `doors.XReload` | `doors.Reload` |
| `Source.XUpdate` / `Source.XMutate` | `Source.Update` / `Source.Mutate` |

### Call rework

| Old | New |
|---|---|
| `XCall[T](ctx, action) <-chan CallResult[T]` | `Call(ctx, action.Into(&dst)) <-chan error` |
| `Call(ctx, action)` (no return) | `Call(ctx, action) <-chan error` |
| `CallResult[T]` | removed |
| `ActionEmit{Name, Arg}` | `ActionEmit[T]{Name, Arg}` (`ActionEmit[any]` to ignore the result) |

### Renames

| Old | New |
|---|---|
| `doors.InstanceId` | `doors.InstanceID` |
| `doors.SessionId` | `doors.SessionID` |

### Removed APIs

| Old | New |
|---|---|
| `AShared` / `NewAShared` | `doors.Setter` |
| `AKeyDown.Filter` / `AKeyUp.Filter` | `AKeyDown.Keys` / `AKeyUp.Keys` |
| `Free` | `DetachedContext` |
| `FreeRoot` | `InstanceContext` |
| `Source.RouteSource` | `Source.Route` |
| `RouteLocationDefault{,Beam,Bind,Comp}` | `RouteDefault{,Beam,Bind,Comp}` |

## Migration

```sh
go get github.com/doors-dev/doors@v0.15.0-rc2
```

Update GoX to v0.3.0 alongside, then apply the renames in the tables above — all mechanical. The only behavioral shifts to review are the completion-channel contracts on `Call`, door operations, and `Source` updates.

---

# Doors `0.14` Release Notes — "Altitude"

Doors `0.14` is a focused ergonomics release. The changes are small in surface area, but they make daily Doors code feel more coherent for developers and users.

## Highlights

### GoX syntax coherence

Doors now uses GoX v0.2 syntax.

**Go snippets** used `~{ ... }` to switch from template mode into plain Go statements. The curly-brace delimiters were easy to confuse with regular Go blocks when scanning templates quickly, especially nested ones, and implied a scope boundary that doesn't actually exist.

```gox
<>
    ~// Old
    ~{
        user := db.Get(id)
        items := user.Items
    }

    ~// New
    ~~
    user := db.Get(id)
    items := user.Items
    ~~
</>
```

The new `~~ ... ~~` delimiters are visually distinct and don't create a scope that isn't actually there.

**Inline expressions** used `~func { ... }` for multi-statement blocks that evaluate at render time and return a value into the template. The form looked like a Go function literal but wasn't one - it was a GoX-specific construct.

```gox
    <>
    ~// Old
    ~func {
        user, err := db.Get(id)
        if err != nil { return <span>error</span> }
        return Card(user)
    }

    ~// New
    ~({
        user, err := db.Get(id)
        if err != nil { return <span>error</span> }
        return Card(user)
    })
</>
```

The new form is just a placeholder `~(...)` with `{ ... }` inside. In attribute values, the `~` is dropped as usual:

```gox
<input checked=({ return ok })>
```

Old syntax is still supported and compiles without issues. To migrate existing templates, make sure you have the latest GoX toolchain. The GoX LSP formatter and `gox fmt` both auto-convert the old syntax.

See [Template Syntax](./docs/03-template-syntax.md).

### Path model segments

Path models now support prefix tags: put the shared URL segment or prefix in the struct-tag key, then list variants under it in the tag value. This removes the repeated path segments that used to make local route groups noisy:

```go
// Old
type ArticlePath struct {
	Section ArticleSection `path:"/article/:ID | /article/:ID/edit | /article/:ID/settings"`
	ID      int
}

// New
type ArticlePath struct {
	Section ArticleSection `"/article/:ID":" | edit | settings"`
	ID      int
}
```

Variants describe only what changes inside that area, while the shared prefix lives in the tag key. This makes path models much easier to split by concern and keeps large applications from accumulating one oversized route model.

Old `path:"..."` tags and bool marker fields remain supported as compatibility syntax.

See [Routing](./docs/05-routing.md).

### Zero-delay reconnect

When the network drops, the sync client backs off with progressive reconnect delays, up to the maximum retry delay. Previously, if the network recovered while the browser was still waiting inside a long retry delay, the next user action could still appear delayed until the scheduled reconnect fired.

Now, user activity resets the pending sync reconnect delays and triggers an immediate reconnect attempt. After a long network interruption, a click, input, or navigation no longer waits behind an old backoff timer.

This closes the last visible gap where Doors' server-driven sync model could surface to the user after network recovery.

### Navigation error handling

Failed Doors link navigation no longer reloads the page by default.

Previously, a failed navigation behaved more like a traditional PHP-style server-rendered app: if the network was lost, the browser could leave the current page and show its own connection error page. That made sense as a conservative SSR fallback, but it is not how users expect a modern web app to behave.

Now there is no default error action. You can still specify your own `OnError` actions for fallback UI, notifications, or recovery behavior. If a Doors link optimistically changes the URL and the navigation hook fails, the URL change is reverted instead of turning into a full browser-level failure page.

## Other improvements

### ⚡ Claude Fable codebase audit

Doors completed a Claude Fable audit covering the entire codebase, including the reactive state layer, Door engine, instance/session HTTP handling, routing/path models, printing/resources, TypeScript client, and public hook APIs. The audit used parallel subsystem reviews, then each finding was re-verified by hand against the actual source.

All findings were fixed. The follow-up work tightened sync and screen cleanup, path decoding/encoding edge cases, form and hook error handling, resource build failures, request-body limits, client recovery paths, and several race/leak-prone lifecycle edges. Regression coverage was added around the fixes so these cases stay closed.

### Ecosystem level-up: official Caddy plugin

[`doors-caddy`](https://github.com/doors-dev/doors-caddy) is the official Caddy plugin and Go integration library for production Doors deployments. It solves three stack-level problems:

- **Rollouts without user interruption.** For stateful server-driven apps, rollouts are usually painful because live sessions must stay connected to the process that owns them. Doors no longer has that pain: Kubernetes + Caddy plugin + Argo Rollouts can provide seamless rollouts. The user keeps a complete, coherent app version throughout the live session, while normal navigation moves them to the fresh deployment.
- **Sticky load balancing.** A Doors instance is live on one server. The balancer is responsible for initial assignment; after that, the user sticks to that server until the session ends. 
- **Geo-based redirects.** UX responsiveness depends on server location. The plugin can redirect users to the closest regional domain by country, while leaving Doors system paths and non-GET requests untouched.

---

# Doors `0.13` Release Notes — "Harmony"

## Highlights

### Performance control

`SolitaireFrameTime` now controls the entire sync cycle — both server-to-client flushes and client-to-server reports. That enables precise performance tuning.

The default is ~33ms (30 FPS). Lower values increase UI responsiveness; higher values reduce syscall frequency and Go scheduler load.

### Streaming reports

Chromium-based browsers can send client reports over persistent HTTP streams instead of individual POST requests, reducing connection overhead.

This is enabled by default when the browser and connection support streaming request bodies. Unsupported browsers fall back to the previous behavior automatically. Some deployment infrastructure, including some reverse-proxy setups, does not handle streaming request bodies correctly; set `SolitaireDisableReportStreaming` after checking the full production path.

### Solitaire engine rewrite

The sync engine (client and server) has been restructured: the connection handler is split into dedicated sender and receiver components, with a new frame-based write controller and adaptive RTT estimation.

## Details

**Separated push and pull.** Server-to-client sync and client-to-server reporting now run over independent connections, each with its own lifecycle. This avoids unnecessary connection churn and keeps the coupling between directions intentional rather than accidental.

**Explicit framing.** Server-to-client data is now organized into deliberate frames. The server controls when a frame goes out and the client responds to complete frames, so behavior no longer depends on incidental network fragmentation or system buffer behavior — making the flow more controllable.

**Report loss detection.** In the previous design, the server's response to a client report served as implicit delivery confirmation. With separated connections, that guarantee is gone. The protocol now carries explicit acknowledgments so the client can detect and retransmit dropped reports.

**Adaptive timing.** The server continuously estimates round-trip time from live traffic and uses it to pace recovery probes.

**WebTransport-ready.** The new implementation is transport-agnostic — ready for WebTransport support when the standard becomes more common, with no structural changes needed.

---

# Doors `0.12` Release Notes — "Romance"

Doors `0.12` is a broad release focused on making the public API more direct: `App` replaced `Router`, page routing is built entirely on reactive state, writable derived views expand the state toolkit, and several verbose APIs were tightened into a cleaner shape.

This release has migration-impacting changes for applications written against `0.8.x`-`0.10.x`. Start with the [migration guide](./MIGRATION.md) when updating an existing codebase.

## Highlights

### App entry point

`doors.NewApp(...)` replaces the old router setup. The page function receives a Doors runtime context plus `doors.Request`, then returns the root `gox.Comp` directly:

```go
app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
	return App{}
})

http.ListenAndServe(":8080", app)
```

The returned app is a regular `http.Handler`, so it can be passed to `net/http`, mounted in another mux, or wrapped with middleware.

See [App](./docs/04-app.md).

### Middleware and static files

Static files and page request pre-processing now use app middleware:

```go
app.Use(
	doors.UseFS("/assets/", assetsFS, doors.CacheControlImmutable),
	doors.UseDir("/public/", "./public", doors.CacheControlStatic),
)
```

`UseFS`, `UseDir`, `UseFile`, and `UseResource` replace the old `RouteFS`, `RouteDir`, `RouteFile`, and `RouteResource` route types. Cache-control presets are exported for common static, HTML, API, private, and immutable responses.

See [App middleware](./docs/04-app.md#middleware).

### All page routing is reactive state

The current URL is now exposed as `doors.Source[doors.Location]`, and path-model routes are declared inside the page component with `doors.Route(...)`:

```gox
~(doors.Route(
	doors.RouteModel(elem(p doors.Source[Path]) {
		~Page{path: p}
	}),
	doors.RouteLocationDefaultComp(NotFound{}),
))
```

The matched route receives a live source or beam. Updating the source updates the browser URL and reroutes the page instance without a full reload.

Path models still support the existing bool-field variant style. `0.12` also adds compact typed `int` variants with `|`-separated path patterns.

See [Routing](./docs/05-routing.md) and [Navigation](./docs/09-navigation.md).

### State routing and derived state

`Source[T]` and `Beam[T]` now share a richer state model:

- `DeriveBeam` / `DeriveBeamEqual` create read-only derived views.
- `DeriveSource` / `DeriveSourceEqual` create writable derived views over a parent source.
- `Source.Route` and `Beam.RouteBeam` branch UI on any reactive value, not only URLs.

This makes route switching, tab panels, feature gates, and nested state views use the same primitive.

See [State](./docs/07-state.md).

### Joinable indicators, scopes, actions, and query matchers

Hook options that used to accept slices now accept a single joinable interface value:

- `doors.Indicators`
- `doors.Scopes`
- `doors.Actions`
- `doors.QueryMatcher`

Single helpers can be passed directly, and multiple values compose with `.And(...)` or the `Join*` helpers:

```go
Indicator: doors.IndicateClass("loading").And(doors.IndicateAttrQuery("#spinner", "aria-busy", "true"))
```

This removes the old `Only` helper families and makes single-value cases less noisy.

See [Indication](./docs/11-indication.md), [Scopes](./docs/10-scopes.md), [Actions](./docs/12-actions.md), and [Navigation active links](./docs/09-navigation.md#active).

### App configuration

Configuration now lives on `NewApp` options:

```go
app := doors.NewApp(page,
	doors.WithConf(doors.Conf{RequestTimeout: 20 * time.Second}),
	doors.WithCSP(doors.CSP{ConnectSources: []string{"https://api.example.com"}}),
	doors.ESProfile{Minify: true, JSX: doors.JSXReact()},
)
```

Use `doors.ESProfile{...}` for simple esbuild settings and `doors.WithESProfiles(...)` when named profiles or full esbuild options are needed.

See [Configuration](./docs/21-configuration.md) and [JavaScript](./docs/15-javascript.md).

### Door method names

The low-level `Door` API was renamed into a more concise and expressive system, making direct Door manipulation cleaner:

| Old | New |
| --- | --- |
| `Update` | `Inner` |
| `Rebase` | `Outer` |
| `Replace` | `Static` |
| `Delete` | `Static(ctx, nil)` |
| `Clear` | `Inner(ctx, nil)` |

The `X*` completion variants follow the same names: `XInner`, `XOuter`, `XStatic`, `XReload`, and `XUnmount`.

See [Door](./docs/06-door.md).

## Breaking Changes

The following old router APIs were removed:

- `doors.NewRouter`
- `doors.UseModel`
- `doors.UseRoute`
- `doors.UseFallback`
- `doors.UseSystemConf`
- `doors.UseCSP`
- `doors.UseESConf`
- `doors.UseServerID`
- `doors.UseSessionCallback`
- `doors.UseErrorPage`

The following old types and helpers were also removed or renamed:

- `doors.RequestModel`
- `doors.Response`, `doors.ResponseComp`, `doors.ResponseRedirect`, `doors.ResponseReroute`
- `doors.RouteFS`, `doors.RouteDir`, `doors.RouteFile`, `doors.RouteResource`
- `doors.SystemConf` -> `doors.Conf`
- `doors.ESOptions` / `doors.ESConf` -> `doors.ESProfile` or `doors.WithESProfiles`
- `doors.NewBeam` / `doors.NewBeamEqual` -> `doors.DeriveBeam` / `doors.DeriveBeamEqual`
- `doors.Sub(...)` -> `beam.Bind(...)`
- `doors.Inject(...)` -> `beam.Effect(ctx)`
- `doors.NewLocation(ctx, model)` -> `doors.NewLocation(model)`
- `doors.IndicatorOnly*`, `doors.ScopeOnly*`, `doors.ActionOnly*`, and `doors.QueryMatcherOnly*`

## Migration

Upgrade the module first:

```sh
go get github.com/doors-dev/doors@latest
```

Then follow the [migration guide](./MIGRATION.md). It includes mechanical rewrite tables, worked examples, and grep checks for finding old API usage after the first pass.

The highest-impact migration areas are:

- move from router registration to `doors.NewApp(...)` plus routes rendered inside the page component
- replace static route declarations with `app.Use(...)` middleware
- change response-returning handlers into components that return `gox.Comp`
- move HTTP redirects to middleware and in-instance reroutes to location-source updates
- update request helper signatures from `doors.Request` to `doors.RequestCommon` outside the page factory
- replace slice fields for indicators, scopes, actions, and query matchers with joinable values
- rename removed older `Door`, beam, and location helpers
