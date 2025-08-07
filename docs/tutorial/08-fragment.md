# Fragment And Styles

##  1. List Categories From DB

> If you used the DB driver code from the previous example, it will populate the categories table with sample data.

`./catalog/main.go`

```templ
package catalog

import (
    "github.com/doors-dev/doors"
	"github.com/derstruct/doors-starter/driver"
)


// helper function to prepare href attribute for category link
func catHref(cat driver.Cat) doors.Attr {
	return doors.AHref{
		Model: Path{
			IsCat: true,
			CatId: cat.Id,
		},
	}
}

templ main() {
	<h1>Catalog</h1>
	// query and iterate categories list
	for _, cat := range driver.Cats.List() {
		<article>
			<header>
				// link
				<a { doors.A(ctx, catHref(cat))... }>
					{ cat.Name }
				</a>
			</header>
			{ cat.Desc }
		</article>
	}
}

```

Now, on the `/catalog` page, you should see a list of clickable cards 

## 2. Category Fragment

Using functional components (individual templ functions) is fine, but in many cases it's easirer to deal with structured data. Let's convert our `category`  to  `Fragment`

`./catalog/cat.templ`

```templ
package catalog

import "github.com/doors-dev/doors"

func newCategoryFragment() *categoryFragment {
	return &categoryFragment{}
}

type categoryFragment struct{}

// fragment must have render method
templ (c *categoryFragment) Render() {
	<h1>Category</h1>
	<a { doors.A(ctx, c.backHref() )... }>Go back</a>
}

func (c *categoryFragment) backHref() doors.Attr {
	return doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
}
```

To render fragment, we use `doors.F(doors.Fragment)`  helper

`./catalog/page.templ`

```templ
func (c *catalogPage) Body() templ.Component {
    b := doors.NewBeam(c.beam, func(p Path) bool {
        return p.IsMain
    })
    return doors.Sub(b, func(isMain bool) templ.Component {
      if isMain {
        return main()
      }
      // return fragment component
      return doors.F(newCategoryFragment())
    })
}

```

## 3. Categories menu

To enable a switch between categories, we add `aside` with navigation

`./catalog/cat.templ`

```templ
package catalog

import (
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

func newCategoryFragment() *categoryFragment {
	return &categoryFragment{}
}

type categoryFragment struct{}

templ (c *categoryFragment) Render() {
	<div>
		<aside>
			// nav component
			@c.nav()
		</aside>
		<div>
			// main content
			<h1>Category</h1>
		</div>
	</div>
}

templ (c *categoryFragment) nav() {
	<a { doors.A(ctx, c.backHref() )... }>Go back</a>
	<ul>
		for _, cat := range driver.Cats.List() {
			<li><a { doors.A(ctx, c.catHref(cat))... }>{ cat.Name }</a></li>
		}
	</ul>
}

func (c *categoryFragment) catHref(cat driver.Cat) doors.Attr {
	return doors.AHref{
		Model: Path{
			IsCat: true,
			CatId: cat.Id,
		},
		Active: doors.Active{
			Indicator: doors.IndicatorClass("contrast"),
		},
	}
}

func (c *categoryFragment) backHref() doors.Attr {
	return doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
}

```

But without styles, it's not displaying how we want it. Let's add some style. Easiest way too add styles is by using `doors.Style` helper component. It will convert inline styles to one with href (cachable).

> You can also write styles in a separate file and import it; we will learn how to do it later.

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) Render() {
	@c.style()
	<div class="cat">
		<aside>
			@c.nav()
		</aside>
		<div class="content">
			<h1>Category</h1>
		</div>
	</div>
}

templ (c *categoryFragment) style() {
  // style helper component
  // doors minifies styles, but does not apply any scoping
	@doors.Style() {
       <style>
            .cat {
                display: flex;
                flex-direction: row;
                gap: calc(var(--pico-typography-spacing-vertical) * 3);
            }
            .cat .content {
                display: flex;
                flex-direction: column;
                align-items: flex-start;
            }
       </style>
	}
}

```

## 4. Dynamic Content

#### Inject Path Beam

`./catalog/cat.templ`

```templ
func newCategoryFragment(path doors.Beam[Path]) *categoryFragment {
	return &categoryFragment{
		path: path,
	}
}

type categoryFragment struct {
		path doors.SourceBeam[Path]
}
```

`./catalog/page.templ`

```templ
func (c *catalogPage) Body() templ.Component {
	b := doors.NewBeam(c.beam, func(p Path) bool {
		return p.IsMain
	})
	return doors.Sub(b, func(isMain bool) templ.Component {
		if isMain {
			return main()
		}
		return doors.F(newCategoryFragment(c.beam))
	})
}
```

### Dynamic Content

Subscribe to our beam to display the selected category

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) Render() {
	@c.style()
	<div class="cat">
		<aside>
			@c.nav()
		</aside>
		<div class="content">
		  // render component
			@c.content()
		</div>
	</div>
}

func (c *categoryFragment) content() templ.Component {
  // subscribe component
	return doors.Sub(
	  // to derived beam
		doors.NewBeam(c.path, func(p Path) string {
			return p.CatId
		}),
		func(catId string) templ.Component {
			cat, ok := driver.Cats.Get(catId)
			if !ok {
				return c.notFound()
			}
			return c.cat(cat)
		},
	)
}


templ (c *categoryFragment) cat(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
}

templ (c *categoryFragment) notFound() {
	<h1>Category Not Found</h1>
}
```

## Conclusion

By this time, you have learned one of the core frameworks' mechanics: `Beam` derivation and subscription.   As you probably already understand, doors allow for establishing an explicit and direct connection between data and HTML updates, without relying on hidden virtual DOM diffs.

In the following chapters, we will cover authentication and event handling.

### Final Code

`./catalog/page.templ`

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

func (c *catalogPage) Body() templ.Component {
    b := doors.NewBeam(c.beam, func(p Path) bool {
        return p.IsMain
    })
	return doors.Sub(b, func(isMain bool) templ.Component {
		if isMain {
			return main()
		}
		return doors.F(newCategoryFragment(c.beam))
	})
}
```



`./catalog/cat.templ`

```templ
package catalog

import (
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

func newCategoryFragment(path doors.Beam[Path]) *categoryFragment {
	return &categoryFragment{
		path: path,
	}
}

type categoryFragment struct {
	path       doors.SourceBeam[Path]
}

templ (c *categoryFragment) Render() {
	@c.style()
	<div class="cat">
		<aside>
			@c.nav()
		</aside>
		<div class="content">
			@c.content()
		</div>
	</div>
}

func (c *categoryFragment) content() templ.Component {
	return doors.Sub(
		doors.NewBeam(c.path, func(p Path) string {
			return p.CatId
		}),
		func(catId string) templ.Component {
			cat, ok := driver.Cats.Get(catId)
			if !ok {
				return c.notFound()
			}
			return c.cat(cat)
		},
	)
}

templ (c *categoryFragment) cat(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
}

templ (c *categoryFragment) notFound() {
	<h1>Category Not Found</h1>
}

templ (c *categoryFragment) style() {
	@doors.Style() {
		<style>
            .cat {
                display: flex;
                flex-direction: row;
                gap: calc(var(--pico-typography-spacing-vertical) * 3);
            }
            .cat .content {
                display: flex;
                flex-direction: column;
                align-items: flex-start;
            }
        </style>
	}
}

templ (c *categoryFragment) nav() {
	<a { doors.A(ctx, c.backHref() )... }>Go back</a>
	<ul>
		for _, cat := range driver.Cats.List() {
			<li><a { doors.A(ctx, c.catHref(cat))... }>{ cat.Name }</a></li>
		}
	</ul>
}

func (c *categoryFragment) catHref(cat driver.Cat) doors.Attr {
    return doors.AHref{
        Model: Path{
            IsCat: true,
            CatId: cat.Id,
        },
        Active: doors.Active{
            Indicator: doors.IndicatorClass("contrast"),
        },
    }
}

func (c *categoryFragment) backHref() doors.Attr {
	return doors.AHref{
		Model: Path{
			IsMain: true,
		},
	}
}

```

`./catalog/main.templ`

```templ
package catalog

import (
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

func catHref(cat driver.Cat) doors.Attr {
	return doors.AHref{
		Model: Path{
			IsCat: true,
			CatId: cat.Id,
		},
	}
}

templ main() {
	<h1>Catalog</h1>
	for _, cat := range driver.Cats.List() {
		<article>
			<header>
				<a { doors.A(ctx, catHref(cat))... }>
					{ cat.Name }
				</a>
			</header>
			{ cat.Desc }
		</article>
	}
}

```

