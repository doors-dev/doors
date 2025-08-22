# Title And Meta

Let's add a dynamic title and metadata to our catalog page

## 1. Dynamic Title With Head Component

`./catalog/page.templ`

```templ
// converted templ to func equivalent
func (c *catalogPage) Head() templ.Component {
	// head component, takes path beam and function to derive HeadData
	return doors.Head(c.path, func(p Path) doors.HeadData {
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

> That works and is totally okay. However, it's not optimal because it runs the derive function on every path change, but not all path changes lead to a title change (for example, the page query parameter). 

## 2. Optimize Update Strategy

```templ
// state that title depends on
type pathState struct {
	Cat  string
	Item int
}

func (c *catalogPage) Head() templ.Component {
	// derive new beam with pathState
	state := doors.NewBeam(c.path, func(p Path) pathState {
		if p.IsMain {
			return pathState{
				Cat:  "",
				Item: -1,
			}
		}
		if p.IsCat {
			return pathState{
				Cat:  p.CatId,
				Item: -1,
			}
		}
		return pathState{
			Cat:  p.CatId,
			Item: p.ItemId,
		}
	})
	// head component, takes path beam and function to derive HeadData
	return doors.Head(state, func(ps pathState) doors.HeadData {
		// no category selected
		if ps.Cat == "" {
			return doors.HeadData{
				Title: "Catalog",
			}
		}
		cat, ok := driver.Cats.Get(ps.Cat)
		// cannout find category
		if !ok {
			return doors.HeadData{
				Title: "Category Not Found",
			}
		}
		// category page
		if ps.Item == -1 {
			return doors.HeadData{
				Title: cat.Name,
			}
		}
		item, ok := driver.Items.Get(ps.Item)
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

## 3. Meta

If you want meta tags to update in response to a path change, include it in the `HeadData`

```go
  /* ... */
  if p.Item == -1 {
    return doors.HeadData{
      Title: cat.Name,
      Meta: map[string]string{
        "description": cat.Desc,
      },
    }
  }
  /* ... */
  return doors.HeadData{
    Title: item.Name,
    Meta: map[string]string{
      "description": item.Desc,
    },
  }

```

> If a meta tag was previously included in `HeadData`, but is absent in the new one, it will be removed from the page.

