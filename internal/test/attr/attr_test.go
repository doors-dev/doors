package attr

import "github.com/doors-dev/doors/internal/test"
import "github.com/doors-dev/doors/internal/common"
import "testing"
import "time"
import "fmt"

func TestData(t *testing.T) {
	data := common.RandId()

	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &dataFragment{
			data: data,
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	test.TestContent(t, page, "#target", data)
}

func TestHook(t *testing.T) {
	data := common.RandId()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &hookFragment{
			data: data,
			r:    test.NewReporter(1),
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestContent(t, page, "#target", fmt.Sprint(len(data)))
	test.TestReport(t, page, data)
}
func TestCall(t *testing.T) {
	data := common.RandId()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &callFragment{
			data: data,
			r:    test.NewReporter(2),
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestContent(t, page, "#target", fmt.Sprint(len(data)))
	test.TestReport(t, page, data)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 1, "response")
}
