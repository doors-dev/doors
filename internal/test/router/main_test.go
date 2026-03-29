package router

import (
	"testing"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	test.RunMain(func() int {
		browser = test.NewBrowser()
		code := m.Run()
		browser.MustClose()
		return code
	})
}
