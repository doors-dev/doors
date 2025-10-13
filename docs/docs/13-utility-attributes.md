# Attribute Utilities

Utility attributes for building and dynamically updating HTML attributes.  
Provid control over attribute rendering, updates, and composition at runtime.

---

## `ADyn`

Dynamic attributes let you update or toggle HTML attributes at runtime without re-rendering the whole element.

```go
// ADyn is a dynamic attribute that can be updated at runtime.
type ADyn interface {
  doors.Attr
  // Value sets the attribute's value.
  Value(ctx context.Context, value string)
  // Enable adds or removes the attribute.
  Enable(ctx context.Context, enable bool)
}

// NewADyn returns a new dynamic attribute with the given name, value, and state.
func NewADyn(name string, value string, enable bool) ADyn
```

Use cases:
- Dynamically show/hide attributes like `disabled`, `hidden`, or `checked`
- Update `aria-*` attributes or custom data attributes while preserving the DOM node

Example:

```templ
{{ disabled := doors.NewADyn("disabled", "", false) }}

<button { doors.A(ctx, disabled)... }>Submit</button>

@doors.AClick{
  On: func(ctx context.Context, ev doors.REvent[doors.PointerEvent]) bool {
    disabled.Enable(ctx, true)
    return false
  },
}
```

---

## `AMap`

Represents a map of HTML attributes.  
Used for rendering or spreading multiple attributes at once.

```go
type AMap map[string]any
```

Example:

```templ
@doors.AMap{"id": "btn", "data-role": "submit"}
<button>Submit</button>
```

---

## `AOne`

Represents a single HTML attribute key-value pair.

```go
type AOne [2]string
```

Example:

```templ
@doors.AOne{"id", "submit-btn"}
<button>Submit</button>
```

---

## `AClass`

Helper for adding one or more CSS classes.  
Equivalent to `AOne{"class", value}`, but joins multiple values with a space.

```go
func AClass(class ...string) AOne
```

Example:

```templ
@doors.AClass("btn", "btn-primary")
<button>Submit</button>
```

---

## Composition and Overwrite Rules

- Attributes overwrite each other when duplicated.
- The `"class"` attribute is the only exception: values are concatenated instead of replaced.



