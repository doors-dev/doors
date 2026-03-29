package attr

import (
	"testing"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	test.RunMain(func() int {
		browser = test.NewBrowser()
		defer browser.MustClose()
		return m.Run()
	})
}
