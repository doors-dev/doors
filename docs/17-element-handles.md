# Element Handles

An element handle is a Go value you attach to elements as an attr and then act on from Go, reaching the live DOM without a rerender.

There are two:

- `doors.Setter` sets attributes on the elements it is attached to
- `doors.Emitter` dispatches synthetic DOM events to the elements it is attached to

Both are stateless. They change live elements only, so a rerendered element goes back to what its template says. Keep durable values in state and templates.

## Use

Reach for a handle when:

- the change is a small, targeted DOM effect and Go already knows what it should be
- several elements should stay in sync, or one element is simplest to manage this way
- a rerender would be unnecessary work

Use normal rendering and state instead when the UI itself should change, and [Events](./08-events.md) for ordinary DOM events coming from the browser.

Prefer `ActionEmit` with a `$on(...)` handler when the browser should own the logic, especially when the change depends on measurements, timers, or third-party widgets. See [JavaScript](./15-javascript.md) and [Actions](./12-actions.md).

## Setter

The zero value is ready to use. Attach it like any other attr, then dispatch `Set(name, value)`, which returns an [Action](./12-actions.md) for `doors.Call` or any action-accepting API.

```gox
<>
	~~
	var locked doors.Setter
	~~

	<header>
		<button (&locked)>Save draft</button>
	</header>

	<footer>
		<button (&locked)>Publish</button>
	</footer>

	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				doors.Call(ctx, locked.Set("disabled", true))
				return false
			},
		})>
		Lock both actions
	</button>
</>
```

Because both buttons share the setter, one call disables both.

The value follows template attribute semantics:

- `nil` and `false` remove the attribute
- `true` sets it bare
- `gox.Output` values serialize themselves
- anything else is formatted with the `fmt` package

```go
doors.Call(ctx, locked.Set("disabled", true))
doors.Call(ctx, s.Set("aria-label", "busy"))
doors.Call(ctx, s.Set("hidden", nil))
```

A `gox.Mutate` value, such as `doors.Class(...)`, fails the action. Composing one needs the previous attribute value, which only a rerender has — pass a plain string instead.

`Set(...).Into(&count)` captures the number of live elements the action reached, `0` when none is live.

Typical fits are `disabled`, `hidden`, `aria-*`, `data-*`, and single-purpose `class` or `style` values.

## Emitter

The zero value is ready to use. Attach it like any other attr, then dispatch an event method's action.

```gox
<>
	~~
	var field doors.Emitter
	~~

	<input
		type="text"
		(&field)
		(doors.AInput{
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				doors.Logger(ctx).Info("input", "data", r.Event().Data)
				return false
			},
		})/>

	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				doors.Call(ctx, field.Input(doors.InputEmit{Data: "hey"}))
				return false
			},
		})>
		Simulate typing
	</button>
</>
```

Each method dispatches a real DOM event that bubbles and is cancelable, so event attrs on the element **and on its ancestors** run, and plain JavaScript listeners see it too.

The event families and their init structs:

| Methods | Init |
| --- | --- |
| `Click`, `PointerDown`, `PointerUp`, `PointerMove`, `PointerOver`, `PointerOut`, `PointerEnter`, `PointerLeave`, `PointerCancel`, `GotPointerCapture`, `LostPointerCapture` | `PointerEmit` |
| `KeyDown`, `KeyUp` | `KeyboardEmit` |
| `Focus`, `Blur`, `FocusIn`, `FocusOut` | `FocusEmit` |
| `Input` | `InputEmit` |
| `Change` | `ChangeEmit` |
| `Submit` | `SubmitEmit` |

Unset init fields use the browser defaults. `FocusEmit` runs the handlers without moving focus, and `SubmitEmit` reads the form data from the event target, so attach the emitter to the form itself.

`Into(&count)` captures the number of hook requests the emitted events triggered, not the number of elements reached: an attached element with no matching event attr adds nothing, and a bubbling event adds one per event attr it reaches. Any failure among them fails the call.

## Sharing

One handle can be attached to many elements, and many handles can be attached to one element — including a setter and an emitter together.

```gox
<div id="row" (&locked) (&highlight) (&row)>…</div>
```

Attaching the same handle to one element twice is harmless; it counts once.

## Rules

- Use a handle when the DOM change is the whole job.
- Reuse one handle when several elements should stay in sync.
- Prefer `ActionEmit` with `$on(...)` for richer client-side DOM work.
- Use normal rendering and state when the UI itself should change.
- Keep the source of truth in state and templates — handles do not survive a rerender.
- Remember which unit `Into` reports: live elements for `Setter`, hook requests for `Emitter`.

## Related

- [Events](./08-events.md) for the event attrs an emitted event triggers.
- [Actions](./12-actions.md) for `doors.Call`, `Into`, and action lists.
- [JavaScript](./15-javascript.md) for `ActionEmit` and `$on(...)`.
- [State](./07-state.md) for changes that should survive a rerender.
