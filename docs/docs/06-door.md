# Door

`doors.Door` controls a dynamic **container** in the DOM tree that can be updated, replaced, or removed at runtime. It is a fundamental building block of *doors* framework, enabling reactive HTML updates without a virtual DOM.

By default, `doors.Door` does not affect layout (it is a custom element with `display: contents`). However, some HTML tags expect specific children (like `<table>`). You can use any tag as a *door* by setting the `Tag` field:

```templ
door := doors.Door{
  Tag: "tr", // usable inside <table> or <tbody>
  A: [string]any{"class": "row"}, // attributes (id will be overwritten)
}
```

✅ Specify the tag only if necessary.

---

## Lifecycle

- A `doors.Door` object always remains alive. It can be attached to or detached from the DOM.
- When detached, you can still call `Update`, `Remove`, `Replace`, or `Clear`. The latest state will apply at render time.
- If replaced or removed before rendering, it stays detached even after render.
- If untouched, updated, or cleared before render, it becomes attached after rendering.
- Once attached, its methods affect the DOM directly. `Remove` or `Replace` make it detached again.
- Re-rendering while still attached removes the previous container from the DOM completely.

---

## API

### Update

Updates content inside the **container**.

```templ
func (c *MyComponent) handleClick(ctx context.Context) {
  c.contentDoor.Update(ctx, c.newContent())
}

templ (c *MyComponent) newContent() {
  <p>Updated content at { time.Now().Format("15:04:05") }</p>
}
```

### Replace

Replaces the **container** with static content. The door becomes detached (static) until `Update` or `Clear`.

```templ
c.door.Replace(ctx, c.staticContent())
```

### Clear

Empties the **container** (equivalent to `Update(ctx, nil)`).

```templ
c.door.Clear(ctx)
```

### Remove

Removes the door’s container and content from the DOM (equivalent to `Replace(ctx, nil)`).

```templ
c.door.Remove(ctx)
```

### Rendering With Children

```templ
c.door {
  @content()
}
```

is equivalent to:

```templ
{{ c.door.Update(ctx, content()) }}
@c.contentDoor
```

---

## Extended API

```templ
XReload(ctx context.Context) <-chan error
XUpdate(ctx context.Context, content templ.Component) <-chan error
XReplace(ctx context.Context, content templ.Component) <-chan error
XRemove(ctx context.Context) <-chan error
XClear(ctx context.Context) <-chan error
```

These return a channel that reports the operation’s completion status.

- If the channel **closes without a value**, the door was detached.
- If it returns **nil**, the frontend applied the change successfully.
- If it returns a **non-nil error**, the operation failed (e.g., overwritten before completion).

> Avoid using extended APIs in non-blocking render contexts; waiting on channels can cause deadlocks. Use them within `@doors.Go(func(context.Context))` goroutines.

---

## Example

```templ
type noticeFragment struct {
  msg doors.Door
}

templ (n *noticeFragment) Render() {
  <div>
    @n.msg {
      Press the button!
    }

    @doors.AClick{
      On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
        n.msg.Update(ctx, doors.Text("Hello there"))
        return false
      },
    }
    <button>Show message</button>
  </div>
}
```

```templ
templ content() {
  @doors.Go(func(context.Context) {
    for {
      err, ok := <-door.XUpdate(ctx, content())
      if !ok {
        break
      }
      if err != nil {
        break
      }
      // do something after successful update
    }
  })
}
```
