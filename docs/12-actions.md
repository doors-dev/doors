# Actions

Actions are client-side effects triggered from Go.

Use them when the browser should do something imperative instead of just rendering different HTML.

Good fits are:

- navigate after a successful request
- scroll to a result or validation error
- show a timed indication
- call a JavaScript handler with `$on(...)`

If the UI should simply render different content, prefer normal rendering.

If one attribute should stay shared across existing elements, `AShared` from [Shared Attr](./17-shared-attr.md) is often a better fit than an action.

## Places

You can schedule actions in five common places:

- `doors.Call(ctx, action)` for fire-and-forget dispatch from Go
- `doors.XCall[T](ctx, action)` when Go needs the client result
- `Before` on a request attr such as an event attr or `ALink`, just before the request is sent
- `r.After(...)` after a successful request
- `OnError` on a request attr when a client-visible hook error happens

Action lists run in the order you give them.

`OnError` is for normal client-visible failures such as network, server, bad request, and similar hook errors.

It does not run for scope cancellations or expired hooks, and a stopped instance is handled by reloading the page instead.

## Direct

Use `doors.Call` when the result does not matter.

```go
doors.Call(ctx, doors.ActionLocationReload{})
```

Use `doors.XCall[T]` when the client handler should return a value to Go.

```go
ch, cancel := doors.XCall[string](ctx, doors.ActionEmit{
	Name: "pick",
	Arg:  "hello",
})
defer cancel()

res, ok := <-ch
if ok && res.Err == nil {
	println(res.Ok)
}
```

Do not wait on `XCall` during rendering.

If you need to wait for the result, do it in a hook, inside `doors.Go(...)`, or
in your own goroutine with `doors.Free(ctx)`.

`doors.Free(ctx)` keeps the original context values, but switches to the root
Doors context and extends cancellation/deadline/lifetime to the instance
runtime.

Cancellation is best-effort.

If a direct `XCall` is canceled, its channel closes without a value.

For most actions, use `json.RawMessage` as `T`.

`ActionEmit` is the main case where you usually want the real result type.

## Emit

`ActionEmit` calls a client handler registered with `$on(name, handler)`.

```gox
<>
	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				doors.Call(ctx, doors.ActionEmit{
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

For normal in-app navigation, prefer [Navigation](./09-navigation.md), especially `ALink` and path model mutation.

Built-ins:

- `doors.ActionLocationAssign{Model: ...}` pushes a new history entry and loads that URL
- `doors.ActionLocationReplace{Model: ...}` replaces the current history entry and loads that URL
- `doors.ActionLocationReload{}` reloads the current page
- `doors.ActionLocationRawAssign{URL: ...}` loads a literal URL

If the target belongs to your **Doors** path model, model-based actions still help you build the URL safely, but they are still hard navigations.

Use `RawAssign` when you already have a full URL or want to leave that model-based routing path.

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

When it runs from direct `Call` or `XCall`, there is no event target, so use explicit selectors like `SelectorQuery(...)`.

Indication details are covered in [Indication](./11-indication.md).

## Helpers

The `ActionOnly...` helpers return a single-item `[]Action`.

They are mostly there for `Before`, `OnError`, and `r.After(...)`.

- `doors.ActionOnlyEmit(...)`
- `doors.ActionOnlyLocationReload()`
- `doors.ActionOnlyLocationReplace(...)`
- `doors.ActionOnlyLocationAssign(...)`
- `doors.ActionOnlyLocationRawAssign(...)`
- `doors.ActionOnlyScroll(...)`
- `doors.ActionOnlyIndicate(...)`

## Rules

- Prefer rendering and state for durable UI changes.
- Prefer `AShared` when one existing attribute should stay shared without rerendering the elements.
- Use `doors.Call` when the result does not matter.
- Use `doors.XCall` mainly with `ActionEmit`.
- Keep `$on(...)` handlers synchronous and scoped intentionally.
- Prefer path model mutation or `ALink` for in-app navigation.
- Use location actions when you intentionally want a full page load.
- Use `r.After(...)` for success-only follow-up and `OnError` for fallback or recovery behavior.
