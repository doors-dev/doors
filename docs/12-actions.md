# Actions

Actions are client-side effects triggered from Go.

Use them when the browser should do something imperative instead of just rendering different HTML.

For example, call a JavaScript handler registered with `$on(...)`.

> If the UI should simply render different content, prefer normal rendering.
> If attributes should stay shared across existing elements, `Setter` from [Element Handles](./17-element-handles.md) is often a better fit than a custom action.

## Places

You can schedule actions in five common places:

- `doors.Call(ctx, action)` to dispatch from Go; the returned completion channel is optional to use
- `Before` on a request attr such as an event attr or `ALink`, just before the request is sent
- `r.After(...)` after a successful request
- `After` on `ALink`, once the navigation succeeds
- `OnError` on a request attr when a client-visible hook error happens

Action lists run in the order you give them.

`OnError` is for normal client-visible failures such as network, server, bad request, and similar hook errors.

It does not run for scope cancellations or expired hooks, and a stopped instance is handled by reloading the page instead.

## Direct

In most cases, ignore the returned channel. There is nothing to check: whether the action runs depends on the lifetime of the content it was called from, and that is managed for you.

```go
doors.Call(ctx, doors.ActionLocationReload{})
```

When the client handler should return a value to Go, capture it with `Into`. The channel first sends `nil` once the call is scheduled, or an error if the action cannot be built. On a detached context a second value follows: `nil` once the browser has answered and the destination holds the value, or an error if the client failed or the result could not be decoded. On any other context the channel closes after the first value and the destination stays unwritten; the closed channel is the only signal, there is no error. If the call is canceled, the channel sends `context.Canceled`. See [Detached Context](./19-session-instance.md#detached-context).

Never wait on the channel during render or in beam subscriptions. When you need the answer, wait for the second value in a goroutine of your own with `doors.DetachedContext(ctx)`; only a detached context delivers it:

```go
On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
	go func(ctx context.Context) {
		var picked string
		ch := doors.Call(ctx, doors.ActionEmit[string]{
			Name: "pick",
			Arg:  "hello",
		}.Into(&picked))
		<-ch // scheduled
		if err, ok := <-ch; !ok || err != nil {
			return
		}
		result.Inner(ctx, Picked{Value: picked})
	}(doors.DetachedContext(ctx))
	return false
},
```

Use `doors.Go(...)` when the wait should follow the lifetime of a rendered subtree, and `doors.InstanceContext(ctx)` when it should outlive the current owner and live with the instance. See [Detached Context](./19-session-instance.md#detached-context).

Canceling `ctx` requests best-effort cancellation; on a detached context, a call canceled after it was scheduled sends `context.Canceled` as its second value.

A Door update and a `doors.Call` made in the same handler are not ordered against each other unless the call is made from that Door's content. Targeting the content of another Door by selector needs care.

`ActionEmit[T]` declares its result type; `T` is what `Into` decodes into. For
fire-and-forget emits, use `ActionEmit[any]` and skip `Into`. `Setter.Set`
supports `Into(&count)` to capture the number of affected elements. `Emitter`
events support `Into(&count)` to capture the number of hook requests the
emitted events triggered.

## Emit

`ActionEmit` calls a client handler registered with `$on(name, handler)`.

```gox
<>
	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				doors.Call(ctx, doors.ActionEmit[any]{
					Name: "alert",
					Arg:  "Hello!",
				})
				return false
			},
		})>
		Alert
	</button>

	<script>
		$on("alert", (message) => {
			window.alert(message)
			return "ok"
		})
	</script>
</>
```

A handler registered by a managed script rendered in the same update may not be ready when the emit arrives; see [Emit and cleanup](./15-javascript.md#emit-and-cleanup).

Handler search is scoped through the Door tree.

**Doors** starts from the Door where the action was created and walks outward through parent Doors until it finds a matching handler.

That means:

- the nearest matching handler wins
- local handlers shadow outer handlers with the same name
- handlers outside that Door ancestry are not visible
- if no handler is found, the action fails

`$on(...)` handlers may return a `Promise`.

When `ActionEmit` is triggered from `OnError`, the handler receives the hook error as its second argument: `(arg, err)`.

## Location

Location actions are hard navigations.

They go through the browser location API and load the target page again.

That makes them useful when you intentionally want a full page load.

For normal in-app navigation, prefer [Navigation](./09-navigation.md), especially `ALink` or updating the `Source` a `RouteSource` receives inside `RouteModel`.

Built-ins:

- `doors.ActionLocationAssign{Model: ...}` pushes a new history entry and loads that URL
- `doors.ActionLocationReplace{Model: ...}` replaces the current history entry and loads that URL
- `doors.ActionLocationReload{}` reloads the current page
- `doors.ActionLocationRawAssign{URL: ...}` loads a literal URL
- `doors.ActionLocationRawReplace{URL: ...}` replaces the current history entry and loads a literal URL

If the target belongs to your **Doors** path model, model-based actions still help you build the URL safely, but they are still hard navigations.

Use `RawAssign` or `RawReplace` when you already have a full absolute URL or want to leave that model-based routing path — they are the only way to navigate to another origin (OAuth, external pages). `RawReplace` drops the current page from history, which fits redirects the user should not navigate back to.

Location actions are deferred to the end of the current client turn.

That means earlier actions in the same list can still run first.

## Scroll

`ActionScroll` scrolls the first matching selector into view.

It is useful for:

- validation jumps
- bringing a changed region into view
- moving the user back to a result block or top section

If nothing matches, the action fails — with `doors.Call`, the error arrives as the second value on the completion channel, on a detached context.

## Indicate

`ActionIndicate` applies indicators for a fixed duration.

Use it when the feedback should be explicit and timed, instead of being tied automatically to the request lifecycle.

Unlike an event attr `Indicator`, it does not stop when the request finishes.

It lasts for the `Duration` you give it.

When `ActionIndicate` runs from `Before`, `r.After(...)`, or `OnError`, `SelectorTarget()` can use the current event element.

When it runs from a direct `Call`, there is no event element, so only the `SelectorQuery(...)` and `SelectorQueryAll(...)` selector variants apply.

Indication details are covered in [Indication](./11-indication.md).

## Combining

`Before`, `OnError`, and `r.After(...)` accept a single `doors.Actions` value. Each action struct (`ActionEmit`, `ActionLocationReload`, `ActionLocationAssign`, `ActionScroll`, `ActionIndicate`, …) is itself a `doors.Actions`, so a single action passes directly. To run several in order, chain them with `.And(...)` or `doors.JoinActions(...)`:

```go
Before: doors.ActionScroll{Selector: "#top"}.
	And(doors.ActionIndicate{
		Indicator: doors.IndicateClass("pending"),
		Duration:  300 * time.Millisecond,
	})

// or
Before: doors.JoinActions(
	doors.ActionScroll{Selector: "#top"},
	doors.ActionIndicate{
		Indicator: doors.IndicateClass("pending"),
		Duration:  300 * time.Millisecond,
	},
)
```

The same shape works for `r.After(...)` and `OnError`.

## Rules

- Prefer rendering and state for durable UI changes.
- Prefer `Setter` when existing attributes should stay shared without rerendering the elements.
- Ignore the `doors.Call` completion channel when the outcome does not matter; capture results with `Into`.
- Scope `$on(...)` handlers intentionally; async handlers hold the action pending until their `Promise` settles.
- Prefer `ALink` or updating the `Source` a `RouteSource` receives inside `RouteModel` for in-app navigation.
- Use location actions when you intentionally want a full page load.
- Use `r.After(...)` for success-only follow-up and `OnError` for fallback or recovery behavior.
