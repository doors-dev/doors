# Door

`doors.Door` is the main primitive for dynamic page regions in **Doors**.

A Door lets you render a place in the page now and change just that place later from handlers, subscriptions, or `doors.Go(...)`. It is the tool you reach for when you do not want a whole-page reroute.

**Doors** are often stored on struct fields when the same Door needs to be reused:

```gox
type Panel struct {
	body doors.Door
}
```

## Rendering

There are two common ways to render a Door.


### Proxy

Use `~>(door)` when you want a real element in the template to become the Door container:

```gox
elem (p *Panel) Main() {
	~>(p.body) <div class="panel">
		Initial content
	</div>
}
```

This mounts the Door and seeds it with that subtree.

If the proxied element has no inner content, **Doors** uses the Door's current stored content instead:

```gox
elem (p *Panel) Main() {
	~>(p.body) <div class="panel"></div>
}
```

A single Door can only be mounted in one place at a time. If you render the same Door again somewhere else, **Doors** removes the previous mounted copy from the HTML and the new render becomes the active one.

### Current State

Use `~(&door)` when you want to render the Door's current state directly:

```gox
elem (p *Panel) Main() {
	~{
		p.body.Rebase(ctx, <div>Prepared before mount</div>)
	}
	~(&p.body)
}
```

This is useful when the Door was prepared earlier and you just want to place it on the page.

### Containers

Every mounted Door needs a DOM container.

- With `~>(door) <tag>...</tag>`, your tag becomes that container.
- With `~>(door) <>...</>`, **Doors** creates its own container element.
- With `~(&door)`, **Doors** uses the last container from the internal state, or creates its own.

That generated container tag is `d0-r`, and **Doors** styles it with `display: contents`, so it usually does not affect layout.

Use an explicit tag with `~>(door)` when the exact HTML parent matters.

## Methods

```go
Reload(ctx context.Context)
Update(ctx context.Context, content any)
Rebase(ctx context.Context, el gox.Elem)
Replace(ctx context.Context, content any)
Delete(ctx context.Context)
Unmount(ctx context.Context)
Clear(ctx context.Context)
```

### Update

`Update` replaces the Door's children while keeping the Door in the same place.

```gox
p.body.Update(ctx, <div id="updated">Updated</div>)
```

Use this when the region should stay mounted and only its contents should change.

### Clear

`Clear` is `Update(ctx, nil)`.

It empties the Door but keeps it available for later updates.

### Reload

`Reload` re-renders the Door's current content.

Use it when the stored content depends on outside state and you want to redraw without swapping in new content.

### Replace

`Replace` swaps the Door itself out for other rendered content, which effectively makes it static.

```gox
p.body.Replace(ctx, <div id="replacement">Replacement</div>)
```

After `Replace`, the original Door is no longer the mounted node.

### Delete

`Delete` removes the Door and forgets its content.

It is equivalent to replacing with `nil`.

### Unmount

`Unmount` removes the Door from the DOM but keeps its current content.

This is the important difference from `Delete`:

- `Delete` removes the Door and forgets its content
- `Unmount` removes the Door but keeps its content for a future mount

### Rebase

`Rebase` not only changes the inner content, but also recomputes and replaces the container. It is like `Replace`, but keeps the result dynamic.

## X Methods

Each mutating method also has an `X*` variant:

```go
XReload(ctx context.Context) <-chan error
XUpdate(ctx context.Context, content any) <-chan error
XRebase(ctx context.Context, el gox.Elem) <-chan error
XReplace(ctx context.Context, content any) <-chan error
XDelete(ctx context.Context) <-chan error
XUnmount(ctx context.Context) <-chan error
XClear(ctx context.Context) <-chan error
```

These report completion:

- `nil` means the operation completed
- a non-nil error means it failed or canceled before finishing (for example, parent is unmounted)
- a closed channel with no value usually means the Door was not mounted by the time the operation was observed

Do not wait on `X*` during rendering.

If you need to wait, do it in a hook, inside `doors.Go(...)`, or in your own
goroutine with `doors.Free(ctx)`.

`doors.Free(ctx)` keeps the original context values, but switches to the root
Doors context and extends cancellation/deadline/lifetime to the instance
runtime. That makes it the right context for long-running goroutines and for
waiting on `X*` completion safely.

Most code should use the regular methods. Reach for `X*` when completion itself matters, such as pacing a fast stream of updates.

## Lifecycle

A Door has two sides:

- stored state on the `doors.Door` value itself
- mounted state on the page

That explains most of its behavior:

1. a new Door starts unmounted
2. you can still call methods on it before it is rendered
3. when the Door is later rendered, it mounts its saved state unless you overwrite it by proxying a container with content, like `~>(door) <div>This content will overwrite whatever was stored in the Door</div>`
4. while mounted, later updates are synchronized to the browser DOM

This means:

- `Update` before mount stores content that will appear later
- `Clear` before mount stores an empty Door
- `Replace` before mount stores replacement content instead of the original container
- `Delete` before mount stores an absent state
- `Unmount` removes the Door now but keeps its content for a later mount

After a Door has been replaced, deleted or unmounted, later calls still update the Door's stored state. They do not automatically put that old Door back into the DOM, but they do affect what will happen if the Door is rendered again later.

## Use Cases

- Use `Update` when the Door should stay in place and only its contents should change.
- Use `Clear` when the Door should stay alive but become empty.
- Use `Reload` when you want to redraw the current content.
- Use `Replace` when the Door itself should be replaced by other content.
- Use `Delete` when the Door should disappear and forget its previous content.
- Use `Unmount` when the Door should disappear for now but keep its internal state for reuse.
- Use `Rebase` when you need a new root element but want to keep the same Door handle.
