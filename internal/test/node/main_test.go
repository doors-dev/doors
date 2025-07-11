package node

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"

	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	browser = rod.New().MustConnect()
	code := m.Run()
	browser.MustClose()
	os.Exit(code)
}
