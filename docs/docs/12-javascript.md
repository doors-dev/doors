# JavaScript Integration

For on-page interactivity, extended event handling, and integration with frontend frameworks.

---

## Reference

Magic variables available inside scripts rendered by `doors.Script()`:

| Variable                                                     | Description                                                  |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| `$data(name: string): any`                                   | Retrieve data provided via `doors.AData`.                    |
| `$hook(name: string, arg: any): Promise<any>`                | Call a Go hook (`doors.AHook`), serialize/deserialize input/output. |
| `$fetch(name: string, arg: any): Promise<Response>`          | Call a Go hook (`doors.AHook` or `doors.ARawHook`) and get the raw `Response`. |
| `HookErr`                                                    | Error class thrown by hooks on failure.                      |
| `$on(name: string, handler: (arg: any, err?: HookErr) => any): void` | Register a handler for `ActionEmit` invoked from Go.         |
| `$G: { [key: string]: any }`                                 | Persistent global object for sharing state between scripts.  |
| `$clean(handler: () => void | Promise<void>): void`          | Register a cleanup function to run when the script is removed from the DOM. |

---

## Data Conversion Rules

When calling `$hook` or `$fetch`, the argument is encoded as follows:

| Input type | Request body |
|-------------|---------------|
| `undefined` | No body |
| `FormData` | multipart/form-data |
| `URLSearchParams` | application/x-www-form-urlencoded |
| `Blob` | Raw blob |
| `File` / `ReadableStream` | application/octet-stream |
| Any other value | JSON |

---

## `HookErr`

Errors thrown by `$hook` or `$fetch` are instances of `HookErr`, a subclass of `Error`.

### Error Reasons

- **canceled** — canceled by scope  
- **not_found** — hook expired (404)  
- **unauthorized** — triggers page reload (401, 410)  
- **bad_request** — invalid input (400)  
- **network** — network or transport issue  
- **server** — server error (5XX)  
- **capture** — client-side handling issue  

### Fields

```ts
status?: number        // HTTP status code, if available
cause?: Error          // Original error
kind: string           // One of the above kinds
message: string
```

### Methods

Convenience checks:

```ts
err.canceled()
err.unauthorized()
err.notFound()
err.network()
err.server()
err.badRequest()
err.capture()
err.isOther()
```

> Errors from event binding attributes are ignored unless handled via `OnError` actions.  
> `$hook` and `$fetch` must be caught manually.

---

## Script Component

The `Script` component integrates inline scripts with framework runtime:

- Provides magic variables  
- Converts inline scripts into cacheable, loadable resources  
- Minifies automatically  
- Supports TypeScript (`type="application/typescript"` or `"text/typescript"`)  
- Wraps code in an async IIFE so you can `await` safely

```templ
@doors.Script() {
  <script>
    console.log("hello world!")
  </script>
}
```

⚠️ `type="module"` is **not supported** inside `doors.Script()`.  
For ES modules, use **Imports** and load with `await import("specifier")`.

### Variants

- `doors.Script()` — public, cacheable  
- `doors.ScriptPrivate()` — session-protected, cached  
- `doors.ScriptDisposable()` — session-protected, not cached (for dynamic scripts)

---

## Call Go from JavaScript

### Structured Hook

`doors.AHook` handles Go hooks with JSON (un)marshaling.

```go
type AHook[T any] struct {
  Name      string
  Scope     []Scope
  Indicator []Indicator
  On        func(ctx context.Context, r RHook[T]) (any, bool)
}
```

Example:

```templ
@doors.AHook[string]{
  Name: "length",
  On: func(ctx context.Context, r doors.RHook[string]) (any, bool) {
    return len(r.Data()), false
  },
}

@doors.Script() {
  <script>
    const length = await $hook("length", "hello!")
    console.log(length) // 6
  </script>
}
```

---

### Raw Hook

`doors.ARawHook` provides low-level access to the HTTP body or form data.

```go
type ARawHook struct {
  Name      string
  Scope     []Scope
  Indicator []Indicator
  On        func(ctx context.Context, r RRawHook) bool
}
```

Example:

```templ
@doors.ARawHook{
  Name: "upload",
  On: func(ctx context.Context, r doors.RRawHook) bool {
    file, _, _ := r.FormFile("file")
    defer file.Close()
    // process file
    return false
  },
}

@doors.Script() {
  <script>
    await $fetch("upload", new FormData(document.querySelector("form")))
  </script>
}
```

---

## Call JavaScript from Go

Use `ActionEmit` or the `Call` API to trigger registered JS handlers.

### Register a Handler

```templ
@doors.Script() {
  <script>
    $on("alert", (msg) => {
      alert(msg)
      return true
    })
  </script>
}
```

### Invoke from Go

```go
// Fire-and-forget
doors.Call(ctx, doors.ActionEmit{Name: "alert", Arg: "Hello!"})

// Await a result
ch, cancel := doors.XCall[string](ctx, doors.ActionEmit{Name: "alert", Arg: "Hello!"})
defer cancel()
res := <-ch
if res.Err == nil {
  log.Println("Handler returned:", res.Ok)
}
```