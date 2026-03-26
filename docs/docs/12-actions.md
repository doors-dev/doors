# Actions

Actions are backend-triggered client operations.

They are the piece of Doors that lets Go tell the browser to do something immediately on the client side without changing the page through normal HTML rendering.

## Choose

Use this quick mapping:

- one-way client effect: `doors.Call(...)`
- need client return value: `doors.XCall(...)`
- run before request: attribute `Before`
- run only after successful request: `r.After(...)`
- run on request failure: attribute `OnError`

## Places

You can use actions in four places:

- `doors.Call(ctx, action)` for fire-and-forget dispatch
- `doors.XCall[T](ctx, action)` when you need a result
- `r.After(...)` to run actions after a request succeeds
- `Before` and `OnError` fields on event attributes

`Before` runs on the client before the request is sent.  
`r.After(...)` runs after a successful request.  
`OnError` runs when the hook request fails.

## Flow

Actions execute on the client.

That has a few important consequences:

- they are not HTML updates
- they run through the Doors client runtime
- they are synchronous from the client dispatcher point of view
- async client action handlers are not allowed

Most actions do not return meaningful data. `ActionEmit` is the main exception.

## Calls

Use `doors.Call` when you do not need a return value:

```go
doors.Call(ctx, doors.ActionLocationReload{})
```

Use `doors.XCall[T]` when you need the client result:

```go
ch, cancel := doors.XCall[string](ctx, doors.ActionEmit{
	Name: "pick",
	Arg:  "hello",
})
defer cancel()

res := <-ch
if res.Err == nil {
	println(res.Ok)
}
```

`XCall` should only be awaited where blocking is acceptable, such as:

- event handlers
- your own goroutines
- `doors.Go(...)`

For most actions, use `json.RawMessage` as `T`. For `ActionEmit`, use the real result type you expect from the client handler.

## Pipelines

Actions also participate in hook pipelines.

`Before` is attached on the attribute:

```gox
<button
	(doors.AClick{
		Before: []doors.Action{
			doors.ActionScroll{
				Selector: "#top",
				Smooth:   true,
			},
		},
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			return false
		},
	})>
	Run
</button>
```

`After` is queued from the backend handler:

```gox
<button
	(doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyLocationAssign(AppPath{
				Home: true,
			}))
			return false
		},
	})>
	Go home
</button>
```

Use `After` when the action should happen only if the request completed successfully.

## Emit

`ActionEmit` calls a client handler registered with `$on(...)`.

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

Handler lookup is scoped through the Door tree. Doors starts from the current Door context and walks outward through parent Doors until it finds a matching `$on(name, ...)`.

That means:

- local handlers shadow outer handlers with the same name
- an action can reach handlers in the current Door scope or an outer one

`$on` can be used for error handling too. If `ActionEmit` is executed from `OnError`, the client handler receives a second argument with the hook error (`HookErr`).

```gox
<script>
	$on("save_error", (arg, err) => {
		if (err) {
			console.log(err.kind, err.message)
		}
	})
</script>
```

## Location

Location actions change browser location:

- `doors.ActionLocationReload{}`
- `doors.ActionLocationReplace{Model: ...}`
- `doors.ActionLocationAssign{Model: ...}`
- `doors.ActionLocationRawAssign{URL: ...}`

Use model-based variants when the target belongs to your Doors routing model. Use `RawAssign` when you already have a literal external or prebuilt URL.

`Replace` does not create a new history entry. `Assign` does.

## Scroll

`ActionScroll` scrolls to the first matching selector.

This is useful for:

- validation jumps
- moving focus to a result block
- bringing a changed region into view

## Indicate

`ActionIndicate` applies indicators for a fixed duration.

Use it when indication should happen as an explicit action rather than as part of a hook lifecycle.

Indicator details are covered in [10-indication.md](/Users/alex/Lib/doors/docs/docs/10-indication.md).

## Helpers

Each built-in action also has a single-item helper for places that want `[]Action`:

- `doors.ActionOnlyEmit(...)`
- `doors.ActionOnlyLocationReload()`
- `doors.ActionOnlyLocationReplace(...)`
- `doors.ActionOnlyLocationAssign(...)`
- `doors.ActionOnlyLocationRawAssign(...)`
- `doors.ActionOnlyScroll(...)`
- `doors.ActionOnlyIndicate(...)`

These are especially convenient for `Before`, `OnError`, and `r.After(...)`.

## Rules

- Use `doors.Call` when the result does not matter.
- Use `doors.XCall` mainly with `ActionEmit`.
- Keep client handlers registered with `$on(...)` synchronous.
- Use model-based location actions when you want routing-safe URLs.
- Use `After` for success-only follow-up work.
- Use `OnError` for recovery or fallback client behavior.
