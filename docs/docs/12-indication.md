# Indication

In **Doors**, indication is temporary DOM feedback that starts on the client when the request actually begins.

Use it when the backend work may take long enough that the user should see something immediately.

Common examples are:

- `aria-busy`
- temporary button text like `Saving...`
- loading classes
- showing or hiding nearby UI while a request is pending

## Why

Indication runs before the server responds, so it is the right tool for immediate feedback around an in-flight request.

It also restores automatically when the request ends.

In practice that means:

- scopes decide whether and when a request runs
- indication shows the user what is happening right now
- if a scope delays or cancels the request, indication follows that decision
- the indication clears after **Doors** applies the request result on the client

## Target

Each indicator chooses what element to affect.

- `doors.SelectorTarget()` uses the element that triggered the event
- `doors.SelectorQuery(query)` uses the first matching element
- `doors.SelectorQueryAll(query)` uses all matching elements
- `doors.SelectorQueryParent(query)` uses the nearest matching ancestor

Use `Target` when the loading state belongs on the clicked or edited element itself.

Use `Query`, `QueryAll`, or `QueryParent` when the feedback belongs somewhere else, like a toolbar, spinner, panel, or form wrapper.

## Kinds

`doors.IndicatorContent` replaces the element content for the duration of the request.

Use it for text like `Saving...` or `Loading...`.

`doors.IndicatorAttr` sets an attribute temporarily.

Use it for `aria-busy`, `disabled`, `hidden`, or testing hooks.

`doors.IndicatorClass` adds classes temporarily.

Use it for loading, dimming, or animation classes.

`doors.IndicatorClassRemove` removes classes temporarily.

Use it when the pending state should reveal something that is normally hidden.

`IndicatorContent` writes to `innerHTML`, so only use trusted HTML there.

## Helpers

For simple cases, use the `IndicatorOnly*` helpers instead of building the struct by hand.

- `doors.IndicatorOnlyContent(...)`, `doors.IndicatorOnlyAttr(...)`, `doors.IndicatorOnlyClass(...)`, and `doors.IndicatorOnlyClassRemove(...)` target the event source
- add `Query`, `QueryAll`, or `QueryParent` when the target is somewhere else

That covers most everyday cases without much setup.

## Restore

Restore is precise.

When an indication ends, **Doors** restores the previous content, attributes, and class presence for the parts that indicator changed.

That means:

- attributes that did not exist before are removed again
- classes removed temporarily are added back if they were there before
- content returns to its previous HTML

You do not need manual cleanup code for normal pending states.

## Overlap

Indications on the same element can overlap.

When that happens, **Doors** queues them on the client.

As one indication finishes, the next one takes over.

If the next indication does not mention a field the previous one changed, that field falls back to the original value instead of leaking stale pending state forward.

## Example

```gox
<>
	<span id="search-loader"></span>

	<input
		type="search"
		placeholder="Country"
		(doors.AInput{
			Indicator: doors.IndicatorOnlyAttrQuery("#search-loader", "aria-busy", "true"),
			Scope:     doors.ScopeOnlyDebounce(300*time.Millisecond, 600*time.Millisecond),
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				<-time.After(500 * time.Millisecond)
				return false
			},
		})/>
</>
```

This keeps the typing interaction responsive, debounces requests, and shows feedback as soon as the current debounced request is actually running.

## Rules

- Use indication for temporary interaction feedback, not as your main UI state.
- Use `SelectorTarget()` when the pending state belongs on the active element; query selectors when it belongs elsewhere.
- Use helpers first, then move to manual `Indicator...` values when one event needs several temporary DOM changes.
- Use [11-scopes.md](/Users/alex/Lib/doors/docs/docs/11-scopes.md) together with indication when request timing matters.
- Use [13-actions.md](/Users/alex/Lib/doors/docs/docs/13-actions.md) when you want to trigger indication as an action instead of from an event attribute.
