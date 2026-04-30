# Indication

Indication is temporary client-side DOM feedback while a hook request is in flight. **Doors** applies it as the request starts and restores the previous DOM when the request ends.

```gox
<button (doors.AClick{
	Indicator: doors.IndicateClass("active"),
	On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
		<-time.After(500 * time.Millisecond)
		return false
	},
})>
	Save
</button>
```

## Kinds

The `Indicate*` family covers four indicator kinds:

| Helper | Effect |
| --- | --- |
| `IndicateContent(html)` | Replace element content with `html` |
| `IndicateAttr(name, value)` | Set attribute `name=value` |
| `IndicateClass(class)` | Add CSS class(es) |
| `IndicateClassRemove(class)` | Remove CSS class(es) |

> `IndicateContent` writes to `innerHTML` — pass only trusted HTML.

## Targets

By default the helper acts on the event source. A suffix redirects it:

| Suffix | Target |
| --- | --- |
| *(none)* | Event source element |
| `Query` | First match of a CSS query |
| `QueryAll` | All matches |
| `QueryParent` | Closest matching ancestor |

```go
doors.IndicateAttrQuery("#loader", "aria-busy", "true")
doors.IndicateClassQueryAll(".row", "pending")
```

The four selectors are also available directly: `doors.SelectorTarget()`, `doors.SelectorQuery(q)`, `doors.SelectorQueryAll(q)`, `doors.SelectorQueryParent(q)` — used with the struct form below.

## Combining

`Indicator` is a single `doors.Indicators` value. Each helper returns one, so a single helper passes directly. Combine with `.And(...)` or `doors.JoinIndicators(...)`:

```go
Indicator: doors.IndicateClass("loading").
	And(doors.IndicateAttrQuery("#search-loader", "aria-busy", "true"))
```

## Struct Form

For shared selectors or several changes on the same element, use the structs directly. They are also `doors.Indicators`:

```go
sel := doors.SelectorQuery("#card")
Indicator: doors.JoinIndicators(
	doors.IndicatorAttr{Selector: sel, Name: "data-state", Value: "saving"},
	doors.IndicatorClass{Selector: sel, Class: "pending"},
	doors.IndicatorContent{Selector: sel, Content: "Saving..."},
)
```

## Behavior

- Indication starts when the request actually begins on the client. Scopes can delay or cancel the request — indication follows that decision.
- When the request ends, **Doors** restores what the indicator changed: temporary attributes removed, added classes removed, removed classes restored, content put back.
- Overlapping indications on the same element queue. When one ends, the next takes over. Fields the next indication doesn't touch fall back to the original value, not the previous indication's value.

## Example

```gox
<>
	<span id="search-loader"></span>

	<input
		type="search"
		placeholder="Country"
		(doors.AInput{
			Indicator: doors.IndicateAttrQuery("#search-loader", "aria-busy", "true"),
			Scope: &doors.ScopeDebounce{
				Duration: 300 * time.Millisecond,
				Limit:    600 * time.Millisecond,
			},
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				<-time.After(500 * time.Millisecond)
				return false
			},
		})/>
</>
```

## Rules

- Use indication for temporary feedback, not as your main UI state.
- Default to `Indicate*` helpers; reach for the struct form when you need shared selectors or several changes on the same element.
- Pair with [Scopes](./10-scopes.md) when request timing matters.
- For programmatic, fixed-duration indication outside the event lifecycle, see `ActionIndicate` in [Actions](./12-actions.md).
