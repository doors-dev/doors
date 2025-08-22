# Catalog Page

Mini app with 3 pages

* Display a list of categories.
* Display a list of items in a category.
* Display item card

> I assume that you have live reloading enabled. If you prefer not to use it, don't forget to run `templ generate` and restart.

## 1. Path Model

Let's add a new pages path model. 

`./common/path.go`

```templ
package common

type CatalogPath struct {
	IsMain bool `path:"/catalog"`                // show categories
	IsCat  bool `path:"/catalog/:CatId"`         // show items of category
	IsItem bool `path:"/catalog/:CatId/:ItemId"` // show item
	CatId  string
	ItemId int
	Page   *int `query:"page"` // query param for pagination (used pointer to avoid showing 0 default value)
}

// prev one, keep it
type HomePath struct {
	Main bool `path:"/"`
}
```

## 2. Page & Handler

In new package `catalog`

`./catalog/page.templ`

```templ
package catalog

import "github.com/doors-dev/doors"
import "github.com/derstruct/doors-tutorial/common"

type Path = common.CatalogPath

type catalogPage struct {
}

templ (c *catalogPage) Head() {
	<title>catalog</title>
}

templ (c *catalogPage) Body() {
	<h1>Catalog</h1>
}

/*
Instead of doing this:

templ (c *catalogPage) Render(b doors.SourceBeam[Path]) {
	@common.Template(c)
}

Because there is only one component `common.Template(c)`, we can return it directly:
*/

func (c *catalogPage) Render(b doors.SourceBeam[Path]) templ.Component {
    return common.Template(c)
}

// page request handler
func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&catalogPage{})
}
```

## 3. Router

`./main.go`

```go
package main

import (
	"net/http"

	"github.com/derstruct/doors-tutorial/catalog"
	"github.com/derstruct/doors-tutorial/home"
	"github.com/doors-dev/doors"
)

func main() {
	r := doors.NewRouter()
	r.Use(doors.ServePage(home.Handler))

	// our new catalog page
	r.Use(doors.ServePage(catalog.Handler))

	err := http.ListenAndServeTLS(":8443", "localhost+2.pem", "localhost+2-key.pem", r)
	if err != nil {
		panic(err)
	}
}
```

## 4. Check

Visit https://localhost:8443/catalog 

## 5. Styling 

To save some time, let's use PicoCSS default styles for our tutorial purposes. Include CDN styles in our template.

`./common/page.templ`

```templ 
package common

import "github.com/doors-dev/doors"

type Page interface {
	Head() templ.Component
	Body() templ.Component
}

templ Template(p Page) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<meta name="color-scheme" content="light dark"/>
			@doors.Include()

			// same as <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"/>
			// but adds CSP headers (useful if you enable CSP later), please refer to docs/ref/imports.md for details
			@doors.Imports(doors.ImportStyleExternal{
				Href: "https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css",
			})
			@p.Head()
		</head>
		<body>
			@p.Body()
		</body>
	</html>
}
```

## 5. Menu 

We have two pages now, let's cook some nav:

`./common/components.templ`

```templ
package common

templ menu() {
	<nav>
		<ul>
			<li><strong>doors tutorial</strong></li>
		</ul>
		<ul>
			<li>
				<a href="/">home</a>
			</li>
			<li>
				<a href="/catalog">services</a>
			</li>
		</ul>
	</nav>
}
```

include it in our template

`./common/page.templ`

```templ
templ Template(p Page) {
		/* ... */
    <body>
        // PicoCSS container
        <main class="container">
            // menu component
            @menu()
            @p.Body()
        </main>
    </body>
		/* ... */
}

```

> Confirm that it works in browser

## 5. Idiomatic Menu

*doors* enables you to prepare href attributes in a type-safe manner and has extensive tooling for active link highlighting built in.

`/common/components.templ`

```templ
package common

import "github.com/doors-dev/doors"

templ menu() {
	<nav>
		<ul>
			<li><strong>doors tutorial</strong></li>
		</ul>
		<ul>
			<li>
			  // Magic attribute syntax, attribute attached to the next element
				@doors.AHref{
					// href Path Model
					Model: HomePath{},
					// active link highlighting settings
					Active: doors.Active{
						// indicate active link with attribute aria-current="page"
						Indicator: doors.IndicatorAttr("aria-current", "page"),
					},
				}
				<a>home</a>
			</li>
			<li>
			  // Magic attribute syntax, attribute attached to the next element
				@doors.AHref{
					// href Path Model
					Model: CatalogPath{
						// marker of variant
						IsMain: true,
					},
					// active link highlighting settings
					Active: doors.Active{
						// indicate active link with attribute aria-current="page"
						Indicator: doors.IndicatorAttr("aria-current", "page"),
						// page path must start with href value to enable highlighting
						PathMatcher: doors.PathMatcherStarts(),
						// ignore query params 
						QueryMatcher: doors.QueryMatcherIgnore(),
					},
				}
				<a>services</a>
			</li>
		</ul>
	</nav>
}
```

> Navigation between the home page and the catalog causes a new page load, which happens because you are navigating between different path models. We learn about dynamic pages in next article.

Please refer to **href-and-src.md** for full details on href. Alternatively to **Magic Attribute** syntax you can use attribute spread, check out **attributes.md** for details.

## 6. Conclusion

Visit https://localhost:8443/ to checkout that switching between pages works and active link indication applied.
