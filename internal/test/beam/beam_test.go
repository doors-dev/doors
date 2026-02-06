package beam

import (
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
)

func TestBeamBasics(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamUpdateFragment{
				r: test.NewReporter(1),
				b: doors.NewSource(state{}),
			}
		})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestReportId(t, page, 0, "0")
	test.Click(t, page, "#update")
	test.TestReportId(t, page, 0, "1")
	test.Click(t, page, "#mutate")
	test.TestReportId(t, page, 0, "2")
	test.Click(t, page, "#mutate-cancel")
	test.TestReportId(t, page, 0, "2")

}

func testConsistency(t *testing.T, f func() test.Fragment) {
	bro := test.NewFragmentBro(browser,
		f,
	)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestReportId(t, page, 1, "0")
	test.TestReportId(t, page, 2, "0")
	test.TestReportId(t, page, 3, "0")
	<-time.After(50 * time.Millisecond)
	test.TestReportId(t, page, 0, "3")
	test.TestReportId(t, page, 4, "3")
	test.Click(t, page, "#reload")
	test.TestReportId(t, page, 1, "3")
	test.TestReportId(t, page, 2, "3")
	test.TestReportId(t, page, 3, "3")
	test.TestReportId(t, page, 0, "6")
	<-time.After(50 * time.Millisecond)
	test.TestReportId(t, page, 4, "6")
}

func TestConsistent(t *testing.T) {
	testConsistency(t, func() test.Fragment {
		return &BeamConsistentFragment{
			r: test.NewReporter(1),
			b: doors.NewSource(state{}),
		}
	})
}
func TestDerive(t *testing.T) {
	testConsistency(t, func() test.Fragment {
		return &BeamDeriveFragment{
			r: test.NewReporter(1),
			b: doors.NewSource(state{}),
		}
	})
}

func TestSkip(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamSkipFragment{
				r: test.NewReporter(1),
				b: doors.NewSource(state{}),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#update1")
	test.ClickNow(t, page, "#update2")
	<-time.After(500 * time.Millisecond)
	test.TestReport(t, page, "init")
}

func TestNoSkip(t *testing.T) {
	b := doors.NewSource(state{})
	b.DisableSkipping()
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamSkipFragment{
				r: test.NewReporter(1),
				b: b,
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#update1")
	<-time.After(500 * time.Millisecond)
	test.TestReport(t, page, "propagated")
}
