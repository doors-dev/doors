package components

import (
	"os"
	"testing"

	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	browser = rod.New().MustConnect()
	code := m.Run()
	browser.MustClose()
	os.Exit(code)
}
