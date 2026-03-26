# Indication

Indication is temporary DOM feedback applied on the client.

It starts immediately when an event begins and is restored automatically when the event completes or fails.

That makes it useful for:

- `aria-busy`
- temporary button text
- loading classes
- quick visual feedback before the backend round trip finishes

## Selectors

Indicators target elements through selectors.

Available selectors:

- `doors.SelectorTarget()`
- `doors.SelectorQuery(query)`
- `doors.SelectorQueryAll(query)`
- `doors.SelectorQueryParent(query)`

Use:

- `Target` for the element that triggered the event
- `Query` for one matching element
- `QueryAll` for many elements
- `QueryParent` for the nearest matching ancestor

## Kinds

Indicator kinds:

- `doors.IndicatorContent`
- `doors.IndicatorAttr`
- `doors.IndicatorClass`
- `doors.IndicatorClassRemove`

Examples:

```gox
doors.IndicatorContent{
	Selector: doors.SelectorTarget(),
	Content: "Saving...",
}
```

```gox
doors.IndicatorAttr{
	Selector: doors.SelectorQuery("#submit"),
	Name: "aria-busy",
	Value: "true",
}
```

```gox
doors.IndicatorClass{
	Selector: doors.SelectorQuery(".panel"),
	Class: "loading dim",
}
```

```gox
doors.IndicatorClassRemove{
	Selector: doors.SelectorQuery("#dialog"),
	Class: "hidden",
}
```

## Helpers

Target helpers:

- `doors.IndicatorOnlyContent(...)`
- `doors.IndicatorOnlyAttr(...)`
- `doors.IndicatorOnlyClass(...)`
- `doors.IndicatorOnlyClassRemove(...)`

Query helpers:

- `doors.IndicatorOnlyContentQuery(...)`
- `doors.IndicatorOnlyAttrQuery(...)`
- `doors.IndicatorOnlyClassQuery(...)`
- `doors.IndicatorOnlyClassRemoveQuery(...)`

Query-all helpers:

- `doors.IndicatorOnlyContentQueryAll(...)`
- `doors.IndicatorOnlyAttrQueryAll(...)`
- `doors.IndicatorOnlyClassQueryAll(...)`
- `doors.IndicatorOnlyClassRemoveQueryAll(...)`

Parent-query helpers:

- `doors.IndicatorOnlyContentQueryParent(...)`
- `doors.IndicatorOnlyAttrQueryParent(...)`
- `doors.IndicatorOnlyClassQueryParent(...)`
- `doors.IndicatorOnlyClassRemoveQueryParent(...)`

## Restore

Indicator restore is precise.

When an indication ends, Doors restores:

- the previous content
- the previous attribute values
- the previous class presence

If an attribute did not exist before, it is removed afterward.  
If a class was removed temporarily, it is added back afterward.

## Queueing

Indicators on the same element can overlap.

In that case, they queue on the client. When one finishes:

- the next queued indication becomes active
- only the changed pieces are updated
- unrelated original values are preserved

## Example

```gox
<button
	id="save"
	(doors.AClick{
		Indicator: []doors.Indicator{
			doors.IndicatorAttr{
				Selector: doors.SelectorTarget(),
				Name: "aria-busy",
				Value: "true",
			},
			doors.IndicatorContent{
				Selector: doors.SelectorTarget(),
				Content: "Saving...",
			},
		},
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			<-time.After(500 * time.Millisecond)
			return false
		},
	})>
	Save
</button>
```

This gives immediate client-side feedback while the backend request is in flight.

## Notes

- Indication is client-side and starts immediately.
- It works naturally with scopes: scopes decide whether the request proceeds, indication shows the interaction state.
- Actions can also trigger indication explicitly, and that is covered in the next doc.
