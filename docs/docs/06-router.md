#  Router

The *doors* framework provides a router that you can plug into the standard Go `http.Server`.
 It handles page routing, static files, hooks, and framework resources.

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

Call `router.Use(...doors.Mod)` to configure.

## Pages

### Page as struct

implements the `doors.Page[M any]` interface

```go
type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}
```

### Page as function

 a function with this signature:

```go
func(SourceBeam[M any]) templ.Component
```

Where `M` is the **Path Model**  (see [Path Model](./03-path-model.md))

### Register a page

Use `UsePage` to bind a handler for a path model:

```go
func UsePage[M any](
  handler func(p PageRouter[M], r RPage[M]) PageRoute,
) Mod
```

Example:

```go
func homePage(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
  	// serve home page (implements doors.Page[BlogPath])
   return p.Page(&homePage{}) 
}
```

```go
r.Use(doors.UsePage(homePage))
```

### Page Router (`doors.PageRouter[M]`)

Provides routing options for serving a given path model:

1. **Page(page Page[M]) PageRoute**
    Serve a page struct implementing `doors.Page[M]`.
2. **PageFunc(func(SourceBeam[M]) templ.Component) PageRoute**
    Serve a page via a function.
3. **Reroute(model any, detached bool)**
    Internally re-route to another path model.
    If `detached = true`, the URL will not be updated on the frontend.
4. **RedirectStatus(model any, status int) PageRoute**
    Perform an HTTP redirect to the URL built from `model`.
5. **StaticPage(content templ.Component, status int) PageRoute**
    Serve a static page with the given status code.
    ⚠️ Using beams/doors inside a static page will panic.

> To set status code on dynamic pages, use `@doors.Status(code)`  component or `doors.SetStatus(ctx, code)` function.

### Page Request `doors.RPage[M any]`

Gives access to cookies, headers, and the requested path in the form of a typed **Path Model**. Use it to check user authentication or inject per-request data before serving a page.

### Please, consider:

* Each **Path Model** type can be registered only once.
* The route handler inside `doors.UsePage` is the right place to enforce cookie access control (retrieve session) for protected resources.

## Static File Serving

The router provides helpers for serving files and directories.  

### UseDir

Serve static files from a **local directory** via `os.DirFS`.

```go
func UseDir(prefix string, path string) Mod
```

**Parameters**:  

- `prefix` (string): URL prefix (e.g., `/public/`)  
- `path` (string): local directory path  

Example:

```go
router.Use(doors.UseDir("/public/", "./public"))
```

### UseFS

Serve files from an `fs.FS` (typically `embed.FS`).

```go
func UseFS(prefix string, fs fs.FS) Mod
```

**Parameters**:  

- `prefix` (string): URL prefix (e.g., `/assets/`)  
- `fs` (fs.FS): filesystem source  

Example:

```go
//go:embed assets/*
var assets embed.FS

router.Use(doors.UseFS("/assets/", assets))
```

### UseFile

Serve a **single file** at a given URL path.

```go
func UseFile(path string, localPath string) Mod
```

**Parameters**:  

- `path` (string): URL path (e.g., `/favicon.ico`)  
- `localPath` (string): local file path  

Example:

```go
router.Use(doors.UseFile("/favicon.ico", "./static/favicon.ico"))
```

---

## Fallback

Set a fallback handler for unmatched requests. Useful for integrating with another router (e.g., Gorilla Mux) or serving a custom 404.

```go
func UseFallback(handler http.Handler) Mod
```

Example:

```go
router.Use(doors.UseFallback(myOtherMux))
```

## System Configuration

Apply system-wide config such as timeouts, resource limits, and session behavior.

```go
type SystemConf = common.SystemConf

func UseSystemConf(conf SystemConf) Mod
```

Example:

```go
router.Use(doors.UseSystemConf(doors.SystemConf{
  SessionInstanceLimit: 12,
}))
```

## Error Page

Provide a custom component for error handling.

```go
func UseErrorPage(page func(message string) templ.Component) Mod
```

Example:

```go
router.Use(doors.UseErrorPage(func(msg string) templ.Component {
  return myErrorPage{Message: msg}
}))
```

## Security (CSP)

Configure Content Security Policy headers.

```go
type CSP = common.CSP

func UseCSP(csp CSP) Mod
```

Example:

```go
router.Use(doors.UseCSP(doors.CSP{
  ScriptSrc: []string{"'self'", "cdn.example.com"},
}))
```

## Build Profiles

Configure esbuild options and profiles for JS/TS

```go
func UseESConf(conf ESConf) Mod
```

---

## Session Hooks

Register callbacks for session lifecycle.

```go
func UseSessionHooks(onCreate func(id string), onDelete func(id string)) Mod
```

## 
