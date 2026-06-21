# Get Started

## Install GoX

> Ensure your Go is at least 1.25.1; you can check in the terminal by running `go version`

**Doors** is built on top of [GoX](https://github.com/doors-dev/gox), a purpose-built Go language extension that turns HTML templates into typed Go expressions and adds `elem` primitives.

**GoX** comes with its own language server, which mostly acts as a [gopls](https://go.dev/gopls/) proxy while adding extra features on top.

Please use the official [VS Code](https://marketplace.visualstudio.com/items?itemName=doors-dev.gox) or [Neovim](https://github.com/doors-dev/nvim-gox) extension. Alternatively, follow the manual installation guide in the [GoX README](https://github.com/doors-dev/gox).

It is also recommended to have the `gox` binary on your `PATH`.

That lets you run commands such as `gox fmt` and `gox gen` yourself, and it also helps editor tooling and code agents trigger the same workflow when needed.

### GoX Workflow

The practical workflow is simple:

- write GoX source in `.gox`
- use `.go` for ordinary Go files when that fits naturally
- treat `.x.go` as generated output, don't edit it

> The language server keeps generated files up to date while you work, and you can use `gox gen` or `gox fmt` when you want to run generation or formatting yourself.

## Setup Project

> For new applications, use [doors-dev/doors-starter](https://github.com/doors-dev/doors-starter), the recommended ready-to-run **Doors** project skeleton. This page walks through a minimal hello-world app by hand.

Create a new directory containing our project:

```bash
mkdir hello-doors
```

Initialize a new Go module there and get **Doors**:

```bash
cd hello-doors
go mod init github.com/doors-dev/doors-examples/hello-doors
go get github.com/doors-dev/doors
```

## App Component

Write a component with the page template to `app.gox`:

```gox
package main

import "github.com/doors-dev/gox"

type App struct{}

elem (a App) Main() {
	<!doctype html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<title>Hello Doors!</title>
		</head>
		<body>
			<main class="container">
				<h1>Hello Doors!</h1>
			</main>
		</body>
	</html>
}
```

Components in **Doors** (and **GoX**) must have a `Main()` method that returns `gox.Elem`. The `elem` keyword lets you write an HTML template directly in the function body.

> The GoX language server generates and manages `.x.go` files automatically.

## Serve The App

Create the app, hand it to Go's HTTP server, and run it. Put this in `main.go`:

```go
package main

import (
	"context"
	"net/http"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

func main() {
	app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
		return App{}
	})

	if err := http.ListenAndServe(":8080", app); err != nil {
		panic(err)
	}
}
```

Start the program with `go run .` and open [http://localhost:8080](http://localhost:8080).

> **Safari on localhost**
>
> **Doors** uses a `Secure` internal session cookie by default. Chrome and Firefox usually accept secure cookies on `http://localhost`, but Safari rejects them on plain HTTP. If you test with Safari locally, either use local HTTPS or disable the `Secure` attribute for development:
>
> ```go
> app := doors.NewApp(page, doors.WithConf(doors.Conf{
> 	ServerSessionCookieNoSecure: true,
> }))
> ```

#### What Just Happened?

`doors.NewApp(...)` builds the **Doors** application. It takes a single page function that, for every request, returns the root component **Doors** should render:

```go
app := doors.NewApp(func(ctx context.Context, r doors.Request) gox.Comp {
	return App{}
})
```

> `doors.Request` exposes request and response headers, cookies, and the underlying HTTP context. The `ctx` is the **Doors** runtime context for the new page instance — use it for **Doors** APIs.

The returned `app` is an `http.Handler`, so it plugs straight into the standard Go server:

```go
http.ListenAndServe(":8080", app)
```

Configuration options (CSP, request timeouts, error pages, etc.) and middleware for static files live on the same `app` value. See [App](./04-app.md) and [Configuration](./21-configuration.md).

## Dynamic Update

At this point you have a static page.

To see the basic **Doors** flow, turn it into a tiny interactive counter.

Update the imports in `app.gox` and add a small counter component:

```gox
import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

type Counter struct {
	count int
	door  doors.Door // dynamic container
}

elem (c *Counter) Main() {
	<button
		(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestPointer) bool {
				c.count += 1
				c.door.Inner(ctx, c.count)
				return false
			},
		})>
		Click Me
	</button>

	~>(c.door) <span>0</span>
}
```

And render it on the page:

```gox
elem (a App) Main() {
	<!doctype html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<title>Hello Doors!</title>
		</head>
		<body>
			<main class="container">
				<h1>Hello Doors!</h1>
				~(&Counter{})
			</main>
		</body>
	</html>
}
```

Now the browser click is sent back to Go, the handler updates the counter state, and the dynamic `<span>` shows the current value.

### Next

- [Core Concepts](./02-core-concepts.md) explains the runtime model behind sessions, instances, doors, hooks, and state.
- [Template Syntax](./03-template-syntax.md) covers the GoX syntax used throughout the docs.
- [Components](./08-components.md) explains component structure, state ownership, and rendering boundaries.
- [App](./04-app.md) and [Routing](./05-routing.md) take the next step into request handling and URL design.
- [State](./07-state.md) covers reactive state.
