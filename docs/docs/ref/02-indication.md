# Indication

The Indication API applies **temporary UI changes** to DOM elements. These changes can modify **content**, **attributes**, and **CSS classes** and are **restored automatically** when the indication ends. 

You can provide a slice of `doors.Indicator` to introduce multiple temporary UI changes.

## Concept

* The `Indicator` type represents a temporary DOM modification.  
* The `Indicator` field in attributes accepts a slice of indicators (to perform multiple temporary UI changes).  
* **You can create a simple indication with helper functions (`IndicatorOnly*`)** or take full control with `[]doors.Indicator{/* indicators */}`.  
* Each `Indicator` struct contains a `Selector` parameter (target itself, query, ancestor query) and specific indication settings.  
* **Indicators can overlap on the same elements**. For example, you may apply a global loader indicator while also indicating local changes.  
  * Later indications queue if an earlier one is active.  
  * Once the first finishes, the next applies if still needed.  
  * After the final indication completes, the element returns to its original baseline.  

## Selectors

Selectors define **which DOM element** an indicator applies to.

### `SelectorTarget()`

Selects the element that triggered the event.

### `SelectorQuery(query string)`

Selects the first element matching a CSS query (e.g. `#id`, `.class`).

### `SelectorParentQuery(query string)`

Selects the closest ancestor matching a CSS query.

## Indicators

Indicators define the **type of temporary UI change**.

---

### Content

Temporarily replaces the `innerHTML` of the selected element.  
⚠️ **No sanitization is applied.**

```go
type IndicatorContent struct {
	Selector Selector // Target element
	Content  string   // Replacement content
}
```

**Helpers:**

- `IndicatorOnlyContent(content string)`  
- `IndicatorOnlyContentQuery(query, content string)`  
- `IndicatorOnlyContentQueryParent(query, content string)`

---

### Attribute

Temporarily sets an attribute on the selected element.  
If the attribute was not present before, it will be removed afterwards.

```go
type IndicatorAttr struct {
	Selector Selector // Target element
	Name     string   // Attribute name
	Value    string   // Attribute value
}
```

**Helpers:**

- `IndicatorOnlyAttr(name, value string)`  
- `IndicatorOnlyAttrQuery(query, name, value string)`  
- `IndicatorOnlyAttrQueryParent(query, name, value string)`

---

### Class (Add)

Temporarily adds CSS classes. They will be removed automatically.  
⚠️ Even if a class was present originally, it will still be removed after indication.

```go
type IndicatorClass struct {
	Selector Selector // Target element
	Class    string   // Space-separated classes
}
```

**Helpers:**

- `IndicatorOnlyClass(class string)`  
- `IndicatorOnlyClassQuery(query, class string)`  
- `IndicatorOnlyClassQueryParent(query, class string)`

---

### Class (Remove)

Temporarily removes CSS classes. They will be restored automatically.  
⚠️ Even if a class was absent originally, it will be added back afterwards.

```go
type IndicatorClassRemove struct {
	Selector Selector // Target element
	Class    string   // Space-separated classes
}
```

**Helpers:**

- `IndicatorOnlyClassRemove(class string)`  
- `IndicatorOnlyClassRemoveQuery(query, class string)`  
- `IndicatorOnlyClassRemoveQueryParent(query, class string)`

---

## Example

### Single indicator with helper

```templ
@doors.ASubmit[loginData]{
    Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
    Scope:     doors.ScopeOnlyBlocking(),
    On: func(ctx context.Context, r doors.RForm[loginData]) bool {
        /* logic */
        return true
    },
}
```

### Multiple indicators manually

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
        /* logic */
        return true
    },
}
```