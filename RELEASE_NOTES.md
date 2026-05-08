# Doors `0.12` Release Notes

Doors `0.12` is a broad release focused on making the public API more direct: an app is now an `http.Handler`, routing is built on reactive state, and the old router/response layer has been removed.

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

Static files and request pre-processing now use app middleware:

```go
app.Use(
	doors.UseFS("/assets/", assetsFS, doors.CacheControlImmutable),
	doors.UseDir("/public/", "./public", doors.CacheControlStatic),
)
```

`UseFS`, `UseDir`, `UseFile`, and `UseResource` replace the old `RouteFS`, `RouteDir`, `RouteFile`, and `RouteResource` route types. Cache-control presets are exported for common static, HTML, API, private, and immutable responses.

See [App middleware](./docs/04-app.md#middleware).

### All page routing is reactive state

The current URL is now exposed as `doors.Source[doors.Location]`. `doors.RouterSource(...)` and `doors.RouterBeam(...)` are shortcuts over that source, and path-model routes are declared inside the page component:

```gox
~(doors.RouterSource(
	doors.RouteModelSource(func(p doors.Source[Path]) gox.Comp {
		return Page{path: p}
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
- `RouteBeam` and `RouteSource` branch UI on any reactive value, not only URLs.

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
Indicator: doors.IndicateClass("loading").
	And(doors.IndicateAttrQuery("#spinner", "aria-busy", "true"))
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

### Request and response cleanup

Page factories now return `gox.Comp` directly. The old `doors.Response*` types are gone.

The old common request interface was renamed to `doors.RequestCommon`, while the new `doors.Request` is the page-function request type for cookies, headers, and response headers.

HTTP-level redirects now belong in middleware before the app handler. In-instance reroutes are normal state updates through the location source, such as `doors.Router(ctx)` or a `Source[Path]` received from `RouteModelSource`.

### Door method names

Deprecated `Door` methods were removed in favor of names that describe what changes:

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
- rename removed deprecated `Door`, beam, GoX cursor, and location helpers
