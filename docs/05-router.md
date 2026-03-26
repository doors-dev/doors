# Router

The **Doors** router is the HTTP entry point for your app.

You create it with `doors.NewRouter()`, register page models and any custom routes you need, then pass it to `http.ListenAndServe`.

```go
router := doors.NewRouter()

doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	return doors.ResponseComp(Page(s))
})

if err := http.ListenAndServe(":8080", router); err != nil {
	panic(err)
}
```

`doors.Router` implements `http.Handler`, so it plugs directly into the standard Go server.

## Flow

For normal app requests, the router works in this order:

1. custom `Route`s added with `doors.UseRoute`
2. page models added with `doors.UseModel`
3. fallback handler, if configured
4. `404 Not Found`

That means custom routes always get the first chance to handle a request.

## Routes

Use `doors.UseRoute` for endpoints that are not page models, such as:

- health checks
- webhooks
- custom file endpoints
- static file mounts

To add one, implement `doors.Route`:

```go
type Route interface {
	Match(r *http.Request) bool
	Serve(w http.ResponseWriter, r *http.Request)
}
```

Example:

```go
type HealthRoute struct{}

func (HealthRoute) Match(r *http.Request) bool {
	return r.URL.Path == "/health"
}

func (HealthRoute) Serve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

doors.UseRoute(router, HealthRoute{})
```

## Files

**Doors** includes a few ready-to-use route types for serving files.

### RouteFS

Serve an `fs.FS` under a URL prefix:

```go
doors.UseRoute(router, doors.RouteFS{
	Prefix: "/assets",
	FS:     assetsFS,
})
```

### RouteDir

Serve a local directory under a URL prefix:

```go
doors.UseRoute(router, doors.RouteDir{
	Prefix:  "/public",
	DirPath: "./public",
})
```

### RouteFile

Serve one local file at a fixed path:

```go
doors.UseRoute(router, doors.RouteFile{
	Path:     "/robots.txt",
	FilePath: "./static/robots.txt",
})
```

### Resources

`doors.RouteResource` and `doors.RouteResourceFS` serve fixed public paths through the **Doors** resource registry.

Use them when you want the resource handling behavior from the framework, but still want a stable URL of your own.

## Fallback

Use `doors.UseFallback` when another `http.Handler` should receive unmatched requests.

```go
doors.UseFallback(router, myMux)
```

This is useful when **Doors** is only part of a larger server.

## Config

Most apps only need `doors.NewRouter()`, `doors.UseModel(...)`, and maybe `doors.UseRoute(...)`.

When you need more control, router-level helpers include:

- `doors.UseFallback(...)`
- `doors.UseSystemConf(...)`
- `doors.UseErrorPage(...)`
- `doors.UseSessionCallback(...)`
- `doors.UseCSP(...)`
- `doors.UseServerID(...)`
- `doors.UseESConf(...)`
- `doors.UseLicense(...)`

Use these when you are configuring the router itself, not when you are defining page URLs. Page URL design belongs in your path models.

For the main router-level settings, see [Configuration](./19-configuration.md).
