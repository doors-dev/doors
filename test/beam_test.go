package test

import (
	"testing"
	"time"

	"github.com/doors-dev/doors"
)

/*
	func TestBeamInit(t *testing.T) {
		bro := newBro(
			doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
				return pr.Page(&Page{
					f: &BeamUpdateFragment{
						r: newReporter(1),
						b: doors.NewSourceBeam(state{}),
					},
				})
			}),
		)
		defer bro.close()
		page := bro.page(t, "/")
		defer page.Close()
		testReportId(t, page, 0, "0")
		click(t, page, "#update")
		testReportId(t, page, 0, "1")
		click(t, page, "#mutate")
		testReportId(t, page, 0, "2")
		click(t, page, "#mutate-cancel")
		testReportId(t, page, 0, "2")

}
*/
func TestConsistent(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				f: &BeamConsistentFragment{
					r: newReporter(1),
					b: doors.NewSourceBeam(state{}),
				},
			})
		}),
	)
	defer bro.close()
	page := bro.page(t, "/")
	defer page.Close()
	testReportId(t, page, 1, "0")
	testReportId(t, page, 2, "0")
	testReportId(t, page, 3, "0")
	<-time.After(50 * time.Millisecond)
	testReportId(t, page, 0, "3")
	click(t, page, "#reload")
	testReportId(t, page, 1, "3")
	testReportId(t, page, 2, "3")
	testReportId(t, page, 3, "3")
	testReportId(t, page, 0, "6")
}
