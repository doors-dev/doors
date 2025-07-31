# [WIP]

## Add authorization to the catalog page

`./catalog/page.templ`

```go
type catalogPage struct {
	beam    doors.SourceBeam[Path]
	// added session field
	session *driver.Session
}
```

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

