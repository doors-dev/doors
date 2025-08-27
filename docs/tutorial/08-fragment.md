# Fragment And Styles

Using functional components (individual templ functions) is fine, but in many cases, it's easier to deal with structured data. 

## 1. Category Data

Our current category functional component:

```templ
templ category() {
	<h1>Category</h1>
	// href to go test category page
	@doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
	<a>Go back</a>
}
```

To query category data from the DB, we need the category ID. Obtain the category ID from the **path beam**: 

```templ
templ category(path doors.SourceBeam[Path]) 
```

`./catalog/page.templ`

```templ
/* ... */
func (c *catalogPage) Body() templ.Component {
	b := doors.NewBeam(c.path, func(p Path) bool {
		return p.IsMain
	})
	return doors.Sub(b, func(isMain bool) templ.Component {
		if isMain {
			return main()
		}
		// add path beam dependency
		return category(c.path)
	})
}
/* ... */
```

Next, derive the category ID **beam** and subscribe to it:

`./catalog/category.templ`

```templ
templ category(path doors.SourceBeam[Path]) {
  // subscribe to the derived beam
	@doors.Sub(doors.NewBeam(path, func(p Path) string {
	  // derive category id
		return p.CatId
	}), func(catId string) templ.Component {
	  // render new category when when ID changes
		return categoryContent(catId)
	})
	@doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
	<a>Go back</a>
}

templ categoryContent(catId string) {
  // query category
	{{ cat, ok := driver.Cats.Get(catId) }}
	// category found
	if ok {
		<hgroup>
			<h1>{ cat.Name }</h1>
			<p>{ cat.Desc } </p>
		</hgroup>
	} else {
		<div>
			<mark>Not Found</mark>
		</div>
	}
}

```

## 2. Refactor to Fragment

Instead of a templ function, we can have a struct type, which will allows to group all pieces nicely in its methods/fields.

```templ
type categoryFragment struct {
	path doors.SourceBeam[Path]
}

// doors.Fragment must have Render() method
templ (f *categoryFragment) Render() {
	@doors.Sub(doors.NewBeam(f.path, func(p Path) string {
		return p.CatId
	}), func(catId string) templ.Component {
		return f.content(catId)
	})
	@doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
	<a class="contrast">Go back</a>
}

templ (f *categoryFragment) content(catId string) {
	{{ cat, ok := driver.Cats.Get(catId) }}
	if ok {
		<hgroup>
			<h1>{ cat.Name }</h1>
			<p>{ cat.Desc } </p>
		</hgroup>
	} else {
		<div>
			<mark>Not Found</mark>
		</div>
	}
}

```

To keep it compatible with the previous implementation, wrap fragment in a component:

```templ
func category(path doors.SourceBeam[Path]) templ.Component {
	// doors.F - fragment renderer, also can be used in templates @doors.F(doors.Fragment)
	return doors.F(&categoryFragment{
		path: path,
	})
}
```

## 3. Layout and Styles

### Styles

Prepare styles for layout with side nav.

```templ
templ (f *categoryFragment) style() {
	// converts inline style to cachable stylesheet with href
	@doors.Style() {
        <style>
            .category {
                display: flex;
                gap: calc(var(--pico-typography-spacing-vertical) * 3);
            }
            aside {
                flex: 0 0 auto; 
            }
            main {
                flex: 1;
            }
        </style>
	}
}

```

### Nav Component

```templ
templ (d *categoryFragment) nav() {
	<nav>
		<ul>
			<li>
				// link to main category variant
				@doors.AHref{
					Model: Path{
						IsMain: true,
					},
				}
				<a class="secondary">Go back</a>
			</li>
			// query categories
			for _, cat := range driver.Cats.List() {
				<li>
					// attach href 
					@doors.AHref{
						// link to category
						Model: Path{
							IsCat: true,
							CatId: cat.Id,
						},
						Active: doors.Active{
							// indicate with attribute
							Indicator: doors.IndicatorAttr("aria-current", "page"),
							// path must start with href
							PathMatcher: doors.PathMatcherStarts(),
							// ignore query params
							QueryMatcher: doors.QueryMatcherIgnore(),
						},
					}
					<a class="contrast">{ cat.Name }</a>
				</li>
			}
		</ul>
	</nav>
}
```

### Layout

Assemble all tougher in fragment render function

```templ
templ (f *categoryFragment) Render() {
	@f.style()
	<div class="category">
		<aside>
			@f.nav()
		</aside>
		<div class="content">
			@doors.Sub(doors.NewBeam(f.path, func(p Path) string {
				return p.CatId
			}), func(catId string) templ.Component {
				return f.content(catId)
			})
		</div>
	</div>
}
```

## Conclusion

By this time, you have learned one of the core frameworks' mechanics: `Beam` derivation and subscription.   As you probably already understand, the framework establishes an explicit and direct connection between data and HTML updates, without relying on hidden virtual DOM diffs (**diff data, not DOM**).

In the following chapters, we will cover authentication and event handling.



---

Next: [Form Auth](./09-form-auth.md)

