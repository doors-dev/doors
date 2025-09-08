# Event Hook Attributes

Attributes that are used to enable and configure the event handler. 

```templ
type AClick struct {
	// If true, stops the event from bubbling up the DOM.
	// Optional.
	StopPropagation bool
	// If true, prevents the browser's default action for the event.
	// Optional.
	PreventDefault bool
	// If true, only fires when the event occurs on this element itself.
	// Optional.
	ExactTarget bool
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Actions to run before the hook request.
	// Optional.
	Before []Action
	// Backend event handler.
	// Receives a typed REvent[PointerEvent].
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(context.Context, REvent[PointerEvent]) bool
	// Actions to run on error.
	// Optional.
	OnError []Action
}
```

## Available Event Hook Attributes

* Pointer Events
  * AClick
  * APointerDown
  * APointerUp
  * APointerMove 
  * APointerOver 
  * APointerOut 
  * APointerEnter 
  * APointerLeave 
  * APointerCancel 
  * ALostPointerCapture 
* Focus Events
  * AFocus
  * ABlur
  * AFocusIn
  * AFocusOut
* Form Events
  * ASubmit[D]
    Where D - struct with unmarshaling annotations from [go-playground/form](https://github.com/go-playground/form)
  * ARawSubmit
    To deal with raw http form request 
* Input Events
  * AChange
  * AInput

> More event bindings are coming soon. For unsupported ones use script hooks and JS bindings

## Concurrency Control

It's guaranteed by the framework that handler function calls happen in series **within a single hook**. 

```templ
@doors.ASubmit[Data]{
  // next call of On function can occur only after previous is completed 
	On: func(ctx context.Context, r doors.RForm[Data]) bool {
	  /* processing */
		return false
	}
}
<form>
/* form */
</form>
```

If multiple events occur on the same handler, they will be queued on the backend. 

> You can attach a single handler to multiple elements using pre-initialization and, therefore, join their queues. Pre-initialization is described further in the article

That's too relaxed for some cases.  So, precise control can be enabled on the frontend via the **Scopes** API:

```templ
@doors.AInput{
  // enable debounce for input event
  Scope: doors.ScopeOnlyDebounce(300 * time.Millisecond, time.Second),
	On: func(ctx context.Context, r doors.REvent[doors.InputEvent]) bool {
	  /* processing */
		return false
	}
}
```

Check out the [Scopes](./ref/03-scopes.md) article for details.

## Indication

For better UX apply pending indication with the **Indicator** API:

```templ
@doors.ASubmit[loginData]{
    // indicate on element #login-submit by temporary setting (adding) attribute 
    // aria-busy with value "true"
		Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
		Scope:     doors.ScopeOnlyBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			/* logic */
			return true
		},
}
```

Refer to the [Indication](./ref/02-indication.md) article for more details.

## Error Handling

`OnError` defines actions that execute when an **error occurs during a hook’s processing**. It supports UI indication, custom client-side callbacks and navigation (check [Actions](./ref/04-actions.md) for the full list). 

>  Please refer to [JavaScript](./10-javascript.md) for details about error object.

