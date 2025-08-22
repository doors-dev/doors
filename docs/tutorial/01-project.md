# Init Project

## 1. System

* Ensure your [Go](https://go.dev/) is at least `1.24.1`; you can check in the terminal by running:

  ```bash
  go version
  ```

- Install the [templ](https://templ.guide/quick-start/installation) CLI tool.
- *Optional:* install [wgo](https://github.com/bokwoon95/wgo) for "live" reload 

##  2. Init Project

1. Initialize a new project.

*Skip if you already have one.*

```bash
mkdir project
cd project
go mod init github.com/derstruct/project # address of your repo (actually can be anything)
```

> You can find the code from this tutorial here [https://github.com/derstruct/doors-tutorial](https://github.com/derstruct/doors-tutorial). Branch name ~= part name.

2. Get *doors* dependency

```bash
go get github.com/doors-dev/doors
```

## 3. Hello World

Create `./main.go`

```go
package main

func main() {
  println("hello world")
}

```

Run:

```bash
$ go run .
hello world
```

