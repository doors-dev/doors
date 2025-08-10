# Title And Meta

Let's add a dynamic title and metadata to our catalog page

## 1. Dynamic Title With Head Component

./catalog/page.templ`

```templ
// converted templ to func equivalent
func (c *catalogPage) Head() templ.Component {
  // head component, takes path beam and function to derive HeadData
	return doors.Head(c.beam, func(p Path) doors.HeadData {
	  // no category selected
		if p.IsMain {
			return doors.HeadData{
				Title: "Catalog",
			}
		}
		cat, ok := driver.Cats.Get(p.CatId)
		// cannout find category
		if !ok {
			return doors.HeadData{
				Title: "Category Not Found",
			}
		}
		// category page
		if p.IsCat {
			return doors.HeadData{
				Title: cat.Name,
			}
		}
		item, ok := driver.Items.Get(p.ItemId)
		// cannot find item
		if !ok {
			return doors.HeadData{
				Title: "Item Not Found",
			}
		}
		return doors.HeadData{
			Title: item.Name,
		}
	})
}

```

> That works and totaly ok. But it's not optimal, because it updates every path change, and not all path changes lead to title update (page query param for example). 

## 2. Optimize Update Straregy

