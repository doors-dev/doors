# Migration Guide: `0.8.x`-`0.10.x` → `0.12.X`

Sections are grouped by the *shape* of the change — each starts with prose, then a mapping table, then a worked example where it helps.

## Contents

1. [App Entry Point](#1-app-entry-point) — `NewRouter` / `UseModel` → `NewApp` + page function + routes inside the page
2. [App Configuration](#2-app-configuration) — `Use*Conf` → `With*` options
3. [Request Types](#3-request-types) — `RequestModel` / `Response*` are gone; old `Request` → `RequestCommon`
4. [Path Source Wiring](#4-path-source-wiring) — same `Source[Path]` type, different entry point
5. [Joiner Shape](#5-joiner-shape-indicators-scopes-actions-querymatcher) — `[]Indicator` / `[]Scope` / `[]Action` / `[]QueryMatcher` → single interface values composed with `.And(...)`
6. [Path Model](#6-path-model) — unchanged for `bool` variants
7. [Static Files](#7-static-files-old-route-types--middleware) — `RouteFS` / `RouteDir` / `RouteFile` / `RouteResource` → `app.Use(...)` middleware
8. [Removed Older APIs](#8-removed-older-apis) — `doors.Sub` / `doors.Inject` / old `Door` method names
9. [Other Removed](#9-other-removed) — full kill list
10. [Sanity Checks](#quick-sanity-checks-for-migrated-code) — grep recipes for verifying a migrated tree

## Before You Start

Bump the `github.com/doors-dev/doors` dependency first. The new API (`NewApp`, `Route`, etc.) is not in older published versions.

```bash
go get github.com/doors-dev/doors@latest
```

Without this, the rest of the migration won't compile no matter how carefully you rewrite the call sites.

---

## 1. App Entry Point

The router went away. **Doors** apps now have one page function that returns the root component, plus middleware via `app.Use(...)`. URL-to-page dispatch is no longer registration on the router — it happens inside the page component using `Route(...)` and route builders.

| Old | New |
| --- | --- |
| `doors.NewRouter()` | `doors.NewApp(page, opts...)` |
| `doors.UseModel(router, h)` | inside the page component: `doors.Route(doors.RouteModel(render), …)` |
| `doors.UseRoute(router, route)` / `RouteFS` / `RouteDir` / `RouteFile` / `RouteResource` | `app.Use(doors.UseFS(…))` / `UseDir` / `UseFile` / `UseResource` |
| `doors.UseFallback(router, h)` | mount the app inside a parent mux, or use `app.Use(...)` middleware |

**Old:**
```go
type Path struct {
    Home bool `path:"/"`
}

func main() {
    r := doors.NewRouter()

    doors.UseModel(r, func(req doors.RequestModel, s doors.Source[Path]) doors.Response {
        return doors.ResponseComp(App{path: s})
    })

    doors.UseRoute(r, doors.RouteFS{Prefix: "/assets", FS: assetsFS})

    http.ListenAndServe(":8080", r)
}
```

**New:**
```go
func main() {
    app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
        return App{}
    })

    app.Use(
        doors.UseFS("/assets/", assetsFS, doors.CacheControlImmutable),
    )

    http.ListenAndServe(":8080", app)
}
```

Routing happens **inside** the page component:

```gox
elem (a App) Main() {
    <html>
        <body>
            ~(doors.Route(
                doors.RouteModel(elem(p doors.Source[Path]) {
                    ~Page{path: p}
                }),
                doors.RouteLocationDefaultComp(NotFound{}),
            ))
        </body>
    </html>
}
```

For each old `UseModel(...)` registration, add a route branch such as `RouteModel` inside one `Route(...)` call. The renderer receives the same `doors.Source[Path]` it used to — only the wiring moves.

### Per-request setup before rendering

What used to live at the top of a `UseModel` handler now goes into the page function. Session and instance stores are reached from the Doors context (`doors.SessionStore(ctx)` / `doors.InstanceStore(ctx)`).

**Old:**
```go
doors.UseModel(r, func(req doors.RequestModel, s doors.Source[Path]) doors.Response {
    state := req.SessionStore().Init(sessionStateKey{}, func() any { … }).(doors.Source[SessionState])
    return doors.ResponseComp(App{state: state})
})
```

**New:**
```go
app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
    state := doors.SessionStore(ctx).Init(sessionStateKey{}, func() any { … }).(doors.Source[SessionState])
    return App{state: state}
})
```

### Redirects and reroutes

`ResponseRedirect` / `ResponseReroute` are gone.

- HTTP-level redirects → middleware via `app.Use(...)`.
- Reroute → change the URL inside a page function or live instance via locations source (`doors.Router(ctx)` returns `doors.Source[doors.Location]`)

---

## 2. App Configuration

Router-level setters (`Use*Conf`, `UseCSP`, `UseSessionCallback`, etc.) are gone. Configuration is now passed as `doors.With...` options to `doors.NewApp(...)`.

| Old | New |
| --- | --- |
| `doors.UseSystemConf(router, doors.SystemConf{…})` | `doors.WithConf(doors.Conf{…})` |
| `doors.UseCSP(router, …)` | `doors.WithCSP(…)` |
| `doors.UseESConf(router, doors.ESOptions{…})` | `doors.WithESProfiles(func(profile string) api.BuildOptions {…})` *or* the convenience `doors.ESProfile{…}` (see below) |
| `doors.UseServerID(router, "id")` | `doors.WithID("id")` |
| `doors.UseSessionCallback(router, cb)` | `doors.WithSessionTracker(t)` (interface methods are now `Create` / `Delete`) |
| `doors.UseErrorPage(router, ep)` | `doors.WithErrorPage(ep)` |
| `doors.SystemConf` (type) | `doors.Conf` |

**Old:**
```go
r := doors.NewRouter()
doors.UseSystemConf(r, doors.SystemConf{RequestTimeout: 20 * time.Second})
doors.UseCSP(r, doors.CSP{ConnectSources: []string{"https://api.example.com"}})
```

**New:**
```go
app := doors.NewApp(page,
    doors.WithConf(doors.Conf{RequestTimeout: 20 * time.Second}),
    doors.WithCSP(doors.CSP{ConnectSources: []string{"https://api.example.com"}}),
)
```

### Esbuild options

`doors.ESOptions` / `doors.ESConf` are gone. Two replacements:

- **`doors.ESProfile{…}`** — drop-in struct option (passed directly to `NewApp`, no wrapper). `JSX`, `External`, `Minify` fields. Closest 1:1 replacement for `ESOptions`.
- **`doors.WithESProfiles(func(profile string) api.BuildOptions)`** — full control via a profile function. Use when you need named profiles or per-profile build options. Your function must handle the default profile `""`.

`doors.JSX`, `doors.JSXReact()`, and `doors.JSXPreact()` are unchanged.

```go
// Simple
doors.ESProfile{Minify: true, JSX: doors.JSXReact()}

// Custom
doors.WithESProfiles(func(profile string) api.BuildOptions {
    return api.BuildOptions{Target: api.ES2022, MinifySyntax: true /* … */}
})
```

`SessionTracker` interface for `WithSessionTracker`:

```go
type SessionTracker interface {
    Create(id string, r *http.Request)
    Delete(id string)
}
```

---

## 3. Request Types

The interface that used to be called `Request` was renamed `RequestCommon`. The new `Request` is specifically the page-function parameter (cookies/headers + `RequestHeader`/`ResponseHeader`). `RequestModel` and the `Response*` machinery are gone — page functions just return `gox.Comp`.

| Old | New |
| --- | --- |
| `doors.RequestModel` | `doors.Request` (page-function param); inside dynamic code use `doors.SessionStore(ctx)` |
| `doors.Request` (inside helpers, event/form/hook handlers) | `doors.RequestCommon` |
| `doors.Response`, `doors.ResponseComp`, `doors.ResponseRedirect`, `doors.ResponseReroute` | the page function returns `gox.Comp` directly |
| `RequestRawForm.W()` / `RequestRawHook.W()` | `.ResponseWriter()` |
| `r.After([]doors.Action{...})` | `r.After(actions)` (a single `doors.Actions` value — see §5) |

A helper that previously took `doors.Request` from a caller should take `doors.RequestCommon` now. `doors.Request` only fits the page-function signature.

**Old:**
```go
func InitFromRequest(store doors.Store, r doors.Request) { … }
```

**New:**
```go
func InitFromRequest(store doors.Store, r doors.RequestCommon) { … }
```

---

## 4. Path Source Wiring

`doors.Source[Path]` is still the type a route renderer receives — that didn't change. What changed is *where* the renderer receives it: as a parameter to a `RouteModel(render)` callback inside the page component, instead of as an argument to a `UseModel` handler.

| Old | New |
| --- | --- |
| `func(req doors.RequestModel, s doors.Source[Path]) doors.Response` (handler) | `func(s doors.Source[Path]) gox.Elem` (route renderer for `RouteModel`); or `func(s doors.Beam[Path]) gox.Elem` for `RouteModelBeam` |
| `doors.NewLocation(ctx, model)` | `doors.NewLocation(model)` (no `ctx`) |

Helper signatures and struct fields holding the path source don't need to change — `doors.Source[Path]` is still the right type. Only the entry-point shape moved.

If a helper truly only reads, narrow to `doors.Beam[Path]` to make the intent explicit.

### Beam Derivation Renamed

`doors.NewBeam` and `doors.NewBeamEqual` were renamed; same shape, same behavior, just clearer naming for the "derive a smaller view from a parent" operation.

| Old | New |
| --- | --- |
| `doors.NewBeam(source, get)` | `doors.DeriveBeam(source, get)` |
| `doors.NewBeamEqual(source, get, equal)` | `doors.DeriveBeamEqual(source, get, equal)` |

`doors.DeriveSource(...)` / `doors.DeriveSourceEqual(...)` are new — they build a writable view over a parent source. See [State](./docs/07-state.md) for the full picture.

---

## 5. Joiner Shape: Indicators, Scopes, Actions, QueryMatcher

This is the biggest mechanical pattern. Four kinds of fields/values that used to be slices are now single interface values that compose with `.And(...)` (or `doors.JoinX(...)`):

- `[]doors.Indicator` → `doors.Indicators`
- `[]doors.Scope` → `doors.Scopes`
- `[]doors.Action` → `doors.Actions`
- `Active.QueryMatcher []QueryMatcher` → `Active.QueryMatcher QueryMatcher`

The struct types themselves (`IndicatorClass`, `ScopeBlocking`, `ActionScroll`, etc.) all implement the corresponding interface, so a single value passes directly. Combine with `.And(...)`:

```go
// Old
Indicator: []doors.Indicator{a, b}
Scope:     []doors.Scope{x, y}
Before:    []doors.Action{p, q}
QueryMatcher: []doors.QueryMatcher{m, n}

// New
Indicator: a.And(b)                 // or doors.JoinIndicators(a, b)
Scope:     x.And(y)                 // or doors.JoinScopes(x, y)
Before:    p.And(q)                 // or doors.JoinActions(p, q)
QueryMatcher: m.And(n)              // or doors.Join(m, n)
```

**Helpers that *return* these need their return type updated too** — from a slice to the interface:

```go
// Old
func myIndicators() []doors.Indicator { return []doors.Indicator{a, b} }

// New
func myIndicators() doors.Indicators { return a.And(b) }
```

Same shape for scopes, actions, and query matchers.

### `XxxOnly*` Helpers Removed

The old `XxxOnly*` helpers existed solely to wrap a single value in a one-element slice. They're gone. Use the regular helper / struct directly — it's already an `Indicators` / `Scopes` / `Actions` / `QueryMatcher`.

| Old | New |
| --- | --- |
| `doors.IndicatorOnlyContent(...)` (and `Attr`, `Class`, `ClassRemove` + `Query`/`QueryAll`/`QueryParent` suffixes) | `doors.IndicateContent(...)` (same suffixes); or use struct types like `doors.IndicatorContent{Selector: …, Content: …}` directly |
| `doors.ScopeOnlyBlocking()` / `Serial` / `Latest` | `&doors.ScopeBlocking{}` / `&doors.ScopeSerial{}` / `&doors.ScopeLatest{}` |
| `doors.ScopeOnlyDebounce(d, l)` | `&doors.ScopeDebounce{Duration: d, Limit: l}` (was a method, now struct fields) |
| `doors.ActionOnlyEmit(...)` / `LocationReload` / `LocationReplace` / `LocationAssign` / `LocationRawAssign` / `Scroll` / `Indicate` | use the struct literal directly: `doors.ActionEmit{Name: …, Arg: …}`, `doors.ActionLocationReload{}`, `doors.ActionScroll{Selector: …}`, etc. |
| `doors.QueryMatcherOnlyIgnoreSome(...)` / `OnlyIgnoreAll` / `OnlySome` / `OnlyIfPresent` | the regular matchers themselves (`doors.QueryMatcherIgnoreSome(...)` etc.) — chain with `.And(doors.QueryMatcherIgnoreAll())` to reproduce the *Only* variant |

`Active.QueryMatcher` example:

```go
// Old
Active: doors.Active{
    QueryMatcher: doors.QueryMatcherOnlyIgnoreSome("page"),
}

// New
Active: doors.Active{
    QueryMatcher: doors.QueryMatcherIgnoreSome("page").And(doors.QueryMatcherIgnoreAll()),
}
```

---

## 6. Path Model

Path tags are unchanged. **No code change required for existing models with `bool` fields** — the old form compiles as-is.

A new compact form is also available (one typed `int` field with `|`-separated patterns) but it's purely additive — not part of the migration. See [Routing](./docs/05-routing.md#variants) if you want to adopt it later.

---

## 7. Static Files (Old `Route*` types → middleware)

`doors.UseRoute` and the `Route*` types are gone. Use `app.Use(...)` middleware:

| Old | New |
| --- | --- |
| `doors.UseRoute(r, doors.RouteFS{Prefix: "/assets", FS: fsys})` | `app.Use(doors.UseFS("/assets/", fsys, doors.CacheControlImmutable))` |
| `doors.UseRoute(r, doors.RouteDir{Prefix: "/public", DirPath: "./public"})` | `app.Use(doors.UseDir("/public/", "./public", doors.CacheControlStatic))` |
| `doors.UseRoute(r, doors.RouteFile{Path: "/robots.txt", FilePath: "./static/robots.txt"})` | `app.Use(doors.UseFile("/robots.txt", "./static/robots.txt", doors.CacheControlStatic))` |
| `doors.UseRoute(r, doors.RouteResource{Path: "assets/sans.ttf", Resource: doors.ResourceFS(fs, "sans.ttf")})` | `app.Use(doors.UseResource("/assets/sans.ttf", doors.ResourceFS(fs, "sans.ttf"), ""))` |

`Cache-Control` constants are exported on the package: `doors.CacheControlImmutable`, `doors.CacheControlStatic`, `doors.CacheControlStaticShort`, `doors.CacheControlHTML`, `doors.CacheControlCDN`, `doors.CacheControlPrivate`, `doors.CacheControlNoCache`, `doors.CacheControlNoStore`, `doors.CacheControlAPI`. Pass `""` for no header.

---

## 8. Removed Older APIs

These older helper APIs have been deleted.

### Free helpers

| Removed | Replacement |
| --- | --- |
| `doors.Sub(beam, el)` | `beam.Bind(el)` |
| `doors.Inject(key, beam)` | `beam.Effect(ctx)` inside a dynamic subtree |

Mechanical rewrite — same callback shape:

```go
// Old
doors.Sub(counter, func(v int) gox.Elem { … })

// New
counter.Bind(func(v int) gox.Elem { … })
```

In `.gox` files, the GoX `elem(...) { … }` shorthand is also accepted:

```gox
~(counter.Bind(elem(v int) {
    <span>~(v)</span>
}))
```

### Door methods

| Removed | Replacement |
| --- | --- |
| `door.Update(ctx, content)` | `door.Inner(ctx, content)` |
| `door.Rebase(ctx, el)` | `door.Outer(ctx, el)` |
| `door.Replace(ctx, content)` | `door.Static(ctx, content)` |
| `door.Delete(ctx)` | `door.Static(ctx, nil)` |
| `door.Clear(ctx)` | `door.Inner(ctx, nil)` |
| `door.XUpdate(ctx, content)` | `door.XInner(ctx, content)` |
| `door.XRebase(ctx, el)` | `door.XOuter(ctx, el)` |
| `door.XReplace(ctx, content)` | `door.XStatic(ctx, content)` |
| `door.XDelete(ctx)` | `door.XStatic(ctx, nil)` |
| `door.XClear(ctx)` | `door.XInner(ctx, nil)` |

---

## 9. Other Removed

- `doors.NewRouter`, `doors.UseModel`, `doors.UseRoute`, `doors.UseFallback`, `doors.UseSystemConf`, `doors.UseCSP`, `doors.UseESConf`, `doors.UseServerID`, `doors.UseSessionCallback`, `doors.UseErrorPage`
- `doors.RouteFS`, `doors.RouteDir`, `doors.RouteFile`, `doors.RouteResource` (use `app.Use(...)` middleware equivalents)
- `doors.SystemConf` (renamed to `doors.Conf`)
- `doors.ESOptions`, `doors.ESConf`
- `doors.RequestModel`
- `doors.Response`, `doors.ResponseComp`, `doors.ResponseRedirect`, `doors.ResponseReroute`
- `doors.IndicatorOnly*`, `doors.ScopeOnly*`, `doors.ActionOnly*`, `doors.QueryMatcherOnly*` helpers

(`doors.JSX`, `doors.JSXReact`, `doors.JSXPreact` are unchanged.)

---

## Quick Sanity Checks for Migrated Code

After a mechanical rewrite, run these greps from the project root. Anything that returns hits is something to look at.

**Should return nothing:**

```sh
grep -rnE 'doors\.(NewRouter|UseModel|UseRoute|UseFallback|UseSystemConf|UseCSP|UseESConf|UseServerID|UseSessionCallback|UseErrorPage)\b' .
grep -rnE 'doors\.(RequestModel|Response|ResponseComp|ResponseRedirect|ResponseReroute|RouteFS|RouteDir|RouteFile|RouteResource|SystemConf|ESOptions|ESConf)\b' .
grep -rnE 'doors\.(IndicatorOnly|ScopeOnly|ActionOnly|QueryMatcherOnly)' .
grep -rnE '\[\]doors\.(Scope|Indicator|Action|QueryMatcher)\b' .   # slice → joiner interface
grep -rn  'doors\.NewLocation(ctx'                              .   # signature dropped ctx
grep -rnE 'doors\.(Sub|Inject)\('                              .
grep -rnE 'doors\.NewBeam(Equal)?\b'                            .   # renamed to DeriveBeam / DeriveBeamEqual
```

**Needs human review** (greps that catch the renamed method but also legitimate code):

```sh
# Old Door methods. Source.Update(ctx, ...) is also legitimate — distinguish by receiver.
grep -rnE '\.(Update|Rebase|Replace|Delete|Clear|XUpdate|XRebase|XReplace|XDelete|XClear)\(ctx' .

# RequestRawForm / RequestRawHook .W() → .ResponseWriter()
grep -rnE '\.W\(\)' .

# Outside the doors.NewApp page-function signature, request params should be doors.RequestCommon.
grep -rnE 'doors\.Request\b' .
```

**Spot-check by hand:**

- Page factories return `gox.Comp` directly (no `doors.Response` / `ResponseComp`).
- `Indicator`, `Scope`, `Before`, `After`, `OnError`, and `Active.QueryMatcher` fields receive a single value, not a slice.
- Helpers that *return* indicators / scopes / actions / query matchers return the interface type (`doors.Indicators`, etc.), not a slice.
