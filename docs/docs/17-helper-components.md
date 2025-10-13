# Helper Components

Small, composable building blocks for dynamic rendering, scripts/styles, and reactive flows.

---

## Include

```go
func Include() templ.Component
```

Injects framework client script and styles. Place in `<head>`.

---

## Fragment and `F`

```go
type Fragment interface{ Render() templ.Component }
func F(doors.Fragment) templ.Component
```

Render a struct that implements `Render()`.

```templ
templ Demo() { @doors.F(&Counter{}) }
```

---

## Reactive Helpers

### `Sub`

```go
func Sub[T any](beam Beam[T], render func(T) templ.Component) templ.Component
```

Subscribe to a Beam and re-render on updates.

### `Inject`

```go
func Inject[T any](key any, beam Beam[T]) templ.Component
```

Writes the Beam value into `context.Context` for children. Re-renders on updates.

### `If`

```go
func If(beam Beam[bool]) templ.Component
```

Conditional rendering based on a Beam.

> It's not recommended to use multiple `If` to change the content. **Only show-hide single block**. For content switching  use `Sub` or `Inject` helpers to guarantee smooth behavior.

---

## Evaluate and Run

```go
func E(func(context.Context) templ.Component) templ.Component
func Run(func(context.Context)) templ.Component
```

Evaluate to a component at render time, or run a side-effect during render.

---

## Goroutine Spawn

```go
func Go(func(context.Context)) templ.Component
```

Start a goroutine tied to the component lifecycle. Context is cancelable on unmount.

---

## Script

```go
func Script() templ.Component
func ScriptPrivate() templ.Component
func ScriptDisposable() templ.Component
```

Convert inline `<script>` into an external, processed resource (via esbuild and async wrapping).  
Private variants scope delivery to the session; disposable disables caching.

---

## Style

```go
func Style() templ.Component
func StylePrivate() templ.Component
func StyleDisposable() templ.Component
```

Convert inline `<style>` into an external, minified resource.  
Private variants scope to session; disposable disables caching.

Example:

```templ
@doors.Style() {
  <style>
    body { background-color: powderblue; }
    h1 { color: blue; }
    p { color: red; }
  </style>
}
```

---

## Text

```go
func Text(any) templ.Component
```

Render any value as escaped text.

---

## Attributes

```go
func Attributes([]Attr) templ.Component
```

Renders slice of attributes sequentialy

---

## Any

```go
func Any(any) templ.Component
```

Type-directed rendering:

- `templ.Component` → render directly  
- `doors.Fragment` → via `F()`  
- `[]templ.Component` → sequential render  
- `[]Attr` → renders attributes
- `func(context.Context) templ.Component` → via `E()`  
- `func(context.Context)` → via `Run()`  
- otherwise → `Text()`

---

## Head Metadata

```go
type HeadData struct {
  Title string
  Meta  map[string]string
}

func Head[M any](b Beam[M], cast func(M) HeadData) templ.Component
```

Render `<title>` and `<meta>` tags that update reactively when the Beam changes.

Example:

```templ
@doors.Head(pathBeam, func(p Path) doors.HeadData {
  return doors.HeadData{
    Title: "Product: " + p.Name,
    Meta: map[string]string{
      "description": "Buy " + p.Name,
      "keywords": p.Name + ", product, buy",
    },
  }
})
```

---

## HTTP Status

```go
func SetStatus(ctx context.Context, statusCode int)
func Status(statusCode int) templ.Component
```

Set HTTP response status code from templates. Takes effect only on the first render.
