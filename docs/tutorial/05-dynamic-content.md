# Dynamic Content

Our structure:

* Main catalog page 
  *Show the list of categories*
* Category catalog page
  *Show a list of items of the category and an item card in a pop-up.*

## 1. Prepare templates for different path variants

### Catalog Main Page

`./catalog/main.teml`

```templ
package catalog

import "github.com/doors-dev/doors"

// test href to show category page
var testCatHref = doors.AHref {
	Model: Path{
		IsCat: true,
		CatId: "test_cat",
	},
}

templ main() {
	<h1>Catalog</h1>
	// attach to <a>
	@testCatHref
	<a>Test Category</a>
}
```

`./catalog/cat.templ`

```templ
package catalog

import "github.com/doors-dev/doors"

// test href to go main catalog page
var testMainHref = doors.AHref{
	Model: Path{
		IsMain: true,
	},
}

templ category() {
	<h1>Category</h1>
	// attach to <a>
	@testMainHref
	<a>Go back</a>
}
```



## 2. Enable dynamic page updates

### Save beam to page field.

```templ
type catalogPage struct {
	// add field
	beam doors.SourceBeam[Path]
}

func (c *catalogPage) Render(b doors.SourceBeam[Path]) templ.Component {
	// save it
	c.beam = b
	return common.Template(c)
}

```

Now we can use `Beam` with Path Model inside Body(). Let's enable dynamic updates explicitly.

`./catalog/page.templ`

```templ
templ (c *catalogPage) Body() {
	// {{ ... }} used write regular golang code inside templ
	{{
	// initialize dynamic element
	door := doors.Door{}
	// subscribe to our beam
	c.beam.Sub(ctx, func(ctx context.Context, p Path) bool {
		// depending on path variant marker set Door content
		if p.IsMain {
			door.Update(ctx, main())
		} else {
			door.Update(ctx, category())
		}
		// false means not done, keep sub active
		return false
	})
	}}
	@door
}

```

> Visit the catalog page and try our first dynamic page update! Congrats!

### Refactor to @doors.Sub helper

Beam and Door are elementary building pieces. But sometimes you don't want so much control (and boilerplate), so let's refactor `Body()` method to use a helper component that combines `Beam` an `Door`

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

It does precisely the same under the hood.

Our final `./catalog/page.templ`

```templ
package catalog

import (
	"github.com/derstruct/doors-starter/common"
	"github.com/doors-dev/doors"
)

type Path = common.CatalogPath

type catalogPage struct {
	beam doors.SourceBeam[Path]
}

func (c *catalogPage) Render(b doors.SourceBeam[Path]) templ.Component {
	c.beam = b
	return common.Template(c)
}

templ (c *catalogPage) Head() {
	<title>catalog</title>
}

templ (c *catalogPage) Body() {
	@doors.Sub(c.beam, func(p Path) templ.Component {
		if p.IsMain {
			return main()
		}
		return category()
	})
}
```





