package components

import (
	"testing"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/test"
)

func TestTitle(t *testing.T) {
	p := common.RandId()
	bro := test.NewBro(browser,
		doors.ServePage(func(pr doors.PageRouter[test.Path], r doors.RPage[test.Path]) doors.PageRoute {
			return pr.Page(&test.Page{
				H: head,
				F: &LinksFragment{
					Param: p,
				},
			})
		}),
	)
	page := bro.Page(t, "/")
	test.TestContent(t, page, "title", "home")
	test.Click(t, page, "#param")
	test.TestContent(t, page, "title", p)
	test.Click(t, page, "#string")
	test.TestContent(t, page, "title", "s")
}
