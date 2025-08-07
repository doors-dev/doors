# Item Creation

##  1. List Items 

Add item listing to category fragment.

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) listItems(cat driver.Cat) {
	// show the first page, we will deal with pagination later
	{{ items := driver.Items.List(cat.Id, 0) }}
	if len(items) == 0 {
		No Items 
	} else {
		<table>
			<thead>
				<tr>
					<th scope="col">Name</th>
					<th scope="col">Rating</th>
				</tr>
			</thead>
			<tbody>
				for _, item := range items {
					<tr>
						<td><a { doors.A(ctx, c.itemHref(cat.Id, item.Id))... }>{ item.Name }</a></td>

						<td>
							// helper component to display any as text (default formatting)
							@doors.Text(item.Rating)
						</td>
					</tr>
				}
			</tbody>
		</table>
	}
}

func (c *categoryFragment) itemHref(catId string, itemId int) doors.Attr {
	return doors.AHref{
		Model: Path{
			IsItem: true,
			CatId:  catId,
			ItemId: itemId,
		},
	}
}

```

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) cat(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
	// add listing
	@c.listItems(cat)
}

```

> It will be empty for now

## 2. Create Item Fragment

### Form Show / Hide

`./catalog/create_item.templ`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

// we can return templ.Component directly from the constructor for cleaner code
func createItem(cat driver.Cat) templ.Component {
    // wrap in fragment render
	return doors.F(&createItemFragment{
		cat:    cat
	})
}

type createItemFragment struct {
    // current category
	cat    driver.Cat
	// node to show pop up with form
	node   doors.Node
}

// main render function
templ (c *createItemFragment) Render() {
	@c.node
	@c.button()
}
templ (c *createItemFragment) button() {
  // button with openForm handler
	<button class="secondary" { doors.A(ctx, c.openForm())... }>
		Create
	</button>
}

func (c *createItemFragment) openForm() doors.Attr {
	return doors.AClick{
		Scope: doors.ScopeBlocking(),
		On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
		  // display form
			c.node.Update(ctx, c.form())
			return false
		},
	}
}

// form component
templ (c *createItemFragment) form() {
	<dialog open>
		<article>
			<header>
			  // close form 
				<button aria-label="Close" rel="prev" { doors.A(ctx, c.closeForm())... }></button>

				<p>
					<strong>Add New Item To <strong>{ c.cat.Name }</strong></strong>
				</p>
			</header>
			<form { doors.A(ctx, c.submit())... }>
				<fieldset>
					<label>
						Name
						<input
							name="name"
						/>
					</label>
					<label>
						Description
						<textarea
							name="desc"
						></textarea>
					</label>
				</fieldset>
				<button id="item-create" class="accent" role="submit">Create</button>
			</form>
		</article>
	</dialog>
}


func (c *createItemFragment) closeForm() doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
		    // clear the nodes content
			c.node.Clear(ctx)
			return true
		},
	}
}

```

`./catalog/cat.templ`

```templ
templ (c *categoryFragment) content(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
	// render our component
	<p>
		@createItem(cat)
	</p>
	@c.listItems(cat)
}

```

> Now, you should see the Create button inside any category (for example, https://localhost:8443/catalog/electronics/). It opens a dialog with a functional close button and a form that does nothing (yet).

## 3. Catalog Page Authorization

### Add `authorized` field to the category fragment

`./catalog/cat.teml`

```templ
// second argument to constructor
func newCategoryFragment(path doors.Beam[Path], authorized bool) *categoryFragment {
	return &categoryFragment{
		// set authorized
		authorized: authorized,
		path:       path,
	}
}

type categoryFragment struct {
	// new field
	authorized bool
	path       doors.SourceBeam[Path]
}

templ (c *categoryFragment) cat(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
	// add "authorization" to the create action
	if c.authorized {
		<p>
			@createItem(cat)
		</p>
	}
	@c.listItems(cat)
}


```

### Add a session field to the catalog page.

`./catalog/page.templ`

```go

type catalogPage struct {
	beam    doors.SourceBeam[Path]
  // new prop
	session *driver.Session
}


func (c *catalogPage) Body() templ.Component {
	b := doors.NewBeam(c.beam, func(p Path) bool {
		return p.IsMain
	})
	return doors.Sub(b, func(isMain bool) templ.Component {
		if isMain {
			return main()
		}
    // add authorized arg
		return doors.F(newCategoryFragment(c.beam, c.session != nil))
	})
}

```

### Update Page Handler

`./catalog/handler.go`

```templ
package catalog

import (
	"github.com/derstruct/doors-starter/common"
	"github.com/doors-dev/doors"
)

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&catalogPage{
    // the same as on home page
		session: common.GetSession(r),
	})
}

```

> The form now opens only if the user is authorized.

## 4. Handle Form Submission

### Add form submission handler

`./catalog/create_item.templ`

```templ
templ (c *createItemFragment) form() {
	<dialog open>
		<article>
			<header>
				<button aria-label="Close" rel="prev" { doors.A(ctx, c.closeForm())... }></button>
				<p>
					<strong>Add New Item To <strong>{ c.cat.Name }</strong></strong>
				</p>
			</header>
			// added submission handler
			<form { doors.A(ctx, c.submit())... }>
				<fieldset>
					<label>
						Name
						<input
							name="name"
							required="true"
						/>
					</label>
					<label>
						Description
						<textarea
							name="desc"
							required="true"
						></textarea>
					</label>
				</fieldset>
				<button id="item-create" class="accent" role="submit">Create</button>
			</form>
		</article>
	</dialog>
}

// form data for decoding
type itemFormData struct {
	Name string `form:"name"`
	Desc string `form:"desc"`
}

func (c *createItemFragment) submit() doors.Attr {
	return doors.ASubmit[itemFormData]{
	    // indicate busy
		Indicator: doors.IndicatorAttrQuery("#item-create", "aria-busy", "true"),
		On: func(ctx context.Context, r doors.RForm[itemFormData]) bool {
			item := driver.Item{
				Name:   r.Data().Name,
				Desc:   r.Data().Desc,
				Cat:    c.cat.Id,
				Rating: 0,
			}
			// create item entry
			driver.Items.Create(item)
			//close form
			c.node.Clear(ctx)
			// remove hook
			return true
		},
	}
}


```

> New entries appear only after reloading the page.

### Add partial page reload

#### `reload` function dependency in the create item fragment

```templ

// reload function as dependency
func createItem(cat driver.Cat, reload func(ctx context.Context)) templ.Component {
	return doors.F(&createItemFragment{
		cat:    cat,
		// save to field
		reload: reload,
	})
}

type createItemFragment struct {
	cat    driver.Cat
	node   doors.Node
	// new field
	reload func(context.Context)
}

func (c *createItemFragment) submit() doors.Attr {
	return doors.ASubmit[itemFormData]{
		Indicator: doors.IndicatorAttrQuery("#item-create", "aria-busy", "true"),
		On: func(ctx context.Context, r doors.RForm[itemFormData]) bool {
			/* ...  */
			// call reload
			c.reload(ctx)
			c.node.Clear(ctx)
			return true
		},
	}
}


```

#### Reload with Node

`./catalog/cat.templ`

```templ
type categoryFragment struct {
	authorized bool
	path       doors.SourceBeam[Path]
	// add new field 
	itemsNode  doors.Node
}

templ (c *categoryFragment) cat(cat driver.Cat) {
	<hgroup>
		<h1>{ cat.Name }</h1>
		<p>{ cat.Desc } </p>
	</hgroup>
	if c.authorized {
		<p>
		  // pass node reload function 
			@createItem(cat, c.itemsNode.Reload)
		</p>
	}
	// render list inside node
	@c.itemsNode {
		@c.listItems(cat)
	}
}

```

## Conclusion

`itemCreateFragment` is an example of how a component can operate without

### Final Files

`./catalog/create_item.teml`

```templ
package catalog

import (
	"context"
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

func createItem(cat driver.Cat, reload func(ctx context.Context)) templ.Component {
	return doors.F(&createItemFragment{
		cat:    cat,
		reload: reload,
	})
}

type createItemFragment struct {
	cat    driver.Cat
	node   doors.Node
	reload func(context.Context)
}

templ (c *createItemFragment) Render() {
	@c.node
	@c.button()
}

templ (c *createItemFragment) button() {
	<button class="secondary" { doors.A(ctx, c.openForm())... }>
		Create
	</button>
}

func (c *createItemFragment) openForm() doors.Attr {
	return doors.AClick{
        Scope: doors.ScopeBlocking(),
		On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
			c.node.Update(ctx, c.form())
			return false
		},
	}
}


templ (c *createItemFragment) form() {
	<dialog open>
		<article>
			<header>
				<button aria-label="Close" rel="prev" { doors.A(ctx, c.closeForm())... }></button>
				<p>
					<strong>Add New Item To <strong>{ c.cat.Name }</strong></strong>
				</p>
			</header>
			<form { doors.A(ctx, c.submit())... }>
				<fieldset>
					<label>
						Name
						<input
							name="name"
							required="true"
						/>
					</label>
					<label>
						Description
						<textarea
							name="desc"
							required="true"
						></textarea>
					</label>
				</fieldset>
				<button id="item-create" class="accent" role="submit">Create</button>
			</form>
		</article>
	</dialog>
}

func (c *createItemFragment) closeForm() doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
			c.node.Clear(ctx)
			return true
		},
	}
}

type itemFormData struct {
	Name string `form:"name"`
	Desc string `form:"desc"`
}

func (c *createItemFragment) submit() doors.Attr {
	return doors.ASubmit[itemFormData]{
		Indicator: doors.IndicatorAttrQuery("#item-create", "aria-busy", "true"),
		On: func(ctx context.Context, r doors.RForm[itemFormData]) bool {
			item := driver.Item{
				Name:   r.Data().Name,
				Desc:   r.Data().Desc,
				Cat:    c.cat.Id,
				Rating: 0,
			}
			driver.Items.Create(item)
			c.reload(ctx)
			c.node.Clear(ctx)
			return true
		},
	}
}

```

`./catalog/cat.templ`

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
	}
}

type categoryFragment struct {
	authorized bool
	path       doors.SourceBeam[Path]
	itemsNode  doors.Node
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
	if c.authorized {
		<p>
			@createItem(cat, c.itemsNode.Reload)
		</p>
	}
	@c.itemsNode {
		@c.listItems(cat)
	}
}

templ (c *categoryFragment) listItems(cat driver.Cat) {
	{{ items := driver.Items.List(cat.Id, 0) }}
	if len(items) == 0 {
		No Items 
	} else {
		<table>
			<thead>
				<tr>
					<th scope="col">Name</th>
					<th scope="col">Rating</th>
				</tr>
			</thead>
			<tbody>
				for _, item := range items {
					<tr>
						<td><a { doors.A(ctx, c.itemHref(cat.Id, item.Id))... }>{ item.Name }</a></td>

						<td>
							@doors.Text(item.Rating)
						</td>
					</tr>
				}
			</tbody>
		</table>
	}
}

func (c *categoryFragment) itemHref(catId string, itemId int) doors.Attr {
	return doors.AHref{
		Model: Path{
			IsItem: true,
			CatId:  catId,
			ItemId: itemId,
		},
	}
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
                align-items: start;
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

`./catalog/page.templ`

```templ
package catalog

import (
	"github.com/derstruct/doors-starter/common"
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

type Path = common.CatalogPath

type catalogPage struct {
	beam    doors.SourceBeam[Path]
	session *driver.Session
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
		return doors.F(newCategoryFragment(c.beam, c.session != nil))
	})
}

```

