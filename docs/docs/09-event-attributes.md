# Event Hook Attributes

Event hook attributes enable and configure backend handlers for DOM events. Each attribute is a struct rendered before or spread onto an element to bind a typed `On` handler and optional client-side behavior.

---

## Common Options

Most event hook attributes share these fields:

- `StopPropagation` · stop event bubbling.
- `PreventDefault` · prevent the browser default.
- `ExactTarget` · fire only on the element itself.
- `Scope []Scope` · control scheduling (blocking, debounce, etc.).
- `Indicator []Indicator` · UI indication while running.
- `Before []Action` · actions before request.
- `OnError []Action` · actions on error.
- `On func(context.Context, REvent[T]) bool` · backend handler, return `true` to complete and remove the hook. Type `T` depends on the event.

Render inline:

```templ
@doors.AClick{
  PreventDefault: true,
  On: func(ctx context.Context, ev doors.REvent[doors.PointerEvent]) bool {
    /* logic */
    return false
  },
}
<button>Click</button>
```

Or spread:

```templ
{{ click := doors.AClick{ On: func(ctx context.Context, ev doors.REvent[doors.PointerEvent]) bool { return true } } }}
<button { doors.A(ctx, click)... }>Click</button>
```

---

## Available Attributes

### Pointer events
- `AClick`
- `APointerDown`, `APointerUp`, `APointerMove`
- `APointerOver`, `APointerOut`, `APointerEnter`, `APointerLeave`
- `APointerCancel`
- `AGotPointerCapture`, `ALostPointerCapture`

All use `REvent[PointerEvent]`. Support propagation and default control.

### Focus events
- `AFocus`, `ABlur`  
  Use `REvent[FocusEvent]`. Basic capture.
- `AFocusIn`, `AFocusOut`  
  Support `StopPropagation` and `ExactTarget`.

### Keyboard events
- `AKeyDown`, `AKeyUp`  
  Use `REvent[KeyboardEvent]`. Support `Filter []string` on `event.key`, plus propagation/default options.

### Input and change
- `AInput` · `REvent[InputEvent]` with optional `ExcludeValue`.
- `AChange` · `REvent[ChangeEvent]` for committed value changes.

### Forms
- `ASubmit[T]` · multipart parsing and decoding into `T` (via go-playground/form v4). Optional `MaxMemory`. Handler gets `RForm[T]`.
- `ARawSubmit` · raw multipart access for streaming/uploads. Handler gets `RRawForm`.

---

## Concurrency

Within a single hook, backend `On` calls are serialized. Multiple events on the same handler queue on the backend. Reusing a pre-initialized attribute across elements joins their queue.

```templ
@doors.ASubmit[Data]{
  On: func(ctx context.Context, r doors.RForm[Data]) bool {
    /* safe, next call runs after this returns */
    return false
  },
}
<form>...</form>
```

---

## Scopes

Use scopes for precise scheduling, such as debounce and blocking.

```templ
@doors.AInput{
  Scope: doors.ScopeOnlyDebounce(300 * time.Millisecond, time.Second),
  On: func(ctx context.Context, r doors.REvent[doors.InputEvent]) bool { return false },
}
```

See [Scopes](./16-scopes.md) for details.

---

## Indicators

Apply pending indication with the `Indicator` API.

```templ
@doors.ASubmit[loginData]{
  Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
  Scope:     doors.ScopeOnlyBlocking(),
  On: func(ctx context.Context, r doors.RForm[loginData]) bool { return true },
}
```

See [Indication](./15-indication.md) for details.

---

## Errors

`OnError` runs when a hook processing error occurs. It supports UI indication, navigation, and custom JS actions. See [Actions](./14-actions.md) and [JavaScript](./12-javascript.md) for error object details.

