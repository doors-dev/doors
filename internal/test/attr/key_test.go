package attr

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod/lib/input"
)

func TestKey(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &keyFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.Click(t, page, "#input")
	page.Keyboard.Press(input.KeyC)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "c")
	test.TestReportId(t, page, 1, "down")
	test.TestReportId(t, page, 2, "")
	test.TestReportId(t, page, 3, "")
	page.Keyboard.Release(input.KeyC)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 2, "c")
	test.TestReportId(t, page, 3, "up")

	test.TestReportId(t, page, 4, "")
	test.TestReportId(t, page, 5, "")
	test.TestReportId(t, page, 6, "")

	page.Keyboard.Press(input.AltLeft)
	page.Keyboard.Press(input.KeyE)
	page.Keyboard.Release(input.KeyE)
	page.Keyboard.Release(input.AltLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 6, "true")

	page.Keyboard.Press(input.ShiftLeft)
	page.Keyboard.Press(input.KeyE)
	page.Keyboard.Release(input.KeyE)
	page.Keyboard.Release(input.ShiftLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 4, "true")

	page.Keyboard.Press(input.ControlLeft)
	page.Keyboard.Press(input.KeyE)
	page.Keyboard.Release(input.KeyE)
	page.Keyboard.Release(input.ControlLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 5, "true")

}
