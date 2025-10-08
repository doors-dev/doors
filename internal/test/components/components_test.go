package components

import (
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/test"
)

func TestTitle(t *testing.T) {
	p := common.RandId()
	bro := test.NewBro(browser,
		doors.UsePage(func(pr doors.PageRouter[test.Path], r doors.RModel[test.Path]) doors.ModelRoute {
			return pr.Page(&test.Page{
				H: head,
				F: &LinksFragment{
					Param: p,
				},
			})
		}),
	)
	page := bro.Page(t, "/")

	// Test initial state (home) - has description, keywords, og:title
	test.TestContent(t, page, "title", "home")
	test.TestAttr(t, page, `meta[name="description"]`, "content", "Welcome to the home page")
	test.TestAttr(t, page, `meta[name="keywords"]`, "content", "home, main, index")
	test.TestAttr(t, page, `meta[name="og:title"]`, "content", "Home Page")

	// Verify home-specific tags don't exist yet
	test.TestMustNot(t, page, `meta[name="author"]`)
	test.TestMustNot(t, page, `meta[name="category"]`)

	// Test param state - has description, keywords, og:title, author (adds author, removes nothing)
	test.Click(t, page, "#param")
	test.TestContent(t, page, "title", p)
	test.TestAttr(t, page, `meta[name="description"]`, "content", "Page for parameter: "+p)
	test.TestAttr(t, page, `meta[name="keywords"]`, "content", "param, "+p)
	test.TestAttr(t, page, `meta[name="og:title"]`, "content", "Param: "+p)
	test.TestAttr(t, page, `meta[name="author"]`, "content", "Parameter Author")
	// Verify category still doesn't exist
	test.TestMustNot(t, page, `meta[name="category"]`)

	// Test string state - has description, category (removes keywords, og:title, author; adds category)
	test.Click(t, page, "#string")
	test.TestContent(t, page, "title", "s")
	test.TestAttr(t, page, `meta[name="description"]`, "content", "String page description")
	test.TestAttr(t, page, `meta[name="category"]`, "content", "text-content")
	// Verify removed tags are gone
	test.TestMustNot(t, page, `meta[name="keywords"]`)
	test.TestMustNot(t, page, `meta[name="og:title"]`)
	test.TestMustNot(t, page, `meta[name="author"]`)

	// Go back to home state - verify proper cleanup and restoration
	test.Click(t, page, "#home")
	<-time.After(200 * time.Millisecond)
	test.TestContent(t, page, "title", "home")
	test.TestAttr(t, page, `meta[name="description"]`, "content", "Welcome to the home page")
	test.TestAttr(t, page, `meta[name="keywords"]`, "content", "home, main, index")
	test.TestAttr(t, page, `meta[name="og:title"]`, "content", "Home Page")
	// Verify string-specific tags are removed
	test.TestMustNot(t, page, `meta[name="category"]`)
	test.TestMustNot(t, page, `meta[name="author"]`)
}
