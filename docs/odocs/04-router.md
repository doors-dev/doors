# Router

The doors framework provides a router that plugs into the standard Go `http.Server`.  
It handles app routing, static files, hooks, and framework resources.

```go
func main() {
  // create doors router
  router := doors.NewRouter()
  
  /* configuration */
  
  // use router as handler
  err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", router)
  if err != nil {
      log.Fatal("ListenAndServe: ", err)
  }
}
```

Call `router.Use(...doors.Use)` to configure.

## Apps

### App as struct

Implements the `doors.App[M any]` interface:

```go
type App[M any] interface {
	Render(SourceBeam[M]) templ.Component
}
```

> App render function must provide full page HTML. 
>
> **You must render `@doors.Include()` component in the page `<head>` block to include framework resources**.

### App as function

Any function with this signature:

```go
func(SourceBeam[M any]) templ.Component
```

Where `M` is the **Path Model** (see [Path Model](./05-path-model.md)).

### Register an app

Use `UseModel` to bind a handler for a path model:

```go
func UseModel[M any](
  handler func(m ModelRouter[M], r RModel[M]) ModelRoute,
) Use
```

Example:

```go
func homeApp(m doors.ModelRouter[Path], r doors.RModel[Path]) doors.ModelRoute {
  	// serve home app (implements doors.App[Path])
   return m.App(&homeApp{}) 
}

router.Use(doors.UseModel(homeApp))
```

### Model Router (`doors.ModelRouter[M]`)

Provides routing options for serving a given path model:

1. **App(app App[M]) ModelRoute**  
   Serve an app struct implementing `doors.App[M]`.

2. **AppFunc(func(SourceBeam[M]) templ.Component) ModelRoute**  
   Serve an app via a function.

3. **Reroute(model any, detached bool) ModelRoute**  
   Internally reroute to another path model.  
   If `detached = true`, the URL is not updated on the frontend.

4. **Redirect(model any, status int) ModelRoute**  
   Perform an HTTP redirect to the URL built from `model`.

5. **StaticPage(content templ.Component, status int) ModelRoute**  
   Serve a static page with the given status code.  
   ⚠️ Using beams/doors inside a static page will panic.

> To set a status code on dynamic apps, use `@doors.Status(code)` component or `doors.SetStatus(ctx, code)` function.

### Model Request `doors.RModel[M any]`

Gives access to cookies, headers, and the requested path in the form of a typed **Path Model**.  
Use it to check user authentication or inject per-request data before serving an app.

### Notes

* Each **Path Model** type can be registered only once.  
* The route handler inside `doors.UseModel` is the right place to enforce cookie/session access control.

---

## Custom Routes

You can define your own route types by implementing the `Route` interface:

```go
type Route interface {
  Match(r *http.Request) bool
  Serve(w http.ResponseWriter, r *http.Request)
}
```

Then register them with:

```go
func UseRoute(r Route) Use
```

Example:

```go
type HealthRoute struct{}

func (h HealthRoute) Match(r *http.Request) bool {
  return r.URL.Path == "/health"
}

func (h HealthRoute) Serve(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("ok"))
}

router.Use(doors.UseRoute(HealthRoute{}))
```

> All routes matched **before** apps

---

## Static File Routes

The router provides `Route` types that can be registered via `UseRoute`.

### RouteFS

Serve files from an `fs.FS` under a URL prefix.

```go
type RouteFS struct {
  Prefix       string // URL prefix (not root "/")
  FS           fs.FS  // filesystem source
  CacheControl string // optional Cache-Control header
}
```

Example:

```go
//go:embed assets/*
var assets embed.FS

router.Use(doors.UseRoute(doors.RouteFS{
  Prefix: "/assets/",
  FS:     assets,
}))
```

### RouteDir

Serve files from a local directory under a URL prefix.

```go
type RouteDir struct {
  Prefix       string // URL prefix (not root "/")
  DirPath      string // local directory path
  CacheControl string // optional Cache-Control header
}
```

Example:

```go
router.Use(doors.UseRoute(doors.RouteDir{
  Prefix: "/public/",
  DirPath: "./public",
}))
```

### RouteFile

Serve a single file at a fixed URL path.

```go
type RouteFile struct {
  Path         string // URL path (not root "/")
  FilePath     string // local file path
  CacheControl string // optional Cache-Control header
}
```

Example:

```go
router.Use(doors.UseRoute(doors.RouteFile{
  Path: "/favicon.ico",
  FilePath: "./static/favicon.ico",
}))
```

---

## Fallback

Set a fallback handler for unmatched requests. Useful for integrating with another router.

```go
func UseFallback(handler http.Handler) Use
```

Example:

```go
router.Use(doors.UseFallback(myOtherMux))
```

---

## System Configuration

Apply system-wide config such as timeouts, resource limits, and session behavior.

```go
type SystemConf = common.SystemConf

func UseSystemConf(conf SystemConf) Use
```

Example:

```go
router.Use(doors.UseSystemConf(doors.SystemConf{
  SessionInstanceLimit: 12,
}))
```

---

## Error Page

Provide a custom component for error handling.

```go
func UseErrorPage(page func(message string) templ.Component) Use
```

Example:

```go
router.Use(doors.UseErrorPage(func(msg string) templ.Component {
  return myErrorPage{Message: msg}
}))
```

---

## Security (CSP)

Configure Content Security Policy headers.

```go
func UseCSP(csp CSP) Use
```

Example:

```go
router.Use(doors.UseCSP(doors.CSP{
  ScriptSrc: []string{"'self'", "cdn.example.com"},
}))
```

---

## Build Profiles

Configure esbuild options and profiles for JavaScript/TypeScript.

```go
func UseESConf(conf ESConf) Use
```

> Check [esbuild](./ref/06-esbuild.md) for details

---

## Session Callback

Register callbacks for the session lifecycle.  
Useful for analytics, auditing, or load balancing.

```go
type SessionCallback interface {
  Create(id string, header http.Header)
  Delete(id string)
}

func UseSessionCallback(callback SessionCallback) Use
```

---

## License

Verifies and adds the license certificate. A license is required for commercial production use. 

```go
func UseLicense(cert string) Use
```
