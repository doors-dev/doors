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
	~~
	p.body.Outer(ctx, <div>Prepared before mount</div>)
	~~
	~(&p.body)
}
```

This is useful when the Door was prepared earlier and you just want to place it on the page.

### Containers

Every mounted Door needs a DOM container.

- With `~>(door) <tag>...</tag>`, your tag becomes that container.
- With `~>(door) <>...</>`, **Doors** creates its own container element.
- With `~(&door)`, **Doors** uses the last container from the internal state, or creates its own.

By default, the generated container tag is `d0-r`, and **Doors** styles it with `display: contents`, so it usually does not affect layout.

Use an explicit tag with `~>(door)` when the exact HTML parent matters.

## Methods

```go
Inner(ctx context.Context, content any)
Outer(ctx context.Context, outer any)
Static(ctx context.Context, content any)
Reload(ctx context.Context)
Unmount(ctx context.Context)
Freeze(ctx context.Context)
```

### Inner

`Inner` replaces the Door's children while keeping the current Door container mounted.

```gox
p.body.Inner(ctx, <div id="updated">Updated</div>)
```

Use this when the region should stay mounted and only its contents should change.

Passing `nil` empties the Door while keeping it available for later changes.

### Outer

`Outer` replaces the rendered Door with a new outer element while keeping the same Go `Door` handle live.

Use it when you need to change the root element, attributes, or wrapper structure and still update the Door later.

```gox
p.body.Outer(ctx, <section class="panel is-open">Updated shell</section>)
```

Like `Inner`, `outer` accepts any renderable value — an element, a component, a string — not only markup.

### Static

`Static` makes the Door static: the content replaces the dynamic element on the page. The target Door loses control over the region, is unmounted, and can be garbage collected.

Later method calls still update the Door's stored state for a future render, but they do not automatically put that Door back into the current DOM.

```gox
p.body.Static(ctx, <div id="done">Done</div>)
```

Passing `nil` removes the mounted Door without rendering replacement content.

Static content may mount new **Doors** of its own — that makes `Static` the building block for unbounded feeds and streams, see [Growing Content](#growing-content).

### Reload

`Reload` re-renders the Door's current content.

Use it when the stored content depends on outside state and you want to redraw without swapping in new content.

### Unmount

`Unmount` removes the Door from the DOM but keeps its current content for a future mount.

### Freeze

`Freeze` keeps the Door's current markup on the page but releases everything behind it: hooks, subscriptions, nested Doors, and scoped background work started with `doors.Go(...)`. Interactive elements inside stay visible but silently stop working.

Unlike `Static`, no content is sent — the page keeps what is already there. The Door keeps its stored state and can be mounted again.

Made for content that is dynamic only for a while and then becomes final: blocks in a growing feed, log or chat entries, streamed output. Freeze the finished block, drop the reference, and server memory stays flat as the page grows. For the append side of that pattern, see [Growing Content](#growing-content).

## Growing Content

Content passed to `Static` may itself mount new **Doors**. That turns `Static` into an append primitive built on three ideas:

- **Edge handoff.** Keep one Door as the growth edge. To append, replace the edge with the new item plus a fresh edge. The old edge is released: the page accumulates without bound, while the server keeps one live Door per chain.
- **Pre-mount buffering.** Door methods work on an unmounted Door by writing its stored state, so appends may arrive before the feed ever renders — the chain builds up behind its head Door, and the first render reproduces it whole. Render mounts the chain's head; appends target the edge.
- **The feed is render-once.** Appended items go to the page and are forgotten — the server does not keep them. Consuming the seed drops the last reference to the chain, so a later render cannot restore the items and starts a fresh chain; when the feed must survive an ancestor re-render, render past items from your data source.

```gox
type Feed struct {
	mu   sync.Mutex
	seed *doors.Door
	tail *doors.Door
}

func (f *Feed) Append(ctx context.Context, item gox.Elem) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// append before the first render — seed a head door
	// that will hold the chain until it mounts
	if f.tail == nil {
		f.seed = new(doors.Door)
		f.tail = f.seed
	}
	// replace the edge with the item plus a fresh edge
	next := new(doors.Door)
	f.tail.Static(ctx, <>
		~(item)
		~(next)
	</>)
	f.tail = next
}

elem (f *Feed) Main() {
	~~
	f.mu.Lock()
	defer f.mu.Unlock()
	// consume the seed — the chain renders once,
	// then the server forgets it
	head := f.seed
	f.seed = nil
	// no seed — start a fresh chain from this edge
	if head == nil {
		head = new(doors.Door)
		f.tail = head
	}
	~~
	<section class="feed">~(head)</section>
}
```

One mutex guards `seed` and `tail` across both paths — Door methods are safe from any goroutine, but the fields tracking the chain are yours. Holding the lock across the whole render body is fine: a Door's content render is scheduled on the runtime, not performed inline, so the lock stays short.

The edge's position sets the growth direction: edge after the item grows the chain down (append), edge before the item grows it up (prepend).

For items that stay dynamic for a while — streamed output, edit-in-place entries — mount the item's dynamic part through its own Door and `Freeze` it when the item becomes final.

To pace a fast producer, read the completion channel `Static` returns before appending the next item — see [Completion Channels](#completion-channels).

Inside `<table>`, `<select>`, and other tag-restricted parents, the generated `d0-r` container is invalid and browsers foster-parent it out. Give each edge a real tag when creating it, and render it as usual:

```gox
next := new(doors.Door)
next.Outer(ctx, <tr></tr>)
```

## Completion Channels

Each mutating method returns a completion channel. The return value is optional to use:

```go
Inner(ctx context.Context, content any) <-chan error
Outer(ctx context.Context, outer any) <-chan error
Static(ctx context.Context, content any) <-chan error
Reload(ctx context.Context) <-chan error
Unmount(ctx context.Context) <-chan error
Freeze(ctx context.Context) <-chan error
```

On success the channel sends **two** `nil` values then closes — the first means the call was scheduled (render was completed), the second means it was applied to the page.

On failure, the channel sends an error then closes. An error of `context.Canceled` means the operation was overwritten by a newer Door operation, unmount, or related lifecycle change.

A closed channel with no value means the Door was not mounted by the time the operation was observed.

Do not wait on the channel during rendering.

If you need to wait, do it in a hook, inside `doors.Go(...)`, or in your own
goroutine with `doors.DetachedContext(ctx)`.

`doors.DetachedContext(ctx)` keeps the current Doors ownership and lifecycle,
so it is useful when you want to wait on the channel safely from that same fragment.

If the work should outlive the current dynamic owner, use
`doors.InstanceContext(ctx)`. It switches Doors ownership to the root of the
current instance and uses the instance runtime lifecycle.

> Most code ignores the returned channel. Read it when completion itself matters, such as pacing a fast stream of updates.

## Lifecycle

A Door has two sides:

- stored state on the `doors.Door` value itself
- mounted state on the page

That explains most of its behavior:

1. a new Door starts unmounted
2. you can still call methods on it before it is rendered
3. when the Door is later rendered, it mounts its saved state unless you overwrite it by proxying a container with content, like `~>(door) <div>This content will overwrite whatever was stored in the Door</div>`
4. while mounted, later changes are synchronized to the browser DOM

This means:

- `Inner` before mount stores children that will appear later
- `Inner(ctx, nil)` before mount stores an empty Door
- `Outer` before mount stores a new outer element
- `Static` before mount stores static content instead of a live Door container
- `Static(ctx, nil)` before mount stores an absent state
- `Unmount` removes the Door now but keeps its content for a later mount
- `Freeze` before mount leaves the stored state unchanged

After a Door has been made static, frozen, or unmounted, later calls still update the Door's stored state. They do not automatically put that Door back into the DOM, but they do affect what will happen if the Door is rendered again later.

### Lifecycle Hooks

`doors.OnReady`, `doors.OnSettle`, and `doors.OnClean` attach callbacks to the content being rendered:

```go
doors.OnReady(ctx context.Context, f func(ctx context.Context))
doors.OnSettle(ctx context.Context, on func(ctx context.Context), ops ...func(ctx context.Context))
doors.OnClean(ctx context.Context, f func())
```

`OnReady` fires when the render cycle that produced the surrounding content completes — the HTML is rendered and the update is on its way to the client. Called with a `ctx` whose content is already on the page (an event handler, for example), it fires promptly.

`OnSettle` runs `ops` inside the current dispatch batch and fires once that batch settles: every Doors operation the batch started has been processed and its updates are enqueued. Where `OnReady` tracks one render cycle, `OnSettle` tracks everything the batch started, including the Door updates and beam propagation it triggered. Settling is server-side — the updates are queued, not yet written to the connection or acknowledged by the browser.

`doors.HoldSettle` keeps the current dispatch batch open after the handler returns, until the returned `release` is called. Use it to hand the batch over to a goroutine:

```go
release := doors.HoldSettle(ctx)
go func() {
	defer release()
	// operations started with the handler ctx join the batch
}()
```

While held, the batch does not settle: `OnSettle` callbacks wait, and in a hook the client keeps the indicator, the scope, and the `$hook` promise pending. Call it synchronously in the handler; `release` is idempotent and may run settle callbacks inline. Operations started with the handler `ctx` join the batch, operations started with `doors.DetachedContext` do not. Always call `release` — a held batch otherwise settles only when the instance ends. Outside a batch, `release` is a no-op.

`OnClean` fires when that content is cleared:

- the enclosing Door is updated (`Inner`, `Outer`, `Reload`)
- the Door is removed (`Static`, `Unmount`) or frozen (`Freeze`)
- an ancestor Door re-renders
- the render fails
- the instance ends

They are not symmetric. `OnReady` is **best-effort**: if the render cycle fails or is superseded by a newer Door operation, it never fires. `OnSettle` and `OnClean` are **exactly-once**: a batch always settles one way or another, and every rendered piece of content is eventually cleared. So acquire in render code, release in `OnClean`:

```gox
elem (c Chat) Main() {
	~~
	sub := c.hub.Subscribe()
	doors.OnClean(ctx, func() {
		sub.Close()
	})
	~~
	<div class="chat">Live</div>
}
```

Do not block in any of these callbacks. `OnReady` and `OnSettle` run on the instance goroutine pool with a context equivalent to `doors.DetachedContext`; `OnSettle` runs inline on the calling goroutine when the instance is shutting down or the owner is already canceled. `OnClean` runs inline on framework goroutines and receives no context — if teardown is slow, start your own goroutine with a context captured beforehand (for example from `doors.InstanceContext(ctx)`).

## Use Cases

- Use `Inner` when the Door should stay in place and only its contents should change.
- Use `Inner(ctx, nil)` when the Door should stay alive but become empty.
- Use `Outer` when you need a new root element but want to keep the same Door handle.
- Use `Outer(ctx, nil)` when you need mounted placeholder that does not affect layout (`d0-r`)
- Use `Static` when the region's content is final and the Door's live container is no longer needed.
- Use `Static(ctx, nil)` when the Door should disappear without replacement content.
- Use `Reload` when you want to redraw the current content.
- Use `Unmount` when the Door should disappear for now but keep its internal state for reuse.
- Use `Freeze` when finished content should stay visible but no longer consume server resources.
- Use a `Static` chain when the page should accumulate items — feeds, logs, chats, streams — while the server keeps only the growth edge.

## Related

- [Components](./08-components.md) for component rendering boundaries and state ownership.
- [State](./07-state.md) for `Bind`, `Effect`, sources, and beams.
