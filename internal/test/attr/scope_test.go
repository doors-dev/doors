package attr

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
)

func TestScopeBlocking(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestReportId(t, page, 0, "0")
	test.TestReportId(t, page, 1, "0")
	test.ClickNow(t, page, "#b1")
	test.ClickNow(t, page, "#b2")
	test.ClickNow(t, page, "#b3")
	<-time.After(400 * time.Millisecond)
	test.TestReportId(t, page, 0, "1")
	if got := test.GetReportContent(t, page, 1); got == "0" {
		t.Fatal("blocking scope should allow exactly one of the overlapping actions to finish")
	}
	test.ClickNow(t, page, "#b3")
	<-time.After(400 * time.Millisecond)
	test.TestReportId(t, page, 0, "2")
	test.TestReportId(t, page, 1, "3")
}

func TestScopeSerial(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.ClickNow(t, page, "#s1")
	test.ClickNow(t, page, "#s2")
	test.ClickNow(t, page, "#s3")
	<-time.After(310 * time.Millisecond)
	test.TestReport(t, page, "1")
	test.TestReportId(t, page, 1, "1")
	<-time.After(310 * time.Millisecond)
	test.TestReport(t, page, "2")
	test.TestReportId(t, page, 1, "2")
	<-time.After(310 * time.Millisecond)
	test.TestReport(t, page, "3")
	test.TestReportId(t, page, 1, "3")
}

func TestScopeDebounce(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#d1")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#d2")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#d1")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#d3")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#d3")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#d2")
	<-time.After(330 * time.Millisecond)

	test.TestReport(t, page, "1")
	test.TestReportId(t, page, 1, "2")

}

func TestScopeDebounceLimit(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#dl1")
	<-time.After(200 * time.Millisecond)
	test.ClickNow(t, page, "#dl2")
	<-time.After(200 * time.Millisecond)
	test.ClickNow(t, page, "#dl1")
	<-time.After(100 * time.Millisecond)
	test.ClickNow(t, page, "#dl3")
	<-time.After(200 * time.Millisecond)
	test.ClickNow(t, page, "#dl3")
	<-time.After(150 * time.Millisecond)
	test.TestReport(t, page, "1")
	test.TestReportId(t, page, 1, "3")

	test.ClickNow(t, page, "#dl2")
	<-time.After(100 * time.Millisecond)
	test.ClickNow(t, page, "#dl1")
	<-time.After(100 * time.Millisecond)
	test.ClickNow(t, page, "#dl2")
	<-time.After(350 * time.Millisecond)

	test.TestReport(t, page, "2")
	test.TestReportId(t, page, 1, "2")

}

func TestScopeFrame(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#f1")
	test.ClickNow(t, page, "#f1")
	test.ClickNow(t, page, "#f2")
	test.ClickNow(t, page, "#f3")
	test.ClickNow(t, page, "#f1")
	test.ClickNow(t, page, "#f4")
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "1")
	test.TestReportId(t, page, 1, "2")
	<-time.After(300 * time.Millisecond)
	test.TestReport(t, page, "2")
	test.TestReportId(t, page, 1, "1")
	<-time.After(250 * time.Millisecond)
	test.TestReport(t, page, "3")
	test.TestReportId(t, page, 1, "1")
	<-time.After(400 * time.Millisecond)
	test.TestReport(t, page, "4")
	test.TestReportId(t, page, 1, "3")
	//
	test.ClickNow(t, page, "#f4")
	<-time.After(50 * time.Millisecond)
	test.TestReport(t, page, "5")
	test.TestReportId(t, page, 1, "4")
	test.ClickNow(t, page, "#f2")
	<-time.After(50 * time.Millisecond)
	test.TestReport(t, page, "6")
	test.TestReportId(t, page, 1, "2")
}

func TestScopePipe(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#p1")
	test.ClickNow(t, page, "#p2")
	test.ClickNow(t, page, "#p4")
	test.ClickNow(t, page, "#p3")
	test.ClickNow(t, page, "#p5")

	// noise
	test.ClickNow(t, page, "#p4")
	test.ClickNow(t, page, "#p2")
	test.ClickNow(t, page, "#p3")

	<-time.After(450 * time.Millisecond)

	//noise
	test.ClickNow(t, page, "#p4")
	test.ClickNow(t, page, "#p2")
	test.ClickNow(t, page, "#p3")

	test.TestReport(t, page, "1")

	<-time.After(300 * time.Millisecond)
	test.TestReport(t, page, "2")
	test.TestReportId(t, page, 1, "3")

	<-time.After(300 * time.Millisecond)
	test.TestReport(t, page, "3")
	test.TestReportId(t, page, 1, "5")
}

func TestScopeConcurrent(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &scopeFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.ClickNow(t, page, "#c1")
	test.ClickNow(t, page, "#c2")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#c3")
	<-time.After(400 * time.Millisecond)

	test.TestReportId(t, page, 0, "2")
	if got := test.GetReportContent(t, page, 1); got == "3" {
		t.Fatal("concurrent scope should block the different group while group 1 is active")
	}

	test.ClickNow(t, page, "#c3")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "3")
	test.TestReportId(t, page, 1, "3")
}
