package test
import (
	"testing"

	"github.com/doors-dev/doors"
)

func TestNodeLoadPage(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				header: "Page Node",
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	h1Text := page.MustElement("h1").MustText()
	if h1Text != "Page Node" {
		t.Fatal("header missmatch")
	}
}

func TestNodeInitialContent(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
}

func TestNodeUpdatedBefore(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#updated")
}
func TestNodeRemovedBefore(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMustNot(t, page, "#removed")
}
func TestNodeReplacedBefore(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMustNot(t, page, "#initReplaced")
	testMust(t, page, "body > #replaced")

}
func TestNodeDynamic(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &DynamicFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
	click(t, page, "#update")
	testMustNot(t, page, "#init")
	testMust(t, page, "#updated")
	click(t, page, "#replace")
	testMustNot(t, page, "#updated")
	testMust(t, page, "#replaced")
	click(t, page, "#remove")
	testMustNot(t, page, "#replaced")
}

func TestEmbedded(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &EmbeddedFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
	click(t, page, "#clear")
	testMustNot(t, page, "#init")
	click(t, page, "#replace")
	testMustNot(t, page, "#temp")
	testMust(t, page, "#replaced")
}

func TestEmbeddedRemove(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &EmbeddedFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
	click(t, page, "#clear")
	testMustNot(t, page, "#init")
	click(t, page, "#remove")
	click(t, page, "#replace")
	testMustNot(t, page, "#temp")
	testMustNot(t, page, "#replaced")
}

func TestUpdateX(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &FragmentX{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
	click(t, page, "#updatex")
	testReport(t, page, "ok")
	testMustNot(t, page, "#init")
	testMust(t, page, "#updated")
	click(t, page, "#removex")
	testReport(t, page, "ok")
	testMustNot(t, page, "#updated")
	click(t, page, "#updatex")
	testReport(t, page, "false update")

}
