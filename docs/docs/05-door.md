# Door

`doors.Door` is the core primitive for dynamic DOM regions in Doors.

A Door lets you render a placeholder now and change it later from event handlers, reactive subscriptions, or lifecycle goroutines. Unlike full-page rerouting, a Door updates only its own part of the page.

## Basics

A `doors.Door` value is stateful. It keeps track of its current virtual content and, when mounted, synchronizes that content with the browser DOM.

You usually keep Doors as struct fields:

```gox
type Panel struct {
	body doors.Door
}
```

## Rendering

In GoX, a Door can be used in two related ways.

### Mount

Use `~>(door)` as a proxy over the following subtree:

```gox
elem (p *Panel) Main() {
	~>(p.body) <div>
		Initial content
	</div>
}
```

This mounts the Door and seeds it with that subtree.

### Current State

Use `~(&door)` to render the Door as a component:

```gox
elem (p *Panel) Main() {
	~{
		p.body.Update(ctx, <div>Prepared before mount</div>)
	}
	~(&p.body)
}
```

This is useful when the Door already has virtual state prepared before it is rendered.

## Containers

Every mounted Door needs a DOM container that Doors can track.

There are two cases:

- If you mount a Door with `~>(door) <tag>...</tag>`, your element becomes the Door container.
- If you render a Door with `~(&door)`, or mount it with `~>(door) <>...</>`, there is no user-provided root element, so Doors creates its own container element.

The internal fallback container tag is `d0-r`. Doors styles that element with `display: contents`, so by default it does not affect layout.

That means:

- Use `~>(door) <div>...</div>` when you want a specific real element to be the mounted Door node.
- Use `~>(door) <>...</>` when you want to seed Door content but do not want to provide a real wrapper element. In that case Doors creates `d0-r`.
- Use `~(&door)` when you only want to render the Door's current state and you do not care about providing the outer container yourself.
- For markup where the exact HTML parent matters, prefer giving the Door an explicit element with `~>(door)`.

## Lifecycle

Doors have both virtual state and mounted state.

Virtual state is the content and lifecycle state stored on the `doors.Door` value itself. Mounted state is whether that Door currently has a live DOM node on the page.

The lifecycle works like this:

1. A new Door starts unmounted, but it can still accept operations.
2. If you call `Update`, `Clear`, `Replace`, `Delete`, or `Unmount` before the Door is rendered, Doors stores that result as virtual state.
3. When the Door is later rendered, Doors uses that stored state to decide what to mount.
4. While mounted, further operations are synchronized to the browser DOM.
5. Some operations keep the Door alive for future reuse, while others reset or detach it more aggressively.

The practical meaning of each state transition is:

- `Update` before mount stores content that will appear when the Door is rendered later.
- `Clear` before mount stores an empty Door.
- `Replace` before mount stores replacement content, so the original Door container is not what ends up rendered.
- `Delete` before mount stores an absent state, so rendering the Door later still shows nothing.
- `Unmount` removes the Door from the DOM but keeps its current content, so a future render can mount that content again.

After a Door has already been replaced, deleted, or otherwise stopped being the mounted node, later operations still apply to the `doors.Door` value itself.

- Calling `Update` after `Replace` or `Delete` updates the Door's virtual state. It does not magically put that Door back into the DOM, but the new content will be used the next time that Door is rendered or mounted somewhere again.
- Calling `Unmount` after `Replace` or `Delete` has no visible immediate effect, because that Door is already not mounted.
- Calling `Clear` after `Replace` or `Delete` clears the Door's stored content, so if it is rendered again later it comes back empty.

This is why Doors work well for deferred fragments, conditional UIs, and long-lived stateful components.

## Methods

All non-`X` methods are fire-and-forget:

```
Reload(ctx context.Context)
Update(ctx context.Context, content any)
Rebase(ctx context.Context, el gox.Elem)
Replace(ctx context.Context, content any)
Delete(ctx context.Context)
Unmount(ctx context.Context)
Clear(ctx context.Context)
```

### Update

Frontend synchronization is automatic for these methods. In normal code, you should treat the Door as already updated immediately after the call and continue your logic from that new state.

The browser-side patch is still delivered asynchronously under the hood, but the Doors programming model is designed so you usually do not need to wait for it.

`Update` changes the Door's children while keeping the same mounted Door node.

```gox
<button
	(doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			p.body.Update(ctx, <div id="updated">Updated</div>)
			return false
		},
	})>
	Update
</button>
```

Use this when you want the same Door position to stay in the DOM.

### Clear

`Clear` is the same as `Update(ctx, nil)`.

It empties the Door's children but keeps the Door itself alive for future updates.

### Reload

`Reload` re-renders the Door's current content.

This is mainly useful when the stored content depends on outside state and you want to redraw it without swapping in a new value.

### Replace

`Replace` swaps the Door itself out for other rendered content.

```gox
type Panel struct {
	body  doors.Door
	other doors.Door
}

elem (p *Panel) replacement() {
	<div id="replacement">Replacement</div>
}

elem (p *Panel) Main() {
	~{
		p.other.Update(ctx, p.replacement())
	}
	<button
		(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				p.body.Replace(ctx, &p.other)
				return false
			},
		})>
		Replace
	</button>
}
```

After `Replace`, the original Door is no longer the mounted node.

### Delete

`Delete` removes the Door completely and resets its content.

It is equivalent to replacing with `nil`.

After `Delete`, a later `XUpdate` can simply close without a value if the Door is not mounted anymore.

### Unmount

`Unmount` removes the Door from the DOM but preserves its current content.

This is the key difference from `Delete`.

- `Delete` removes the Door and forgets its content.
- `Unmount` removes the Door but keeps its content for a future mount.

### Rebase

`Rebase` swaps the Door to a new root element while keeping the same `Door` object live.

It takes `gox.Elem`, not `any`, because it replaces the actual Door container with a newly rooted dynamic subtree.

## Blocking Variants

Each mutating operation also has an `X*` variant:

```
XReload(ctx context.Context) <-chan error
XUpdate(ctx context.Context, content any) <-chan error
XRebase(ctx context.Context, el gox.Elem) <-chan error
XReplace(ctx context.Context, content any) <-chan error
XDelete(ctx context.Context) <-chan error
XUnmount(ctx context.Context) <-chan error
XClear(ctx context.Context) <-chan error
```

These return a channel that reports completion.

- If you receive `nil`, the operation completed successfully.
- If you receive a non-nil error, the operation failed.
- If the channel closes without a value, the Door was not mounted by the time the operation was observed.

### X Usage

Do not block on `X*` directly during rendering.

If you need to wait for completion, do it inside an event handler or inside `doors.Go`, which gives you a Door-lifecycle-aware context:

```gox
elem (p *Panel) Main() {
	~(doors.Go(func(ctx context.Context) {
		for {
			select {
			case <-time.After(time.Second):
				err, ok := <-p.body.XUpdate(ctx, <div>Tick</div>)
				if !ok || err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}))

	~(&p.body)
}
```

Most code should use the fire-and-forget methods. Reach for `X*` only when completion itself matters.

One important use case is pacing a fast producer so you do not overwhelm the system with pending updates. If you are streaming very frequent real-time data, waiting for `XUpdate` lets you send the next frame only after the previous one has actually completed, instead of blindly queueing more work.

When plain fire-and-forget updates overlap, Doors does best-effort synchronization toward the latest state. In practice, if a newer update arrives before an older one finishes rendering, Doors prefers delivering the newest version rather than forcing every intermediate state to be rendered.

## Choosing

- Use `Update` when the Door should stay in place and only its children should change.
- Use `Clear` when the Door should stay mounted but empty.
- Use `Replace` when the Door itself should be replaced by different content.
- Use `Delete` when the Door should disappear and forget its previous content.
- Use `Unmount` when the Door should disappear for now but keep its content for later reuse.
- Use `Reload` when you want to redraw the current content.
- Use `Rebase` when you need a new dynamic root while keeping the same `Door` handle.

## Notes

- A Door is best stored on a long-lived struct, not recreated on every update.
- `doors.Go` is the right place for long-running work tied to a Door's lifetime.
- `Update`, `Replace`, and friends accept `any`, so you can pass GoX elements, components, strings, or another `Door`.
- `return false` in click handlers is a good default when no extra client navigation is needed.
