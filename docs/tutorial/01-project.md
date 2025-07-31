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
go mod init github.com/account/project # address of your repo (actually can be anything)
```

2. Get *doors* library

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

