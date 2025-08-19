# Attributes

Besides attribute syntax and features from [templ](https://templ.guide/syntax-and-usage/attributes), *doors* provides an API for system (framework-related) attributes.

System attributes are used for:

* event handlers (hooks)
* data passing
* dynamic links 
* dynamic href/src for files
* ....and more

>  Most of the system attributes are represented by struct types that start with `doors.A...`

## Magic Attributes

You can attach system attributes using magic attribute syntax. 
**Just render the attribute before the element to which you want to add it.**

```templ
// magic attribute insertion to handle click event on the button
@doors.AClick{
  // handler function, doors.REvent is http request wrapper
	On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
	   /* event processing */
	   
		// not done, keep hook active
		return false
	}
	/* optional fields */
}
<button>Click Me!</button> 
```

In this example, the framework will add event binding attributes to the following button. 

It's also allowed to **stack multiple magic attributes**

```temp
// attach onchange handler
@doors.AChange{
/* setup */
}
// attach onfocys handler
@doors.AFocus{
/* setup */
}
<input type="text" name="name">
```

Additionally,  you can **prepare attributes as an array in a separate function**:

```templ
func attrs() []doors.Attr {
  return []{doors.AChange{/* setup */}, doors.AFocus{/* setup */}}
}
```

To attach it, use `doors.Attributes([]doors.Attr)` component:

```templ
@doors.Attributes(attrs())
<input type="text" name="name">
```

>❌ If magic attribute is not followed by a valid HTML tag, it will be dropped, for example
>
>```templ
>@doors.AClick{
>/* setup */
>}
>Some text
><button>Click Me!</button> 
>```
>
>"Some text" interrupts attribute insertion!

### Pre-initialization

Sometimes you want to use the same handler for multiple elements. To do this, pre-initialize the attribute and then render it where you need it

```templ
// initialization
{{ onchange := doors.A(ctx, doors.AChange{
	/* setup */
}) }}

@onchange
<input type="radio" name="radio" value="option1"/>
@onchange
<inputtype="radio" name="radio" value="option2"/>

```

> Before rendering or initializing, there is no actual attribute object, `doors.A..` is only a  structure for configuration .

## Attribute Spread

Alternatively to **Magic Attributes,** you can utilize templ spread syntax

```templ
{{  onclick := doors.AClick{
	/* setup */
} }}

<button { doors.A(ctx, onclick)... }>Click Me!</button>
```

> ⚠️ One element can't have magic and spread attributes; one overwrites another



