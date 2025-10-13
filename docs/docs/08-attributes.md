# Attributes

Besides attribute syntax and features from [templ](https://templ.guide/syntax-and-usage/attributes), *doors* provides a special attribute system for:

- event handlers (hooks)
- data passing
- dynamic links
- dynamic href/src for files
- and more

> Most system attributes are struct types prefixed with `doors.A`. 

---

## Magic Attributes

You can attach attributes using *magic attribute* syntax.  
Render the attribute immediately before the element you want to modify.

```templ
// magic attribute insertion to handle click event on the button
@doors.AClick{
  // handler function, doors.REvent is the http request wrapper
  On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
     /* event processing */
     // not done, keep hook active
     return false
  }
  /* optional fields */
}
<button>Click Me!</button> 
```

In this example, the framework adds the event binding attributes to the following button.

You can **stack multiple magic attributes**:

```templ
// attach onchange handler
@doors.AChange{
  /* setup */
}
// attach onfocus handler
@doors.AFocus{
  /* setup */
}
<input type="text" name="name">
```

You can also **prepare attributes as an array in a separate function**:

```go
func attrs() []doors.Attr {
  return []doors.Attr{
    doors.AChange{/* setup */},
    doors.AFocus{/* setup */},
  }
}
```

Attach them using `@doors.Any()`:

```templ
@doors.Any(attrs())
<input type="text" name="name">
```

❌ If a valid HTML tag does **not** follow a magic attribute, it is dropped:

```templ
@doors.AClick{
  /* setup */
}
Some text // interrupts attribute attachment!
<button>Click Me!</button> 
```

---

### Pre-initialization

When you want to reuse the same handler for multiple elements, pre-initialize the attribute and then render it where needed.

```templ
// initialization
{{ onchange := doors.A(ctx, doors.AChange{
  /* setup */
}) }}

@onchange
<input type="radio" name="radio" value="option1"/>
@onchange
<input type="radio" name="radio" value="option2"/>
```

> Before rendering or initializing, there is no actual attribute object;  
> `doors.A...{}` is only a configuration structure.

---

## Attribute Spread

Alternatively to *Magic Attributes*, you can use templ’s **spread syntax**.

```templ
{{ onclick := doors.AClick{
  /* setup */
} }}

<button { doors.A(ctx, onclick)... }>Click Me!</button>
```

> ⚠️ One element cannot have both magic and spread attributes;  
> one form overwrites the other.
