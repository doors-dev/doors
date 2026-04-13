# Router

The **Doors** router is the HTTP entry point for your app.

Most apps use it in three steps:

1. create a router with `doors.NewRouter()`
2. register page models with `doors.UseModel(...)`
3. pass the router to `http.ListenAndServe`

```go
router := doors.NewRouter()

/* configuration */

if err := http.ListenAndServe(":8080", router); err != nil {
	panic(err)
}
```

`doors.Router` implements `http.Handler`, so it plugs directly into the standard Go server.

In practice, the router is where you decide:

- which page models belong to this app
- which non-page endpoints should live beside them
- which router-wide settings apply to the whole app

## Start

Start with `doors.UseModel(...)`.

That is the normal path for **Doors** pages:

```go
type Path struct {
	Home bool `path:"/"`
}

doors.UseModel(router, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	return doors.ResponseComp(Page(s))
})
```

If your app is mainly pages, you can often stop there.

Path matching, decoding, and URL generation are covered in [Path Model](./04-path-model.md).

## Custom Routes

Good fits are:

- health checks
- static file mounts
- simple public `GET` endpoints

`doors.UseRoute(...)` takes a `doors.Route`:

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

The important practical rule is simple:

- use `UseModel` for app pages
- use `UseRoute` for public `GET` endpoints that should bypass page-model routing

`UseRoute` participates in the router's normal content handling, which means it is used for `GET` requests.

If you need a broader HTTP surface such as webhooks, APIs, or another mux, hand unmatched requests to something else with [Fallback](#fallback) or mount **Doors** inside a larger server.


### `RouteFS`

Serve an `fs.FS` under a URL prefix:

```go
doors.UseRoute(router, doors.RouteFS{
	Prefix: "/assets",
	FS:     assetsFS,
})
```

### `RouteDir`

Serve a local directory under a URL prefix:

```go
doors.UseRoute(router, doors.RouteDir{
	Prefix:  "/public",
	DirPath: "./public",
})
```

### `RouteFile`

Serve one local file at a fixed path:

```go
doors.UseRoute(router, doors.RouteFile{
	Path:     "/robots.txt",
	FilePath: "./static/robots.txt",
})
```

### Resource Routes

`doors.RouteResource` serves a fixed public path through the **Doors** resource registry. You can use it to serve an asset at a known path with caching and gzip support:

```go
doors.UseRoute(r, doors.RouteResource{
	Path:     "assets/sans.ttf",
	Resource: doors.ResourceFS(assets.Get(), "sans.ttf"),
})
```

## Fallback

Use `doors.UseFallback` when another `http.Handler` should receive unmatched requests.

```go
doors.UseFallback(router, myMux)
```

This is useful when **Doors** is only part of a larger server.

## Matching

When a request comes in, the router first handles its own framework URLs:

- resource URLs
- hook URLs
- internal sync URL

After that, normal app content is handled in this order:

1. routes added with `doors.UseRoute(...)`
2. models added with `doors.UseModel(...)`
3. the fallback handler, if one is configured
4. `404 Not Found`

Matching within each category happens in the order of registration.



## Router Settings

Most apps only need `doors.NewRouter()`, `doors.UseModel(...)`, and maybe `doors.UseRoute(...)`.

Add router-wide settings only when you need them.

Common helpers include:

- `doors.UseFallback(...)`
- `doors.UseSystemConf(...)`
- `doors.UseErrorPage(...)`
- `doors.UseSessionCallback(...)`
- `doors.UseCSP(...)`
- `doors.UseServerID(...)`
- `doors.UseESConf(...)`

These configure the router itself. They are separate from path-model design and page behavior.

For the main router-level settings, see [Configuration](./21-configuration.md).
