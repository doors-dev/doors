package attr

import (
	"os"
	"testing"

	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	browser = rod.New().MustConnect()
	defer browser.MustClose()
	code := m.Run()
	os.Exit(code)
}
