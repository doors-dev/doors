# Indication

The Indication API applies **temporary UI changes** to DOM elements. These changes can modify **content**, **attributes**, and **CSS classes** and are **restored automatically** when the indication ends. 

You can provide a slice of `doors.Indicator` to introduce multiple temporary UI changes.

## Concept

* `Indicator` field accepts a slice of indicators (to perform multiple temporary UI changes) 
* **You can create a simple indication with helper functions, such as `doors.IndicatorClass("active")`** or take full control with `[]doors.Indicator{/* indicators */}`
* Indicator contains a selector parameter (target itself, query, ancestor query) and specific indication settings
* **Indicators can overlap on the same elements** (for example, you have a global "loader" animation  for all actions) 
  * A later indication may queue while an earlier one is active
  * Upon completion of the first, the next indication applies if it is still required
  * After the final indication completes, the element returns to its original baseline.


## Selectors

### `SelectTarget`

Apply the indication to the event target

### `SelectQuery(query: string)`

Apply the indication to the element found by the `document.querySelector``

### `SelectParentQuery(query: string)`

Apply the indication to the closest target acnestor that matches query

## Indicators

Indicators define the **type of temporary UI change**. They all follow the same pattern:

- A struct with `Selector` and specific fields.
- Helper functions for common cases:
  - Target itself (`Indicator*`)
  - Target by query (`Indicator*Query`)
  - Target by parent query (`Indicator*QueryParent`)

### Content 

Temporary sets the provided string to innerHTML property of element

> ❗**WARNING** ❗ no sanitization applied

```templ
type ContentIndicator struct {
	Selector Selector // Element selector
	Content  string   // HTML content to display
}
```

**Helpers:**

* `IndicatorContent(content string) []Indicator`
   Changes the content of the target element.
* `IndicatorContentQuery(query string, content string) []Indicator`
   Changes the content of the first element matching query.
* `IndicatorContentQueryParent(query string, content string) []Indicator`
   Changes the content of the closest matching ancestor.

### Attribute

Temporarily sets an element attribute to a value. Attributes not present before will be removed afterwards.

```templ
type AttrIndicator struct {
	Selector Selector // Element selector
	Name     string   // Attribute name
	Value    string   // Temporary attribute value
}
```

**Helpers:**

- `IndicatorAttr(name string, value string) []Indicator`
   Sets attribute on the target element.
- `IndicatorAttrQuery(query string, name string, value string) []Indicator`
   Sets attribute on an element found via query.
- `IndicatorAttrQueryParent(query string, name string, value string) []Indicator`
   Sets attribute on the closest matching ancestor.

### Class

Temporarily adds a CSS class or classes. They will be removed automatically.

```templ
type ClassIndicator struct {
	Selector Selector // Element selector
	Class    string   // One or more classes to add
}
```

> ⚠️ Even if class was present originally, it will be removed after the indication completes.

**Helpers:**

- `IndicatorClass(class string) []Indicator`
   Adds class to the target element.
- `IndicatorClassQuery(query string, class string) []Indicator`
   Adds class to element found via query.
- `IndicatorClassQueryParent(query string, class string) []Indicator`
   Adds class to closest matching ancesto

### Remove class

Temporarily removes a CSS class or classes. They will be restored automatically.

```templ
type ClassRemoveIndicator struct {
	Selector Selector // Element selector
	Class    string   // One or more classes to remove
}
```

> ⚠️ Even if class was absent originally, it will be added after the indication completes.

**Helpers:**

- `IndicatorClassRemove(class string) []Indicator`
   Removes class from the target element.
- `IndicatorClassRemoveQuery(query string, class string)` []Indicator
   Removes class from element found via query.
- `IndicatorClassRemoveQueryParent(query string, class string) []Indicator`
   Removes class from closest matching ancestor.

## Example

### Single indicator with helper function

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

### Multiple indicators with manual slice creation

```templ
@doors.ASubmit[loginData]{
		Indicator: []doors.Indicator{
		  // indicate with attribute
			doors.AttrIndicator{
				// select by query
				Selector: doors.SelectorQuery("#login-submit"),
				Name:     "aria-busy",
				Value:    "true",
			},
			// indicate with content
			doors.ContentIndicator{
				Selector: doors.SelectorQuery("#login-submit"),
				Content:  "Wait...",
			},
		},
		Scope: doors.ScopeBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			/* logic */
			return true
		},
}
```



