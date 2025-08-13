package attr

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
)

func TestCaptureError(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &errorFragment{}
	})
	page := bro.Page(t, "/")
	
	// initial
	test.TestContent(t, page, "#report", "initial")

	// root error handler
	test.Click(t, page, "#err_1")
	test.TestContent(t, page, "#report", "root_error/err_1")

	// root normal handler
	test.Click(t, page, "#err_2")
	test.TestContent(t, page, "#report", "root/err_2")

	// n1 scope: error routed to n1 expectation
	test.Click(t, page, "#err_3")
	test.TestContent(t, page, "#report", "n1_error/err_3")

	// n1 scope: normal event
	test.Click(t, page, "#err_4")
	test.TestContent(t, page, "#report", "n1/err_4")

	// n2 scope: normal event
	test.TestContent(t, page, "#indicator", "init")
	test.TestAttrNo(t, page, "#indicator","data-indicator")
	test.Click(t, page, "#err_5")
	test.TestContent(t, page, "#report", "n2/err_5")
	test.TestContent(t, page, "#indicator", "indicator")
	test.TestAttr(t, page, "#indicator","data-indicator", "true")
	<-time.After(500 * time.Millisecond)
	test.TestContent(t, page, "#indicator", "init")
	test.TestAttrNo(t, page, "#indicator","data-indicator")


	// n2 scope: error routed to n1 expectation
	test.Click(t, page, "#err_6")
	test.TestContent(t, page, "#report", "n1_error/err_6")
}
