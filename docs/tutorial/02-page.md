# The First Page

Let’s build a dynamic, reactive dashboard app — written entirely in Go. Along the way, you’ll see how each core piece of the framework fits together: from live HTML updates and event handling to state management, concurrency control, and navigation. 

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/7l7jc73xeqzfyvcwsn54.gif)

> For tutorial purposes, I deliberately omitted error handling to reduce LOC; that's not how it should be done!

### 0. SSL Certs (optional, but recommended)

Framework is optimized for HTTP/2/3. Without SSL, the browser limits the number of simultaneous requests to 6, which can cause issues in some rare, highly interactive and heavy-sync scenarios.

> _6 requests are not enough?_ 
> Each event goes via an individual HTTP request (it has benefits, e.g., native form data support). With some long-running processing and no concurrency control enabled, it's easy to hit the limit.
> _What about overhead?_
> The HTTP/2/3 multiplexing and header compression keep the cost of additional requests low; we are cool. 

Cook self-signed SSL certs:

```bash
# install package
$ go install filippo.io/mkcert@latest

# makes generated certs trustable by the system (removes browser warning for you), optional 
$ mkcert -install

# create certs in the current folder
$ mkcert localhost 127.0.0.1 ::1
...
The certificate is at "./localhost+2.pem" and the key at "./localhost+2-key.pem" 
```

> In a production environment behind a reverse proxy, there is no need for SSL on a Go app itself.

## 2. General Page Template

`./page_template.templ`

This app has one page with multiple path variants, so a separate template isn’t needed.

Still, it's nice to have concerns separated.

### Page Interface

The page must provide `head` and `body` content to the template:

```templ
// all our pages must be of this shape
type Page interface {
	// method returns component for head insertion
	Head() templ.Component
	// method returns component for body insertion
	Body() templ.Component
}

```

### Template

```templ
// template that takes `Page` as arg
templ Template(p Page) {
	<!DOCTYPE html>
	<html lang="en" data-theme="dark">
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<meta name="color-scheme" content="light dark"/>
			// IMPORTANT: include frameworks' assets (~10KB)
			@doors.Include()
			// Generates <link rel="stylesheet" ... > while collecting info for CSP
			@doors.ImportStyleExternal{
				Href: "https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css",
			}
			// page title, meta-data, etc
			@p.Head()
		</head>
		<body>
			<main class="container">
				// page content
				@p.Body()
			</main>
		</body>
	</html>
}

```

Two notes:
1. We include the framework’s assets; that’s crucial.
2. Instead of just `<link rel="stylesheet" href="...">`, we used `doors.ImportExternalStyle`, which also collects information for. [CSP](https://content-security-policy.com/) header generation. CSP is disabled by default, but this prepares us for it.

> `doors.Import...` handles local, embedded, and external CSS and JS assets. For JavaScript/TypeScript modules, it eenables build/bundle steps and generates an import map

## 3. Page and Path

`./page.templ`

### Page Path

In _doors_, the URI is decoded into a **Path Model**. It supports path variants, parameters, and query values. 

Our path will have two variants:
* `/` location selector
* `/:Id` dashboard for selected location

One parameter:
* Id of the city
And two query values: 
* forecast days 
* units (metric/imperial)

We’ll omit query values for now and add them later.

Our path model:

```templ
type Path struct {
	Selector  bool `path:"/"`    // the first variant
	Dashboard bool `path:"/:Id"` // the second variant
	Id        int  // path parameter with City Id
}
```

* The framework uses `path` tags to match the request path against the provided pattern. 
* The matched variant’s field is set to true.

### Page Component

The path structure is wrapped in the state primitive (**Beam**) and passed to the page render function:

```templ
type page struct{}

// page render function, follows doors.Page interface
func (hp *page) Render(path doors.SourceBeam[Path]) templ.Component {
	return Template(hp)
}

// head component for our template
templ (hp *page) Head() {
	<title>Dashboard App</title>
}

// body component for our template
templ (hp *page) Body() {
	<h1>Hello <i>doors</i>!</h1>
}
```

> To be compatible with the framework, the page type must implement Render(), return a component, and accept a **Beam** with the path model.

### Page Handler

A function that runs when the path matches. It reads request data (`doors.RPage`) and chooses the response (`doors.PageRouter`).

In our case, it's straightforward:

```templ
func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	// just serve the page instance, no checks
	return p.Page(&page{})
}
```

> `doors.PageRouter` also supports soft (internal) and hard (HTTP) redirects and serving static pages.

## 4. Router

`./main.go`

Create a *doors* router, provide a page handler, and launch the server.

```templ
package main

import (
	"github.com/doors-dev/doors"
	"net/http"
)

func main() {
	// create doors router
	dr := doors.NewRouter()
	dr.Use(
		// attach page handler function
		doors.UsePage(Handler),
	)

	// start standard Go server with our self signed cert
	// and router
	err := http.ListenAndServeTLS(":8443", "localhost+2.pem", "localhost+2-key.pem", dr)
	if err != nil {
		panic(err)
	}
}
```

> Notice how `doors.Router` just plugs into the Go standard server! Go is awesome.

## 4. Launch

Build & launch

```bash
templ generate && go run .
```

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/4u2ulst63r75togk3dfg.png)

---

Next: [Live Reloading](./03-live-reloading.md)

---

## Code

`./page_template.templ`

```templ
package main

import "github.com/doors-dev/doors"

// all our pages must be of this shape
type Page interface {
	// method returns component for head insertion
	Head() templ.Component
	// method returns component for body insertion
	Body() templ.Component
}

// template that takes `Page` as arg
templ Template(p Page) {
	<!DOCTYPE html>
	<html lang="en" data-theme="dark">
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<meta name="color-scheme" content="light dark"/>
			// IMPORTANT: include frameworks' assets (~10KB)
			@doors.Include()
			// Generates <link rel="stylesheet" ... > while collecting info for CSP
			@doors.ImportStyleExternal{
				Href: "https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css",
			}
			// page title, meta-data, etc
			@p.Head()
		</head>
		<body>
			<main class="container">
				// page content
				@p.Body()
			</main>
		</body>
	</html>
}
```

`page.templ`

```templ
package main

import "github.com/doors-dev/doors"

type Path struct {
	Selector  bool `path:"/"`    // the first variant
	Dashboard bool `path:"/:Id"` // the second variant
	Id        int  // path parameter with City Id
}

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	// just serve the page instance, no checks
	return p.Page(&page{})
}

type page struct{}

// page render function, follows doors.Page interface
func (hp *page) Render(path doors.SourceBeam[Path]) templ.Component {
	return Template(hp)
}

// head component for our template
templ (hp *page) Head() {
	<title>Dashboard App</title>
}

// body component for our template
templ (hp *page) Body() {
	<h1>Hello <i>doors</i>!</h1>
}
```

`./main.go`

```templ
package main

import (
	"github.com/doors-dev/doors"
	"net/http"
)

func main() {
	// create doors router
	dr := doors.NewRouter()
	dr.Use(
		// attach page handler function
		doors.UsePage(Handler),
	)

	// start standard Go server with our self signed cert
	// and router
	err := http.ListenAndServeTLS(":8443", "localhost+2.pem", "localhost+2-key.pem", dr)
	if err != nil {
		panic(err)
	}
}
```

