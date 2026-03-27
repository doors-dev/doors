package attr

import (
	"os"
	"testing"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	browser = test.NewBrowser()
	defer browser.MustClose()
	code := m.Run()
	os.Exit(code)
}
