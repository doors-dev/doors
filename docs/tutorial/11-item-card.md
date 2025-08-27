# Item Card

By now, we can view the table of items and create new items. However, the item card view has not yet been implemented, and no edit functionality is available.

## 1. Card Fragment

`./catalog/card.templ`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-tutorial/driver"
	"github.com/doors-dev/doors"
)

// path to show card when user clicks the link
// reload to update items list after card is closed
func itemCard(path doors.SourceBeam[Path], reload func(context.Context)) templ.Component {
	return doors.F(&cardFragment{
		path:   path,
		reload: reload,
	})
}

type cardFragment struct {
	path    doors.SourceBeam[Path]
	reload  func(context.Context)
	content doors.Door
}

templ (c *cardFragment) Render() {
	// Run - helper to execute a function during render
	@doors.Run(func(ctx context.Context) {
		// derive beam with itemId
		itemId := doors.NewBeam(c.path, func(p Path) int {
			// item pattern match check
			if !p.IsItem {
				return -1
			}
			return p.ItemId
		})
		// this time we subscribe directly (not by using @doors.Sub(..))
		// because we need direct access to the door to also update to/after edit
		itemId.Sub(ctx, func(ctx context.Context, id int) bool {
			// item not delected - the door is empty
			if id == -1 {
				c.content.Clear(ctx)
				return false
			}
			// load card into the door
			c.content.Update(ctx, c.card(id))
			return false
		})
	})
	@c.content
}

templ (c *cardFragment) card(id int) {
	{{ item, ok := driver.Items.Get(id) }}
	<dialog open>
		<article>
			<header>
				@doors.AClick{
					On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
						// reload category items table
						c.reload(ctx)
						// manually switch path to Cat from Item
						c.path.Mutate(ctx, func(p Path) Path {
							p.IsCat = true
							p.IsItem = false
							return p
						})
						// remove hook
						return true
					},
				}
				<button aria-label="Close" rel="prev"></button>
				<p>
					if ok {
						<strong>{ item.Name }</strong>
					} else {
						<strong>Item Not Found</strong>
					}
				</p>
			</header>
			if ok {
				<p>
					{ item.Desc }
				</p>
				<kbd>
					Rating: 
					@doors.Text(item.Rating)
				</kbd>
			}
		</article>
	</dialog>
}
```

> Notice how `SourceBeam[Path]` is synchronized with the path in the browser when we close the pop-up. 

`./catalog/cat.templ`

```templ
templ (f *categoryFragment) Render() {
	@f.style()
	// render card fragment
	@itemCard(f.path, f.items.Reload)
	<div class="category">
		<aside>
			@f.nav()
		</aside>
		<main>
			@doors.Sub(doors.NewBeam(f.path, func(p Path) string {
				return p.CatId
			}), func(catId string) templ.Component {
				return f.content(catId)
			})
		</main>
	</div>
}
```



## 2. Edit

At this point, you have all the concepts to do it yourself. 

Steps:

1. If authorized (common.IsAuthorized(ctx)), show the edit button inside the card pop-up, which switches the **door** content with the edit form component
2. Add submit handler attribute — update item via driver and load the card into **door**.



---

Next: [Pagination](./12-pagination.md)

