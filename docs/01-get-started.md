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

> GoX language server compiles it and manages `.x.go` files automatically.

## Serve The App

Declare path, create router, attach handler and serve in `main.go`:

```go
package main

import (
	"net/http"

	"github.com/doors-dev/doors"
)

type Path struct {
	Home bool `path:"/"` 
}

func main() {
	r := doors.NewRouter()

	doors.UseModel(r, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
		return doors.ResponseComp(App{})
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
```

Start the program with `go run .` and open [http://localhost:8080](http://localhost:8080).

#### What Just Happened?

We declared a **model**. **Doors** uses a struct with tagged fields to match, decode, and encode a path:

```go
type Path struct {
	Home    bool `path:"/"` // first path pattern
	Catalog bool `path:"/catalog/:ID?"` // second path pattern with optional ID param
	ID      *string
}
```

Next, we added a model handler to the router and served our `App` component:

```go
doors.UseModel(r, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
	return doors.ResponseComp(App{})
})
```
> `doors.RequestModel` provides access to HTTP data such as cookies and headers, while `doors.Source[Path]` is the reactive state primitive that holds the current model value. Usually you store that `Source` on the component so you can use it during rendering.

Finally, the **Doors** router plugs straight into Go's standard HTTP server:

```go
http.ListenAndServe(":8080", r)
```

### Next

- [Core Concepts](./02-core-concepts.md) explains the runtime model behind sessions, instances, doors, hooks, and state.
- [Template Syntax](./03-template-syntax.md) covers the GoX syntax used throughout the docs.
- [Path Model](./04-path-model.md) and [Router](./05-router.md) take the next step into URL design and request handling.
