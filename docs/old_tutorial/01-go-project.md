# Initialize the Project

> An alternative tutorial is available [here](https://dev.to/derstruct/go-devs-just-got-superpowers-2lb3).

## 1. Prepare the system

* Ensure your [Go](https://go.dev/) is at least `1.24.1`; you can check in the terminal by running:

  ```bash
  go version
  ```

- Install the [templ](https://templ.guide/quick-start/installation) CLI tool.
- Optional: install wgo for "live" reload 

##  2. Initialize the Go project and add *doors*

1. Initialize a new project.

```bash
mkdir project
cd project
go mod init github.com/derstruct/project # address of your repo (actually can be anything)
```

2. Get the *doors* dependency

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



---

Next: [Home Page](./02-home-page.md)
