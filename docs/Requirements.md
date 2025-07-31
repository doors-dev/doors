# Requirements

## System

* Ensure your [Go](https://go.dev/) is at least `1.24.1`; you can check in the terminal by running:
  ```bash
  go version
  ```

- Install the [templ](https://templ.guide/quick-start/installation) CLI tool (there are LSP and IDE plugins).
- *Optional:* install [wgo](https://github.com/bokwoon95/wgo) for "live" reload 

## Skills

* A basic understanding of HTML and HTTP serving is required.

* **Golang**. 
  Go is considered an easy language to learn. If you are going from other languages, you can jump in right away after a brief overview of the type system, pointers, channels, and generics.

* **JavaScript** is recommended, but not necessary.
  
* I recommend going briefly through the [templ](https://templ.guide) docs. 

  > *For those who are in a hurry: to write HTML components, create .templ files instead of .go, where you can use the magic templ keyword instead of func to enable HTML syntax/templating (everything else like in a regular go file) just straight in a function body (btw don't specify return value). Then use `templ generate` to generate regular .go files.*

