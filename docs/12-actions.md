# Actions

Actions are client-side effects triggered from Go.

Use them when the browser should do something imperative instead of just rendering different HTML.

For example, call a JavaScript handler registered with `$on(...)`.

> If the UI should simply render different content, prefer normal rendering.
> If attributes should stay shared across existing elements, `Setter` from [Element Handles](./17-element-handles.md) is often a better fit than a custom action.

## Places

You can schedule actions in five common places:

- `doors.Call(ctx, action)` to dispatch from Go; the returned error channel is optional to use
- `Before` on a request attr such as an event attr or `ALink`, just before the request is sent
- `r.After(...)` after a successful request
- `OnError` on a request attr when a client-visible hook error happens

Action lists run in the order you give them.

`OnError` is for normal client-visible failures such as network, server, bad request, and similar hook errors.

It does not run for scope cancellations or expired hooks, and a stopped instance is handled by reloading the page instead.

## Direct

Ignore the returned channel when the outcome does not matter.

```go
doors.Call(ctx, doors.ActionLocationReload{})
```

When the client handler should return a value to Go, capture it with `Into`:

```go
var picked string
ch := doors.Call(ctx, doors.ActionEmit[string]{
	Name: "pick",
	Arg:  "hello",
}.Into(&picked))

err, ok := <-ch
if ok && err == nil {
	println(picked)
}
```

The destination is valid after the channel delivers nil.

Do not wait on the result channel during rendering.

If you need to wait for the result, do it in a hook, inside `doors.Go(...)`, or
in your own goroutine with `doors.DetachedContext(ctx)`.

`doors.DetachedContext(ctx)` keeps the current Doors ownership and lifecycle.

If the work should outlive that owner, use `doors.InstanceContext(ctx)`. It
switches Doors ownership to the root of the current instance and uses the
instance runtime lifecycle.

Canceling `ctx` requests best-effort cancellation. If a direct `Call` is canceled, its channel closes without a value.

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

Handler search is scoped through the Door tree.

**Doors** starts from the Door where the action was created and walks outward through parent Doors until it finds a matching handler.

That means:

- the nearest matching handler wins
- local handlers shadow outer handlers with the same name
- handlers outside that Door ancestry are not visible
- if no handler is found, the action fails

`$on(...)` handlers used by actions must stay synchronous.

Returning a `Promise` makes the action fail.

When `ActionEmit` is triggered from `OnError`, the handler receives the hook error as its second argument: `(arg, err)`.

## Location

Location actions are hard navigations.

They go through the browser location API and load the target page again.

That makes them useful when you intentionally want a full page load.

For normal in-app navigation, prefer [Navigation](./09-navigation.md), especially `ALink` or updating the `Source` from `RouteModel`.

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

If nothing matches, nothing happens.

## Indicate

`ActionIndicate` applies indicators for a fixed duration.

Use it when the feedback should be explicit and timed, instead of being tied automatically to the request lifecycle.

Unlike an event attr `Indicator`, it does not stop when the request finishes.

It lasts for the `Duration` you give it.

When `ActionIndicate` runs from `Before`, `r.After(...)`, or `OnError`, `SelectorTarget()` can use the current event element.

When it runs from a direct `Call`, there is no event target, so use explicit selectors like `SelectorQuery(...)`.

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
- Ignore the `doors.Call` error channel when the outcome does not matter; capture results with `Into`.
- Keep `$on(...)` handlers synchronous and scoped intentionally.
- Prefer `ALink` or updating the `Source` from `RouteModel` for in-app navigation.
- Use location actions when you intentionally want a full page load.
- Use `r.After(...)` for success-only follow-up and `OnError` for fallback or recovery behavior.
