# Pagination

## 1. Pagination 

quick & dirty

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) pages(cat driver.Cat) {
	@doors.Style() {
		<style>
        .pages {
            display: flex;
            justify-content: start;
            gap: 1rem;
        }
    </style>
	}
	<div class="pages">
	  // range page count from driver
		for i := range driver.Items.CountPages(cat.Id) {
			<a { doors.A(ctx, c.pageHref(cat, i))... } class="secondary">
				@doors.Text(i + 1)
			</a>
		}
	</div>
}

func (c *categoryFragment) pageHref(cat driver.Cat, page int) doors.Attr {
	var p *int = nil
	// first page is just no page param at all
	if page != 0 {
		p = &page
	}
	return doors.AHref{
		Active: doors.Active{
		  // indicate with an attribute
			Indicator:    doors.IndicatorAttr("aria-current", "true"),
			// match path start (/catalog/:CatId)
			PathMatcher:  doors.PathMatcherStarts(),
			// match query param page (not necessary in this case, just for demo)
			QueryMatcher: doors.QueryMatcherSome("page"),
		},
		Model: Path{
			IsCat: true,
			CatId: cat.Id,
			Page:  p,
		},
	}
}

```

> Paginator changes query param, but not the content for now.

## 2. Derive & Inject Page Beam

`./catalog/cat.templ`

```templ
func newCategoryFragment(path doors.SourceBeam[Path], authorized bool) *categoryFragment {
	return &categoryFragment{
		authorized: authorized,
		path:       path,
		// derive page beam
		page: doors.NewBeam(path, func(p Path) int {

			if p.Page == nil || *p.Page < 0 {
				return 0
			}
			return *p.Page
		}),
	}
}

type categoryFragment struct {
	authorized bool
	path       doors.SourceBeam[Path]
	// add page beam field
	page       doors.Beam[int]
	itemsNode  doors.Node
	loadNode   doors.Node
}

templ (c *categoryFragment) listItems(cat driver.Cat) {
	// inject beam value into context, alternative to @doors.Sub
	// it will rerender children on each update of the beam
	@doors.Inject("page", c.page) {
	  // use beam value from context to 
		{{ items := driver.Items.List(cat.Id, ctx.Value("page").(int)) }}
		if len(items) == 0 {
			No Items 
		} else {
			<div class="grid">
				<div>
					for i, item := range items {
						if i % 2 == 0 {
							@c.item(item, page)
						}
					}
				</div>
				<div>
					for i, item := range items {
						if i % 2 == 1 {
							@c.item(item, page)
						}
					}
				</div>
			</div>

		}
	}
	// pagination
	@c.pages(cat)
}

// added page param, to maintain page number when user clicks card link
templ (c *categoryFragment) item(item driver.Item, page int) {
	<article>
		<header><a { doors.A(ctx, c.itemHref(item.Cat, item.Id, page))... }>{ item.Name }</a></header>
		<kbd>
			Rating 
			@doors.Text(item.Rating)
		</kbd>
	</article>
}

// item href with page param
func (c *categoryFragment) itemHref(catId string, itemId int, page int) doors.Attr {
  var p *int 
  // hide 0 page in query
  if page != 0 {
      p = &page
  }
	return doors.AHref{
		Model: Path{
			IsItem: true,
			CatId:  catId,
			ItemId: itemId,
      Page: p,
		},
	}
}

```

> We did it with `@doors.Inject(...)` for learning purposes. It's actually better to reuse `c.itemsNode` by updating it in the page beam sub. 
>
> ```
>templ (c *categoryFragment) cat(cat driver.Cat) {
> 	...
> 	// doors.E helper renders output of the provided function
> 	@doors.E(func(ctx context.Context) templ.Component {
> 	  //derive page beam
> 		doors.NewBeam(c.path, func(p Path) int {
> 			if p.Page == nil || *p.Page < 0 {
> 				return 0
> 			}
> 			return *p.Page
> 		// subsribe
> 		}).Sub(ctx, func(ctx context.Context, page int) bool {
> 		  // update items node with actual page
> 			c.itemsNode.Update(ctx, c.listItems(cat, page))
> 			return false
> 		})
> 		// render items node
> 		return &c.itemsNode
> 	})
> }
> // add page as arg, no need to derive / inject
> templ (c *categoryFragment) listItems(cat driver.Cat, page int) {
> ...
> 		{{ items := driver.Items.List(cat.Id, page) }}
>  ...
> }
>  ```
> 
> 

## 3. Load More

Many aspects can be improved in pagination, but let's use a 'load more' button instead to learn a new, useful pattern.

#### First, load all pages up to the current

#### `./catalog/cat.templ`

```templ
templ (c *categoryFragment) listItems(cat driver.Cat) {
	// derive page beam from our path
	{{
  // read current page value
	page, _ := c.page.Read(ctx)
  // count pages
	pageCount := driver.Items.CountPages(cat.Id)
	}}
  // means no content
	if pageCount == 0 {
		No Items
	} else {
    // render all pages
		for i := range page + 1 {
			@c.itemsPage(cat, i)
		}
	}
	@c.pages(cat)
}


templ (c *categoryFragment) itemsPage(cat driver.Cat, page int) {
	{{ items := driver.Items.List(cat.Id, page) }}
	if len(items) > 0 {
		<div class="grid">
			<div>
				for i, item := range items {
					if i % 2 == 0 {
						@c.item(item, page)
					}
				}
			</div>
			<div>
				for i, item := range items {
					if i % 2 == 1 {
						@c.item(item, page)
					}
				}
			</div>
		</div>
	}
}

```

> Now there is no reaction to change of page query param — that's intentional, we removed HTML dependency on page beam.

### Load More With Node Replace

We will implement loading new page by placeholder node replacement

##### Add Struct Field

```templ
type categoryFragment struct {
	authorized bool
	path       doors.SourceBeam[Path]
	page       doors.Beam[int]
	itemsNode  doors.Node
	// new node
	loadNode   doors.Node
}


```

#### Add Page Render

```templ

// new bool argument last, because we will append only to the last rendered page
templ (c *categoryFragment) itemsPage(cat driver.Cat, page int, last bool) {
	{{ items := driver.Items.List(cat.Id, page) }}
	if len(items) > 0 {
		<div class="grid">
			<div>
				for i, item := range items {
					if i % 2 == 0 {
						@c.item(item, page)
					}
				}
			</div>
			<div>
				for i, item := range items {
					if i % 2 == 1 {
						@c.item(item, page)
					}
				}
			</div>
		</div>
		// if last page 
		if last {
		  // to avoid recursive render
			{{ c.loadNode.Clear(ctx) }}
      @c.loadNode
		}
	}
}
```

#### Load More Button And Method

```templ
templ (c *categoryFragment) listItems(cat driver.Cat) {
	{{
	page, _ := c.page.Read(ctx)
	pageCount := driver.Items.CountPages(cat.Id)
	}}
	if pageCount == 0 {
		No Items
	} else {
		for i := range page + 1 {
			@c.page(cat, i, i == page)
		}
	}
	// "load more" depends on page beam
	@doors.Inject("page", c.page) {
		{{ page := ctx.Value("page").(int) }}
    // If there is no mo pages, we don't need the button
		if page < pageCount - 1 {
		  @c.attachLoadMore(cat, page + 1)
			<button>Load More</button>
		}
	}
}

func (c *categoryFragment) attachLoadMore(cat driver.Cat, page int) doors.Attr {
    return doors.AClick {
        On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
        		// update page number
            c.path.Mutate(ctx, func(p Path) Path {
                p.Page = &page
                return p
            })
            
            //append new page
            c.loadNode.Replace(ctx, c.itemsPage(cat, page, true))
            return true
        },
    }
}
```

> But because old pages do not update when a new one loads, item links are not updated either, so we lose the page parameter query when clicking on some items.
>
> You can verify this by loading more content and then clicking on the first item, and you will see that the query parameter is removed.

#### Links Rerender

We need item links to be dependent on the page beam.

```templ
// we don't need page argument anymore
templ (c *categoryFragment) item(item driver.Item) {
	<article>
		<header>
		  // inject page beam
			@doors.Inject("page", c.page) {
			  // read it
				{{ page := ctx.Value("page").(int) }}
				<a { doors.A(ctx, c.itemHref(item.Cat, item.Id, page))... }>{ item.Name }</a>
			}
		</header>
		<kbd>
			Rating 
			@doors.Text(item.Rating)
		</kbd>
	</article>
}
/*  also, you need to remove the extra argument in itemsPage*/

```

## 4. Refactor

Current catalog fragment code:

`./catalog/cat.teml`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

func newCategoryFragment(path doors.SourceBeam[Path], authorized bool) *categoryFragment {
	return &categoryFragment{
		authorized: authorized,
		path:       path,
		page: doors.NewBeam(path, func(p Path) int {

			if p.Page == nil || *p.Page < 0 {
				return 0
			}
			return *p.Page
		}),
	}
}

type categoryFragment struct {
	authorized bool
	path       doors.SourceBeam[Path]
	page       doors.Beam[int]
	itemsNode  doors.Node
	loadNode   doors.Node
}

templ (c *categoryFragment) Render() {
	@c.style()
	@card(c.path, c.itemsNode.Reload, c.authorized)
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
	if c.authorized {
		<p>
			@createItem(cat, func(ctx context.Context) {
				c.itemsNode.Reload(ctx)
			})
		</p>
	}
	@c.itemsNode {
		@c.listItems(cat)
	}
}

templ (c *categoryFragment) listItems(cat driver.Cat) {
	{{
	page, _ := c.page.Read(ctx)
	pageCount := driver.Items.CountPages(cat.Id)
	}}
	if pageCount == 0 {
		No Items
	} else {
		for i := range page + 1 {
			@c.itemsPage(cat, i, i == page)
		}
	}
	@doors.Inject("page", c.page) {
		{{ page := ctx.Value("page").(int) }}
		if page < pageCount - 1 {
			@c.attachLoadMore(cat, page + 1)
			<button>Load More</button>
		}
	}
}

func (c *categoryFragment) attachLoadMore(cat driver.Cat, page int) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
			c.path.Mutate(ctx, func(p Path) Path {
				p.Page = page
				return p
			})
			c.loadNode.Replace(ctx, c.itemsPage(cat, page, true))
			return true
		},
	}
}

templ (c *categoryFragment) itemsPage(cat driver.Cat, page int, last bool) {
	{{ items := driver.Items.List(cat.Id, page) }}
	if len(items) > 0 {
		<div class="grid">
			<div>
				for i, item := range items {
					if i % 2 == 0 {
						@c.item(item)
					}
				}
			</div>
			<div>
				for i, item := range items {
					if i % 2 == 1 {
						@c.item(item)
					}
				}
			</div>
		</div>
		if last {
			{{ c.loadNode.Clear(ctx) }}
			@c.loadNode
		}
	}
}

templ (c *categoryFragment) item(item driver.Item) {
	<article>
		<header>
			@doors.Inject("page", c.page) {
				{{ page := ctx.Value("page").(int) }}
				<a { doors.A(ctx, c.itemHref(item.Cat, item.Id, page))... }>{ item.Name }</a>
			}
		</header>
		<kbd>
			Rating 
			@doors.Text(item.Rating)
		</kbd>
	</article>
}

func (c *categoryFragment) itemHref(catId string, itemId int, page int) doors.Attr {
	var p *int
	if page != 0 {
		p = &page
	}
	return doors.AHref{
		Model: Path{
			IsItem: true,
			CatId:  catId,
			ItemId: itemId,
			Page:   p,
		},
	}
}

templ (c *categoryFragment) notFound() {
	<h1>Category Not Found</h1>
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
			Indicator:    doors.IndicatorClass("contrast"),
			QueryMatcher: doors.QueryMatcherIgnore(),
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
                align-items: start;
                flex-grow: 1;
            }
            .cat .grid {
                width: 100%;
            }
        </style>
	}
}

```

That works, and seems ok. But we can do much better.

1. Move `listItems` to a separate fragment
2. Load more automatically on query param change
3. Add some safety checks (we are loading pages in a cycle, don't want to do it quintillion times)

`./catalog/items.templ`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)


type pageState struct {
	// current page
	current int
	// max page value
	max int
}

type itemsFragment struct {
	// current category
	cat   driver.Cat
	state doors.Beam[pageState]
	// loadNode
	node doors.Node
}

func items(cat driver.Cat, path doors.SourceBeam[Path]) templ.Component {
	// derive page  number
	page := doors.NewBeam(path, func(p Path) int {
		if p.Page == nil || *p.Page < 0 {
			return 0
		}
		return *p.Page
	})
	// derive page state
	state := doors.NewBeam(page, func(page int) pageState {
		return pageState{
			current: page,
			// if there are no pages, max will be -1
			max: driver.Items.CountPages(cat.Id) - 1,
		}
	})
	return doors.F(&itemsFragment{
		cat:   cat,
		state: state,
	})
}

templ (f *itemsFragment) Render() {
	// run at render time
	@doors.Run(func(ctx context.Context) {
		// previos page state, to calculate how many pages to load
		var loadedPages = -1
		// subscribe to page state changes, to determain range of pages to load
		f.state.Sub(ctx, func(ctx context.Context, s pageState) bool {
      // no entries and no page loaded
      if s.max == -1 && loadedPages == -1 {
          f.node.Replace(ctx, f.empty())
          return false
      }
      // limit to max pages 
		  end := min(s.max, s.current)
      // means already loaded
		  if loadedPages >= end {
		  		return false
			}
			start := loadedPages + 1
			// replace with pages from start to end
			f.node.Replace(ctx, f.pages(start, end))
			loadedPages = end
			// not done, keep sub
			return false
		})
	})
	// render our node (it's has content beacuse first beam.Sub func call is blocking)
	@f.node
	// depend on state
	@doors.Inject(0, f.state) {
		{{ s := ctx.Value(0).(pageState) }}
		// if there are pages to show, render link
		if s.current < s.max {
				<a { doors.A(ctx, f.loadHref(s.current + 1))... } role="button" class="contrast">Load More</a>
		}
	}
}

func (f *itemsFragment) loadHref(page int) doors.Attr {
	return 
}

templ (f *itemsFragment) pages(start int, end int) {
  // render all pages from start to end
	for page := start; page <= end; page++ {
		{{ items := driver.Items.List(f.cat.Id, page) }}
		if len(items) > 0 {
			<div class="grid">
				<div>
					for i, item := range items {
						if i % 2 == 0 {
							@f.item(item)
						}
					}
				</div>
				<div>
					for i, item := range items {
						if i % 2 == 1 {
							@f.item(item)
						}
					}
				</div>
			</div>
		}
	}
	// clear node to prevent infinite render loop
	{{ f.node.Clear(ctx) }}
	@f.node
}


templ (f *itemsFragment) empty() {
	@f.node {
		No Items
	}
}

templ (f *itemsFragment) item(item driver.Item) {
	<article>
		<header>
			@doors.Inject(0, f.state) {
				{{ p := ctx.Value(0).(pageState) }}
				<a { doors.A(ctx, f.itemHref(item.Cat, item.Id, p.current))... }>{ item.Name }</a>
			}
		</header>
		<kbd>
			Rating 
			@doors.Text(item.Rating)
		</kbd>
	</article>
}

func (f *itemsFragment) itemHref(catId string, itemId int, page int) doors.Attr {
	var p *int
	if page != 0 {
		p = &page
	}
	return doors.AHref{
		Model: Path{
			IsItem: true,
			CatId:  catId,
			ItemId: itemId,
			Page:   p,
		},
	}
}
```

`./catalog/cat.templ`

```templ
	/* ... */
type categoryFragment struct {
	authorized bool
	path       doors.SourceBeam[Path]
	itemsNode  doors.Node
	// extra fields removed
}

/* ... */

templ (c *categoryFragment) cat(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
	if c.authorized {
		<p>
			@createItem(cat, c.itemsNode.Reload)
		</p>
	}
	@c.itemsNode {
	  // render items component
		@items(cat, c.path)
	}
}
/*
remove methods: listItems, itemsPage, loadMore, item, itemHref
*/

```

> There is still room for optimization, but it is sufficient to grasp the core principles.
