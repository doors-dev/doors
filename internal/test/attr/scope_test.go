package attr

import (
	"strconv"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

type reportSnapshot struct {
	count  int
	marker string
}

func reportSnapshotNow(t *testing.T, page *rod.Page) reportSnapshot {
	t.Helper()
	value := page.MustEval(`() => {
		const count = document.querySelector("#report-0")?.textContent ?? ""
		const marker = document.querySelector("#report-1")?.textContent ?? ""
		return [count, marker]
	}`)
	items := value.Arr()
	if len(items) != 2 {
		t.Fatalf("expected two report values, got %v", value)
	}
	count, err := strconv.Atoi(items[0].Str())
	if err != nil {
		t.Fatalf("report 0 expected integer, got %q", items[0].Str())
	}
	return reportSnapshot{
		count:  count,
		marker: items[1].Str(),
	}
}

func waitSnapshot(t *testing.T, page *rod.Page, timeout time.Duration, cond func(reportSnapshot) bool, message string) reportSnapshot {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		snapshot := reportSnapshotNow(t, page)
		if cond(snapshot) {
			return snapshot
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s before timeout, got count=%d marker=%q", message, snapshot.count, snapshot.marker)
		}
		<-time.After(25 * time.Millisecond)
	}
}

func waitSnapshotExact(t *testing.T, page *rod.Page, timeout time.Duration, count int, marker string, message string) {
	t.Helper()
	waitSnapshot(t, page, timeout, func(snapshot reportSnapshot) bool {
		return snapshot.count == count && snapshot.marker == marker
	}, message)
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
	test.ClickBurst(t, page, "#f1", "#f1", "#f2", "#f3", "#f1", "#f4")
	waitSnapshotExact(t, page, 400*time.Millisecond, 1, "2", "frame scope first completion")
	waitSnapshotExact(t, page, 700*time.Millisecond, 3, "1", "frame scope queued completions")
	waitSnapshotExact(t, page, 700*time.Millisecond, 4, "3", "frame scope barrier completion")
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
	test.ClickBurst(t, page, "#p1", "#p2", "#p4", "#p3", "#p5")

	// noise
	test.ClickBurst(t, page, "#p4", "#p2", "#p3")

	<-time.After(450 * time.Millisecond)
	snapshot := reportSnapshotNow(t, page)
	if snapshot.count != 0 || snapshot.marker != "0" {
		t.Fatalf("pipe scope should not complete anything before debounced serial work matures, got count=%d marker=%q", snapshot.count, snapshot.marker)
	}

	//noise
	test.ClickBurst(t, page, "#p4", "#p2", "#p3")
	waitSnapshotExact(t, page, 700*time.Millisecond, 1, "2", "pipe scope first serial completion")
	waitSnapshotExact(t, page, 700*time.Millisecond, 2, "3", "pipe scope second serial completion")
	waitSnapshotExact(t, page, 700*time.Millisecond, 3, "5", "pipe scope barrier completion")
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
