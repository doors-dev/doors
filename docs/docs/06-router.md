# Router

*doors* framework provides a router that you can plug into standard go http server:

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

Call `router.Use(...doors.Mod)` to configureate.

## Page Serving

### Page

**Page as struct** must implement page ` doors.Page[M any]` interface

```go
type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}
```

**Page as function** must follow the signature:

```go
func(SourceBeam[M any]) templ.Component
```

Where M is `Path Model` (check out [Path Model](./03-path-model.md))

### Page Handler

To serve a page, attach a page handler to the router with `Mod`: 

```go
func ServePage[M any](handler func(p PageRouter[M], r RPage[M]) PageRoute) Mod
```

Example:

```go

router.Use(doors.ServePage(
  // page handler function
  func (p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
  	// home page implements doors.Page[M any] interface)
  	return p.Page(&homePage{})
	},
))

```

##### Page Router `doors.PageRouter[M any]`

Provides various routing options via methods:

1. `Page(page Page[M]) PageRoute` and `PageStatus(..., status int) PageRoute`
   Serves a dynamic page with 200 or the provided  status code

2. `PageFunc(pageFunc func(SourceBeam[M]) templ.Component) PageRoute` and `PageFuncStatus(..., status int)`
   Serves a dynamic page with 200 or the provided  status code

3. `Reroute(model any, detached bool)`
   Internally, routes the request to a different Path Model, causing a new handler matching process. If detached = true, the path string will not be updated and synced on the frontend.

   > For example, this is useful to serve the Unauthorized page. 
   > ```go
   > router.Use(doors.ServePage(
   >   func (p doors.PageRouter[AdminPath], r doors.RPage[AdminPath]) doors.PageRoute {
   >     if !isAuthorized(r) {
   >         // will keep the original address in the browser, while rendering 
   >       	 // Unauthorized page (it must also be registered with
   >       	 // doors.ServePage[UnauthorizedPath])
   >         return p.Reroute(UnauthorizedPath{}, true)
   >     } 
   >      // home page implements doors.Page[M any] interface)
   >   	return p.Page(&adminPage{})
   >   },
   > ))
   > ```

4. `Redirect(model any) PageRoute` and `RedirectStatus(..., status int) PageRoute`                                                  

   Performs HTTP level redirect with 302 or the provided status to the URL constructed from the provided model
   
5. `StaticPage(content templ.Component) PageRoute` and `StaticPage(...,status int) PageRoute`
   Serves a static page with 200 or the provided status code. The usage of beams, doors, etc, on a static page will cause panic.

##### Page Request `doors.RPage[M any]`

Gives you access to cookies, headers, and the requested path in the form of a `Path Model`.  You can use it to check user authentication before serving a page.

### Please, consider:

* Specific `Path Model` type can be registered only once.
* Page route handler inside `doors.ServePage` is the only place where it's neccessary to check user access for protected resources. 

## Static Serving

### ServeDir

Serves static files from a local directory using `os.DirFS`.
 This creates a filesystem from the directory and serves it at the specified prefix.

**Parameters**:

- `prefix` (string): URL prefix (e.g., `/public/`)
- `path` (string): Local directory path

```go
func ServeDirPath(prefix string, localPath string) Mod
```

### ServeFS

Serves static files from a filesystem at the specified URL prefix.  This is useful for serving embedded assets using Go’s `embed.FS`.

**Parameters**:

- `prefix` (string): URL prefix (e.g., `/assets/`)
- `fs` (fs.FS): Filesystem to serve from (typically `embed.FS`)

```go
func ServeDir(prefix string, path string) Mod
```

### ServeFS

Serves static files from a filesystem at the specified URL prefix.  This is useful for serving embedded assets using Go’s `embed.FS`.

- `prefix` (string): URL prefix (e.g., `/assets/`)
- `fs` (fs.FS): Filesystem to serve from (typically `embed.FS`)

```go
func ServeFS(prefix string, fs fs.FS) Mod
```

#### ServeFile

Serves a single file at the specified URL path.

**Parameters**:

- `path` (string): URL path (e.g., `/favicon.ico`)
- `localPath` (string): Local file path

```
func ServeFile(path string, localPath string) Mod
```

## Custom Serving

### ServeFallback

Sets a fallback router for requests that do not match any routes. For example, you can use `gorilla.Mux` to serve whatever you want.

**Parameters**:

- `handler` (http.Handler): HTTP handler to execute for unmatched routes

**Signature**:

```go
func ServeFallback(handler http.Handler) Mod
```

