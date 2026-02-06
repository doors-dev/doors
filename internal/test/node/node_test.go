package door

import (
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"testing"
	"time"
)

func TestDoorLoadPage(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(pr doors.ModelRouter[test.Path], r doors.RModel[test.Path]) doors.ModelRoute {
			return pr.App(&test.Page{
				Header: "Page Door",
			})
		})
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	h1Text := page.MustElement("h1").MustText()
	if h1Text != "Page Door" {
		t.Fatal("header missmatch")
	}
}


func TestDoorInitialContent(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
}


func DoorUpdatedBefore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#updated")
}

func TestDoorRemovedBefore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMustNot(t, page, "#removed")
}


func TestDoorReplacedBefore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#initReplaced")

	// test.TestMust(t, page, "body > #replaced")

}


func TestDoorDynamic(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &DynamicFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#update")
	test.TestMustNot(t, page, "#init")
	test.TestMust(t, page, "#updated")
	test.Click(t, page, "#replace")
	test.TestMustNot(t, page, "#updated")
	test.TestMust(t, page, "#replaced")
	test.Click(t, page, "#remove")
	test.TestMustNot(t, page, "#replaced")
}

func TestDoorEmbedded(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &EmbeddedFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	
}

func TestDoorEmbeddedRemove(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &EmbeddedFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#clear")
	test.TestMustNot(t, page, "#init")
	test.Click(t, page, "#remove")
	test.Click(t, page, "#replace")
	test.TestMustNot(t, page, "#temp")
	test.TestMust(t, page, "#replaced")
}


func TestDoorUpdateX(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentX{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#updatex")
	test.TestReport(t, page, "ok")
	test.TestMustNot(t, page, "#init")
	test.TestMust(t, page, "#updated")
	test.Click(t, page, "#removex")
	test.TestReport(t, page, "ok")
	test.TestMustNot(t, page, "#updated")
	test.Click(t, page, "#updatex")
	test.TestReport(t, page, "channel closed")
}

func TestDoorMultiple(t *testing.T ){
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentMany{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	c := test.Count(page, ".sample")
	if c != 1 {
		println(page.MustHTML())
		t.Fatal("Counted before upated, need 1, got", c)
	}
	test.Click(t, page, "#replace")
	c = test.Count(page, ".sample")
	if c != 100 {
		t.Fatal("Counted after reaplce, need 100, got", c)
	}
}
