# Dynamic Content

Desired catalog page structure:

* **Main**: Show the list of categories
* **Category**: Show a list of items in the category
  * **Item Card**: Pop-up

## 1. Create templates for all path options

### Catalog Main Page

`./catalog/main.templ`

```templ
package catalog

import "github.com/doors-dev/doors"

templ main() {
	<h1>Catalog</h1>

	// href to  test category page
	@doors.AHref{
		Model: Path{
			IsCat: true,
			CatId: "test_cat",
		},
	}
	<a>Test Category</a>
}
```

`./catalog/category.templ`

```templ
package catalog

import "github.com/doors-dev/doors"

templ category() {
	<h1>Category</h1>
	// back to main
	@doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
	<a>Go back</a>
}

```



## 2. Enable dynamic page updates on path change

### Save the path beam to the page field.

**Beam** represents a reactive, changing value stream. The framework provides a **beam** that holds the **path model**.

```templ
type catalogPage struct {
  // add fields
	path doors.SourceBeam[Path]
}

/* ... */

func (c *catalogPage) Render(b doors.SourceBeam[Path]) templ.Component {
	// save it
	c.path = b
	return common.Template(c)
}

```

Now use the **path beam** inside the body to enable dynamic updates explicitly.

`./catalog/page.templ`

```templ
templ (c *catalogPage) Body() {
	// doors.E evaluates the function and renders the return value
	@doors.E(func(ctx context.Context) templ.Component {
		// initialize dynamic container
		door := doors.Door{}
		// subscribe to path updates
		c.path.Sub(ctx, func(ctx context.Context, p Path) bool {
			// depending on the path variant marker, set the door content.
			if p.IsMain {
				door.Update(ctx, main())
			} else {
				door.Update(ctx, category())
			}
			// false means not done; keep sub active
			return false
		})
		// render dynamic container
		return &door
	})
}
```

> Visit the catalog page and experience our first dynamic page update! Congrats!

### Refactor to @doors.Sub helper

**Beam** and **door** are basic building blocks. But sometimes you want less control (and boilerplate), so let's refactor the `Body()` method to use a helper component that combines **beam** and  **door**.

```templ
templ (c *catalogPage) Body() {
    // subscribe helper component, updates node based on func output
	@doors.Sub(c.path, func(p Path) templ.Component {
		if p.IsMain {
			return main()
		}
		return category()
	})
}
```

It does precisely the same under the hood.

> Now the page updates dynamically when we navigate inside the catalog page.

Since we have only one component, we can replace the templ function with a standard one that follows the interface.

```templ
func (c *catalogPage) Body() templ.Component {
	return doors.Sub(c.path, func(p Path) templ.Component {
		if p.IsMain {
			return main()
		}
		return category()
	})
}
```



