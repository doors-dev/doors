# Binding Attributes

Binding attributes connect Go logic and runtime data to JavaScript.  
They allow two-way data exchange and data binding between the backend and client runtime.

---

## `AData`

Expose server-side values to JavaScript via `$data(name)`.

```go
type AData struct {
  // Name of the data entry to read via JavaScript with $data(name).
  Name  string
  // Value to expose to the client. Marshaled to JSON.
  Value any
}
```

Example:

```templ
@doors.Script() {
  @doors.AData{
    Name:  "user_profile",
    Value: user, // Go struct or map
  }
  <script>
    const user = $data("user_profile")
    console.log(user.name)
  </script>
}
```

You can also define multiple entries at once with `ADataMap`:

```go
type ADataMap map[string]any
```

```templ
@doors.ADataMap{
  "config": config,
  "user":   user,
}
```

---

## `AHook`

Bind a Go function callable from JavaScript via `$hook(name, arg)`.

```go
type AHook[T any] struct {
  Name      string
  Scope     []Scope
  Indicator []Indicator
  On        func(ctx context.Context, r RHook[T]) (any, bool)
}
```

- Input `arg` is sent as JSON, unmarshaled into `T`.  
- Return value is marshaled back to JSON.  
- Return `true` to remove the hook after completion.  

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
    const len = await $hook("length", "hello")
    console.log(len) // 5
    // or await $fetch("length", "hello")
    // to deal with a raw fetch response
  </script>
}
```

---

## `ARawHook`

Low-level variant of `AHook` that provides direct access to the HTTP request and body streams.

```go
type ARawHook struct {
  Name      string
  Scope     []Scope
  Indicator []Indicator
  On        func(ctx context.Context, r RRawHook) bool
}
```

Use for file uploads, raw binary streams, or custom encodings.

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
```

And from JavaScript:

```js
await $hook("upload", new FormData(document.querySelector("form")))
```

---

## Notes

- Hooks and data bindings are scoped to the closest door parent lifecycle.  
- `$hook` calls serialize automatically and return a `Promise`.  
- `$fetch` bypasses JSON and works with any streamable payload.  
- Both APIs are available only inside `<script>` rendered by `doors.Script()`.