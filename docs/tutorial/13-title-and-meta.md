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
type metaState struct {
	Cat  string
	Item int
}

func (c *catalogPage) Head() templ.Component {
	// derive new beam with metaState
	state := doors.NewBeam(c.path, func(p Path) metaState {
		if p.IsMain {
			return metaState{
				Cat:  "",
				Item: -1,
			}
		}
		if p.IsCat {
			return metaState{
				Cat:  p.CatId,
				Item: -1,
			}
		}
		return metaState{
			Cat:  p.CatId,
			Item: p.ItemId,
		}
	})
	// head component, takes path beam and function to derive HeadData
	return doors.Head(state, func(m metaState) doors.HeadData {
		// no category selected
		if m.Cat == "" {
			return doors.HeadData{
				Title: "Catalog",
			}
		}
		cat, ok := driver.Cats.Get(m.Cat)
		// cat not found
		if !ok {
			return doors.HeadData{
				Title: "Category Not Found",
			}
		}
		// category page
		if m.Item == -1 {
			return doors.HeadData{
				Title: cat.Name,
			}
		}
		item, ok := driver.Items.Get(m.Item)
		// item not found
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
  if m.Item == -1 {
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

