# Get Started


## Install GoX

> Ensure your Go is at least 1.25.1; you can check in the terminal by running `go version`


Doors is build on top of [GoX](https://github.com/doors-dev/gox) - purposely designed Go language extenstion that turns HTML templates into typed Go expressions. 

GoX comes with it's own language server, that mostly acts as Gopls proxy while adding some extra features on top.

Please use the official [VS Code](https://marketplace.visualstudio.com/items?itemName=doors-dev.gox) or [Neovim](https://github.com/doors-dev/nvim-gox) extensions, alternatively you can follow the manual installation guide in [README](https://github.com/doors-dev/gox).


## Setup Project

Create a new directory containing our project:

```bash
mkdir hello-doors
```

Initialize a new Go module there and get Doors framework:

```bash
cd hello-doors
go mod init github.com/doors-dev/doors-examples/hello-doors
go get github.com/doors-dev/doors
```

## Create App Component 

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

## Serve App




