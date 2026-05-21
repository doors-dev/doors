# Configuration

Most **Doors** apps can start with the defaults.

Reach for configuration when you need to change a few app-level things:

- session or instance lifetime and runtime limits
- Content Security Policy
- esbuild behavior for scripts, modules, and stylesheets
- the server ID used in **Doors** runtime URLs and session cookie naming
- error pages, session tracking, ...

All of these are passed to `doors.NewApp(...)` as app options:

```go
app := doors.NewApp(page,
	doors.WithConf(doors.Conf{
		RequestTimeout: 20 * time.Second,
		InstanceTTL:    30 * time.Minute,
	}),
	doors.WithCSP(doors.CSP{
		ConnectSources: []string{"https://api.example.com"},
	}),
	doors.ESProfile{
		Minify: true,
		JSX:    doors.JSXReact(),
	},
)
```

## Options

Common app options are:

- `doors.WithConf(...)` — runtime, session, and serving behavior
- `doors.WithCSP(...)` — Content Security Policy
- `doors.ESProfile{...}` — simple esbuild profile settings
- `doors.WithESProfiles(...)` — esbuild profile selection
- `doors.WithID(...)` — server ID for runtime URLs and cookie names
- `doors.WithSessionTracker(...)` — observe session create/delete
- `doors.WithErrorPage(...)` — custom error page

**Doors** fills in defaults automatically, so you usually set only the values you want to change.

## Server ID

Use `doors.WithID(...)` when this app should have its own **Doors** runtime URL prefix and session cookie namespace.
If unset, the ID defaults to `doors`.

```go
app := doors.NewApp(page, doors.WithID("blue"))
```

This value is used in two places:

- **Doors** runtime URLs are built under a prefix like `/~/blue/...`
- the **Doors** session cookie name becomes `blue` by default

That separation is especially useful when you run multiple **Doors** deployments side by side, for example:

- sticky load-balancing setups
- blue/green or canary rollouts
- migrations where old and new deployments should not steal each other's **Doors** session

The ID must already be URL-safe. If it needs escaping, **Doors** will panic during setup.

## System

Use `doors.WithConf(...)` for runtime and serving behavior. The value is `doors.Conf`.

```go
doors.WithConf(doors.Conf{
	SessionInstanceLimit: 6,
	RequestTimeout:       20 * time.Second,
	InstanceTTL:          30 * time.Minute,
})
```

The fields that matter most in practice are:

- `SessionInstanceLimit`: max live page instances per session. Default `12`. If exceeded, older inactive instances are suspended.
- `SessionTTL`: how long the session lives after activity. If unset or too small, **Doors** raises it to at least `InstanceTTL`.
- `InstanceConnectTimeout`: how long a new dynamic page can wait for its first client connection. Default `RequestTimeout`.
- `InstanceTTL`: how long an inactive instance is kept. Default `40m`, and never below `2 * RequestTimeout`.
- `InstanceGoroutineLimit`: max goroutines per page instance for runtime work. Default `8`.
- `DisconnectHiddenTimer`: how long hidden pages stay connected before disconnecting. Default `InstanceTTL / 2`.
- `RequestTimeout`: max duration of a client request or hook call. Default `30s`.
- `ServerCacheControl`: cache header for **Doors**-served JS and CSS resources. Default `public, max-age=31536000, immutable`.
- `ServerDisableGzip`: disables gzip for HTML, JS, and CSS.
- `ServerSessionCookiePrefix`: optional prefix for the internal **Doors** session cookie name. Empty by default, so with `doors.WithID("blue")` the cookie is named `blue`. Set it explicitly when you want browser-enforced cookie prefix rules such as `__Host-` or `__Secure-`.
- `ServerSessionCookieNoSecure`: omits the `Secure` attribute from the internal **Doors** session cookie. Use only for plain HTTP development.

The `Solitaire*` fields tune the sync transport between server and browser:

- `SolitaireSyncTimeout` limits how long a pending server-to-client sync task may wait. If it is exceeded, the instance is ended.
- `SolitaireQueue` and `SolitairePending` control queue depth and backpressure.
- `SolitairePing`, `SolitaireRollTimeout`, `SolitaireFlushSizeLimit`, and `SolitaireFlushTimeout` control how the sync connection is kept alive and flushed.
- `SolitaireDisableGzip` disables gzip for solitaire sync payloads without affecting HTML, JS, or CSS compression.

Most apps should leave the `Solitaire*` settings alone unless they are debugging runtime behavior or tuning under load.

## Logging

**Doors** uses Go's `log/slog` for internal diagnostics. By default, it writes through `slog.Default()`.

Use `doors.WithLogger(...)` when the app should route framework logs through your own logger:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

app := doors.NewApp(page, doors.WithLogger(logger))
```

Pass `nil` or omit the option to use `slog.Default()`.

## CSP

CSP is off until you call `doors.WithCSP(...)`.

```go
doors.WithCSP(doors.CSP{
	ConnectSources:      []string{"https://api.example.com"},
	ScriptStrictDynamic: true,
})
```

When enabled, **Doors** builds the `Content-Security-Policy` header per page and automatically collects hashes and sources from **Doors**-managed resources.

In practice, that means:

- `script-src` always includes `'self'` plus collected script hashes and sources
- `style-src` always includes `'self'` plus collected style hashes and sources
- `connect-src` always includes `'self'`

External script and style resources added through **Doors** also register their source automatically for CSP.

The field groups behave like this:

| Fields | `nil` | `[]` | values |
| --- | --- | --- | --- |
| `ScriptSources`, `StyleSources`, `ConnectSources` | keep only the **Doors** defaults | keep only the **Doors** defaults | append your values |
| `DefaultSources` | use the built-in default | omit the directive | emit your values |
| `FormActions`, `ObjectSources`, `FrameSources`, `FrameAcestors`, `BaseURIAllow` | default to `'none'` | omit the directive | emit your values |
| `ImgSources`, `FontSources`, `MediaSources`, `Sandbox`, `WorkerSources` | omit the directive | omit the directive | emit your values |

`ReportTo` only emits the `report-to` directive. You still need to send the matching `Report-To` response header yourself.

## Esbuild

**Doors** already has a default esbuild profile. The base profile targets `ES2022` and minifies output.

This is used for:

- the main **Doors** client bundle
- managed inline `<script>...</script>` resources
- buildable `<script src=(...)>` resources
- buildable `<link rel="stylesheet" href=(...)>` and `<style>...</style>` resources

Use `doors.ESProfile{...}` for simple esbuild settings:

```go
doors.ESProfile{
	Minify: true,
	JSX:    doors.JSXReact(),
}
```

Use `doors.WithESProfiles(...)` when you need named profiles or full esbuild options. The argument is `func(profile string) api.BuildOptions` — return the `esbuild` options that should apply to that named profile.

The `profile` value comes from the `profile` attribute on script resources. Your function must support the default profile `""`, because **Doors** uses it for its own main client build too.

One important rule: resource types still apply the entry-point and output settings they require, so your esbuild options can be supplemented or overridden to make that resource type work.

## Error Page

Use `doors.WithErrorPage(...)` to render your own page for internal runtime errors:

```go
doors.WithErrorPage(func(r *http.Request, err error) gox.Elem {
	return ErrorPage{Err: err}.Main()
})
```

The callback receives the original `*http.Request` and the runtime error.

## Session Tracker

Use `doors.WithSessionTracker(...)` to observe **Doors** session create/delete events:

```go
type tracker struct{}

func (tracker) Create(id string, r *http.Request) {
	// e.g. seed a server-side store
}

func (tracker) Delete(id string) {
	// e.g. clean up the server-side store
}

doors.WithSessionTracker(tracker{})
```

`Create` receives the new session ID and the request that triggered creation. The request must not be retained beyond the call, and its body must not be read.

## Rules

- Start with defaults and change only the settings you actually need.
- Use `WithConf` for lifetime, timeout, sync, and serving behavior.
- Turn on CSP with `WithCSP` when you want browser-enforced loading rules, then add only the extra sources your app really needs.
- Use `ESProfile` for simple esbuild settings.
- Use `WithESProfiles` only when one esbuild profile is not enough or the defaults don't fit.
