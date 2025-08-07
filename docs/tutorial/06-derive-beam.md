# State Derive

In the previous part, we subscribed to path changes:

```templ
templ (c *catalogPage) Body() {
	@doors.Sub(c.beam, func(p Path) templ.Component {
		if p.IsMain {
			return main()
		}
		return category()
	})
}
```

It's simple, but not optimal. It will update the whole page body on any path changes.

However, at the <body> level, we only need updates when the path switches between the categories list and the category page.

To achieve this, we derive a new state piece and subscribe to it.

```templ
templ (c *catalogPage) Body() {
	{{
	  // create new beam, cast bool value from Path
    b := doors.NewBeam(c.beam, func(p Path) bool {
        return p.IsMain
    })
	}}
	@doors.Sub(b, func(isMain bool) templ.Component {
		if isMain {
			return main()
		}
		return category()
	})
}
```

For cleaner code, we can convert it to a regular function (because the template has one component, we can just return it)

```templ
func(c * catalogPage) Body() templ.Component {
    b := doors.NewBeam(c.beam, func(p Path) bool {
        return p.IsMain
    })
    return doors.Sub(b, func(isMain bool) templ.Component {
        if isMain {
            return main()
        }
        return category()
    })
}
```

> `Beam` by default triggers update only if `newValue != oldValue`. You can use doors.NewBeamExt constructor to provide a custom equality check function. 
>
> Path Model beam uses `reflect.DeepEqual` under the hood.



