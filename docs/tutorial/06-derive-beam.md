# State Derive

In the previous part, we subscribed to the path updates:

```templ
func (c *catalogPage) Body() templ.Component {
	return doors.Sub(c.path, func(p Path) templ.Component {
		if p.IsMain {
			return main()
		}
		return category()
	})
}.
```

It's simple, but not the most optimal solution. It will rerender the whole page content on **any path model update**. 

However, at this level, we only need updates when the path switches between the categories list and the category page.

To achieve this, we derive a new state piece and subscribe to it.

```templ
func (c *catalogPage) Body() templ.Component {
	// derive beam, that will only trigger updates when
	// we change from/to main variant. It does not depend
	// on page query param or item ID.
	b := doors.NewBeam(c.path, func(p Path) bool {
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

Beam triggers an update only if newValue != oldValue. Additionaly, you can use doors.NewBeamExt constructor to provide a custom distinction check function. Check out **beam.md**  for details on **beam** and **source beam**.

> Path Model beam uses `reflect.DeepEqual` under the hood.

As a general rule, always derive **beam** into the most specific pieces. It will not only optimize DOM updates, but also reduce DB load in many cases.

