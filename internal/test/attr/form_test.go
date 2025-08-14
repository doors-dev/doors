package attr

import (
	"fmt"
	"testing"

	"github.com/doors-dev/doors/internal/test"
)

func TestFormSimple(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &formFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	f := formData{
		Name:      "aaaa",
		Email:     "b@b.b",
		Age:       113,
		Subscribe: "on",
	}
	test.TestInput(t, page, "#name", f.Name)
	test.TestInput(t, page, "#email", f.Email)
	test.TestInput(t, page, "#age", fmt.Sprint(f.Age))
	test.ClickNow(t, page, "#subscribe")
	test.Click(t, page, "#submit")
	test.TestReportId(t, page, 0, f.Name)
	test.TestReportId(t, page, 1, f.Email)
	test.TestReportId(t, page, 2, fmt.Sprint(f.Age))
	test.TestReportId(t, page, 3, fmt.Sprint(f.Subscribe))

}
func TestFormFile(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &formFragment{
			r:   test.NewReporter(10),
			raw: true,
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	file := test.NewRandFile(1000_0000)
	fileInput := page.MustElement("#file")
	fileInput.MustSetFiles(file.Path)
	test.Click(t, page, "#submit")
	hash := test.GetReportContent(t, page, 0)
	if hash != file.Hash {
		t.Fatal("hash missmatch")
	}
}
