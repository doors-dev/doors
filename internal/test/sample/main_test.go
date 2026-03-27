package test

import (
	"os"
	"testing"

	testsupport "github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	browser = testsupport.NewBrowser()
	code := m.Run()
	browser.MustClose()
	os.Exit(code)
}
