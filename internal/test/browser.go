package test

import (
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func NewBrowser() *rod.Browser {
	l := launcher.New()
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		l = l.NoSandbox(true)
	}
	return rod.New().ControlURL(l.MustLaunch()).MustConnect()
}
