# Pagination

## 1. Pagination 

quick & dirty

`./catalog/item_list.templ`

```templ
templ (f *itemListFragment) Render() {
	<div>
		@doors.Inject("page", f.page) {
			/* ... */
		}
		// render pagination outside injection, we don't 
		// need to rerender it on page number change
		@f.pagination()
	</div>
}

templ (f *itemListFragment) pagination() {
	@doors.Style() {
		<style>
        .item-pages {
            display: flex;
            justify-content: start;
            gap: 1rem;
        }
    </style>
	}
	<div class="item-pages">
		// range page count from driver
		for i := range driver.Items.CountPages(f.cat.Id) {
			@f.pageHref(i)
			<a class="secondary">
				@doors.Text(i + 1)
			</a>
		}
	</div>
}

func (f *itemListFragment) pageHref(page int) doors.Attr {
	var p *int = nil
	// first page - no page param at all
	if page != 0 {
		p = &page
	}
	return doors.AHref{
		Active: doors.Active{
			// indicate with an attribute
			Indicator: doors.IndicatorAttr("aria-current", "true"),
			// match path start (/catalog/:CatId)
			PathMatcher: doors.PathMatcherStarts(),
			// match query param page (not necessary in this case, just for demo)
			QueryMatcher: doors.QueryMatcherSome("page"),
		},
		Model: Path{
			IsCat: true,
			CatId: f.cat.Id,
			Page:  p,
		},
	}
}
```

## 2. Load More

Many aspects can be improved in pagination, but let's use a 'load more' button instead to learn a new, useful pattern.

`./catalog/item_list.templ`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-tutorial/driver"
	"github.com/doors-dev/doors"
)

// combined page state (incude maximum page value)
type pageState struct {
	current int
	last    bool
}

func itemList(cat driver.Cat, path doors.SourceBeam[Path]) templ.Component {

	// derive page beam as before
	page := doors.NewBeam(path, func(p Path) int {
		if p.Page == nil {
			return 0
		}
		return max(*p.Page, 0)
	})

	// enrich page beam with
	state := doors.NewBeam(page, func(p int) pageState {
		maxPage := driver.Items.CountPages(cat.Id) - 1 // if max = -1, means no pages to show
		current := min(p, maxPage)
		return pageState{
			current: current,
			last:    maxPage == current,
		}
	})

	return doors.F(&itemListFragment{
		cat:  cat,
		path: path,
		page: state,
	})
}

type itemListFragment struct {
	cat  driver.Cat
	path doors.SourceBeam[Path]
	// change type of page beam to pageState
	page doors.Beam[pageState]
	// placeholder we will use to insert new pages
	placeHolder doors.Door
}

templ (f *itemListFragment) Render() {
	// helper component to run provided function at runtime
	@doors.Run(func(ctx context.Context) {
		// state to track last loaded page
		loadedPage := -1
		f.page.Sub(ctx, func(ctx context.Context, ps pageState) bool {
			// no pages available and no pages loaded
			if ps.current == -1 && loadedPage == -1 {
				f.placeHolder.Update(ctx, doors.Text("No Items"))
				return false
			}
			// already loaded
			if loadedPage >= ps.current {
				return false
			}
			// load from loaded page up to current
			from := loadedPage + 1
			loadedPage = ps.current
			f.placeHolder.Replace(ctx, f.pages(from, loadedPage))
			return false
		})
	})
	// placeholder for the loaded pages
	@f.placeHolder
	// subsribe to pageState change
	@doors.Sub(f.page, func(ps pageState) templ.Component {
		// nothing to load, it's the last page
		if ps.last {
			return nil
		}
		return f.loadMore(ps.current + 1)
	})
}

templ (f *itemListFragment) loadMore(next int) {
	@doors.AHref{
		Model: Path{
			IsCat: true,
			CatId: f.cat.Id,
			Page:  &next,
		},
	}
	<a role="button" class="contrast">Load More</a>
}

templ (f *itemListFragment) pages(start int, end int) {
	// render all pages from start to end
	for page := start; page <= end; page++ {
		{{ items := driver.Items.List(f.cat.Id, page) }}
		<div class="grid">
			<div>
				for i, item := range items {
					if i % 2 == 0 {
						@f.item(item, page)
					}
				}
			</div>
			<div>
				for i, item := range items {
					if i % 2 == 1 {
						@f.item(item, page)
					}
				}
			</div>
		</div>
	}
	// clear the placeholder to prevent infinite render loop
	{{ f.placeHolder.Clear(ctx) }}
	// render placeholder for the next portion of pages
	@f.placeHolder
}

// no change here
templ (f *itemListFragment) item(item driver.Item, page int) {
	<article>
		<header>
			@doors.AHref{
				Model: Path{
					IsItem: true,
					CatId:  item.Cat,
					ItemId: item.Id,
					Page: func() *int {
						if page == 0 {
							return nil
						}
						return &page
					}(),
				},
			}
			<a>{ item.Name }</a>
		</header>
		<kbd>
			Rating 
			@doors.Text(item.Rating)
		</kbd>
	</article>
}
```

We learned how to use the **door** as a "cursor" for HTML insertion.

 But, there is an issue, steprs to reproduce:

1. Load More
2. Click the card from the top of the list
3. Close card
4. The content has been reloaded, and we are back on the first page again.



## 3. Dynamic links

The issue is that the item link is rendered with the page number to which this item belongs. We need item links to be updated dynamically on page change.

Fix:

`./catalog/item_list.templ`

```templ
// remove page argument, don't need it anymore
templ (f *itemListFragment) item(item driver.Item) {
	<article>
		<header>
			// inject page beam to rerender the link
			@doors.Inject(0, f.page) {
				// read the beam value
				{{ ps := ctx.Value(0).(pageState) }}
				@doors.AHref{
					Model: Path{
						IsItem: true,
						CatId:  item.Cat,
						ItemId: item.Id,
						// use the beam value in the path model
						Page: func() *int {
							if ps.current == 0 {
								return nil
							}
							return &ps.current
						}(),
					},
				}
				<a>{ item.Name }</a>
			}
		</header>
		<kbd>
			Rating 
			@doors.Text(item.Rating)
		</kbd>
	</article>
}
```

>  There is still room for optimization, but it is sufficient to grasp the core principles.
