# Get Started


## Install GoX

> Ensure your Go is at least 1.25.1; you can check in the terminal by running `go version`


**Doors** is build on top of [GoX](https://github.com/doors-dev/gox) - purposely designed Go language extenstion that turns HTML templates into typed Go expressions and adds `elem` primitives.

**GoX** comes with it's own language server, that mostly acts as [Gopls](https://go.dev/gopls/) proxy while adding some extra features on top.

Please use the official [VS Code](https://marketplace.visualstudio.com/items?itemName=doors-dev.gox) or [Neovim](https://github.com/doors-dev/nvim-gox) extensions, alternatively you can follow the manual installation guide in [README](https://github.com/doors-dev/gox).


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

Write component with page template to `app.gox`

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

Component in **Doors** (and **GoX**) must have `Main()` method that returns `gox.Elem`. Keyword `elem` allows to write html template immidiately in the function body. 

> GoX language server compiles it and manages `.x.go` file automatically! 

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

Start go programm with `go run .` and take a look at [http://localhost:8080](http://localhost:8080)!

#### what just happened:

We declared a **model**. **Doors** uses struct with tagged fields to match, decode and encode path:

```go
type Path struct {
	Home    bool `path:"/"` // first path pattern
	Catalog bool `path:"/catalog/:ID?"`  // second path pattern with optional ID param
    ID      *string
}
```

Next, we added model handler to the router and served our App component:

```go
doors.UseModel(r, func(r doors.RequestModel, s doors.Source[Path]) doors.Response {
    return doors.ResponseComp(App{})
})
```
> `doors.RequestModel` provides access to http data (cookies, headers), while `doors.Source[Path]` is reactive state primitive
with **model** value. Usually you store **Source** to the component field to use it in rendering.

Finaly, **Doors** router just plugs in into Go's standart http server:
```go
 http.ListenAndServe(":8080", r)
```

### Next 
- [Learn GoX syntax](https://doors.dev)
- [Learn about model and routing](https://doors.dev)






