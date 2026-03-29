package test

import (
	"testing"

	testsupport "github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	testsupport.RunMain(func() int {
		browser = testsupport.NewBrowser()
		code := m.Run()
		browser.MustClose()
		return code
	})
}
