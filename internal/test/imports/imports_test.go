package imports

import (
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func checkColor(t *testing.T, page *rod.Page) {
	element := page.MustElement("h1")
	styleValue, err := element.Eval(`() => getComputedStyle(this).color`)
	if err != nil {
		t.Fatalf("")
	}
	if styleValue.Value.Str() != "rgb(255, 0, 0)" {
		t.Fatal("h1 expected color", " red ", "got ", styleValue.Value.Str() )
	}
}

func testStyle(t *testing.T, h func(doors.SourceBeam[test.Path]) templ.Component) {
	bro := test.NewBro(browser,
		doors.UsePage(func(pr doors.PageRouter[test.Path], r doors.RPage[test.Path]) doors.PageRoute {
			return pr.Page(&test.Page{
				Header: "Testing Imports",
				H: h,
				F: &ModuleFragment{},
			})
		}),
		doors.UseDir("module", modulePath),
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	checkColor(t, page)
}
func testModule(t *testing.T, h func(doors.SourceBeam[test.Path]) templ.Component) {
	bro := test.NewBro(browser,
		doors.UsePage(func(pr doors.PageRouter[test.Path], r doors.RPage[test.Path]) doors.PageRoute {
			return pr.Page(&test.Page{
				Header: "Testing Imports",
				H: h,
				F: &ModuleFragment{},
			})
		}),
		doors.UseDir("module", modulePath),
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
}

func TestModule(t *testing.T) {
	testModule(t, moduleHead)
}

func TestModuleBytes(t *testing.T) {
	testModule(t, moduleBytesHead)
}
func TestModuleRaw(t *testing.T) {
	testModule(t, moduleRawHead)
}
func TestModuleRawBytes(t *testing.T) {
	testModule(t, moduleRawBytesHead)
}
func TestModuleFS(t *testing.T) {
	testModule(t, moduleBundleFSHead)
}

func TestModulHosted(t *testing.T) {
	testModule(t, moduleBundleHostHead)
}

func TestModuleExternal(t *testing.T) {
	testModule(t, moduleExternalHead)
}
func TestStyleHosted(t *testing.T) {
	testStyle(t, styleHostedHead)
}
func TestStyleExternal(t *testing.T) {
	testStyle(t, styleExternalHead)
}

func TestStyleBytes(t *testing.T) {
	testStyle(t, styleBytesHead)
}
func TestStyle(t *testing.T) {
	testStyle(t, styleHead)
}

func TestReact(t *testing.T) {
	bro := test.NewBro(browser,
		doors.UsePage(func(pr doors.PageRouter[test.Path], r doors.RPage[test.Path]) doors.PageRoute {
			return pr.Page(&test.Page{
				H: reactHead,
				F: &ReactFragment{},
			})
		}),
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	<-time.After(100 * time.Millisecond)
	test.TestContent(t, page, "#h2", "React")
	test.TestContent(t, page, "#ph2", "Preact")
	test.Click(t, page, "#inc")
	test.Click(t, page, "#inc")
	test.Click(t, page, "#dec")
	test.TestReportId(t, page, 0, "1")
	test.Click(t, page, "#pinc")
	test.Click(t, page, "#pinc")
	test.Click(t, page, "#pdec")
	test.Click(t, page, "#pdec")
	test.Click(t, page, "#pdec")
	test.TestReportId(t, page, 1, "-1")
}

func TestFiles(t *testing.T) {
	bro := test.NewBro(browser,
		doors.UsePage(func(pr doors.PageRouter[test.Path], r doors.RPage[test.Path]) doors.PageRoute {
			return pr.Page(&test.Page{
				H: staticFiles,
				F: &Empty{},
			})
		}),
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
}
