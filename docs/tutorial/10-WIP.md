# [WIP]

##  List Items 





## 1. Catalog Page Authorization

### Add `authorized` field to category fragment

It will get useful later

`./catalog/cat.teml`

```go
func newCategoryFragment(b doors.Beam[Path], authorized bool) *categoryFragment {
	return &categoryFragment{
    // set authorized 
		authorized: authorized,
		catId: doors.NewBeam(b, func(p Path) string {
			return p.CatId
		}),
	}
}

type categoryFragment struct {
  // new field
	authorized  bool
	catId       doors.Beam[string]
}


```

### Add a session field to the catalog page

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

```go
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

## 2. Item Card Fragment
