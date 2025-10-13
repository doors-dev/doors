# Actions

Actions are client-side operations triggered from the backend.  
They can be dispatched in different contexts:

- Directly via the Calls API (`doors.Call(...)`, `doors.XCall(...)`)
- In a hook lifecycle **after** the request (`r.After(...)`)
- In a hook lifecycle **before** the request (`Before` field on attributes)
- In a hook lifecycle on **error** (`OnError` field on attributes)

  > [error object](./12-javascript.md) is provided as second argument to the handler function 

---


## `ActionEmit`

Invokes a client-side handler registered with  
`$on(name: string, func: (arg: any, err?: HookErr) => any)`.

```go
type ActionEmit struct {
  Name string // Handler name
  Arg  any    // Argument passed to handler
}
```

Helper:

```go
ActionOnlyEmit(name string, arg any) []Action
```

### Emit Resolution Rules

Handler resolution is scoped to the closest dynamic parent (Door).  
When invoking:

- A hook can only call handlers **in its parent door or above**  
- Inner handlers shadow outer ones if names collide

---

## `ActionLocationReload`

Reloads the current page.

```go
type ActionLocationReload struct{}
```

Helper:

```go
ActionOnlyLocationReload() []Action
```

---

## `ActionLocationReplace`

Replaces the current location with a URL derived from a model.

```go
type ActionLocationReplace struct {
  Model any // Path model
}
```

Helper:

```go
ActionOnlyLocationReplace(model any) []Action
```

---

## `ActionLocationAssign`

Navigates to a URL derived from a model.

```go
type ActionLocationAssign struct {
  Model any // Path model
}
```

Helper:

```go
ActionOnlyLocationAssign(model any) []Action
```

---

## `ActionScroll`

Scrolls to the first element matching the CSS selector.

```go
type ActionScroll struct {
  Selector string // CSS selector
  Smooth   bool   // Smooth scrolling if true
}
```

Helper:

```go
ActionOnlyScroll(selector string, smooth bool) []Action
```

---

## `ActionIndicate`

Applies visual indicators for a fixed duration.

```go
type ActionIndicate struct {
  Indicator []Indicator   // Indicators to apply
  Duration  time.Duration // How long they remain active
}
```

Helper:

```go
ActionOnlyIndicate(indicator []Indicator, duration time.Duration) []Action
```

