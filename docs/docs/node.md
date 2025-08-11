## Node

`Node` controls a dynamic **container** in the DOM tree that can be updated, replaced, or removed at runtime. It is a fundamental building block of *doors* framework, enabling reactive HTML updates without a virtual DOM.

Nodes act as insertion points in your templates where content can be dynamically manipulated from Go code. All changes made to a Node are automatically synchronized with the frontend DOM.

## Lifecycle

 `doors.Node` object always remains alive. However, it can be attached to or detached from the DOM.

If Node is detached, you can still `Update`, `Remove`, `Replace`, or `Clear`  it. When rendering occurs, only the latest state will be applied.

If node `replaced` or `removed` before rendering, it stays detached after render (acts like a static component) 

If the node is untouched, `updated`, or `cleared` before rendering, it becomes attached after.

After the `Node` becomes attached, calling its methods will affect the DOM. `Remove, Replace` methods make the `Node` detached again.

If the `Node` is rendered a second time while still being attached, the previous **container** will be completely removed from the DOM. 

## API

### Update

`Update` changes the content inside the **container**.

```templ
func (c *MyComponent) handleClick(ctx context.Context) {
    c.contentNode.Update(ctx, c.newContent())
}

templ (c *MyComponent) newContent() {
    <p>Updated content at { time.Now().Format("15:04:05") }</p>
}
```

### Replace

`Replace` replaces the  **container**  with static content. The node behaves like a static component after replacement (until `Update` or `Clear`) and loses control over its initial DOM position (becomes detached).

```go
c.contentNode.Replace(ctx, c.staticContent())
```

### Clear

`Clear` empties the Nodes  **container**. Equivalent to calling `Update(ctx, nil)`

```go
c.contentNode.Clear(ctx)
```

### Remove

`Remove` removes the Node container with content from the DOM.  Equivalent to calling `Replace(ctx, nil)`

```go
c.contentNode.Remove(ctx)
```

### Rendering With Children

```templ
c.contentNode {
	@content()
}
```

is equivalent to

```templ
{{ 
c.contentNode.Update(ctx, content())
}}
@c.contentNode
```

## Extra API

```
XReload(ctx context.Context) <-chan error
XUpdate(ctx context.Context, content templ.Component) <-chan error
XReplace(ctx context.Context, content templ.Component) <-chan error
XRemove(ctx context.Context) <-chan error
XClear(ctx context.Context) <-chan error
```

Do the same, but return a channel that can be used to track operation progress.

* If the channel closes without a value, it means the Node is detached.

* If the channel returns nil, it means the frontend successfully applied the DOM change. 

* If a channel returns a non-nil value (error), it indicates that an error occurred during the operation (for example, if the operation was overwritten by a new one before it was applied).

  > It's not recommended to use extended APIs outside a blocking-allowed context, because waiting on a channel in the render runtime environment can cause a deadlock. 
  > Use inside @doors.Go(func(context.Context)) (gorourine independent from the render runtime)

## Example

```templ
type noticeFragment struct {
  msg doors.Node
}

templ (n *noticeFragment) Render() {
  <div>
    @n.msg {
    		Press the button!
    }
    <button { doors.A(ctx, n.show())... }>Show message</button>
  </div>
}

func (n *noticeFragment) show() doors.Attr {
  return doors.AClick{
    On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
      n.msg.Update(ctx, doors.Text("Hello there"))
      // n.msg.Remove(ctx)
      // n.msg.Replace(ctx, doors.Text("Hello there"))
      return false
    },
  }
}

```

```templ
templ content() {
  // spawn goroutine with blocking-allowed context
  @doors.Go(func (context.Context) {
    for {
      err, ok := <- node.XUpdate(ctx, content())
      if !ok {
        break
      }
      if err != nil {	
        break
      }
      // do something
    }
  })
}
```



