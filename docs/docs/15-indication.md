# Indication

The Indication API applies **temporary UI changes** to DOM elements.  
These changes can modify **content**, **attributes**, and **CSS classes**, and they are **restored automatically** when the indication ends.

---

## Concept

- The `Indicator` type represents a temporary DOM modification.  
- The `Indicator` field in attributes accepts a slice of indicators to perform multiple changes.  
- Simple indications can be created via helper functions (`IndicatorOnly*`).  
- Each `Indicator` specifies a `Selector` targeting an element and an operation.  
- Multiple indicators can overlap and queue; once the first finishes, the next executes if still needed.

---

## Selectors

Selectors define **which element** the indicator affects.

### `SelectorTarget()`
Selects the element that triggered the event.

### `SelectorQuery(query string)`
Selects the first element matching a CSS selector (e.g., `#id`, `.class`).

### `SelectorQueryAll(query string)`
Selects **all** elements matching a CSS selector.

### `SelectorParentQuery(query string)`
Selects the closest ancestor matching a CSS selector.

---

## Indicator Types

### Content

Temporarily replaces the `innerHTML` of the target element.  
No sanitization is performed.

```go
type IndicatorContent struct {
  Selector Selector
  Content  string
}
```

Helpers:
- `IndicatorOnlyContent(content string)`
- `IndicatorOnlyContentQuery(query, content string)`
- `IndicatorOnlyContentQueryParent(query, content string)`
- `IndicatorOnlyContentQueryAll(query, content string)`

---

### Attribute

Temporarily sets an attribute. If it was missing, it will be removed afterwards.

```go
type IndicatorAttr struct {
  Selector Selector
  Name     string
  Value    string
}
```

Helpers:
- `IndicatorOnlyAttr(name, value string)`
- `IndicatorOnlyAttrQuery(query, name, value string)`
- `IndicatorOnlyAttrQueryParent(query, name, value string)`
- `IndicatorOnlyAttrQueryAll(query, name, value string)`

---

### Class (Add)

Adds CSS classes temporarily, then removes them even if originally present.

```go
type IndicatorClass struct {
  Selector Selector
  Class    string
}
```

Helpers:
- `IndicatorOnlyClass(class string)`
- `IndicatorOnlyClassQuery(query, class string)`
- `IndicatorOnlyClassQueryParent(query, class string)`
- `IndicatorOnlyClassQueryAll(query, class string)`

---

### Class (Remove)

Removes CSS classes temporarily, then restores them even if not originally present.

```go
type IndicatorClassRemove struct {
  Selector Selector
  Class    string
}
```

Helpers:
- `IndicatorOnlyClassRemove(class string)`
- `IndicatorOnlyClassRemoveQuery(query, class string)`
- `IndicatorOnlyClassRemoveQueryParent(query, class string)`
- `IndicatorOnlyClassRemoveQueryAll(query, class string)`

---

## Example

### Single Indicator (Helper)

```templ
@doors.ASubmit[loginData]{
  Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
  Scope:     doors.ScopeOnlyBlocking(),
  On: func(ctx context.Context, r doors.RForm[loginData]) bool {
    // logic
    return true
  },
}
```

---

### Multiple Indicators (Manual)

```templ
@doors.ASubmit[loginData]{
  Indicator: []doors.Indicator{
    doors.IndicatorAttr{
      Selector: doors.SelectorQuery("#login-submit"),
      Name:     "aria-busy",
      Value:    "true",
    },
    doors.IndicatorContent{
      Selector: doors.SelectorQuery("#login-submit"),
      Content:  "Wait...",
    },
  },
  On: func(ctx context.Context, r doors.RForm[loginData]) bool {
    // logic
    return true
  },
}
```
