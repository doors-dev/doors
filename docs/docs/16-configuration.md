# Configuration

Most **Doors** apps can start with the router defaults.

Reach for configuration when you need to change a few router-level things:

- session or instance lifetime and runtime limits
- Content Security Policy
- esbuild behavior for scripts and modules
- the server ID used in **Doors** system URLs and session cookie naming

All of these are router-level settings:

```go
r := doors.NewRouter()

doors.UseSystemConf(r, doors.SystemConf{
	RequestTimeout: 20 * time.Second,
	InstanceTTL:    30 * time.Minute,
})

doors.UseCSP(r, doors.CSP{
	ConnectSources: []string{"https://api.example.com"},
})

doors.UseESConf(r, doors.ESOptions{
	JSX:    doors.JSXReact(),
	Minify: true,
})
```

## Start

Configuration is applied to the router with:

- `doors.UseSystemConf(...)`
- `doors.UseCSP(...)`
- `doors.UseESConf(...)`
- `doors.UseServerID(...)`
- `doors.UseSessionCallback(...)`
- `doors.UseErrorPage(...)`
- `doors.UseLicense(...)`

**Doors** fills in defaults automatically, so you usually set only the values you want to change.

## Server ID

Use `doors.UseServerID(...)` when this router should have its own framework URL prefix and session cookie namespace.

```go
doors.UseServerID(r, "blue")
```

This value is used in two places:

- framework system URLs are built under a prefix like `/~/blue/...`
- the **Doors** session cookie name becomes `d0rblue`

That separation is especially useful when you run multiple **Doors** deployments side by side, for example:

- sticky load-balancing setups
- blue/green or canary rollouts
- migrations where old and new deployments should not steal each other's framework session

The ID must already be URL-safe. If it needs escaping, **Doors** will panic during setup.

## System

Use `doors.UseSystemConf(...)` for runtime and serving behavior.

```go
doors.UseSystemConf(r, doors.SystemConf{
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
- `InstanceGoroutineLimit`: max goroutines per page instance for runtime work. Default `16`.
- `DisconnectHiddenTimer`: how long hidden pages stay connected before disconnecting. Default `InstanceTTL / 2`.
- `RequestTimeout`: max duration of a client request or hook call. Default `30s`.
- `ServerCacheControl`: cache header for framework-served JS and CSS resources. Default `public, max-age=31536000, immutable`.
- `ServerDisableGzip`: disables gzip for HTML, JS, and CSS.

The `Solitaire*` fields tune the sync transport between server and browser:

- `SolitaireSyncTimeout` limits how long a pending server-to-client sync task may wait. If it is exceeded, the instance is ended.
- `SolitaireQueue` and `SolitairePending` control queue depth and backpressure.
- `SolitairePing`, `SolitaireRollTimeout`, `SolitaireFlushSizeLimit`, and `SolitaireFlushTimeout` control how the sync connection is kept alive and flushed.

Most apps should leave the `Solitaire*` settings alone unless they are debugging runtime behavior or tuning under load.

## CSP

CSP is off until you call `doors.UseCSP(...)`.

```go
doors.UseCSP(r, doors.CSP{
	ConnectSources:      []string{"https://api.example.com"},
	ScriptStrictDynamic: true,
})
```

When enabled, **Doors** builds the `Content-Security-Policy` header per page and automatically collects hashes and sources from framework-managed resources.

In practice, that means:

- `script-src` always includes `'self'` plus collected script hashes and sources
- `style-src` always includes `'self'` plus collected style hashes and sources
- `connect-src` always includes `'self'`

External script and style resources added through **Doors** also register their source automatically for CSP.

The field groups behave like this:

| Fields | `nil` | `[]` | values |
| --- | --- | --- | --- |
| `ScriptSources`, `StyleSources`, `ConnectSources` | keep only the framework defaults | keep only the framework defaults | append your values |
| `DefaultSources` | use the built-in default | omit the directive | emit your values |
| `FormActions`, `ObjectSources`, `FrameSources`, `FrameAcestors`, `BaseURIAllow` | default to `'none'` | omit the directive | emit your values |
| `ImgSources`, `FontSources`, `MediaSources`, `Sandbox`, `WorkerSources` | omit the directive | omit the directive | emit your values |

`ReportTo` only emits the `report-to` directive. You still need to send the matching `Report-To` response header yourself.

## Esbuild

**Doors** already has a default esbuild profile. The router starts with a base profile that targets `ES2022` and minifies output.

This is used for:

- the main **Doors** client bundle
- buildable `ScriptInline`
- buildable `ScriptCommon`
- buildable `ScriptModule`

Use `doors.UseESConf(...)` when you want different esbuild options.

The simplest option is `doors.ESOptions`:

```go
doors.UseESConf(r, doors.ESOptions{
	Minify: false,
	JSX:    doors.JSXPreact(),
})
```

`doors.ESOptions` is one profile applied the same way for every profile name.

Its fields are:

- `Minify`: turns esbuild minification on or off for this profile object
- `External`: package names that should stay external instead of being bundled
- `JSX`: JSX transform settings

Use the JSX helpers when you need them:

- `doors.JSXReact()`
- `doors.JSXPreact()`

If the presets are not enough, build `doors.JSX` yourself:

- `JSX`: the esbuild JSX mode
- `Factory` and `Fragment`: names for classic JSX runtimes
- `ImportSource`: package used by automatic JSX runtime
- `Dev`: enables development JSX output
- `SideEffects`: preserves JSX side effects

If you need named profiles, implement `doors.ESConf` yourself:

```go
type ESConf interface {
	Options(profile string) api.BuildOptions
}
```

The `profile` value comes from the `Profile` field on `ScriptInline`, `ScriptCommon`, or `ScriptModule`.

Your implementation must support the default profile `""`, because **Doors** uses it for its own main client build too.

One important rule: resource types still apply the entry-point and output settings they require, so your esbuild options can be supplemented or overridden to make that resource type work.

## Other

Three smaller router-level helpers are worth knowing about:

- `doors.UseSessionCallback(...)`: observe **Doors** session create/delete events. `Create` receives the new session ID and the headers from the request that created it.
- `doors.UseErrorPage(...)`: render your own page for internal framework errors. The callback receives the requested `doors.Location` and the `error`.
- `doors.UseLicense(...)`: load the license certificate used for non-AGPL production use.

## Rules

- Start with defaults and change only the settings you actually need.
- Use `SystemConf` for lifetime, timeout, sync, and serving behavior.
- Turn on CSP when you want browser-enforced loading rules, then add only the extra sources your app really needs.
- Use `ESOptions` first; move to custom named profiles only when one set of build options is not enough.
