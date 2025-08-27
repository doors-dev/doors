# Event Hook Attributes

Attributes that are used to enable and configure the event handler. 

```templ
type AClick struct {

	// StopPropagation, if true, stops the event from bubbling up the DOM.
	StopPropagation bool

	// PreventDefault, if true, prevents the browser's default action for the event.
	PreventDefault bool

	// Scope determines how this hook is scheduled (e.g., blocking, debounce).
	Scope []Scope

	// Indicator specifies how to visually indicate that the hook is running on the frontent.
	Indicator []Indicator

	// On is the required backend handler for the click event.
	// It receives a typed REvent[PointerEvent] and should return true
	// when the hook is considered complete and can be removed.
	On func(context.Context, REvent[PointerEvent]) bool

	// OnError determines front-end should do if an error occurred during the hook request
	OnError []OnError
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
  Scope: doors.ScopeDebounce(300 * time.Millisecond, time.Second),
	On: func(ctx context.Context, r doors.REvent[doors.InputEvent]) bool {
	  /* processing */
		return false
	}
}
```

Check out the [Scopes](./ref/04-scopes.md) article for details.

## Indication

For better UX apply pending indication with the **Indicator** API:

```templ
@doors.ASubmit[loginData]{
    // indicate on element #login-submit by temporary setting (adding) attribute 
    // aria-busy with value "true"
		Indicator: doors.IndicatorAttrQuery("#login-submit", "aria-busy", "true"),
		Scope:     doors.ScopeBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			/* logic */
			return true
		},
}
```

Refer to the [Indication](./ref/03-indication.md) article for more details.

## Error Handling

Event processing can fail due to several reasons. Some reasons are automatically ignored (canceled by scope), some should be handled (such as network issues) for a better user experience, and some are not expected to ever occur (front-end exceptions during event handling).

To configure front-end action on error use the `OnError` field:

```templ
@doors.ASubmit[loginData]{
    /* attr setup */
    
    // use indication on error, add class "show" to
    // #error_message for 3 seconds
    OnError: doors.OnErrorIndicate(3 * time.Second, doors.IndicatorClassQuery("#error_message", "show")),
}
```

You can set up an indicator or a JavaScript function call, or any combination thereof. Please refer to the [Error Handling](./ref/02-error-handling.md) article.

