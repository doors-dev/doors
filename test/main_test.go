package test

import (
	"os"
	"testing"

	"github.com/go-rod/rod"
)

var browser *rod.Browser

// TestMain: Setup and teardown for all tests in this package
func TestMain(m *testing.M) {
	browser = rod.New().MustConnect()
	defer browser.MustClose()
	code := m.Run()
	os.Exit(code)
}
