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
