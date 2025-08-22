# Item Creation

 ## 1. Create Item Fragment

Form in a pop-up

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-tutorial/driver"
	"github.com/doors-dev/doors"
)

func createItem(cat driver.Cat) templ.Component {
	// wrap in a fragment render
	return doors.F(&createItemFragment{
		cat: cat,
	})
}

type createItemFragment struct {
	// current category
	cat driver.Cat
	// door to show pop up with form
	door doors.Door
}

// main render function
templ (f *createItemFragment) Render() {
	// door for the pop-up form, empty by default
	@f.door
	// render form in the door on click
	@doors.AClick{
		Scope: doors.ScopeBlocking(),
		On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
			// display form
			f.door.Update(ctx, f.form())
			return false
		},
	}
	<button class="contrast">
		Add Item
	</button>
}

// form component
templ (f *createItemFragment) form() {
	<dialog open>
		<article>
			<header>
				@doors.AClick{
					On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
						// clear the form door (hide form)
						f.door.Clear(ctx)
						return true
					},
				}
				<button aria-label="Close" rel="prev"></button>
				<p>
					<strong>Add New Item To <strong>{ f.cat.Name }</strong></strong>
				</p>
			</header>
			<form>
				<fieldset>
					<label>
						Name
						<input name="name"/>
					</label>
					<label>
						Description
						<textarea name="desc"></textarea>
					</label>
				</fieldset>
				// button with id, will use for indication later
				<button id="item-create" role="submit">Add</button>
			</form>
		</article>
	</dialog>
}
```

Render in category fragment instead of the dummy button:

`./catalog/category.templ`

```templ
/* ... */

templ (f *categoryFragment) content(catId string) {
	{{ cat, ok := driver.Cats.Get(catId) }}
	if ok {
		<hgroup>
			<h1>{ cat.Name }</h1>
			<p>{ cat.Desc } </p>
		</hgroup>
		if common.IsAuthorized(ctx) {
		  // our new fragment
			@createItem(cat)
		}
	} else {
		<div>
			<mark>Not Found</mark>
		</div>
	}
}

/* ... */
```

> Check any category page, button "Add Item" now can shows creation form (if authorized)  

## 2. Dynamic Attribute (optional)

Instead of loading the form into the door, we can utilize a dynamic attribute. 

```templ
func createItem(cat driver.Cat) templ.Component {
	return doors.F(&createItemFragment{
		cat: cat,
		// initialize dynamic attrubute, args: name, value, enabled (false means don't render)
		open: doors.NewADyn("open", "", false),
	})
}

type createItemFragment struct {
	cat driver.Cat
	// add field
	open doors.ADyn
}

// main render function
templ (f *createItemFragment) Render() {
    // render form in closed state
	@f.form()
	@doors.AClick{
		Scope: doors.ScopeBlocking(),
		On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
			// display form with dynamic attribute
			f.open.Enable(ctx, true)
			return false
		},
	}
	<button class="contrast">
		Add Item
	</button>
}

templ (f *createItemFragment) form() {
	// attach attribute
	@f.open
	// remove "open" attribute (closed state)
	<dialog>
		<article>
			<header>
				@doors.AClick{
					On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
						// hide form byd disabling attr
						f.open.Enable(ctx, false)
						// keep hook active
						return false
					},
				}
				<button aria-label="Close" rel="prev"></button>
				<p>
					<strong>Add New Item To <strong>{ f.cat.Name }</strong></strong>
				</p>
			</header>
			<form>
				<fieldset>
					<label>
						Name
						<input name="name"/>
					</label>
					<label>
						Description
						<textarea name="desc"></textarea>
					</label>
				</fieldset>
				// button with id, will use for indication later
				<button id="item-create" role="submit">Add</button>
			</form>
		</article>
	</dialog>
}

```

## 3. List Items

### Listing Fragment

Before we proceed with the form handler, let's add item listing, so we can see the result.

```templ
package catalog

import (
	"github.com/derstruct/doors-tutorial/driver"
	"github.com/doors-dev/doors"
)

func itemList(cat driver.Cat, path doors.SourceBeam[Path]) templ.Component {
	return doors.F(&itemListFragment{
		cat:  cat,
		path: path,
		// derive beam with page number
		page: doors.NewBeam(path, func(p Path) int {
			// page is *int (so it's removed from url for the first page)
			if p.Page == nil {
				return 0
			}
			return min(*p.Page, 0)
		}),
	})
}

type itemListFragment struct {
	cat  driver.Cat
	path doors.SourceBeam[Path]
	page doors.Beam[int]
}

templ (f *itemListFragment) Render() {
	// inject beam value into the context under "page" key
	// alternative to doors.Sub
	@doors.Inject("page", f.page) {
		// extract page value
		{{ page := ctx.Value("page").(int) }}
		// query items using page value
		{{ items := driver.Items.List(f.cat.Id, page) }}
		if len(items) == 0 {
			No Items 
		} else {
			// split between two columns
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
	}
}

templ (f *itemListFragment) item(item driver.Item, page int) {
	<article>
		<header>
			@doors.AHref{
				Model: Path{
					IsItem: true,
					CatId:  item.Cat,
					ItemId: item.Id,
					// keep page query param when item opens
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

### Category Fragment Update

After fragment creation, the item list must be reloaded to display the new entry. To achieve that, we will wrap the item fragment in **door**:

`./catalog/create_item.templ`

```templ

/* ... */

// add reload dependency
func createItem(cat driver.Cat, reload func(context.Context)) templ.Component {
	return doors.F(&createItemFragment{
		cat:    cat,
		open:   doors.NewADyn("open", "", false),
		reload: reload,
	})
}

type createItemFragment struct {
	cat  driver.Cat
	open doors.ADyn
	// reload func
	reload func(context.Context)
}

/* ... */
```



`./catalog/category.templ`

```templ
/* ... */
type categoryFragment struct {
	path doors.SourceBeam[Path]
	// prop with the new door
	items doors.Door
}

/* ... */

templ (f *categoryFragment) content(catId string) {
	{{ cat, ok := driver.Cats.Get(catId) }}
	if ok {
		<hgroup>
			<h1>{ cat.Name }</h1>
			<p>{ cat.Desc } </p>
		</hgroup>
		// render items door
		@f.items {
			if common.IsAuthorized(ctx) {
				<p>
				  // pass reload function to create item
					@createItem(cat, f.items.Reload)
				</p>
			}
			@itemList(cat, f.path)
		}
	} else {
		<div>
			<mark>Not Found</mark>
		</div>
	}
}

/* ... */
```



## 4. Create Form Handler

```templ
/* ... */

// form data for decoding
type itemFormData struct {
	Name string `form:"name"`
	Desc string `form:"desc"`
}

func (f *createItemFragment) submit(ctx context.Context, r doors.RForm[itemFormData]) bool {
	item := driver.Item{
		Name:   r.Data().Name,
		Desc:   r.Data().Desc,
		Cat:    f.cat.Id,
		Rating: 0,
	}
	// create item entry
	driver.Items.Create(item)
	//reload (will also close form)
	f.reload(ctx)
	// remove hook
	return true
}

templ (f *createItemFragment) form() {
	// attach attribute
	@f.open
	// remove "open" attribute (closed state)
	<dialog>
		<article>
			<header>
				/* ... */
			</header>
			// add submit handler
			@doors.ASubmit[itemFormData]{
				// indicate pending state on the button
				Indicator: doors.IndicatorAttrQuery("#item-create", "aria-busy", "true"),
				// block rapid resubmittion
				Scope: doors.ScopeBlocking(),
				On:    f.submit,
			}
			<form>
				<fieldset>
					<label>
						Name
						<input name="name"/>
					</label>
					<label>
						Description
						<textarea name="desc"></textarea>
					</label>
				</fieldset>
				<button id="item-create" role="submit">Add</button>
			</form>
		</article>
	</dialog>
}

```



