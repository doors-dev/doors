# Events

Doors event handling is built from attributes.

An event attribute attaches client capture, request scheduling, optional client-side feedback, and a backend handler to an element.

The same system also powers:

- forms
- keyboard and pointer handlers
- client hooks and `$data(...)`

## Basics

Event attributes are regular Go values that work as both:

- a GoX proxy: `~>doors.AClick{...} <button>Save</button>`
- an attribute modifier: `(doors.AClick{...})`

Doors attribute types follow the `doors.A...` naming pattern.

These two forms are related, but they are not identical in how they attach.

- Modifier form attaches to the element whose attribute list you are editing directly.
- Proxy form walks through the following subtree until it reaches the real rendered element and attaches there.

That proxy behavior drills through components and containers.

There is no rule that says one form is the default and the other is only for special cases. Use whichever style you prefer in a given component.

Example:

```gox
~>doors.AClick{
	On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
		return false
	},
} <button>Click</button>
```

Or:

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

## `doors.A`

`doors.A(ctx, ...)` activates Doors attributes in the current context so the resulting value can be attached safely, including to multiple elements.

This is important because `doors.AClick{...}` is attribute configuration, not a standalone handler instance.

For one-off attributes on one element, you usually do not need `doors.A` at all, because GoX already allows multiple attribute modifiers directly:

```gox
<button
	id="save"
	(doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			return false
		},
	},
	doors.AData{
		Name: "kind",
		Value: "primary",
	})>
	Save
</button>
```

Use `doors.A` when you want to prepare one activated attribute value and reuse it.

For example, radio buttons often share the same handler:

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

## Shape

Most event attributes follow the same shape:

- `On`: backend handler
- `Scope`: scheduling rules, covered in [09-scopes.md](/Users/alex/Lib/doors/docs/docs/09-scopes.md)
- `Indicator`: temporary DOM feedback, covered in [10-indication.md](/Users/alex/Lib/doors/docs/docs/10-indication.md)
- `Before`: client-side actions before the request
- `OnError`: client-side actions when the request fails

Some families also expose:

- `StopPropagation`
- `PreventDefault`
- `ExactTarget`
- event-specific filters such as `Filter` on key handlers

The `On` handler returns `bool`:

- `false` keeps the hook active
- `true` marks it done and allows the hook to be removed

For normal DOM events, the request parameter is `doors.RequestEvent[T]`. Use `r.Event()` to read the typed payload.

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
~>doors.AClick{
	PreventDefault: true,
	On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
		x := r.Event().PageX
		y := r.Event().PageY
		_ = x
		_ = y
		return false
	},
} <button>Track click</button>
```

The server receives real pointer coordinates and event-specific data.

## Keys

Keyboard attributes:

- `doors.AKeyDown`
- `doors.AKeyUp`

Use `Filter` to limit by `event.key`.

```gox
<input
	(doors.AKeyDown{
		Filter: []string{"Enter"},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			return false
		},
	})/>
```

The keyboard payload includes key and modifier state such as `AltKey`, `CtrlKey`, and `ShiftKey`.

## Focus

Focus attributes:

- `doors.AFocus`
- `doors.ABlur`
- `doors.AFocusIn`
- `doors.AFocusOut`

Use `AFocusIn` / `AFocusOut` when bubbling behavior matters. Use `AFocus` / `ABlur` for plain focus events.

## Input

Input-related attributes:

- `doors.AInput`
- `doors.AChange`

`AInput` fires continuously as the user edits.  
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

The input and change payloads include the browser-friendly fields you would expect for forms:

- `Name`
- `Value`
- `Number`
- `Date`
- `Selected`
- `Checked`

For `AInput`, `ExcludeValue: true` skips the text value in the payload.

## Forms

Form submission attributes:

- `doors.ASubmit[T]`
- `doors.ARawSubmit`

`ASubmit[T]` parses multipart form data and decodes it into a Go struct using `go-playground/form`.

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

Use `ARawSubmit` when you want full access to the multipart reader, parsed files, or raw upload handling.

## Hooks

Hooks and `$data(...)` are covered separately because they are not part of the normal DOM event attribute families.

Use them when you want:

- manual event wiring from JavaScript to Go
- integration points for embedded mini JavaScript apps
- direct client/server collaboration through `$hook(...)` and `$data(...)`

See [11-hooks.md](/Users/alex/Lib/doors/docs/docs/11-hooks.md).

## Navigation

`doors.AHref` has its own doc: [08-navigation.md](/Users/alex/Lib/doors/docs/docs/08-navigation.md).

Use it when a link should participate in Doors navigation instead of being a plain static `href`.

## Scopes

Scopes are covered in [09-scopes.md](/Users/alex/Lib/doors/docs/docs/09-scopes.md).  
Indication is covered in [10-indication.md](/Users/alex/Lib/doors/docs/docs/10-indication.md).

## Actions

Event attributes can trigger client-side actions before the request or on error through the `Before` and `OnError` fields.

Actions are covered in [12-actions.md](/Users/alex/Lib/doors/docs/docs/12-actions.md), so this page treats them only as part of the event attribute pipeline.

## Rules

- Use the `ctx` supplied by Doors in handlers.
- Prefer scopes for interaction policy instead of rebuilding that logic manually in handlers.
- Use indication for instant feedback; it starts on the client and does not need to wait for the server.
- Use `AInput` for live updates and `AChange` for committed values.
- Use `ASubmit[T]` when structured decoding is enough; use `ARawSubmit` for upload-heavy or custom parsing flows.
- Use `AHook[T]` and `AData` when JavaScript needs to collaborate with Go code directly.
- Keep hook handlers short when possible. Scopes and indication are there to shape the request flow before the backend runs.
