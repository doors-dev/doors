# Item Card

By now, we can view the table of items and create new items. However, the item card view has not yet been implemented, and no edit functionality is available.

## 1. Card Fragment

`./catalog/card.templ`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

// path to show card when user clicks the link
// reload to update items table after card is closed
func card(path doors.SourceBeam[Path], reload func(context.Context)) templ.Component {
	return doors.F(&cardFragment{
		path: path,
    reload: reload,
	})
}

type cardFragment struct {
	path   doors.SourceBeam[Path]
	reload func(context.Context)
	door   doors.Door
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
			if id == -1 {
				c.door.Clear(ctx)
				return false
			}
			// call method to load item card into the door
			c.loadCard(ctx, id)
			return false
		})
	})
	@c.door
}

// query item and update the door with card
func (c *cardFragment) loadCard(ctx context.Context, id int) {
	item, ok := driver.Items.Get(id)
	if !ok {
		c.door.Update(ctx, c.card(nil))
		return
	}
	c.door.Update(ctx, c.card(&item))
}

templ (c *cardFragment) card(item *driver.Item) {
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
					if item != nil {
						<strong>Item { item.Name }</strong>
					} else {
						<strong>Item Not Found</strong>
					}
				</p>
			</header>
			if item != nil {
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

> Notice how `SourceBeam[Path]` is synchronized with the path in the browser. 

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) Render() {
	@c.style()
	// render card fragment
	@card(c.path, c.itemsDoor.Reload)	
  <div class="cat">
		<aside>
			@c.nav()
		</aside>
		<div class="content">
			@c.content()
		</div>
	</div>
}

```



## 2. Edit

At this point, you have all the concepts to do it yourself. 

Steps:

1. Add authorized property to the card fragment and pass it as a constructor argument.
2. If authorized, show the edit button inside the card pop-up, which switches the Door content with the edit form HTML
3. Add submit handler attribute — update item via driver and call loadCard method to show new data

