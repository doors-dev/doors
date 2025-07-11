package imports

import (
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	preactDir, _ := fs.Sub(preactFS, "preact_src")
	reactDir, _ := fs.Sub(reactFS, "react_src")
	moduleDir, _ := fs.Sub(moduleFS, "module_src")
	moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")
	var err error
	preactPath, err = cookModule(preactDir)
	if err != nil {
		log.Fatal(err.Error())
	}
	reactPath, err = cookModule(reactDir)
	if err != nil {
		clean()
		log.Fatal(err.Error())
	}
	modulePath, err = copyTemp(moduleDir)
	if err != nil {
		clean()
		log.Fatal(err.Error())
	}
	moduleBundlePath, err = copyTemp(moduleBundleDir)
	if err != nil {
		clean()
		log.Fatal(err.Error())
	}
	browser = rod.New().MustConnect()
	code := m.Run()
	clean()
	browser.MustClose()
	os.Exit(code)
}
