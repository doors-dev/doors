package attr

import (
	"strconv"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func reportInt(t *testing.T, page *rod.Page, id int) int {
	t.Helper()
	value := test.GetReportContent(t, page, id)
	n, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("report %d expected integer, got %q", id, value)
	}
	return n
}

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
	count := reportInt(t, page, 0)
	if count < 1 || count > 2 {
		t.Fatalf("frame scope expected early progress between 1 and 2, got %d", count)
	}
	marker := test.GetReportContent(t, page, 1)
	if count == 1 && marker != "2" {
		t.Fatalf("frame scope first completion should be marker 2, got %q", marker)
	}
	if count == 2 && marker != "1" {
		t.Fatalf("frame scope second completion should be marker 1, got %q", marker)
	}
	<-time.After(300 * time.Millisecond)
	count = reportInt(t, page, 0)
	if count < 2 || count > 3 {
		t.Fatalf("frame scope expected mid progress between 2 and 3, got %d", count)
	}
	if marker = test.GetReportContent(t, page, 1); marker != "1" {
		t.Fatalf("frame scope should still be on marker 1 before the final queued frame completes, got %q", marker)
	}
	<-time.After(250 * time.Millisecond)
	count = reportInt(t, page, 0)
	if count < 3 || count > 4 {
		t.Fatalf("frame scope expected late progress between 3 and 4, got %d", count)
	}
	marker = test.GetReportContent(t, page, 1)
	if count == 3 && marker != "1" {
		t.Fatalf("frame scope third completion should still be marker 1, got %q", marker)
	}
	if count == 4 && marker != "3" {
		t.Fatalf("frame scope final completion should be marker 3 once it lands early, got %q", marker)
	}
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

	count := reportInt(t, page, 0)
	if count < 1 || count > 2 {
		t.Fatalf("pipe scope expected early progress between 1 and 2, got %d", count)
	}

	<-time.After(300 * time.Millisecond)
	count = reportInt(t, page, 0)
	if count < 2 || count > 3 {
		t.Fatalf("pipe scope expected mid progress between 2 and 3, got %d", count)
	}
	marker := test.GetReportContent(t, page, 1)
	if count == 2 && marker != "3" {
		t.Fatalf("pipe scope should still be on marker 3 before the final frame-only action completes, got %q", marker)
	}
	if count == 3 && marker != "5" {
		t.Fatalf("pipe scope final completion should be marker 5 once it lands early, got %q", marker)
	}

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
