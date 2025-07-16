package attr

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func moveTo(page *rod.Page, id string) (float64, float64) {
	box := page.MustElement("#" + id)
	xy := box.MustShape().Quads[0]
	page.Mouse.MustMoveTo(xy[0], xy[1])
	return xy[0], xy[1]
}

func TestPointer(t *testing.T) {
	println("DDDD")
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &pointerFragment{
			r: test.NewReporter(3),
		}
	})
	page := bro.Page(t, "/")
	println(page.MustHTML())
	test.TestReportId(t, page, 0, "")
	x, y := moveTo(page, "down")
	page.Mouse.MustDown(proto.InputMouseButtonLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "DOWN")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "up")
	test.TestReportId(t, page, 0, "DOWN")
	page.Mouse.MustUp(proto.InputMouseButtonLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "UP")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "enter")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "ENTER")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	moveTo(page, "leave")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "ENTER")
	x, y = moveTo(page, "beforeLeave")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "LEAVE")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "move")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "MOVE")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "over")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "OVER")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	moveTo(page, "out")
	x, y = moveTo(page, "beforeOut")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "OUT")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
}
