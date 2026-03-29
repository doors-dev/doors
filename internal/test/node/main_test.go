package door

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	test.RunMain(func() int {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		browser = test.NewBrowser()
		code := m.Run()
		browser.MustClose()
		return code
	})
}
