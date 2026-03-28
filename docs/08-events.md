# Events

In **Doors**, DOM events are handled through attributes.

An event attribute connects a browser event to a backend handler, with optional client-side scheduling, indication, and actions around the request.

That same system is used for:

- pointer events
- keyboard events
- focus events
- input and change events
- form submission

## Attach

Event attributes are regular Go values with names like `doors.AClick`, `doors.AInput`, or `doors.ASubmit[T]`.

You can attach them in two ways.

As an attribute modifier:

```gox
<button
	(doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			return false
		},
	})>
	Click
</button>
```

Or as a proxy:

```gox
<>
	~>doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			return false
		},
	} <button>Click</button>
</>
```

The modifier form attaches directly to the element you are editing.

The proxy form walks through the following subtree until it reaches the real rendered element and attaches there. That is useful when the final element is inside another component.

## Flow

When an event fires, the client/runtime flow is roughly:

1. capture the browser event and build the payload
2. apply client-side event options such as `PreventDefault`, `StopPropagation`, `ExactTarget`, or key `Filter`
3. run client-side scopes
4. start indication
5. run any `Before` actions
6. send the request to the server handler
7. run `After` actions if the request succeeds
8. run `OnError` actions if the request fails

That is why scopes and indication feel immediate: they start on the client before the server finishes the request.

## Common

Most event attributes share the same core fields:

- `On`: backend handler
- `Scope`: request scheduling rules, covered in [Scopes](./10-scopes.md)
- `Indicator`: temporary client-side feedback, covered in [Indication](./11-indication.md)
- `Before`: client-side actions before the request
- `OnError`: client-side actions if the request fails

`After` is different: it is not an attribute field. You schedule it from inside the handler with `r.After(...)`.

Some families also add fields like:

- `PreventDefault`
- `StopPropagation`
- `ExactTarget`
- `Filter`
- `ExcludeValue`

The `On` handler returns `bool`:

- `false` keeps the handler active
- `true` marks it done so it can be removed

For normal DOM events, `false` is the common default.

## Request

For DOM events, the handler receives `doors.RequestEvent[T]`.

That gives you:

- `r.Event()` for the typed event payload
- `r.SetCookie(...)` and `r.GetCookie(...)`
- `r.After(...)` to schedule client-side actions after the request

Example:

```go
On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
	r.After(doors.ActionOnlyScroll("#top", true))
	return false
}
```

Form handlers use:

- `doors.RequestForm[T]` for decoded form data
- `doors.RequestRawForm` for raw multipart access

## Execution

Each activated event attribute in **Doors** has its own backend hook instance.

Calls to that same instance are serialized.

That means rapid repeated events on the same active handler do not run concurrently on the backend.

If you prepare one activated attribute with `doors.A(ctx, ...)` and reuse it across multiple elements, those elements share the same hook instance and the same execution queue.

## Pointer

Pointer attributes include:

- `doors.AClick`
- `doors.APointerDown`
- `doors.APointerUp`
- `doors.APointerMove`
- `doors.APointerOver`
- `doors.APointerOut`
- `doors.APointerEnter`
- `doors.APointerLeave`
- `doors.APointerCancel`
- `doors.AGotPointerCapture`
- `doors.ALostPointerCapture`

Example:

```gox
<button
	(doors.AClick{
		PreventDefault: true,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			x := r.Event().PageX
			y := r.Event().PageY
			_ = x
			_ = y
			return false
		},
	})>
	Track click
</button>
```

The pointer payload includes the usual browser pointer fields, including coordinates, button state, pointer type, pressure, and timestamp.

## Keys

Keyboard attributes are:

- `doors.AKeyDown`
- `doors.AKeyUp`

Use `Filter` to limit by `event.key`:

```gox
<input
	(doors.AKeyDown{
		Filter: []string{"Enter"},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			return false
		},
	})/>
```

The keyboard payload includes `Key`, `Code`, `Repeat`, and modifier state such as `CtrlKey`, `ShiftKey`, `AltKey`, and `MetaKey`.

## Focus

Focus attributes are:

- `doors.AFocus`
- `doors.ABlur`
- `doors.AFocusIn`
- `doors.AFocusOut`

Use `AFocusIn` and `AFocusOut` when bubbling behavior matters.

Use `AFocus` and `ABlur` for the plain focus events.

## Input

Input-related attributes are:

- `doors.AInput`
- `doors.AChange`

`AInput` fires as the user edits.

`AChange` fires when the value is committed.

```gox
<input
	type="text"
	(doors.AInput{
		On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
			value := r.Event().Value
			_ = value
			return false
		},
	})/>
```

The input and change payloads include browser-style fields such as:

- `Name`
- `Value`
- `Number`
- `Date`
- `Selected`
- `Checked`

`AInput{ExcludeValue: true}` omits the normal input-derived fields from the payload. Use it when you want the event itself without sending the current input value data.

## Forms

Form submission attributes are:

- `doors.ASubmit[T]`
- `doors.ARawSubmit`

`ASubmit[T]` parses the multipart form and decodes it into your Go type.

```gox
type LoginForm struct {
	Email string `form:"email"`
	Code  string `form:"code"`
}

<form
	(doors.ASubmit[LoginForm]{
		On: func(ctx context.Context, r doors.RequestForm[LoginForm]) bool {
			data := r.Data()
			_ = data
			return false
		},
	})>
	<input name="email"/>
	<input name="code"/>
	<button>Send</button>
</form>
```

Use `ARawSubmit` when you want direct multipart access for streaming, custom parsing, or uploads.

For form decoding, **Doors** uses [go-playground/form v4](https://github.com/go-playground/form/tree/v4.2.1).

## Unsupported

If the browser event you need is not supported by the built-in `doors.A...` event attributes, wire it yourself in JavaScript and call a custom hook.

That is the normal extension path for:

- browser events **Doors** does not expose directly
- custom DOM integrations
- third-party widgets that already have their own client-side event system

The usual pattern is:

1. listen to the event in JavaScript
2. call `$hook(...)` or `$fetch(...)`
3. handle it on the Go side with `doors.AHook[...]` or `doors.ARawHook`

Example:

```gox
<script
	(doors.AHook[string]{
		Name: "visibility",
		On: func(ctx context.Context, r doors.RequestHook[string]) (any, bool) {
			println(r.Data())
			return nil, false
		},
	})>
	document.addEventListener("visibilitychange", async () => {
		await $hook("visibility", document.visibilityState)
	})
</script>
```

See [Custom Attrs](./13-custom-attrs.md) and [JavaScript](./15-javascript.md).

## Reuse

Use `doors.A(ctx, ...)` when you want to prepare one activated attribute value and reuse it.

```gox
<>
	~{
		radio := doors.A(ctx, doors.AChange{
			On: func(ctx context.Context, r doors.RequestEvent[doors.ChangeEvent]) bool {
				return false
			},
		})
	}

	<input type="radio" name="pick" value="a" (radio)/>
	<input type="radio" name="pick" value="b" (radio)/>
</>
```

For a one-off attribute on one element, you usually do not need `doors.A(...)`.

## Related

- Use [Navigation](./09-navigation.md) for `doors.ALink`.
- Use [Custom Attrs](./13-custom-attrs.md) for `doors.AHook[...]`, `doors.ARawHook`, `doors.AData`, and `doors.ADyn`.
- Use [Scopes](./10-scopes.md) for request scheduling.
- Use [Indication](./11-indication.md) for client-side feedback.
- Use [Actions](./12-actions.md) for `Before`, `OnError`, and `After` actions.

## Rules

- Use the `ctx` that **Doors** gives you in the handler.
- Use `false` as the normal return value unless the handler really should be one-shot.
- Use `AInput` for live edits and `AChange` for committed values.
- Use `ASubmit[T]` when typed decoding is enough; use `ARawSubmit` for upload-heavy or custom multipart flows.
- Use scopes for interaction policy instead of rebuilding debounce/blocking logic by hand.
- Use indication when the user needs immediate feedback before the server responds.
