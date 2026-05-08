# App

`doors.App` is the HTTP entry point for your **Doors** application.

Most apps follow three small steps:

1. create the app with `doors.NewApp(page, options...)`
2. attach middleware with `app.Use(...)` for static files or anything else that should bypass page rendering
3. pass the app to `http.ListenAndServe`

```go
app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
	return App{}
})

app.Use(
	doors.UseDir("/assets/", "./assets", doors.CacheControlStatic),
)

if err := http.ListenAndServe(":8080", app); err != nil {
	panic(err)
}
```

`doors.App` implements `http.Handler`, so it plugs straight into the standard Go server, a `chi`/`mux` parent, or any other HTTP setup.

In practice, the app pulls together:

- middleware that runs in front of page rendering (static files, logging, CORS, ...)
- app-wide settings (CSP, esbuild, server ID, error pages, session tracking, system limits)
- where the binary plugs into your HTTP setup

## Page Factory

The function passed to `doors.NewApp(...)` is a **per-request page factory**, not a router. It runs once each time a fresh page instance starts (a navigation, a reload, an open-in-new-tab) and returns the root component for that instance.

```go
app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
	auth := doors.SessionStore(ctx).Init(authKey{}, func() any {
		c, err := r.GetCookie("session")
		if err != nil {
			return doors.NewSource(false)
		}
		_, ok := store.Get(c.Value)
		return doors.NewSource(ok)
	}).(doors.Source[bool])

	return App{auth: auth}
})
```

Most apps return the same root component every time — what changes per request is the state you attach to it. The factory exists so you can:

- bootstrap session-scoped state from cookies or headers (auth, theme, locale) before the page renders
- read or write the location for this instance
- inspect request headers, set cookies, set response headers
- decide which component to return when there are very few top-level shapes (e.g. an error page in odd cases) — but day-to-day routing is *not* this; routing happens inside the returned component (see [Routing](#routing) below).

For the auth-bootstrap pattern in full, see [Storage & Auth](./18-storage-auth.md).

The function receives:

- `ctx` — the **Doors** runtime context for the new instance. Pass this to **Doors** APIs.
- `r doors.Request` — request/response headers, cookies, and the underlying HTTP context (`r.Context()`).

## Configuration Options

App-wide settings are passed as `doors.With...` options to `NewApp`:

```go
app := doors.NewApp(page,
	doors.WithConf(doors.Conf{
		RequestTimeout: 20 * time.Second,
	}),
	doors.WithCSP(doors.CSP{
		ConnectSources: []string{"https://api.example.com"},
	}),
)
```

See [Configuration](./21-configuration.md) for the full list.

## Middleware

`app.Use(...)` adds standard `func(http.Handler) http.Handler` middleware in front of the page handler. It can short-circuit static files, set headers, gate access, log, or hand off to another mux before **Doors** renders a page. **Doors** system endpoints under `/~/...` are handled by the app itself; wrap the app in outer middleware if those requests must also be intercepted.

### UseFS

Serve an embedded `fs.FS` (or any `fs.FS`) under a URL prefix:

```go
//go:embed assets/*
var assetsFS embed.FS

app.Use(
	doors.UseFS("/assets/", assetsFS, doors.CacheControlImmutable),
)
```

### UseDir

Serve a directory from disk under a URL prefix:

```go
app.Use(
	doors.UseDir("/public/", "./public", doors.CacheControlStatic),
)
```

### UseFile

Serve a single file at a fixed path:

```go
app.Use(
	doors.UseFile("/robots.txt", "./static/robots.txt", doors.CacheControlStatic),
)
```

### UseResource

`doors.UseResource` exposes a **Doors** static resource at a fixed public path. This goes through the resource registry and gets caching, gzip, and CSP integration:

```go
app.Use(
	doors.UseResource(
		"/assets/sans.ttf",
		doors.ResourceFS(assetsFS, "sans.ttf"),
		"font/ttf",
	),
)
```

The `contentType` argument is optional — pass an empty string to let the registry detect it.

### Cache-Control Presets

The middleware helpers take a `cacheControl` string. Common presets are exported as constants:

| Constant | Value | Use for |
| --- | --- | --- |
| `doors.CacheControlImmutable` | `public, max-age=31536000, immutable` | Fingerprinted assets that never change at a URL |
| `doors.CacheControlStatic` | `public, max-age=3600, must-revalidate` | Long-lived but non-fingerprinted files (`/favicon.ico`, `/robots.txt`) |
| `doors.CacheControlStaticShort` | `public, max-age=300, must-revalidate` | Static files that may change occasionally |
| `doors.CacheControlHTML` | `public, max-age=0, must-revalidate` | HTML entry points (pair with an ETag) |
| `doors.CacheControlCDN` | `public, max-age=3600, s-maxage=86400` | Browser short, CDN longer |
| `doors.CacheControlPrivate` | `private, max-age=0, must-revalidate` | Per-user responses |
| `doors.CacheControlNoCache` | `no-cache` | Always revalidate |
| `doors.CacheControlNoStore` | `no-store` | Sensitive responses |
| `doors.CacheControlAPI` | `private, no-cache` | JSON APIs |

Pass an empty string when you don't want **Doors** to set a `Cache-Control` header.

### Custom Middleware

`app.Use(...)` accepts any `func(http.Handler) http.Handler`, so existing Go middleware composes naturally:

```go
app.Use(
	logger.Handler,
	cors.New(cors.Options{...}).Handler,
	doors.UseDir("/assets/", "./assets", doors.CacheControlImmutable),
)
```

Middleware runs in registration order. Anything that doesn't short-circuit falls through to the **Doors** page handler.

## Mounting Inside Another Server

`doors.App` is a regular `http.Handler`, so you can mount it inside any router:

```go
mux := http.NewServeMux()
mux.Handle("/admin/", adminHandler)
mux.Handle("/", app)
http.ListenAndServe(":8080", mux)
```

For request matching, **Doors** uses the URL it sees, so mount it at `/` unless you also configure your model paths accordingly.

## Routing

`doors.NewApp` itself does not match URLs. The page function returns a single root component for every request, and routing happens *inside* that component using `doors.RouterSource(...)` (or `doors.RouterBeam(...)`):

```gox
elem (a App) Main() {
	<!doctype html>
	<html lang="en">
		<body>
			~(doors.RouterSource(
				doors.RouteModelSource(func(p doors.Source[Path]) gox.Comp {
					return Page{path: p}
				}),
				doors.RouteLocationDefaultComp(NotFound{}),
			))
		</body>
	</html>
}
```

Path models, route builders, default fallbacks, and routing on derived state are covered in [Routing](./05-routing.md).
