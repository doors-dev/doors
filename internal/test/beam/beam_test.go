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
				b: doors.NewSourceBeam(state{}),
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
			b: doors.NewSourceBeam(state{}),
		}
	})
}
func TestDerive(t *testing.T) {
	testConsistency(t, func() test.Fragment {
		return &BeamDeriveFragment{
			r: test.NewReporter(1),
			b: doors.NewSourceBeam(state{}),
		}
	})
}
